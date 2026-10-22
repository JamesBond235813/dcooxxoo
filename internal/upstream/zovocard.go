package upstream

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	siteconnectiondomain "github.com/dujiao-next/internal/modules/siteconnection/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/shopspring/decimal"
)

// ZovoCardAdapter ZovoCard / 兑换桥协议适配器
type ZovoCardAdapter struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	uploadsDir string
	client     *http.Client
}

// NewZovoCardAdapter 创建 ZovoCard / 兑换桥适配器
func NewZovoCardAdapter(conn *siteconnectiondomain.Connection, uploadsDir string) *ZovoCardAdapter {
	base := strings.TrimRight(strings.TrimSpace(conn.BaseURL), "/")
	base = strings.TrimSuffix(base, "/openapi/v1")
	base = strings.TrimSuffix(base, "/openapi")
	if base == "" {
		base = "https://zovocard.com"
	}

	return &ZovoCardAdapter{
		baseURL:    base,
		apiKey:     strings.TrimSpace(conn.ApiKey),
		apiSecret:  strings.TrimSpace(conn.ApiSecret),
		uploadsDir: uploadsDir,
		client: &http.Client{
			Timeout: 45 * time.Second,
		},
	}
}

// Ping 连接测试 (检查余额及连通性)
func (a *ZovoCardAdapter) Ping(ctx context.Context) (*PingResult, error) {
	// 优先调用 /openapi/v1/balance
	var balResp struct {
		Code      int    `json:"code"`
		Msg       string `json:"msg"`
		ErrorCode string `json:"error_code"`
		Data      struct {
			Balance              float64 `json:"balance"`
			SpendableBalance     float64 `json:"spendable_balance"`
			AccountReserveAmount float64 `json:"account_reserve_amount"`
		} `json:"data"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/openapi/v1/balance", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		// 尝试作为自托管兑换桥检测
		var healthResp struct {
			Status  string `json:"status"`
			Message string `json:"message"`
		}
		if hErr := a.doGet(ctx, "/health", &healthResp); hErr == nil && healthResp.Status == "ok" {
			return &PingResult{
				SiteName:        "ZovoCard CDK 兑换桥 (自托管)",
				ProtocolVersion: "1.0",
				Balance:         "0.00",
				Currency:        "USD",
			}, nil
		}
		return nil, fmt.Errorf("ping upstream error: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("upstream returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &balResp); err != nil {
		return nil, fmt.Errorf("unmarshal balance response: %w", err)
	}

	if balResp.Code != 0 {
		return nil, fmt.Errorf("upstream error (%s): %s", balResp.ErrorCode, balResp.Msg)
	}

	spendable := fmt.Sprintf("%.2f", balResp.Data.SpendableBalance)
	return &PingResult{
		SiteName:        "ZovoCard / 兑换桥 (CDK直充平台)",
		ProtocolVersion: "1.0",
		Balance:         spendable,
		Currency:        "USD",
	}, nil
}

// ListCategories 分类列表
func (a *ZovoCardAdapter) ListCategories(ctx context.Context) (*CategoryListResult, error) {
	return &CategoryListResult{
		Supported: true,
		Categories: []UpstreamCategory{
			{
				ID:        1,
				Slug:      "ai-subscriptions",
				Name:      jsonmap.JSON{"zh-CN": "AI 订阅与额度直充", "en-US": "AI Subscriptions & Credits", "zh-TW": "AI 訂閱與額度直充"},
				SortOrder: 1,
			},
		},
	}, nil
}

// ListProducts 拉取可用套餐列表并转换为上游商品
func (a *ZovoCardAdapter) ListProducts(ctx context.Context, opts ListProductsOpts) (*ProductListResult, error) {
	var plansResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PaymentRegions []zovoPaymentRegion `json:"payment_regions"`
			Registry       []struct {
				Key                string   `json:"key"`
				Name               string   `json:"name"`
				Product            string   `json:"product"`
				Purchasable        bool     `json:"purchasable"`
				ServiceFeeUsdMinor int64    `json:"service_fee_usd_minor"`
				PaymentRegions     []string `json:"payment_regions"`
			} `json:"registry"`
			Plans map[string]struct {
				Key                string   `json:"key"`
				Name               string   `json:"name"`
				Product            string   `json:"product"`
				Enabled            *bool    `json:"enabled"`
				ServiceFeeUsdMinor int64    `json:"service_fee_usd_minor"`
				PaymentRegions     []string `json:"payment_regions"`
			} `json:"plans"`
		} `json:"data"`
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/openapi/v1/gpt-direct/plans?product=gpt", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch plans: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("plans API returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	if err := json.Unmarshal(body, &plansResp); err != nil {
		return nil, fmt.Errorf("unmarshal plans response: %w", err)
	}

	items := make([]UpstreamProduct, 0)
	// 表单定义：支持用户下单时填入 ChatGPT session / 账号
	formSchema := jsonmap.JSON{
		"fields": []map[string]interface{}{
			{
				"key":         "session",
				"label":       map[string]interface{}{"zh-CN": "ChatGPT Session", "zh-TW": "ChatGPT Session", "en-US": "ChatGPT Session"},
				"type":        "textarea",
				"required":    true,
				"placeholder": map[string]interface{}{"zh-CN": "请在付款后提交完整的 ChatGPT Session JSON", "zh-TW": "請在付款後提交完整的 ChatGPT Session JSON", "en-US": "Submit the complete ChatGPT Session JSON after payment"},
			},
		},
	}

	for _, reg := range plansResp.Data.Registry {
		planConfig, hasPlanConfig := plansResp.Data.Plans[reg.Key]
		active := reg.Purchasable && (!hasPlanConfig || planConfig.Enabled == nil || *planConfig.Enabled)
		if !opts.IncludeInactive && !active {
			continue
		}
		paymentRegions := plansResp.Data.PaymentRegions
		if len(reg.PaymentRegions) > 0 {
			// A plan-specific list takes precedence when the supplier provides one.
			paymentRegions = make([]zovoPaymentRegion, 0, len(reg.PaymentRegions))
			for _, country := range reg.PaymentRegions {
				currency, _ := zovoCurrencyForCountry(country)
				paymentRegions = append(paymentRegions, zovoPaymentRegion{Country: country, Currency: currency})
			}
		}

		priceUsd := fmt.Sprintf("%.2f", float64(reg.ServiceFeeUsdMinor)/100.0)
		stockQuantity := 0
		if active {
			// The registry reports whether issuance is enabled, not a numeric stock
			// count. Treat it as on-demand/untracked stock instead of inventing 9999.
			stockQuantity = -1
		}
		skuID := stablePlanID(reg.Key)
		prodID := skuID

		titleZH := "ChatGPT " + reg.Name
		titleEN := "ChatGPT " + reg.Name
		if reg.Key == "plus_half" {
			titleZH = "ChatGPT Plus 半价资格套餐"
			titleEN = "ChatGPT Plus Promotion Eligible Plan"
		} else if reg.Key == "pro_50x" {
			titleZH = "ChatGPT Pro (50x) 官方直充/CDK"
			titleEN = "ChatGPT Pro (50x) Subscription / CDK"
		} else if reg.Key == "pro_20x" {
			titleZH = "ChatGPT Pro (20x) 官方直充/CDK"
			titleEN = "ChatGPT Pro (20x) Subscription / CDK"
		} else if reg.Key == "plus" {
			titleZH = "ChatGPT Plus 官方直充/CDK"
			titleEN = "ChatGPT Plus Subscription / CDK"
		}

		items = append(items, UpstreamProduct{
			ID:          prodID,
			CategoryID:  1,
			Title:       jsonmap.JSON{"zh-CN": titleZH, "en-US": titleEN, "zh-TW": titleZH},
			Description: jsonmap.JSON{"zh-CN": "支持自动直充与 CDK 卡密兑换", "en-US": "Supports auto recharge and CDK exchange"},
			Content:     jsonmap.JSON{"zh-CN": "<p>下单支付后系统将自动完成发货与订阅交付。</p>", "en-US": "<p>System automatically fulfills subscription after payment.</p>"},
			PriceAmount: priceUsd,
			Currency:    "USD",
			// ZovoCard creates and delivers the CDK automatically. Mapping import
			// normalizes provider fulfillment as auto/manual; "upstream" is the
			// local fulfillment mode and would be misclassified as manual here.
			FulfillmentType:  constants.FulfillmentTypeAuto,
			ManualFormSchema: formSchema,
			IsActive:         active,
			SKUs: []UpstreamSKU{
				{
					ID:      skuID,
					SKUCode: reg.Key,
					SpecValues: jsonmap.JSON{
						"plan":                  reg.Key,
						"name":                  reg.Name,
						"payment_regions":       paymentRegions,
						"service_fee_usd_minor": reg.ServiceFeeUsdMinor,
					},
					PriceAmount:   priceUsd,
					StockStatus:   "in_stock",
					StockQuantity: stockQuantity,
					IsActive:      active,
				},
			},
			UpdatedAt: time.Now(),
		})
	}

	return &ProductListResult{
		Total:            len(items),
		Items:            items,
		IncludesInactive: opts.IncludeInactive,
	}, nil
}

// GetProduct 获取单个商品详情
func (a *ZovoCardAdapter) GetProduct(ctx context.Context, productID uint) (*UpstreamProduct, error) {
	list, err := a.ListProducts(ctx, ListProductsOpts{IncludeInactive: true})
	if err != nil {
		return nil, err
	}
	for _, p := range list.Items {
		if p.ID == productID {
			return &p, nil
		}
	}
	return nil, ErrUpstreamProductUnavailable
}

// PreflightSession checks a ChatGPT session with ZovoCard before the customer
// confirms submission. It returns only the upstream account summary; the
// short-lived token and quote data are deliberately not exposed to the client.
func (a *ZovoCardAdapter) PreflightSession(ctx context.Context, session string) (map[string]interface{}, error) {
	session = strings.TrimSpace(session)
	if session == "" {
		return nil, errors.New("session is required")
	}
	body, err := json.Marshal(map[string]interface{}{
		"product":          "gpt",
		"credential":       map[string]string{"mode": "session", "session": session},
		"payment_country":  "PH",
		"payment_currency": "PHP",
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/openapi/v1/gpt-direct/preflight", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", a.apiKey)
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ZovoCard preflight request: %w", err)
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, fmt.Errorf("read ZovoCard preflight response: %w", err)
	}
	var envelope struct {
		Code int                    `json:"code"`
		Msg  string                 `json:"msg"`
		Data map[string]interface{} `json:"data"`
	}
	if err := json.Unmarshal(raw, &envelope); err != nil {
		return nil, errors.New("ZovoCard preflight response is invalid")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || envelope.Code != 0 {
		message := strings.TrimSpace(envelope.Msg)
		if message == "" {
			message = fmt.Sprintf("HTTP %d", resp.StatusCode)
		}
		return nil, fmt.Errorf("ZovoCard preflight rejected: %s", message)
	}
	// Whitelist user-facing account state. Never return preflight_token, quotes,
	// payment routing details, or the submitted session to the browser.
	allowed := []string{
		"email", "account_email", "currentPlan", "current_plan", "plan",
		"subscription_has_active", "subscriptionHasActive", "subscription_active_until", "subscriptionActiveUntil",
		"can_purchase_at", "canPurchaseAt", "subscription_will_renew", "subscriptionWillRenew",
		"subscription_recovery_required", "subscriptionRecoveryRequired", "subscription_is_delinquent", "subscriptionIsDelinquent",
		"last_payment", "lastPayment", "payment_method", "paymentMethod",
	}
	result := make(map[string]interface{}, len(allowed))
	for _, key := range allowed {
		if value, exists := envelope.Data[key]; exists {
			switch key {
			case "last_payment", "lastPayment":
				result[key] = pickPublicFields(value, "amount", "amountMinor", "amount_minor", "currency", "status", "created", "created_at", "paidAt", "paid_at")
			case "payment_method", "paymentMethod":
				result[key] = pickPublicFields(value, "type", "brand", "name", "last4", "last_4", "expMonth", "exp_month", "expYear", "exp_year", "isDefault", "is_default")
			default:
				result[key] = value
			}
		}
	}
	if result["email"] == nil && result["account_email"] == nil {
		return nil, errors.New("ZovoCard preflight did not return account email")
	}
	return result, nil
}

// CheckPlusHalfEligibility asks ZovoCard for the account-specific promotion.
// A missing or indeterminate answer is never treated as eligible.
func (a *ZovoCardAdapter) CheckPlusHalfEligibility(ctx context.Context, session string) (bool, string, error) {
	session = strings.TrimSpace(session)
	if session == "" {
		return false, "", errors.New("session is required")
	}
	body, err := json.Marshal(map[string]interface{}{
		"product": "gpt", "plan": "plus_half",
		"credential":      map[string]string{"mode": "session", "session": session},
		"payment_country": "PH", "payment_currency": "PHP",
	})
	if err != nil {
		return false, "", err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/openapi/v1/gpt-direct/preflight", bytes.NewReader(body))
	if err != nil {
		return false, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", a.apiKey)
	resp, err := a.client.Do(req)
	if err != nil {
		return false, "", fmt.Errorf("ZovoCard half-price preflight: %w", err)
	}
	defer resp.Body.Close()
	var result struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Email          string `json:"email"`
			AccountEmail   string `json:"account_email"`
			BlockingReason string `json:"blocking_reason"`
			Promotions     map[string]struct {
				Eligible *bool `json:"eligible"`
			} `json:"promotions"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&result); err != nil {
		return false, "", fmt.Errorf("decode ZovoCard half-price preflight: %w", err)
	}
	if resp.StatusCode != http.StatusOK || result.Code != 0 {
		msg := strings.TrimSpace(result.Msg)
		if msg == "" {
			msg = "ZovoCard half-price preflight rejected the account"
		}
		return false, "", errors.New(msg)
	}
	email := strings.TrimSpace(result.Data.Email)
	if email == "" {
		email = strings.TrimSpace(result.Data.AccountEmail)
	}
	if email == "" {
		return false, "", errors.New("ZovoCard half-price preflight omitted account email")
	}
	promo, exists := result.Data.Promotions["plus_half"]
	return exists && promo.Eligible != nil && *promo.Eligible && result.Data.BlockingReason == "", email, nil
}

