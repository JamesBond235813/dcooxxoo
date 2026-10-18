package fivesim

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"

	"github.com/dujiao-next/internal/logger"
)

var ErrMissingToken = errors.New("5sim API token is not configured")

type SMSItem struct {
	CreatedAt string `json:"created_at"`
	Date      string `json:"date"`
	Sender    string `json:"sender"`
	Text      string `json:"text"`
	Code      string `json:"code"`
}

type ActivationOrder struct {
	ID               int64     `json:"id"`
	Phone            string    `json:"phone"`
	Operator         string    `json:"operator"`
	Product          string    `json:"product"`
	Price            float64   `json:"price"`
	Status           string    `json:"status"` // PENDING, RECEIVED, CANCELED, TIMEOUT, FINISHED, BANNED
	Expires          string    `json:"expires"`
	SMS              []SMSItem `json:"sms"`
	CreatedAt        string    `json:"created_at"`
	Country          string    `json:"country"`
	Forwarding       bool      `json:"forwarding"`
	ForwardingNumber string    `json:"forwarding_number"`
}

type Profile struct {
	ID      int64   `json:"id"`
	Email   string  `json:"email"`
	Balance float64 `json:"balance"`
	Rating  float64 `json:"rating"`
}

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func NewClient(baseURL, token string) *Client {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	if base == "" {
		base = "https://5sim.net/v1"
	}
	return &Client{
		baseURL: base,
		token:   strings.TrimSpace(token),
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
		},
	}
}

func (c *Client) request(ctx context.Context, method, endpoint string) ([]byte, int, error) {
	if strings.TrimSpace(c.token) == "" {
		return nil, 0, ErrMissingToken
	}
	url := fmt.Sprintf("%s%s", c.baseURL, endpoint)
	req, err := http.NewRequestWithContext(ctx, method, url, nil)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, nil
}

// GetProfile 获取个人信息与余额
func (c *Client) GetProfile(ctx context.Context) (*Profile, error) {
	body, code, err := c.request(ctx, http.MethodGet, "/user/profile")
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("5sim get profile failed (HTTP %d): %s", code, string(body))
	}
	var p Profile
	if err := json.Unmarshal(body, &p); err != nil {
		return nil, err
	}
	return &p, nil
}

// BuyActivation 购买激活接码号码
func (c *Client) BuyActivation(ctx context.Context, country, operator, product string) (*ActivationOrder, error) {
	country = strings.TrimSpace(country)
	operator = strings.TrimSpace(operator)
	product = strings.TrimSpace(product)
	if country == "" {
		country = "any"
	}
	if operator == "" {
		operator = "any"
	}
	if product == "" {
		product = "openai"
	}

	endpoint := fmt.Sprintf("/user/buy/activation/%s/%s/%s", country, operator, product)
	body, code, err := c.request(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		msg := strings.TrimSpace(string(body))
		return nil, fmt.Errorf("5sim buy failed (HTTP %d): %s", code, msg)
	}

	var order ActivationOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("unmarshal 5sim response: %w, raw: %s", err, string(body))
	}
	return &order, nil
}

// CheckOrder 查询订单状态与短信
func (c *Client) CheckOrder(ctx context.Context, id int64) (*ActivationOrder, error) {
	endpoint := fmt.Sprintf("/user/check/%d", id)
	body, code, err := c.request(ctx, http.MethodGet, endpoint)
	if err != nil {
		return nil, err
	}
	if code != http.StatusOK {
		return nil, fmt.Errorf("5sim check order failed (HTTP %d): %s", code, string(body))
	}

	var order ActivationOrder
	if err := json.Unmarshal(body, &order); err != nil {
		return nil, fmt.Errorf("unmarshal 5sim response: %w, raw: %s", err, string(body))
	}
	return &order, nil
}

// FinishOrder 结单（收码成功后释放）
func (c *Client) FinishOrder(ctx context.Context, id int64) error {
	endpoint := fmt.Sprintf("/user/finish/%d", id)
	body, code, err := c.request(ctx, http.MethodGet, endpoint)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("5sim finish failed (HTTP %d): %s", code, string(body))
	}
	return nil
}

