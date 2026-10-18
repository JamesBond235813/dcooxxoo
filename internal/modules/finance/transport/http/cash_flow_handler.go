package financehttp

import (
	"encoding/json"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	financedomain "github.com/dujiao-next/internal/modules/finance/domain"
	"github.com/dujiao-next/internal/modules/siteconnection/contract"
	"github.com/dujiao-next/internal/platform/database/gormdb"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type pendingSettlement struct {
	PaymentID uint            `json:"payment_id"`
	OrderID   uint            `json:"order_id"`
	OrderNo   string          `json:"order_no"`
	Channel   string          `json:"channel"`
	Amount    decimal.Decimal `json:"amount"`
	Currency  string          `json:"currency"`
	PaidAt    time.Time       `json:"paid_at"`
}

type cashFlowOrderOption struct {
	OrderNo string `json:"order_no"`
}

type cashFlowResellerOption struct {
	ID          uint   `json:"id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

// CashFlowOptions returns delivered orders not yet linked to a cash-flow entry,
// and active reseller identities for the optional cash-flow form selectors.
func (h *AdminHandler) CashFlowOptions(c *gin.Context) {
	var orders []cashFlowOrderOption
	if err := gormdb.DB.Table("orders").
		Select("orders.order_no").
		Where("orders.deleted_at IS NULL AND orders.parent_id IS NULL AND orders.order_no <> '' AND orders.status IN ?", []string{constants.OrderStatusDelivered, constants.OrderStatusCompleted}).
		Where("NOT EXISTS (SELECT 1 FROM finance_cash_flow_entries f WHERE f.deleted_at IS NULL AND f.order_no = orders.order_no)").
		Order("orders.created_at DESC").Scan(&orders).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	var resellers []cashFlowResellerOption
	if err := gormdb.DB.Table("reseller_profiles rp").
		Select("rp.id, u.email, u.display_name").
		Joins("JOIN users u ON u.id = rp.user_id AND u.deleted_at IS NULL").
		Where("rp.deleted_at IS NULL AND rp.status IN ?", []string{"active", "disabled"}).
		Order("u.display_name ASC, u.email ASC").Scan(&resellers).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, gin.H{"orders": orders, "resellers": resellers})
}

type zovoLedgerRow struct {
	ID           uint      `json:"id"`
	EventTime    time.Time `json:"event_time"`
	BalanceAfter string    `json:"balance_after,omitempty"`
	Income       string    `json:"income,omitempty"`
	Expense      string    `json:"expense,omitempty"`
	Currency     string    `json:"currency"`
	Nature       string    `json:"nature"`
}

type zovoBalanceSummary struct {
	EventTime time.Time `json:"event_time"`
	Balance   string    `json:"balance"`
	Available string    `json:"available,omitempty"`
	Currency  string    `json:"currency"`
}

// CashFlowReport separates confirmed cash movements from online payments that
// have succeeded at the gateway but have not yet been recorded as settled.
func (h *AdminHandler) CashFlowReport(c *gin.Context) {
	start, end := strings.TrimSpace(c.Query("start_date")), strings.TrimSpace(c.Query("end_date"))
	if start == "" || end == "" || !validDate(start) || !validDate(end) || start > end {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	var rows []financedomain.CashFlowEntry
	if err := gormdb.DB.Where("deleted_at IS NULL AND event_date >= ? AND event_date <= ?", start, end).Order("event_date DESC, id DESC").Limit(5000).Find(&rows).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	var zovoRows []financedomain.UpstreamFinanceRecord
	if err := gormdb.DB.Where("event_date >= ? AND event_date <= ?", start, end).Order("event_time DESC, id DESC").Limit(5000).Find(&zovoRows).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	var lastSyncText string
	if err := gormdb.DB.Model(&financedomain.UpstreamFinanceRecord{}).Select("MAX(synced_at)").Scan(&lastSyncText).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	var pending []pendingSettlement
	startAt, _ := time.ParseInLocation("2006-01-02", start, financeLoc())
	endAt, _ := time.ParseInLocation("2006-01-02", end, financeLoc())
	endExclusive := endAt.AddDate(0, 0, 1)
	err := gormdb.DB.Table("payments p").Select("p.id AS payment_id, p.order_id, o.order_no, p.channel_type AS channel, p.amount, p.currency, p.paid_at").Joins("JOIN orders o ON o.id = p.order_id").Where("p.deleted_at IS NULL AND p.status = ? AND p.provider_type <> ? AND p.paid_at >= ? AND p.paid_at < ?", constants.PaymentStatusSuccess, "wallet", startAt, endExclusive).Order("p.paid_at DESC").Limit(2000).Scan(&pending).Error
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	var settlements []financedomain.CashFlowEntry
	if err := gormdb.DB.Select("order_no, reference").Where("deleted_at IS NULL AND category = ?", "channel_settlement").Find(&settlements).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	settled := make(map[string]bool, len(settlements)*2)
	for _, row := range settlements {
		if row.OrderNo != "" {
			settled["order:"+row.OrderNo] = true
		}
		if row.Reference != "" {
			settled["payment:"+row.Reference] = true
		}
	}
	unsettled := pending[:0]
	for _, row := range pending {
		if settled["order:"+row.OrderNo] || settled["payment:"+strconv.FormatUint(uint64(row.PaymentID), 10)] {
			continue
		}
		unsettled = append(unsettled, row)
	}
	pending = unsettled
	incoming, manualOutgoing, transfer := decimal.Zero, decimal.Zero, decimal.Zero
	for _, row := range rows {
		if row.Direction == "in" {
			incoming = incoming.Add(row.CNYAmount)
		}
		if row.Direction == "out" {
			manualOutgoing = manualOutgoing.Add(row.CNYAmount)
		}
		if row.Direction == "transfer" {
			transfer = transfer.Add(row.CNYAmount)
		}
	}
	zovoOutgoing := decimal.Zero
	for _, row := range zovoRows {
		if isZovoCashOut(row) {
			zovoOutgoing = zovoOutgoing.Add(row.CNYAmount)
		}
	}
	outgoing := manualOutgoing.Add(zovoOutgoing)
	var syncedAt any
	if lastSyncText != "" {
		for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05.999999999Z07:00", "2006-01-02 15:04:05.999999999-07:00", "2006-01-02 15:04:05.999999999"} {
			if parsed, err := time.Parse(layout, lastSyncText); err == nil {
				syncedAt = parsed.UTC().Format(time.RFC3339Nano)
				break
			}
		}
		if syncedAt == nil {
			syncedAt = lastSyncText
		}
	}
	zovoLedger, zovoBalance := summarizeZovoLedger(zovoRows)
	response.Success(c, gin.H{"start_date": start, "end_date": end, "currency": "CNY", "cash_in": incoming, "cash_out": outgoing, "manual_cash_out": manualOutgoing, "zovo_cash_out": zovoOutgoing, "net_cash_flow": incoming.Sub(outgoing), "internal_transfers": transfer, "items": rows, "zovo_movements": zovoLedger, "zovo_balance": zovoBalance, "zovo_synced_at": syncedAt, "pending_settlements": pending})
}

// Only confirmed Zovo outflows reduce consolidated platform cash. Moving funds
// from the Zovo wallet onto a platform card is an asset transfer; the card
// purchase is recognized when Zovo reports a settled transaction. Card opening
// fees and Zovo balance-ledger charges are already actual debits.
func isZovoCashOut(row financedomain.UpstreamFinanceRecord) bool {
	if !row.CNYAmount.IsPositive() || row.Direction != "out" {
		return false
	}
	switch row.Source {
	case "balance_log":
		kind := strings.ToLower(strings.TrimSpace(row.EventType))
		if strings.Contains(kind, "open_card") || strings.Contains(kind, "card_recharge") || strings.Contains(kind, "transfer") {
			return false
		}
		return true
	case "card_opening":
		return strings.EqualFold(strings.TrimSpace(row.Status), "active")
	case "card_transaction":
		switch strings.ToLower(strings.TrimSpace(row.Status)) {
		case "success", "completed", "settled", "posted":
			return true
		}
	}
	// CDK issue/order rows duplicate charges already represented in the
	// authoritative Zovo account balance log.
	return false
}

func summarizeZovoLedger(records []financedomain.UpstreamFinanceRecord) ([]zovoLedgerRow, *zovoBalanceSummary) {
	var balance *zovoBalanceSummary
	ledger := make([]zovoLedgerRow, 0, len(records))
	for _, row := range records {
		if row.Source == "balance_snapshot" {
			var raw map[string]any
			if row.RawJSON != "" {
				decoder := json.NewDecoder(strings.NewReader(row.RawJSON))
				decoder.UseNumber()
				_ = decoder.Decode(&raw)
			}
			available := financeAmount(raw["spendable_balance"])
			candidate := &zovoBalanceSummary{EventTime: row.EventTime, Balance: row.Amount.StringFixed(2), Currency: row.Currency}
			if !available.IsZero() {
				candidate.Available = available.StringFixed(2)
			}
			if balance == nil || candidate.EventTime.After(balance.EventTime) {
				balance = candidate
			}
			continue
		}
		// The account statement is the authoritative movement source. CDK fees,
		// card authorizations and card-opening records overlap with these entries.
		if row.Source != "balance_log" || row.Amount.IsZero() {
			continue
		}
		var raw map[string]any
		if row.RawJSON != "" {
			decoder := json.NewDecoder(strings.NewReader(row.RawJSON))
			decoder.UseNumber()
			_ = decoder.Decode(&raw)
		}
		item := zovoLedgerRow{ID: row.ID, EventTime: row.EventTime, Currency: row.Currency, Nature: zovoLedgerNature(row.EventType, row.Category)}
		if after, ok := raw["after"]; ok {
			item.BalanceAfter = financeAmount(after).StringFixed(2)
		}
		if row.Direction == "in" {
			item.Income = row.Amount.StringFixed(2)
		}
		if row.Direction == "out" || row.Direction == "transfer" {
			item.Expense = row.Amount.StringFixed(2)
		}
		ledger = append(ledger, item)
	}
	for _, row := range records {
		if !isZovoCashOut(row) || row.Source == "balance_log" {
			continue
		}
		nature := "已结算卡片消费"
		if row.Source == "card_opening" {
			nature = "开卡费"
		}
		ledger = append(ledger, zovoLedgerRow{ID: row.ID, EventTime: row.EventTime, Expense: row.Amount.StringFixed(2), Currency: row.Currency, Nature: nature})
	}
	sort.Slice(ledger, func(i, j int) bool { return ledger[i].EventTime.After(ledger[j].EventTime) })
	return ledger, balance
}

func zovoLedgerNature(eventType, category string) string {
	kind := strings.ToLower(strings.TrimSpace(eventType))
	switch {
	case strings.Contains(kind, "card_recharge"):
		return "转入卡片"
	case strings.Contains(kind, "recharge") || strings.Contains(kind, "deposit"):
		return "上游账户充值"
	case strings.Contains(kind, "open_card"):
		return "开卡及初次充值"
	case strings.Contains(kind, "cdk"):
		return "卡密服务费"
	case strings.Contains(kind, "cslive"):
		return "附加服务费"
	case strings.Contains(kind, "service") || strings.Contains(kind, "hold"):
		return "直充服务费"
	case strings.Contains(kind, "fee") || strings.Contains(kind, "手续费") || strings.Contains(kind, "手續費"):
		return "手续费"
	case category == "zovo_internal_transfer":
		return "账户内部转账"
	case category == "zovo_fee":
		return "上游手续费"
	default:
		return "其他资金变动"
	}
}

type cashFlowInput struct {
	EventDate            string          `json:"event_date" binding:"required"`
	Direction            string          `json:"direction" binding:"required"`
	Category             string          `json:"category" binding:"required"`
	Account              string          `json:"account" binding:"required"`
	CounterAccount       string          `json:"counter_account"`
	Amount               decimal.Decimal `json:"amount" binding:"required"`
	Currency             string          `json:"currency"`
	ExchangeRateSnapshot decimal.Decimal `json:"exchange_rate_snapshot"`
	Counterparty         string          `json:"counterparty"`
	Reference            string          `json:"reference"`
	OrderNo              string          `json:"order_no"`
	RelatedSource        string          `json:"related_source"`
	RelatedSourceID      string          `json:"related_source_id"`
	Note                 string          `json:"note"`
}

func (h *AdminHandler) ListCashFlows(c *gin.Context) {
	q := gormdb.DB.Model(&financedomain.CashFlowEntry{}).Where("deleted_at IS NULL")
	start, end := strings.TrimSpace(c.Query("start_date")), strings.TrimSpace(c.Query("end_date"))
	if start != "" {
		if !validDate(start) {
			ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		q = q.Where("event_date >= ?", start)
	}
	if end != "" {
		if !validDate(end) {
			ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
			return
		}
		q = q.Where("event_date <= ?", end)
	}
	var rows []financedomain.CashFlowEntry
	if err := q.Order("event_date DESC, id DESC").Limit(1000).Find(&rows).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, gin.H{"items": rows})
}

func (h *AdminHandler) CreateCashFlow(c *gin.Context) { h.saveCashFlow(c, 0) }
func (h *AdminHandler) UpdateCashFlow(c *gin.Context) {
	id, ok := parseID(c)
	if ok {
		h.saveCashFlow(c, id)
	}
}

func (h *AdminHandler) saveCashFlow(c *gin.Context, id uint) {
	var in cashFlowInput
	if err := c.ShouldBindJSON(&in); err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	in.EventDate = strings.TrimSpace(in.EventDate)
	in.Direction = strings.TrimSpace(in.Direction)
	in.Category = strings.TrimSpace(in.Category)
	in.Account = strings.TrimSpace(in.Account)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	if in.Currency == "" {
		in.Currency = "CNY"
	}
	if !validDate(in.EventDate) || (in.Direction != "in" && in.Direction != "out" && in.Direction != "transfer") || in.Category == "" || in.Account == "" || !in.Amount.GreaterThan(decimal.Zero) || (in.Direction == "transfer" && strings.TrimSpace(in.CounterAccount) == "") {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	fx := in.ExchangeRateSnapshot
	if fx.LessThanOrEqual(decimal.Zero) {
		if in.Currency == "CNY" {
			fx = decimal.NewFromInt(1)
		} else if in.Currency == "USD" || in.Currency == "USDT" {
			fx = h.configuredUSDRate()
		} else {
			ginutil.RespondErrorWithMsg(c, response.CodeBadRequest, "不支持该币种的自动折算，请手动填写折算汇率", nil)
			return
		}
	}
	if fx.LessThanOrEqual(decimal.Zero) {
		ginutil.RespondErrorWithMsg(c, response.CodeBadRequest, "未配置 USD/CNY 汇率，请先配置汇率或手动填写折算汇率", nil)
		return
	}
	row := financedomain.CashFlowEntry{
		EventDate: in.EventDate, Direction: in.Direction, Category: in.Category, Account: in.Account,
		CounterAccount: strings.TrimSpace(in.CounterAccount), Amount: in.Amount.Round(6), Currency: in.Currency, ExchangeRateSnapshot: fx.Round(6),
		CNYAmount: in.Amount.Mul(fx).Round(6), Counterparty: strings.TrimSpace(in.Counterparty),
		Reference: strings.TrimSpace(in.Reference), OrderNo: strings.TrimSpace(in.OrderNo),
		RelatedSource: strings.TrimSpace(in.RelatedSource), RelatedSourceID: strings.TrimSpace(in.RelatedSourceID),
		Note: strings.TrimSpace(in.Note), CreatedBy: c.GetUint("admin_id"),
	}
	if id == 0 {
		if err := gormdb.DB.Create(&row).Error; err != nil {
			ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
			return
		}
	} else {
		row.ID = id
		updates := map[string]any{
			"event_date": row.EventDate, "direction": row.Direction, "category": row.Category,
			"account": row.Account, "counter_account": row.CounterAccount, "amount": row.Amount, "currency": row.Currency,
			"exchange_rate_snapshot": row.ExchangeRateSnapshot, "cny_amount": row.CNYAmount,
			"counterparty": row.Counterparty, "reference": row.Reference, "order_no": row.OrderNo,
			"related_source": row.RelatedSource, "related_source_id": row.RelatedSourceID,
			"note": row.Note, "updated_at": time.Now(),
		}
		result := gormdb.DB.Model(&financedomain.CashFlowEntry{}).Where("id = ? AND deleted_at IS NULL", id).Updates(updates)
		if result.Error != nil {
			ginutil.RespondError(c, response.CodeInternal, "error.internal", result.Error)
			return
		}
		if result.RowsAffected == 0 {
			ginutil.RespondError(c, response.CodeNotFound, "error.not_found", nil)
			return
		}
	}
	response.Success(c, row)
}

func (h *AdminHandler) configuredUSDRate() decimal.Decimal {
	if h == nil || h.connections == nil {
		return decimal.Zero
	}
	rows, _, err := h.connections.List(contract.ListFilter{Status: "active", Page: 1, PageSize: 100})
	if err != nil {
		return decimal.Zero
	}
	for _, conn := range rows {
		if conn.Protocol == constants.ConnectionProtocolZovoCard && conn.ExchangeRate.GreaterThan(decimal.Zero) {
			return conn.ExchangeRate
		}
	}
	return decimal.Zero
}

func (h *AdminHandler) DeleteCashFlow(c *gin.Context) {
	h.softDelete(c, &financedomain.CashFlowEntry{})
}