func pickPublicFields(value interface{}, keys ...string) interface{} {
	object, ok := value.(map[string]interface{})
	if !ok {
		if text, isText := value.(string); isText {
			return text
		}
		return nil
	}
	selected := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		if field, exists := object[key]; exists {
			selected[key] = field
		}
	}
	return selected
}

type zovoPaymentRegion struct {
	Country  string `json:"country"`
	Currency string `json:"currency"`
}

type zovoRegionQuote struct {
	Country           string `json:"country"`
	Currency          string `json:"currency"`
	AmountMinor       int64  `json:"amountMinor"`
	MinorUnitExponent int    `json:"minorUnitExponent"`
}

// chooseLowestPricedRegion preflights every currently available GPT region,
// converts the exact upstream quote using ZovoCard's current reference FX
// snapshot, and fails closed if any candidate cannot be compared.
func (a *ZovoCardAdapter) chooseLowestPricedRegion(ctx context.Context, plan, session string, regions []zovoPaymentRegion) (zovoPaymentRegion, error) {
	if len(regions) == 0 {
		return zovoPaymentRegion{}, errors.New("ZovoCard returned no payment regions; refusing to procure without a complete comparison")
	}

	quotes := make([]zovoRegionQuote, 0, len(regions))
	currencies := make(map[string]struct{})
	seenCountries := make(map[string]struct{})
	for _, rawRegion := range regions {
		country := strings.ToUpper(strings.TrimSpace(rawRegion.Country))
		currency := strings.ToUpper(strings.TrimSpace(rawRegion.Currency))
		expectedCurrency, ok := zovoCurrencyForCountry(country)
		if !ok || currency != expectedCurrency {
			return zovoPaymentRegion{}, fmt.Errorf("unsupported ZovoCard payment region %q/%q; refusing to procure without comparing it", country, currency)
		}
		if _, exists := seenCountries[country]; exists {
			continue
		}
		seenCountries[country] = struct{}{}

		var response struct {
			Code int    `json:"code"`
			Msg  string `json:"msg"`
			Data struct {
				QuoteError string                     `json:"quote_error"`
				Quotes     map[string]zovoRegionQuote `json:"quotes"`
			} `json:"data"`
		}
		requestBody, _ := json.Marshal(map[string]interface{}{
			"product":          "gpt",
			"credential":       map[string]string{"mode": "session", "session": session},
			"payment_country":  country,
			"payment_currency": currency,
		})
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/openapi/v1/gpt-direct/preflight", bytes.NewReader(requestBody))
		if err != nil {
			return zovoPaymentRegion{}, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-API-Key", a.apiKey)
		resp, err := a.client.Do(req)
		if err != nil {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard region preflight %s failed: %w", country, err)
		}
		body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
		_ = resp.Body.Close()
		if readErr != nil {
			return zovoPaymentRegion{}, fmt.Errorf("read ZovoCard region quote %s: %w", country, readErr)
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard region preflight %s returned HTTP %d", country, resp.StatusCode)
		}
		if err := json.Unmarshal(body, &response); err != nil {
			return zovoPaymentRegion{}, fmt.Errorf("decode ZovoCard region quote %s: %w", country, err)
		}
		if response.Code != 0 {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard region preflight %s rejected the account", country)
		}
		if strings.TrimSpace(response.Data.QuoteError) != "" {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard region %s did not return a confirmed quote", country)
		}
		quote, exists := response.Data.Quotes[plan]
		if !exists || quote.AmountMinor <= 0 || quote.MinorUnitExponent < 0 || quote.MinorUnitExponent > 6 {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard region %s has no comparable quote for plan %s", country, plan)
		}
		if quote.Country != "" && !strings.EqualFold(quote.Country, country) {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard quote country mismatch for region %s", country)
		}
		if quote.Country == "" {
			quote.Country = country
		}
		if quote.Currency == "" || !strings.EqualFold(quote.Currency, currency) {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard quote currency mismatch for region %s", country)
		}
		quotes = append(quotes, quote)
		currencies[currency] = struct{}{}
	}
	if len(quotes) == 0 {
		return zovoPaymentRegion{}, errors.New("ZovoCard returned no comparable regional quotes")
	}

	rates, err := a.getUSDPerUnitRates(ctx, currencies)
	if err != nil {
		return zovoPaymentRegion{}, err
	}
	var selected zovoPaymentRegion
	var lowest decimal.Decimal
	for index, quote := range quotes {
		currency := strings.ToUpper(quote.Currency)
		usdPerUnit, ok := rates[currency]
		if !ok || !usdPerUnit.GreaterThan(decimal.Zero) {
			return zovoPaymentRegion{}, fmt.Errorf("ZovoCard has no current USD conversion rate for %s", currency)
		}
		localAmount := decimal.NewFromInt(quote.AmountMinor).Shift(-int32(quote.MinorUnitExponent))
		usdAmount := localAmount.Mul(usdPerUnit)
		if index == 0 || usdAmount.LessThan(lowest) {
			lowest = usdAmount
			selected = zovoPaymentRegion{Country: quote.Country, Currency: currency}
		}
	}
	return selected, nil
}

func (a *ZovoCardAdapter) getUSDPerUnitRates(ctx context.Context, currencies map[string]struct{}) (map[string]decimal.Decimal, error) {
	rates := map[string]decimal.Decimal{"USD": decimal.NewFromInt(1)}
	requested := make([]string, 0, len(currencies))
	for currency := range currencies {
		currency = strings.ToUpper(currency)
		if currency != "USD" {
			requested = append(requested, currency)
		}
	}
	if len(requested) == 0 {
		return rates, nil
	}
	// The currencies are three-letter codes obtained from the region map above.
	for i := 0; i < len(requested); i++ {
		for j := i + 1; j < len(requested); j++ {
			if requested[j] < requested[i] {
				requested[i], requested[j] = requested[j], requested[i]
			}
		}
	}
	endpoint := a.baseURL + "/openapi/v1/fx/rates?currency=" + url.QueryEscape(strings.Join(requested, ","))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", a.apiKey)
	var response struct {
		Code int `json:"code"`
		Data struct {
			Stale bool `json:"stale"`
			Rates []struct {
				Currency   string  `json:"currency"`
				USDPerUnit float64 `json:"usd_per_unit"`
			} `json:"rates"`
		} `json:"data"`
	}
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch ZovoCard currency rates: %w", err)
	}
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	_ = resp.Body.Close()
	if readErr != nil {
		return nil, fmt.Errorf("read ZovoCard currency rates: %w", readErr)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ZovoCard currency rates returned HTTP %d", resp.StatusCode)
	}
	if err := json.Unmarshal(body, &response); err != nil || response.Code != 0 {
		return nil, errors.New("ZovoCard currency rates response is invalid")
	}
	if response.Data.Stale {
		return nil, errors.New("ZovoCard currency rates are stale; refusing to choose a payment region")
	}
	for _, rate := range response.Data.Rates {
		if rate.USDPerUnit > 0 {
			rates[strings.ToUpper(rate.Currency)] = decimal.NewFromFloat(rate.USDPerUnit)
		}
	}
	for _, currency := range requested {
		if _, ok := rates[currency]; !ok {
			return nil, fmt.Errorf("ZovoCard omitted the current USD conversion rate for %s", currency)
		}
	}
	return rates, nil
}