// CancelOrder 取消订单（未收码退款）
func (c *Client) CancelOrder(ctx context.Context, id int64) error {
	endpoint := fmt.Sprintf("/user/cancel/%d", id)
	body, code, err := c.request(ctx, http.MethodGet, endpoint)
	if err != nil {
		return err
	}
	if code != http.StatusOK {
		return fmt.Errorf("5sim cancel failed (HTTP %d): %s", code, string(body))
	}
	return nil
}

type priceQuote struct {
	Cost   float64 `json:"cost"`
	Count  int     `json:"count"`
	Rate   float64 `json:"rate"`
	Rate24 float64 `json:"rate24"`
}

type numberCandidate struct {
	country  string
	operator string
	price    float64
	rate     float64
	score    float64
}

// AutoBuyFromPool 基于实时库存、价格和短信接收率选号。国家池含 any 时允许全部国家。
func (c *Client) AutoBuyFromPool(ctx context.Context, pool []string, product string, maxPrice float64, priorFailures map[string]int) (*ActivationOrder, error) {
	if maxPrice <= 0 {
		maxPrice = 0.5
	}
	if product == "" {
		product = "openai"
	}
	body, status, err := c.request(ctx, http.MethodGet, "/guest/prices?product="+url.QueryEscape(product))
	if err != nil {
		return nil, fmt.Errorf("fetch number prices: %w", err)
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("fetch number prices: HTTP %d", status)
	}
	var prices map[string]map[string]map[string]priceQuote
	if err := json.Unmarshal(body, &prices); err != nil {
		return nil, fmt.Errorf("decode number prices: %w", err)
	}
	allowed := map[string]bool{}
	for _, country := range pool {
		allowed[strings.ToLower(strings.TrimSpace(country))] = true
	}
	allowAll := len(allowed) == 0 || allowed["any"]
	candidates := make([]numberCandidate, 0)
	for country, operators := range prices[product] {
		if !allowAll && !allowed[country] {
			continue
		}
		for operator, quote := range operators {
			if quote.Count < 100 || quote.Cost <= 0 || quote.Cost > maxPrice || quote.Rate < 25 {
				continue
			}
			if quote.Rate24 > 0 && quote.Rate24 < 20 {
				continue
			}
			rate := quote.Rate
			if quote.Rate24 > 0 {
				rate = 0.6*quote.Rate + 0.4*quote.Rate24
			}
			if rate < 30 {
				continue
			}
			failures := priorFailures[country+"/"+operator]
			candidates = append(candidates, numberCandidate{
				country: country, operator: operator, price: quote.Cost, rate: rate,
				score: quote.Cost / (rate / 100) * (1 + float64(failures)),
			})
		}
	}
	if len(candidates) == 0 {
		return nil, errors.New("no number pool meets stock, delivery rate and price limits")
	}
	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].score == candidates[j].score {
			return candidates[i].rate > candidates[j].rate
		}
		return candidates[i].score < candidates[j].score
	})
	var lastErr error
	for _, candidate := range candidates {
		logger.Infow("5sim_attempting_pool", "country", candidate.country, "operator", candidate.operator, "quoted_price", candidate.price, "delivery_rate", candidate.rate)
		order, buyErr := c.BuyActivation(ctx, candidate.country, candidate.operator, product)
		if buyErr != nil {
			lastErr = buyErr
			continue
		}
		if order == nil || order.ID == 0 || order.Phone == "" {
			if order != nil && order.ID != 0 {
				if cancelErr := c.CancelOrder(ctx, order.ID); cancelErr != nil {
					return nil, fmt.Errorf("empty activation and cancel failed: %w", cancelErr)
				}
			}
			lastErr = errors.New("provider returned an empty activation")
			continue
		}
		if order.Price > maxPrice {
			if cancelErr := c.CancelOrder(ctx, order.ID); cancelErr != nil {
				return nil, fmt.Errorf("number price exceeded cap and cancel failed: %w", cancelErr)
			}
			lastErr = errors.New("number price exceeded cap")
			continue
		}
		return order, nil
	}
	return nil, fmt.Errorf("all eligible number pools exhausted: %w", lastErr)
}
