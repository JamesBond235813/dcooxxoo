package manualcryptoadapter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	"github.com/dujiao-next/internal/shared/jsonmap"
	"github.com/shopspring/decimal"
)

const binanceP2PAdsEndpoint = "https://www.binance.com/bapi/c2c/v1/public/c2c/agent/ad-list"

var txHashPattern = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{64}$`)
var evmAddressPattern = regexp.MustCompile(`^0x[0-9a-fA-F]{40}$`)
var tronAddressPattern = regexp.MustCompile(`^T[1-9A-HJ-NP-Za-km-z]{33}$`)

type adapter struct {
	client *http.Client
}

func NewManualCryptoAdapter() paymentcontract.GatewayProvider {
	return &adapter{client: &http.Client{Timeout: 6 * time.Second}}
}

// CurrentSellPrice exposes the same Binance P2P CNY/USDT quote to gateways
// that perform their own on-chain settlement. No Binance credentials are used.
func CurrentSellPrice(ctx context.Context, cnyAmount decimal.Decimal) (decimal.Decimal, error) {
	return (&adapter{client: &http.Client{Timeout: 6 * time.Second}}).currentSellPrice(ctx, cnyAmount)
}

func (a *adapter) Type() string { return constants.PaymentProviderManualCrypto + ":" }

func (a *adapter) ValidateConfig(cfg jsonmap.JSON, channelType string) error {
	if strings.ToLower(strings.TrimSpace(channelType)) != "usdt" {
		return fmt.Errorf("%w: manual crypto supports USDT only", paymentcontract.ErrGatewayUnsupportedChannel)
	}
	address := strings.TrimSpace(asString(cfg["address"]))
	network := strings.ToLower(strings.TrimSpace(asString(cfg["network"])))
	if network != "bep20" && network != "trc20" && network != "erc20" {
		return fmt.Errorf("%w: network must be BEP20, TRC20, or ERC20", paymentcontract.ErrGatewayConfigInvalid)
	}
	if (network == "trc20" && !tronAddressPattern.MatchString(address)) ||
		((network == "bep20" || network == "erc20") && !evmAddressPattern.MatchString(address)) {
		return fmt.Errorf("%w: address format does not match the selected network", paymentcontract.ErrGatewayConfigInvalid)
	}
	return nil
}

type adsResponse struct {
	Code string `json:"code"`
	Data struct {
		Items []struct {
			Price          decimal.Decimal `json:"price"`
			MinTransAmount decimal.Decimal `json:"minTransAmount"`
			MaxTransAmount decimal.Decimal `json:"maxTransAmount"`
			TradableAmount decimal.Decimal `json:"tradableAmount"`
			TradeMethods   []string        `json:"tradeMethods"`
			Advertiser     struct {
				UserType string `json:"userType"`
			} `json:"advertiser"`
		} `json:"items"`
	} `json:"data"`
}

func (a *adapter) CreatePayment(ctx context.Context, cfg jsonmap.JSON, input paymentcontract.GatewayCreateInput) (*paymentcontract.GatewayCreateResult, error) {
	if err := a.ValidateConfig(cfg, input.ChannelType); err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(input.Currency), "CNY") || !input.Amount.Decimal.IsPositive() {
		return nil, fmt.Errorf("%w: manual USDT collection requires a positive CNY amount", paymentcontract.ErrGatewayUnsupportedChannel)
	}

	price, err := a.currentSellPrice(ctx, input.Amount.Decimal)
	if err != nil {
		return nil, err
	}
	// P2P price is CNY per USDT. Divide the CNY due amount by that quote;
	// round up to cents so the received USDT cannot under-cover the order.
	amount := input.Amount.Decimal.Div(price).RoundCeil(2)
	if !amount.IsPositive() {
		return nil, fmt.Errorf("%w: converted USDT amount is zero", paymentcontract.ErrGatewayResponseInvalid)
	}
	// Store the effective conversion rate after USDT cent-rounding. The shared
	// payment settlement logic divides the callback amount by this value to
	// recover the CNY amount covered by the payment.
	rate := amount.Div(input.Amount.Decimal)
	network := strings.ToUpper(strings.TrimSpace(asString(cfg["network"])))
	address := strings.TrimSpace(asString(cfg["address"]))
	chain, tokenID := "", ""
	switch network {
	case "TRC20":
		chain, tokenID = "TRON", "tron-usdt"
	case "BEP20":
		chain, tokenID = "BNB Smart Chain", "bsc-usdt"
	case "ERC20":
		chain, tokenID = "Ethereum", "ethereum-usdt"
	}
	return &paymentcontract.GatewayCreateResult{
		ProviderRef:  input.OrderNo,
		QRCodeURL:    address,
		CurrencySent: "USDT",
		AmountSent:   amount.StringFixed(2),
		Payload: jsonmap.JSON{
			"address":             address,
			"chain":               chain,
			"token_id":            tokenID,
			"chain_amount":        amount.StringFixed(2),
			"exchange_rate":       rate.String(),
			"cny_per_usdt":        price.String(),
			"quote_source":        "binance_p2p_sell_alipay",
			"quote_time":          time.Now().UTC().Format(time.RFC3339Nano),
			"original_amount":     input.Amount.Decimal.StringFixed(2),
			"original_currency":   input.Currency,
			"manual_review":       true,
			"manual_proof_status": "not_submitted",
		},
	}, nil
}

func (a *adapter) currentSellPrice(ctx context.Context, cnyAmount decimal.Decimal) (decimal.Decimal, error) {
	endpoint, _ := url.Parse(binanceP2PAdsEndpoint)
	query := endpoint.Query()
	query.Set("fiat", "CNY")
	query.Set("asset", "USDT")
	query.Set("tradeType", "SELL")
	query.Set("limit", "20")
	query.Set("order", "price_desc")
	query.Add("tradeMethodIdentifiers", "ALIPAY")
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w: quote request could not be prepared", paymentcontract.ErrGatewayRequestFailed)
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "Dujiao-Next/1.0 (P2P quote)")
	client := a.client
	if client == nil {
		client = &http.Client{Timeout: 6 * time.Second}
	}
	response, err := client.Do(request)
	if err != nil {
		return decimal.Zero, fmt.Errorf("%w: Binance P2P quote unavailable", paymentcontract.ErrGatewayRequestFailed)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return decimal.Zero, fmt.Errorf("%w: Binance P2P quote returned HTTP %d", paymentcontract.ErrGatewayResponseInvalid, response.StatusCode)
	}
	var result adsResponse
	if err := json.NewDecoder(response.Body).Decode(&result); err != nil || result.Code != "000000" {
		return decimal.Zero, fmt.Errorf("%w: Binance P2P quote response invalid", paymentcontract.ErrGatewayResponseInvalid)
	}
	prices := make([]decimal.Decimal, 0, len(result.Data.Items))
	for _, item := range result.Data.Items {
		if !item.Price.IsPositive() || !item.TradableAmount.IsPositive() || item.MinTransAmount.GreaterThan(item.MaxTransAmount) {
			continue
		}
		if !hasAlipay(item.TradeMethods) || !strings.EqualFold(strings.TrimSpace(item.Advertiser.UserType), "merchant") {
			continue
		}
		// This is a reference conversion quote, not a Binance P2P order for
		// each shop purchase. Small shop orders may fall below every ad's
		// minimum; applying that limit here would make checkout unavailable.
		prices = append(prices, item.Price)
	}
	if len(prices) == 0 {
		return decimal.Zero, errors.New("no Binance P2P USDT sell merchant quote accepting Alipay")
	}
	sort.Slice(prices, func(i, j int) bool { return prices[i].LessThan(prices[j]) })
	return prices[len(prices)/2], nil
}

func hasAlipay(methods []string) bool {
	for _, method := range methods {
		if strings.EqualFold(strings.TrimSpace(method), "ALIPAY") {
			return true
		}
	}
	return false
}

func asString(value interface{}) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}