func zovoCurrencyForCountry(country string) (string, bool) {
	switch strings.ToUpper(country) {
	case "PH":
		return "PHP", true
	case "US":
		return "USD", true
	case "JP":
		return "JPY", true
	case "CL":
		return "CLP", true
	case "EG":
		return "EGP", true
	default:
		return "", false
	}
}

func paymentRegionsFromSpec(spec jsonmap.JSON) []zovoPaymentRegion {
	value, ok := spec["payment_regions"]
	if !ok {
		return nil
	}
	regions, ok := value.([]zovoPaymentRegion)
	if ok {
		return regions
	}
	if countries, ok := value.([]string); ok {
		regions = make([]zovoPaymentRegion, 0, len(countries))
		for _, country := range countries {
			currency, _ := zovoCurrencyForCountry(country)
			regions = append(regions, zovoPaymentRegion{Country: country, Currency: currency})
		}
		return regions
	}
	items, ok := value.([]interface{})
	if !ok {
		return nil
	}
	regions = make([]zovoPaymentRegion, 0, len(items))
	for _, item := range items {
		switch region := item.(type) {
		case zovoPaymentRegion:
			regions = append(regions, region)
		case string:
			currency, _ := zovoCurrencyForCountry(region)
			regions = append(regions, zovoPaymentRegion{Country: region, Currency: currency})
		case map[string]interface{}:
			country, _ := region["country"].(string)
			currency, _ := region["currency"].(string)
			regions = append(regions, zovoPaymentRegion{Country: country, Currency: currency})
		}
	}
	return regions
}

