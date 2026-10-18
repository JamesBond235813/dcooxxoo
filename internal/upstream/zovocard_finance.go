package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
)

// ZovoFinanceSnapshot contains only safe accounting fields; CDK codes and card PANs are never retained.
type ZovoFinanceSnapshot struct {
	Balance      map[string]any   `json:"balance"`
	BalanceLogs  []map[string]any `json:"balance_logs"`
	CDKs         []map[string]any `json:"cdks"`
	CDKOrders    []map[string]any `json:"cdk_orders"`
	Transactions []map[string]any `json:"transactions"`
	Cards        []map[string]any `json:"cards"`
}

func (a *ZovoCardAdapter) GetFinanceSnapshot(ctx context.Context) (*ZovoFinanceSnapshot, error) {
	out := &ZovoFinanceSnapshot{BalanceLogs: []map[string]any{}, CDKs: []map[string]any{}, CDKOrders: []map[string]any{}, Transactions: []map[string]any{}, Cards: []map[string]any{}}
	balance, err := a.getFinanceData(ctx, "/openapi/v1/balance")
	if err != nil {
		return nil, fmt.Errorf("读取上游余额失败: %w", err)
	}
	balanceData, ok := balance.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("读取上游余额失败: 上游响应 data 格式异常")
	}
	out.Balance = safeFinanceFields(balanceData, "balance", "spendable_balance", "account_reserve_amount", "currency")
	for _, endpoint := range []struct {
		path   string
		fields []string
		target *[]map[string]any
	}{
		{"/openapi/v1/balance-logs", []string{"id", "log_id", "transaction_id", "created_at", "type", "amount", "before", "after", "currency", "remark"}, &out.BalanceLogs},
		{"/openapi/v1/gpt-direct/cdks", []string{"id", "plan", "status", "created_at", "fee_amount_minor", "payment_country"}, &out.CDKs},
		{"/openapi/v1/gpt-direct/cdk-orders", []string{"id", "order_id", "cdk_id", "plan", "status", "stage", "created_at", "completed_at", "updated_at", "amount", "currency", "service_fee_usd_minor"}, &out.CDKOrders},
		{"/openapi/v1/cards/all-transactions", []string{"id", "transaction_id", "auth_id", "auth_time", "auth_amount", "auth_currency", "settle_amount", "settle_currency", "status", "type", "merchant_name", "create_time", "description", "local_card_id"}, &out.Transactions},
		{"/openapi/v1/cards", []string{"id", "product_code", "status", "open_fee", "created_at", "available_amount", "currency"}, &out.Cards},
	} {
		rows, fetchErr := a.getFinanceList(ctx, endpoint.path, 20)
		if fetchErr != nil {
			return nil, fmt.Errorf("读取上游流水失败 (%s): %w", endpoint.path, fetchErr)
		}
		for _, row := range rows {
			*endpoint.target = append(*endpoint.target, safeFinanceFields(row, endpoint.fields...))
		}
	}
	return out, nil
}

func (a *ZovoCardAdapter) getFinanceList(ctx context.Context, path string, maxPages int) ([]map[string]any, error) {
	all := make([]map[string]any, 0)
	for page := 1; page <= maxPages; page++ {
		q := url.Values{}
		q.Set("page", strconv.Itoa(page))
		q.Set("page_size", "100")
		payload, err := a.getFinanceData(ctx, path+"?"+q.Encode())
		if err != nil {
			return nil, err
		}
		rows := financeRows(payload)
		all = append(all, rows...)
		if len(rows) < 100 {
			break
		}
	}
	return all, nil
}

func (a *ZovoCardAdapter) getFinanceData(ctx context.Context, path string) (any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, a.baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-API-Key", a.apiKey)
	req.Header.Set("Accept", "application/json")
	resp, err := a.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("上游 HTTP %d", resp.StatusCode)
	}
	var envelope map[string]any
	if err := json.Unmarshal(body, &envelope); err != nil {
		return nil, err
	}
	if code, ok := envelope["code"].(float64); ok && code != 0 {
		return nil, fmt.Errorf("上游返回错误: %v", envelope["msg"])
	}
	data, ok := envelope["data"]
	if !ok || data == nil {
		keys := make([]string, 0, len(envelope))
		for key := range envelope {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		return nil, fmt.Errorf("上游响应缺少 data（顶层字段：%s）", strings.Join(keys, ", "))
	}
	return data, nil
}

func financeRows(data any) []map[string]any {
	var raw []any
	switch payload := data.(type) {
	case []any:
		raw = payload
	case map[string]any:
		if list, ok := payload["list"].([]any); ok {
			raw = list
		} else if list, ok := payload["data"].([]any); ok {
			raw = list
		}
	}
	rows := make([]map[string]any, 0, len(raw))
	for _, value := range raw {
		if row, ok := value.(map[string]any); ok {
			rows = append(rows, row)
		}
	}
	return rows
}

func safeFinanceFields(source map[string]any, fields ...string) map[string]any {
	out := make(map[string]any, len(fields))
	for _, field := range fields {
		if value, exists := source[field]; exists {
			out[field] = value
		}
	}
	return out
}

func financeString(row map[string]any, key string) string {
	if value, ok := row[key].(string); ok {
		return strings.TrimSpace(value)
	}
	return ""
}
