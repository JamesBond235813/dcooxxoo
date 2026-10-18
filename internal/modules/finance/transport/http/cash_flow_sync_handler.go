package financehttp

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strconv"
	"strings"
	"time"

	financedomain "github.com/dujiao-next/internal/modules/finance/domain"
	"github.com/dujiao-next/internal/platform/database/gormdb"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
	"gorm.io/gorm/clause"
)

var financeLongNumberPattern = regexp.MustCompile(`\b(?:[0-9][ -]?){12,19}\b`)

// SyncZovoFinance is the only report operation that calls ZovoCard. It stores
// sanitized, idempotent rows in this service's database for future report reads.
func (h *AdminHandler) SyncZovoFinance(c *gin.Context) {
	snapshot, connectionID, err := h.fetchUpstreamWithID(c.Request.Context())
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.upstream", err)
		return
	}
	rate := h.configuredUSDRate()
	if rate.LessThanOrEqual(decimal.Zero) {
		rate = decimal.NewFromInt(1)
	}
	now := time.Now().UTC()
	currency := strings.ToUpper(financeValueString(snapshot.Balance, "currency"))
	if currency == "" {
		currency = "USD"
	}
	rows := make([]financedomain.UpstreamFinanceRecord, 0, 1+len(snapshot.BalanceLogs)+len(snapshot.CDKs)+len(snapshot.CDKOrders)+len(snapshot.Transactions)+len(snapshot.Cards))
	balance := financeAmount(snapshot.Balance["balance"])
	available := financeAmount(snapshot.Balance["spendable_balance"]).StringFixed(2)
	rows = append(rows, buildZovoRecord(connectionID, "balance_snapshot", map[string]any{"id": now.Format(time.RFC3339Nano), "created_at": now.Format(time.RFC3339Nano), "balance": balance.String(), "spendable_balance": available, "currency": currency}, now, rate, currency, balance, "snapshot", "zovo_balance_snapshot", "ZovoCard 账户", "可用余额 "+available))
	for _, raw := range snapshot.BalanceLogs {
		amount := financeAmount(raw["amount"])
		before, hasBefore := optionalFinanceAmount(raw["before"])
		after, hasAfter := optionalFinanceAmount(raw["after"])
		direction := "unknown"
		if hasBefore && hasAfter {
			delta := after.Sub(before)
			if delta.IsPositive() {
				direction = "in"
				amount = delta
			}
			if delta.IsNegative() {
				direction = "out"
				amount = delta.Abs()
			}
		} else if amount.IsPositive() {
			if strings.Contains(strings.ToLower(financeValueString(raw, "type")), "recharge") || strings.Contains(financeValueString(raw, "type"), "充值") {
				direction = "in"
			} else {
				direction = "out"
			}
		}
		kind := strings.ToLower(financeValueString(raw, "type"))
		category := "zovo_balance_change"
		if strings.Contains(kind, "fee") || strings.Contains(kind, "手续费") || strings.Contains(kind, "手續費") {
			category = "zovo_fee"
		}
		if strings.Contains(kind, "transfer") || strings.Contains(kind, "card_recharge") || strings.Contains(kind, "转卡") || strings.Contains(kind, "轉卡") {
			category = "zovo_internal_transfer"
			direction = "transfer"
		}
		cur := strings.ToUpper(financeValueString(raw, "currency"))
		if cur == "" {
			cur = currency
		}
		rows = append(rows, buildZovoRecord(connectionID, "balance_log", raw, now, rate, cur, amount, direction, category, "ZovoCard 账户", ""))
	}
	for _, raw := range snapshot.CDKs {
		fee := financeAmount(raw["fee_amount_minor"]).Div(decimal.NewFromInt(100))
		rows = append(rows, buildZovoRecord(connectionID, "cdk_issue", raw, now, rate, "USD", fee, "out", "zovo_cdk_service_fee", "ZovoCard CDK", financeFirstString(raw, "plan", "payment_country")))
	}
	for _, raw := range snapshot.CDKOrders {
		amount := financeAmount(raw["service_fee_usd_minor"]).Div(decimal.NewFromInt(100))
		rows = append(rows, buildZovoRecord(connectionID, "cdk_order", raw, now, rate, "USD", amount, "unknown", "zovo_cdk_order", "ZovoCard CDK", financeFirstString(raw, "plan", "stage")))
	}
	for _, raw := range snapshot.Transactions {
		amount, cur := financeMoneyAmount(raw, "settle_amount", "settle_currency")
		if amount.IsZero() {
			amount, cur = financeMoneyAmount(raw, "auth_amount", "auth_currency")
		}
		if cur == "" {
			cur = "USD"
		}
		direction := "unknown"
		status := strings.ToLower(financeValueString(raw, "status"))
		if status == "success" || status == "completed" || status == "settled" || status == "posted" {
			direction = "out"
		}
		rows = append(rows, buildZovoRecord(connectionID, "card_transaction", raw, now, rate, cur, amount.Abs(), direction, "zovo_card_transaction", "ZovoCard 卡片", financeValueString(raw, "merchant_name")))
	}
	for _, raw := range snapshot.Cards {
		amount := financeAmount(raw["open_fee"])
		if amount.IsZero() {
			continue
		}
		cur := strings.ToUpper(financeValueString(raw, "currency"))
		if cur == "" {
			cur = "USD"
		}
		rows = append(rows, buildZovoRecord(connectionID, "card_opening", raw, now, rate, cur, amount.Abs(), "out", "zovo_card_opening_fee", "ZovoCard 卡片", "开卡费"))
	}
	if len(rows) > 0 {
		if err := gormdb.DB.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "connection_id"}, {Name: "source"}, {Name: "source_key"}}, DoUpdates: clause.AssignmentColumns([]string{"event_date", "event_time", "event_type", "status", "direction", "category", "account", "amount", "currency", "exchange_rate_snapshot", "cny_amount", "reference", "description", "raw_json", "synced_at", "updated_at"})}).Create(&rows).Error; err != nil {
			ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
			return
		}
	}
	response.Success(c, gin.H{"connection_id": connectionID, "synced_at": now, "upserted": len(rows), "balance_logs": len(snapshot.BalanceLogs), "cdks": len(snapshot.CDKs), "cdk_orders": len(snapshot.CDKOrders), "card_transactions": len(snapshot.Transactions), "cards": len(snapshot.Cards)})
}