// currentPaymentRegions reads the supplier's current region list at purchase
// time. Product/SKU specifications may be imported or normalized by another
// layer, so they are not an authoritative source for this decision.
func (a *ZovoCardAdapter) currentPaymentRegions(ctx context.Context, plan string) ([]zovoPaymentRegion, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+"/openapi/v1/gpt-direct/plans?product=gpt", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch current ZovoCard payment regions: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("ZovoCard payment regions returned HTTP %d", resp.StatusCode)
	}
	var result struct {
		Code int `json:"code"`
		Data struct {
			PaymentRegions []zovoPaymentRegion `json:"payment_regions"`
			Registry       []struct {
				Key            string   `json:"key"`
				PaymentRegions []string `json:"payment_regions"`
			} `json:"registry"`
		} `json:"data"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 2<<20)).Decode(&result); err != nil || result.Code != 0 {
		return nil, errors.New("ZovoCard payment regions response is invalid")
	}
	for _, item := range result.Data.Registry {
		if item.Key != plan {
			continue
		}
		if len(item.PaymentRegions) == 0 {
			return result.Data.PaymentRegions, nil
		}
		regions := make([]zovoPaymentRegion, 0, len(item.PaymentRegions))
		for _, country := range item.PaymentRegions {
			currency, ok := zovoCurrencyForCountry(country)
			if !ok {
				return nil, fmt.Errorf("unsupported ZovoCard payment country %q", country)
			}
			regions = append(regions, zovoPaymentRegion{Country: country, Currency: currency})
		}
		return regions, nil
	}
	return nil, fmt.Errorf("ZovoCard plan %q is missing from current registry", plan)
}

func (a *ZovoCardAdapter) loadRegionSelection(traceID string) (zovoPaymentRegion, bool, error) {
	path, err := a.regionSelectionPath(traceID)
	if err != nil {
		return zovoPaymentRegion{}, false, err
	}
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return zovoPaymentRegion{}, false, nil
	}
	if err != nil {
		return zovoPaymentRegion{}, false, err
	}
	var selection zovoPaymentRegion
	if err := json.Unmarshal(data, &selection); err != nil || selection.Country == "" || selection.Currency == "" {
		return zovoPaymentRegion{}, false, errors.New("saved ZovoCard payment region is invalid")
	}
	return selection, true, nil
}

func (a *ZovoCardAdapter) saveRegionSelection(traceID string, selection zovoPaymentRegion) error {
	path, err := a.regionSelectionPath(traceID)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create ZovoCard region selection directory: %w", err)
	}
	data, err := json.Marshal(selection)
	if err != nil {
		return err
	}
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0600); err != nil {
		return fmt.Errorf("save ZovoCard region selection: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		_ = os.Remove(tempPath)
		return fmt.Errorf("commit ZovoCard region selection: %w", err)
	}
	return nil
}

func (a *ZovoCardAdapter) regionSelectionPath(traceID string) (string, error) {
	traceID = strings.TrimSpace(traceID)
	if traceID == "" {
		return "", errors.New("procurement trace ID is required for idempotent regional pricing")
	}
	if strings.TrimSpace(a.uploadsDir) == "" {
		return "", errors.New("uploads directory is required to persist ZovoCard payment region")
	}
	digest := sha256.Sum256([]byte(traceID))
	return filepath.Join(a.uploadsDir, ".zovocard-payment-regions", hex.EncodeToString(digest[:])+".json"), nil
}

// CreateOrder 发起采购单：自动发码并根据是否提供 session 自动执行兑换
func (a *ZovoCardAdapter) CreateOrder(ctx context.Context, req CreateUpstreamOrderReq) (*CreateUpstreamOrderResp, error) {
	if req.Quantity < 1 {
		return nil, errors.New("quantity must be positive")
	}
	// SKU ID 是适配器从上游 registry 派生的稳定 ID。未知 SKU 必须失败关闭，
	// 不能像旧逻辑一样静默按 Plus 发码，否则 Pro 订单可能采购成错误套餐。
	plans, err := a.ListProducts(ctx, ListProductsOpts{IncludeInactive: true})
	if err != nil {
		return nil, fmt.Errorf("resolve ZovoCard plan: %w", err)
	}
	planKey := ""
	for _, product := range plans.Items {
		for _, sku := range product.SKUs {
			if sku.ID == req.SKUID {
				planKey = sku.SKUCode
				break
			}
		}
	}
	if planKey == "" {
		return nil, fmt.Errorf("unknown ZovoCard SKU id %d", req.SKUID)
	}
	sessionToken, _ := req.ManualFormData["session"].(string)
	sessionToken = strings.TrimSpace(sessionToken)
	if sessionToken != "" && req.Quantity != 1 {
		return nil, errors.New("direct recharge supports quantity 1 per order")
	}
	idemKey := strings.TrimSpace(req.TraceID)
	if idemKey == "" && strings.TrimSpace(req.DownstreamOrderNo) != "" {
		idemKey = "dj-" + strings.TrimSpace(req.DownstreamOrderNo)
	}
	if idemKey == "" {
		if sessionToken != "" {
			return nil, errors.New("downstream order number is required for idempotent ZovoCard procurement")
		}
		idemKey = fmt.Sprintf("dj-direct-%d", time.Now().UnixNano())
	}
	paymentRegion := zovoPaymentRegion{}
	if planKey == "plus_half" {
		if sessionToken == "" {
			return nil, errors.New("Plus half-price purchase requires a verified session")
		}
		eligible, _, checkErr := a.CheckPlusHalfEligibility(ctx, sessionToken)
		if checkErr != nil {
			return nil, checkErr
		}
		if !eligible {
			return &CreateUpstreamOrderResp{OK: false, ErrorCode: "HALF_PRICE_INELIGIBLE", ErrorMessage: "ChatGPT account is not confirmed eligible for Plus half-price"}, nil
		}
		paymentRegion = zovoPaymentRegion{Country: "PH", Currency: "PHP"}
	} else if sessionToken != "" && planKey != "pro_20x_renew" && planKey != "pro_5x_renew" {
		var selected bool
		paymentRegion, selected, err = a.loadRegionSelection(idemKey)
		if err != nil {
			return nil, fmt.Errorf("load saved ZovoCard payment region: %w", err)
		}
		if !selected {
			regions, regionErr := a.currentPaymentRegions(ctx, planKey)
			if regionErr != nil {
				return nil, fmt.Errorf("fetch current ZovoCard payment regions: %w", regionErr)
			}
			paymentRegion, err = a.chooseLowestPricedRegion(ctx, planKey, sessionToken, regions)
			if err != nil {
				return nil, fmt.Errorf("select lowest-cost ZovoCard payment region: %w", err)
			}
			if err := a.saveRegionSelection(idemKey, paymentRegion); err != nil {
				return nil, err
			}
		}
	}

	// 2. 调用卡台 / 兑换桥发码接口: POST /openapi/v1/gpt-direct/cdks
	issueReqBody := map[string]interface{}{
		"plan":              planKey,
		"count":             req.Quantity,
		"funding_confirmed": true,
	}
	if paymentRegion.Country != "" {
		issueReqBody["payment_country"] = paymentRegion.Country
		issueReqBody["payment_currency"] = paymentRegion.Currency
	}
	reqData, _ := json.Marshal(issueReqBody)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/openapi/v1/gpt-direct/cdks", bytes.NewReader(reqData))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-API-Key", a.apiKey)
	httpReq.Header.Set("Idempotency-Key", idemKey)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return &CreateUpstreamOrderResp{
			OK:           false,
			ErrorCode:    "UPSTREAM_NETWORK_ERROR",
			ErrorMessage: fmt.Sprintf("failed to request CDK issuance: %v", err),
		}, nil
	}
	defer resp.Body.Close()

	respBytes, _ := io.ReadAll(resp.Body)
	var issueResp struct {
		Code      int    `json:"code"`
		Msg       string `json:"msg"`
		ErrorCode string `json:"error_code"`
		Data      struct {
			Requested int `json:"requested"`
			Issued    []struct {
				ID             uint   `json:"id"`
				Code           string `json:"code"`
				CodePrefix     string `json:"code_prefix"`
				Plan           string `json:"plan"`
				FeeAmountMinor int64  `json:"fee_amount_minor"`
			} `json:"issued"`
		} `json:"data"`
	}

	if err := json.Unmarshal(respBytes, &issueResp); err != nil {
		return &CreateUpstreamOrderResp{
			OK:           false,
			ErrorCode:    "UPSTREAM_INVALID_RESPONSE",
			ErrorMessage: "invalid CDK issuance response",
		}, nil
	}

	if issueResp.Code != 0 || len(issueResp.Data.Issued) == 0 {
		return &CreateUpstreamOrderResp{
			OK:           false,
			ErrorCode:    issueResp.ErrorCode,
			ErrorMessage: issueResp.Msg,
		}, nil
	}

	codes := make([]string, 0, len(issueResp.Data.Issued))
	var totalFeeMinor int64
	for _, issued := range issueResp.Data.Issued {
		code := strings.TrimSpace(issued.Code)
		if code != "" {
			codes = append(codes, code)
		}
		totalFeeMinor += issued.FeeAmountMinor
	}
	if len(codes) != len(issueResp.Data.Issued) {
		return nil, errors.New("ZovoCard issued response contains a CDK without a deliverable code")
	}
	firstCDK := issueResp.Data.Issued[0]
	cdkCode := codes[0]

	// 3. 检查是否有用户提交的 session 进行全自动代充兑换
	if sessionToken != "" {
		// 自动执行兑换流程: Preview -> Preflight -> Redeem
		redeemOrderID, err := a.executeAutoRedeem(ctx, cdkCode, sessionToken, req.DownstreamOrderNo)
		if err != nil {
			if planKey == "plus_half" {
				logger.Warnw("half_price_redeem_requires_review", "order_no", req.DownstreamOrderNo, "error", err)
				return &CreateUpstreamOrderResp{OK: false, ErrorCode: "HALF_PRICE_REDEEM_PENDING_REVIEW", ErrorMessage: "Plus half-price redemption requires reconciliation before retry"}, nil
			}
			logger.Warnw("auto_redeem_failed_fallback_to_cdk", "order_no", req.DownstreamOrderNo, "error", err)
			// 自动兑换如果遇到鉴权/参数问题，以包含 CDK 凭证的方式交付给用户手工兑换
			return &CreateUpstreamOrderResp{
				OK:           true,
				OrderID:      firstCDK.ID,
				OrderNo:      fmt.Sprintf("ZOVO-CDK-%d", firstCDK.ID),
				Status:       "delivered",
				Amount:       fmt.Sprintf("%.2f", float64(totalFeeMinor)/100.0),
				Currency:     "USD",
				ErrorMessage: "automatic redeem failed; CDK delivered for manual redemption",
				Fulfillment:  &UpstreamFulfillment{Type: "manual", Status: "delivered", Payload: cdkCode, DeliveryData: jsonmap.JSON{"cdk_codes": []string{cdkCode}, "plan": planKey}},
			}, nil
		}

		// 兑换订单已受理
		return &CreateUpstreamOrderResp{
			OK:       true,
			OrderID:  redeemOrderID,
			OrderNo:  req.DownstreamOrderNo,
			Status:   "fulfilling",
			Amount:   fmt.Sprintf("%.2f", float64(totalFeeMinor)/100.0),
			Currency: "USD",
		}, nil
	}

	// 4. 无 session 模式：直接作为 CDK 激活码完成交付
	return &CreateUpstreamOrderResp{
		OK:          true,
		OrderID:     firstCDK.ID,
		OrderNo:     fmt.Sprintf("ZOVO-CDK-%d", firstCDK.ID),
		Status:      "delivered",
		Amount:      fmt.Sprintf("%.2f", float64(totalFeeMinor)/100.0),
		Currency:    "USD",
		Fulfillment: &UpstreamFulfillment{Type: "manual", Status: "delivered", Payload: strings.Join(codes, "\n"), DeliveryData: jsonmap.JSON{"cdk_codes": codes, "plan": planKey}},
	}, nil
}

func stablePlanID(key string) uint {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(strings.TrimSpace(key))))
	return uint(h.Sum32() & 0x7fffffff)
}

// executeAutoRedeem 执行四步 CDK 自动化兑换流程
func (a *ZovoCardAdapter) executeAutoRedeem(ctx context.Context, cdkCode, sessionToken, orderNo string) (uint, error) {
	deviceID := fmt.Sprintf("dj-bridge-%s", orderNo)

	// Step 1: Preview (POST /api/v1/cdk/preview)
	previewReq := map[string]string{"code": cdkCode}
	previewData, _ := json.Marshal(previewReq)
	req1, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/v1/cdk/preview", bytes.NewReader(previewData))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("X-Redemption-Device", deviceID)

	resp1, err := a.client.Do(req1)
	if err != nil {
		return 0, fmt.Errorf("preview request error: %w", err)
	}
	defer resp1.Body.Close()
	b1, _ := io.ReadAll(resp1.Body)

	var previewResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			RedemptionToken string `json:"redemption_token"`
			Plan            string `json:"plan"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b1, &previewResp); err != nil || previewResp.Data.RedemptionToken == "" {
		return 0, errors.New("preview failed")
	}
	if resp1.StatusCode < 200 || resp1.StatusCode >= 300 || previewResp.Code != 0 {
		return 0, errors.New("preview rejected by ZovoCard")
	}

	// Step 2: Preflight (POST /api/v1/cdk/preflight)
	preflightReq := map[string]interface{}{
		"redemption_token": previewResp.Data.RedemptionToken,
		"credential": map[string]string{
			"mode":    "session",
			"session": sessionToken,
		},
	}
	preflightData, _ := json.Marshal(preflightReq)
	req2, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/api/v1/cdk/preflight", bytes.NewReader(preflightData))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("X-Redemption-Device", deviceID)

	resp2, err := a.client.Do(req2)
	if err != nil {
		return 0, fmt.Errorf("preflight request error: %w", err)
	}
	defer resp2.Body.Close()
	b2, _ := io.ReadAll(resp2.Body)

	var preflightResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			PreflightToken string `json:"preflight_token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b2, &preflightResp); err != nil || preflightResp.Data.PreflightToken == "" {
		return 0, errors.New("preflight failed")
	}
	if resp2.StatusCode < 200 || resp2.StatusCode >= 300 || preflightResp.Code != 0 {
		return 0, errors.New("preflight rejected by ZovoCard")
	}

	// Step 3: Redeem (POST /openapi/v1/gpt-direct/cdks/redeem or /api/v1/cdk/redeem)
	redeemReq := map[string]interface{}{
		"redemption_token":  previewResp.Data.RedemptionToken,
		"preflight_token":   preflightResp.Data.PreflightToken,
		"client_request_id": orderNo,
	}
	redeemData, _ := json.Marshal(redeemReq)
	req3, _ := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/openapi/v1/gpt-direct/cdks/redeem", bytes.NewReader(redeemData))
	req3.Header.Set("Content-Type", "application/json")
	req3.Header.Set("X-API-Key", a.apiKey)
	req3.Header.Set("X-Redemption-Device", deviceID)

	resp3, err := a.client.Do(req3)
	if err != nil {
		return 0, fmt.Errorf("redeem request error: %w", err)
	}
	defer resp3.Body.Close()
	b3, _ := io.ReadAll(resp3.Body)

	var redeemResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Order struct {
				OrderID uint   `json:"order_id"`
				Status  string `json:"status"`
			} `json:"order"`
			ID uint `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(b3, &redeemResp); err != nil {
		return 0, errors.New("redeem response invalid")
	}
	if resp3.StatusCode < 200 || resp3.StatusCode >= 300 || redeemResp.Code != 0 {
		return 0, errors.New("redeem rejected by ZovoCard")
	}

	orderID := redeemResp.Data.Order.OrderID
	if orderID == 0 {
		orderID = redeemResp.Data.ID
	}
	if orderID == 0 {
		return 0, errors.New("redeem response omitted order id")
	}

	return orderID, nil
}