func buildZovoRecord(connectionID uint, source string, raw map[string]any, syncedAt time.Time, rate decimal.Decimal, currency string, amount decimal.Decimal, direction, category, account, desc string) financedomain.UpstreamFinanceRecord {
	eventTime := financeEventTime(raw, source)
	safeRaw := sanitizeFinanceRow(raw)
	b, _ := json.Marshal(safeRaw)
	key := financeFirstString(raw, "id", "log_id", "transaction_id", "auth_id")
	if key == "" {
		sum := sha256.Sum256(append([]byte(source+":"), b...))
		key = hex.EncodeToString(sum[:])
	}
	fx := decimal.NewFromInt(1)
	if currency == "USD" {
		fx = rate
	}
	if currency == "CNY" {
		fx = decimal.NewFromInt(1)
	}
	ref := financeFirstString(safeRaw, "auth_id", "transaction_id", "id", "log_id", "local_card_id")
	if desc == "" {
		desc = financeFirstString(safeRaw, "description", "remark", "type")
	}
	return financedomain.UpstreamFinanceRecord{ConnectionID: connectionID, Source: source, SourceKey: key, EventDate: eventTime.In(financeLoc()).Format("2006-01-02"), EventTime: eventTime, EventType: financeFirstString(safeRaw, "type"), Status: financeFirstString(safeRaw, "status"), Direction: direction, Category: category, Account: sanitizeFinanceText(account), Amount: amount.Round(6), Currency: currency, ExchangeRateSnapshot: fx.Round(6), CNYAmount: amount.Mul(fx).Round(6), Reference: sanitizeFinanceText(ref), Description: sanitizeFinanceText(desc), RawJSON: string(b), SyncedAt: syncedAt}
}

func financeEventTime(raw map[string]any, source string) time.Time {
	keys := []string{"created_at", "create_time", "auth_time", "completed_at", "updated_at"}
	if source == "card_transaction" {
		keys = []string{"auth_time", "create_time", "created_at"}
	}
	if source == "cdk_order" {
		keys = []string{"completed_at", "updated_at", "created_at"}
	}
	for _, key := range keys {
		v := raw[key]
		if v == nil {
			continue
		}
		if n, ok := v.(float64); ok {
			sec := int64(n)
			if sec > 1e12 {
				sec /= 1000
			}
			return time.Unix(sec, 0).UTC()
		}
		s := strings.TrimSpace(fmt.Sprint(v))
		if s == "" {
			continue
		}
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			if n > 1e12 {
				n /= 1000
			}
			return time.Unix(n, 0).UTC()
		}
		for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
			if t, err := time.Parse(layout, s); err == nil {
				return t.UTC()
			}
		}
	}
	return time.Now().UTC()
}

func financeAmount(v any) decimal.Decimal {
	if v == nil {
		return decimal.Zero
	}
	switch n := v.(type) {
	case float64:
		if math.IsNaN(n) || math.IsInf(n, 0) {
			return decimal.Zero
		}
		return decimal.NewFromFloat(n)
	case json.Number:
		d, _ := decimal.NewFromString(n.String())
		return d
	default:
		d, _ := decimal.NewFromString(strings.TrimSpace(fmt.Sprint(v)))
		return d
	}
}
func optionalFinanceAmount(v any) (decimal.Decimal, bool) {
	if v == nil {
		return decimal.Zero, false
	}
	return financeAmount(v), true
}
func financeMoneyAmount(row map[string]any, amountKey, currencyKey string) (decimal.Decimal, string) {
	return financeAmount(row[amountKey]), strings.ToUpper(financeValueString(row, currencyKey))
}
func financeValueString(row map[string]any, key string) string {
	if row == nil || row[key] == nil {
		return ""
	}
	if n, ok := row[key].(float64); ok {
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	return strings.TrimSpace(fmt.Sprint(row[key]))
}
func financeFirstValue(row map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := financeValueString(row, key); value != "" {
			return value
		}
	}
	return ""
}
func sanitizeFinanceText(value string) string {
	return financeLongNumberPattern.ReplaceAllString(value, "[redacted]")
}
func sanitizeFinanceRow(row map[string]any) map[string]any {
	clean := make(map[string]any, len(row))
	for key, value := range row {
		if text, ok := value.(string); ok {
			clean[key] = sanitizeFinanceText(text)
		} else {
			clean[key] = value
		}
	}
	return clean
}
func financeFirstString(row map[string]any, keys ...string) string {
	for _, key := range keys {
		v := row[key]
		if v != nil && strings.TrimSpace(fmt.Sprint(v)) != "" {
			return strings.TrimSpace(fmt.Sprint(v))
		}
	}
	return ""
}