// GetOrder 查询订单状态
func (a *ZovoCardAdapter) GetOrder(ctx context.Context, orderID uint) (*UpstreamOrderDetail, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/openapi/v1/gpt-direct/cdk-orders/%d", a.baseURL, orderID), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch order: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("order %d not found", orderID)
	}

	var detailResp struct {
		Code int    `json:"code"`
		Msg  string `json:"msg"`
		Data struct {
			Order struct {
				OrderID       uint   `json:"order_id"`
				CDKCodePrefix string `json:"cdk_code_prefix"`
				Status        string `json:"status"`
				Stage         string `json:"stage"`
				AccountEmail  string `json:"account_email"`
				CardLastFour  string `json:"card_last_four"`
				Plan          string `json:"plan"`
				AmountMinor   int64  `json:"amount_minor"`
				RenewalStatus string `json:"renewal_status"`
				UpdatedAt     string `json:"updated_at"`
			} `json:"order"`
			Events []struct {
				Stage      string `json:"stage"`
				Status     string `json:"status"`
				Type       string `json:"type"`
				EventType  string `json:"event_type"`
				CreatedAt  string `json:"created_at"`
				OccurredAt string `json:"occurred_at"`
			} `json:"events"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &detailResp); err != nil {
		return nil, fmt.Errorf("unmarshal order response: %w", err)
	}

	ord := detailResp.Data.Order
	st := strings.ToLower(strings.TrimSpace(ord.Status))

	var fulfillment *UpstreamFulfillment
	localStatus := "pending"

	switch st {
	case "completed", "success", "paid":
		localStatus = "delivered"
		payload := fmt.Sprintf("✅ ChatGPT 官方订阅充值已完成！\n套餐：%s\n账号：%s\n卡号后四位：%s\n自动续费状态：%s\n交付时间：%s",
			ord.Plan, ord.AccountEmail, ord.CardLastFour, ord.RenewalStatus, time.Now().Format("2006-01-02 15:04:05"))
		fulfillment = &UpstreamFulfillment{
			Type:    "auto",
			Status:  "delivered",
			Payload: payload,
			DeliveryData: jsonmap.JSON{
				"plan":           ord.Plan,
				"card_last_four": ord.CardLastFour,
				"account_email":  ord.AccountEmail,
				"renewal_status": ord.RenewalStatus,
			},
		}
	case "failed_precharge", "declined", "cancelled", "failed":
		localStatus = "canceled"
	default:
		localStatus = "fulfilling"
	}

	result := &UpstreamOrderDetail{
		OrderID:     ord.OrderID,
		OrderNo:     strconv.Itoa(int(ord.OrderID)),
		Status:      localStatus,
		Stage:       strings.TrimSpace(ord.Stage),
		Amount:      fmt.Sprintf("%.2f", float64(ord.AmountMinor)/100.0),
		Currency:    "USD",
		Fulfillment: fulfillment,
	}
	for _, event := range detailResp.Data.Events {
		stage := event.Stage
		if stage == "" {
			stage = event.EventType
		}
		if stage == "" {
			stage = event.Type
		}
		at := event.CreatedAt
		if at == "" {
			at = event.OccurredAt
		}
		result.Events = append(result.Events, UpstreamProgressEvent{Stage: stage, Status: event.Status, At: at})
	}
	return result, nil
}

// CancelOrder 取消订单
func (a *ZovoCardAdapter) CancelOrder(ctx context.Context, orderID uint) error {
	return errors.New("ZovoCard CDK cancellation is not supported by this adapter")
}

// DownloadImage 下载图片
func (a *ZovoCardAdapter) DownloadImage(ctx context.Context, imageURL string) (string, error) {
	if imageURL == "" {
		return "", errors.New("empty image url")
	}
	resp, err := a.client.Get(imageURL)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download image failed: HTTP %d", resp.StatusCode)
	}

	b := make([]byte, 16)
	_, _ = rand.Read(b)
	filename := hex.EncodeToString(b) + ".png"
	localPath := filepath.Join(a.uploadsDir, filename)

	out, err := os.Create(localPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return "/uploads/" + filename, err
}

func (a *ZovoCardAdapter) doGet(ctx context.Context, path string, target interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+path, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return json.NewDecoder(resp.Body).Decode(target)
}
