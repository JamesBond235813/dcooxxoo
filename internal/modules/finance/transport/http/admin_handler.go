package financehttp

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	financedomain "github.com/dujiao-next/internal/modules/finance/domain"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	procurementdomain "github.com/dujiao-next/internal/modules/procurement/domain"
	"github.com/dujiao-next/internal/modules/siteconnection/contract"
	"github.com/dujiao-next/internal/platform/database/gormdb"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/dujiao-next/internal/upstream"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

const financeLocation = "Africa/Accra"

var balanceFeePattern = regexp.MustCompile(`(?i)(?:fee|手续费|手續費)\s*[:=]?\s*\$?\s*([0-9]+(?:\.[0-9]+)?)`)

type AdminHandler struct{ connections contract.Repository }

func NewAdminHandler(connections contract.Repository) *AdminHandler {
	return &AdminHandler{connections: connections}
}

type receiptInput struct {
	EntryDate    string          `json:"entry_date" binding:"required"`
	EntryType    string          `json:"entry_type" binding:"required"`
	Amount       decimal.Decimal `json:"amount" binding:"required"`
	Currency     string          `json:"currency"`
	Counterparty string          `json:"counterparty"`
	Reference    string          `json:"reference"`
	OrderNo      string          `json:"order_no"`
	Note         string          `json:"note"`
}
type deliveryInput struct {
	DeliveryDate string `json:"delivery_date" binding:"required"`
	Plan         string `json:"plan" binding:"required"`
	Recipient    string `json:"recipient" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required"`
	Note         string `json:"note"`
}

func (h *AdminHandler) ListReceipts(c *gin.Context) {
	date := strings.TrimSpace(c.Query("date"))
	q := gormdb.DB.Model(&financedomain.ManualReceipt{}).Where("deleted_at IS NULL")
	if date != "" {
		q = q.Where("entry_date = ?", date)
	}
	var rows []financedomain.ManualReceipt
	if err := q.Order("entry_date DESC, id DESC").Limit(500).Find(&rows).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, gin.H{"items": rows})
}
func (h *AdminHandler) CreateReceipt(c *gin.Context) { h.saveReceipt(c, 0) }
func (h *AdminHandler) UpdateReceipt(c *gin.Context) {
	id, ok := parseID(c)
	if ok {
		h.saveReceipt(c, id)
	}
}
func (h *AdminHandler) saveReceipt(c *gin.Context, id uint) {
	var in receiptInput
	if err := c.ShouldBindJSON(&in); err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	if !validDate(in.EntryDate) || !validReceiptType(in.EntryType) || !in.Amount.GreaterThan(decimal.Zero) {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	currency := strings.ToUpper(strings.TrimSpace(in.Currency))
	if currency == "" {
		currency = "CNY"
	}
	row := financedomain.ManualReceipt{EntryDate: in.EntryDate, EntryType: in.EntryType, Amount: in.Amount.Round(2), Currency: currency, Counterparty: strings.TrimSpace(in.Counterparty), Reference: strings.TrimSpace(in.Reference), OrderNo: strings.TrimSpace(in.OrderNo), Note: strings.TrimSpace(in.Note), CreatedBy: c.GetUint("admin_id")}
	if id > 0 {
		row.ID = id
		if err := gormdb.DB.Model(&financedomain.ManualReceipt{}).Where("id = ? AND deleted_at IS NULL", id).Updates(map[string]any{"entry_date": row.EntryDate, "entry_type": row.EntryType, "amount": row.Amount, "currency": row.Currency, "counterparty": row.Counterparty, "reference": row.Reference, "order_no": row.OrderNo, "note": row.Note, "updated_at": time.Now()}).Error; err != nil {
			ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
			return
		}
	} else if err := gormdb.DB.Create(&row).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, row)
}
func (h *AdminHandler) DeleteReceipt(c *gin.Context) { h.softDelete(c, &financedomain.ManualReceipt{}) }
func (h *AdminHandler) ListDeliveries(c *gin.Context) {
	q := gormdb.DB.Model(&financedomain.CDKDelivery{}).Where("deleted_at IS NULL")
	if date := strings.TrimSpace(c.Query("date")); date != "" {
		q = q.Where("delivery_date = ?", date)
	}
	var rows []financedomain.CDKDelivery
	if err := q.Order("delivery_date DESC, id DESC").Limit(500).Find(&rows).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, gin.H{"items": rows})
}
func (h *AdminHandler) CreateDelivery(c *gin.Context) {
	adminID, ok := ginutil.GetAdminID(c)
	if !ok {
		return
	}
	var in deliveryInput
	if err := c.ShouldBindJSON(&in); err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", err)
		return
	}
	in.Plan = strings.TrimSpace(in.Plan)
	in.Recipient = strings.TrimSpace(in.Recipient)
	if !validDate(in.DeliveryDate) || in.Plan == "" || in.Recipient == "" || in.Quantity < 1 || in.Quantity > 200 {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	snapshot, connectionID, err := h.fetchUpstreamWithID(c.Request.Context())
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.dashboard_fetch_failed", err)
		return
	}
	var existing []financedomain.CDKDelivery
	if err := gormdb.DB.Where("deleted_at IS NULL").Find(&existing).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	used := map[string]bool{}
	for _, row := range existing {
		var ids []string
		_ = json.Unmarshal([]byte(row.UpstreamCDKIDs), &ids)
		for _, id := range ids {
			used[id] = true
		}
	}
	var direct []procurementdomain.Order
	if err := gormdb.DB.Where("deleted_at IS NULL AND connection_id = ? AND upstream_order_id > 0", connectionID).Find(&direct).Error; err == nil {
		for _, row := range direct {
			used[strconv.FormatUint(uint64(row.UpstreamOrderID), 10)] = true
		}
	}
	candidates := make([]map[string]any, 0)
	for _, item := range snapshot.CDKs {
		id := idString(item["id"])
		if id != "" && !used[id] && stringValue(item["plan"]) == in.Plan && stringValue(item["status"]) == "unused" {
			candidates = append(candidates, item)
		}
	}
	sort.Slice(candidates, func(i, j int) bool {
		return timeValue(candidates[i]["created_at"]).Before(timeValue(candidates[j]["created_at"]))
	})
	if len(candidates) < in.Quantity {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", fmt.Errorf("该套餐本站可分配留存 CDK 只有 %d 张", len(candidates)))
		return
	}
	ids := make([]string, 0, in.Quantity)
	for _, item := range candidates[:in.Quantity] {
		ids = append(ids, idString(item["id"]))
	}
	encoded, _ := json.Marshal(ids)
	row := financedomain.CDKDelivery{DeliveryDate: in.DeliveryDate, Plan: in.Plan, Recipient: in.Recipient, Quantity: in.Quantity, UpstreamCDKIDs: string(encoded), Note: strings.TrimSpace(in.Note), CreatedBy: adminID}
	if err := gormdb.DB.Create(&row).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, row)
}
func (h *AdminHandler) DeleteDelivery(c *gin.Context) { h.softDelete(c, &financedomain.CDKDelivery{}) }
func (h *AdminHandler) softDelete(c *gin.Context, model any) {
	id, ok := parseID(c)
	if !ok {
		return
	}
	if err := gormdb.DB.Model(model).Where("id = ? AND deleted_at IS NULL", id).Update("deleted_at", time.Now()).Error; err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	response.Success(c, gin.H{"deleted": true})
}
func (h *AdminHandler) Report(c *gin.Context) {
	date := strings.TrimSpace(c.Query("date"))
	if date == "" {
		date = time.Now().In(financeLoc()).Format("2006-01-02")
	}
	if !validDate(date) {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return
	}
	nowDate := time.Now().In(financeLoc()).Format("2006-01-02")
	manualDaily, manualTotal, err := manualReceiptTotals(date)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	manualIncomeDay, manualIncomeTotal, err := queryManualTotals(date, true)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	localProfitDay, localProfitTotal, err := localProfitTotals(date)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	resellerDay, resellerTotal, err := resellerProfitTotals(date)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.internal", err)
		return
	}
	remote, connID, remoteErr := h.fetchUpstreamWithID(c.Request.Context())
	cdk := cdkReport(remote, date, connID)
	profitDay, profitTotal := map[string]float64{}, map[string]float64{}
	exchangeRate := 1.0
	if connID > 0 && h.connections != nil {
		if conn, getErr := h.connections.GetByID(connID); getErr == nil && conn != nil && conn.ExchangeRate.GreaterThan(decimal.Zero) {
			exchangeRate = conn.ExchangeRate.InexactFloat64()
		}
	}
	for k, v := range localProfitDay {
		addConverted(profitDay, k, v, exchangeRate)
	}
	for k, v := range localProfitTotal {
		addConverted(profitTotal, k, v, exchangeRate)
	}
	for k, v := range manualIncomeDay {
		parts := strings.SplitN(k, ":", 2)
		if len(parts) == 2 && parts[0] != "upstream_transfer" {
			addConverted(profitDay, parts[1], v, exchangeRate)
		}
	}
	for k, v := range manualIncomeTotal {
		parts := strings.SplitN(k, ":", 2)
		if len(parts) == 2 && parts[0] != "upstream_transfer" {
			addConverted(profitTotal, parts[1], v, exchangeRate)
		}
	}
	if remoteErr == nil {
		for k, v := range cdk["costs_day"].(map[string]float64) {
			addConverted(profitDay, k, -v, exchangeRate)
		}
		for k, v := range cdk["costs_total"].(map[string]float64) {
			addConverted(profitTotal, k, -v, exchangeRate)
		}
	}
	for k, v := range resellerDay {
		addConverted(profitDay, k, -v, exchangeRate)
	}
	for k, v := range resellerTotal {
		addConverted(profitTotal, k, -v, exchangeRate)
	}
	payload := gin.H{"date": date, "timezone": financeLocation, "today": date == nowDate, "exchange_rate_usd_to_cny": exchangeRate, "manual_receipts": gin.H{"day": manualDaily, "cumulative": manualTotal}, "manual_income_excluding_linked_orders": gin.H{"day": manualIncomeDay, "cumulative": manualIncomeTotal}, "local_order_margin": gin.H{"day": localProfitDay, "cumulative": localProfitTotal}, "reseller_earnings": gin.H{"day": resellerDay, "cumulative": resellerTotal}, "cdk": cdk, "platform_net_profit": gin.H{"day": profitDay, "cumulative": profitTotal}, "upstream_connection_id": connID, "upstream_synced_at": time.Now().Format(time.RFC3339), "upstream_error": ""}
	if remoteErr != nil {
		payload["upstream_error"] = remoteErr.Error()
		payload["upstream_synced_at"] = ""
	}
	response.Success(c, payload)
}

func (h *AdminHandler) fetchUpstream(ctx context.Context) (*upstream.ZovoFinanceSnapshot, error) {
	snap, _, err := h.fetchUpstreamWithID(ctx)
	return snap, err
}
func (h *AdminHandler) fetchUpstreamWithID(ctx context.Context) (*upstream.ZovoFinanceSnapshot, uint, error) {
	if h == nil || h.connections == nil {
		return nil, 0, fmt.Errorf("未配置 ZovoCard 连接")
	}
	rows, _, err := h.connections.List(contract.ListFilter{Status: "active", Page: 1, PageSize: 100})
	if err != nil {
		return nil, 0, err
	}
	for _, conn := range rows {
		if conn.Protocol == constants.ConnectionProtocolZovoCard {
			adapter := upstream.NewZovoCardAdapter(&conn, "")
			snap, err := adapter.GetFinanceSnapshot(ctx)
			return snap, conn.ID, err
		}
	}
	return nil, 0, fmt.Errorf("没有启用的 ZovoCard 上游连接")
}
func financeLoc() *time.Location {
	loc, err := time.LoadLocation(financeLocation)
	if err != nil {
		return time.UTC
	}
	return loc
}
func validDate(v string) bool { _, err := time.Parse("2006-01-02", v); return err == nil }
func validReceiptType(v string) bool {
	switch v {
	case "upstream_transfer", "wechat", "alipay", "bank":
		return true
	}
	return false
}
func addConverted(target map[string]float64, currency string, amount, usdToCNY float64) {
	currency = strings.ToUpper(strings.TrimSpace(currency))
	if currency == "USD" {
		target["CNY"] += amount * usdToCNY
		return
	}
	if currency == "" {
		currency = "CNY"
	}
	target[currency] += amount
}
func parseID(c *gin.Context) (uint, bool) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil || id == 0 {
		ginutil.RespondError(c, response.CodeBadRequest, "error.bad_request", nil)
		return 0, false
	}
	return uint(id), true
}
func manualReceiptTotals(date string) (map[string]float64, map[string]float64, error) {
	return queryManualTotals(date, false)
}
func queryManualTotals(date string, excludeLinked bool) (map[string]float64, map[string]float64, error) {
	day := map[string]float64{}
	all := map[string]float64{}
	orderFilter := ""
	if excludeLinked {
		orderFilter = " AND (order_no = '' OR order_no IS NULL)"
	}
	var rows []struct {
		EntryType string
		Currency  string
		Total     decimal.Decimal
	}
	if err := gormdb.DB.Model(&financedomain.ManualReceipt{}).Select("entry_type, currency, SUM(amount) total").Where("deleted_at IS NULL AND entry_date <= ?"+orderFilter, date).Group("entry_type,currency").Scan(&rows).Error; err != nil {
		return day, all, err
	}
	for _, r := range rows {
		all[r.EntryType+":"+r.Currency] = r.Total.InexactFloat64()
	}
	rows = nil
	if err := gormdb.DB.Model(&financedomain.ManualReceipt{}).Select("entry_type, currency, SUM(amount) total").Where("deleted_at IS NULL AND entry_date = ?"+orderFilter, date).Group("entry_type,currency").Scan(&rows).Error; err != nil {
		return day, all, err
	}
	for _, r := range rows {
		day[r.EntryType+":"+r.Currency] = r.Total.InexactFloat64()
	}
	return day, all, nil
}
func localProfitTotals(date string) (map[string]float64, map[string]float64, error) {
	day := map[string]float64{}
	all := map[string]float64{}
	if err := queryLocalProfit(date, day, all); err != nil {
		return day, all, err
	}
	return day, all, queryPaymentFees(date, day, all)
}
func queryPaymentFees(date string, day, all map[string]float64) error {
	loc := financeLoc()
	start, _ := time.ParseInLocation("2006-01-02", date, loc)
	end := start.AddDate(0, 0, 1)
	query := func(from, to *time.Time, dest map[string]float64) error {
		var rows []struct {
			Currency string
			Amount   decimal.Decimal
		}
		q := gormdb.DB.Model(&paymentdomain.Payment{}).Select("currency,COALESCE(SUM(fee_amount),0) amount").Where("deleted_at IS NULL AND status = ? AND provider_type <> ? AND fee_policy = ?", constants.PaymentStatusSuccess, constants.PaymentProviderWallet, constants.PaymentFeePolicyMerchantAbsorbed).Group("currency")
		if from != nil {
			q = q.Where("created_at >= ? AND created_at < ?", *from, *to)
		} else {
			q = q.Where("created_at < ?", *to)
		}
		if err := q.Scan(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			dest[r.Currency] -= r.Amount.InexactFloat64()
		}
		return nil
	}
	if err := query(&start, &end, day); err != nil {
		return err
	}
	return query(nil, &end, all)
}
func queryLocalProfit(date string, day, all map[string]float64) error {
	loc := financeLoc()
	start, _ := time.ParseInLocation("2006-01-02", date, loc)
	end := start.AddDate(0, 0, 1)
	query := func(from, to *time.Time, dest map[string]float64) error {
		var rows []struct {
			Currency string
			Revenue  decimal.Decimal
			Cost     decimal.Decimal
		}
		q := gormdb.DB.Table("order_items oi").Select("orders.currency, COALESCE(SUM(oi.total_price-oi.coupon_discount),0) revenue, COALESCE(SUM(CASE WHEN sc.protocol = ? THEN 0 ELSE oi.cost_price*oi.quantity END),0) cost", constants.ConnectionProtocolZovoCard).Joins("JOIN orders ON orders.id=oi.order_id").Joins("LEFT JOIN product_mappings pm ON pm.local_product_id=oi.product_id AND pm.deleted_at IS NULL").Joins("LEFT JOIN site_connections sc ON sc.id=pm.connection_id").Where("oi.deleted_at IS NULL AND orders.deleted_at IS NULL AND orders.status IN ?", []string{constants.OrderStatusPaid, constants.OrderStatusFulfilling, constants.OrderStatusPartiallyDelivered, constants.OrderStatusPartiallyRefunded, constants.OrderStatusDelivered, constants.OrderStatusCompleted}).Group("orders.currency")
		if from != nil {
			q = q.Where("orders.created_at >= ? AND orders.created_at < ?", *from, *to)
		} else {
			q = q.Where("orders.created_at < ?", *to)
		}
		if err := q.Scan(&rows).Error; err != nil {
			return err
		}
		for _, r := range rows {
			dest[r.Currency] = r.Revenue.Sub(r.Cost).InexactFloat64()
		}
		return nil
	}
	if err := query(&start, &end, day); err != nil {
		return err
	}
	return query(nil, &end, all)
}
func resellerProfitTotals(date string) (map[string]float64, map[string]float64, error) {
	day := map[string]float64{}
	all := map[string]float64{}
	loc := financeLoc()
	start, _ := time.ParseInLocation("2006-01-02", date, loc)
	end := start.AddDate(0, 0, 1)
	for _, target := range []struct {
		dest     map[string]float64
		from, to *time.Time
	}{{day, &start, &end}, {all, nil, &end}} {
		q := gormdb.DB.Table("reseller_ledger_entries").Select("currency, COALESCE(SUM(amount),0) amount").Where("deleted_at IS NULL AND status <> ? AND type IN ?", "canceled", []string{"order_profit", "refund_deduct"})
		if target.from != nil {
			q = q.Where("created_at >= ? AND created_at < ?", *target.from, *target.to)
		} else {
			q = q.Where("created_at < ?", *target.to)
		}
		var rows []struct {
			Currency string
			Amount   decimal.Decimal
		}
		if err := q.Group("currency").Scan(&rows).Error; err != nil {
			return day, all, err
		}
		for _, r := range rows {
			target.dest[r.Currency] = r.Amount.InexactFloat64()
		}
	}
	return day, all, nil
}
func cdkReport(s *upstream.ZovoFinanceSnapshot, date string, connectionID uint) gin.H {
	result := gin.H{"plans": []any{}, "daily": gin.H{}, "cumulative": gin.H{}, "deliveries": []any{}, "outstanding_deliveries": []any{}, "upstream_balance": gin.H{}, "expenses": gin.H{}, "costs_day": map[string]float64{}, "costs_total": map[string]float64{}, "sync_available": s != nil}
	if s == nil {
		return result
	}
	result["upstream_balance"] = s.Balance
	loc := financeLoc()
	start, _ := time.ParseInLocation("2006-01-02", date, loc)
	end := start.AddDate(0, 0, 1)
	byPlan := map[string]map[string]float64{}
	ensure := func(plan string) map[string]float64 {
		if byPlan[plan] == nil {
			byPlan[plan] = map[string]float64{"purchased_day": 0, "purchased_total": 0, "delivered_day": 0, "delivered_total": 0, "redeemed_day": 0, "redeemed_old_day": 0, "redeemed_same_day": 0, "new_outstanding_day": 0, "retained": 0, "outstanding": 0}
		}
		return byPlan[plan]
	}
	idPlan := map[string]string{}
	idAssigned := map[string]string{}
	unused := map[string]bool{}
	// 自动订阅履约对应的 CDK 采购与核销是成对业务流水，不属于这里要
	// 追踪的“手动建码库存”。依据本站采购订单的上游 ID 将其从手工库存账中排除。
	var proc []procurementdomain.Order
	_ = gormdb.DB.Where("deleted_at IS NULL AND connection_id = ? AND upstream_order_id > 0", connectionID).Find(&proc).Error
	automaticCDKIDs := map[string]bool{}
	for _, o := range proc {
		automaticCDKIDs[strconv.FormatUint(uint64(o.UpstreamOrderID), 10)] = true
	}
	for _, x := range s.CDKs {
		id := idString(x["id"])
		p := stringValue(x["plan"])
		status := stringValue(x["status"])
		created := timeValue(x["created_at"])
		idPlan[id] = p
		if automaticCDKIDs[id] {
			continue
		}
		if status == "unused" {
			unused[id] = true
		}
		if p != "" {
			v := ensure(p)
			if dateOf(created, loc) == date {
				v["purchased_day"]++
			}
			if !created.IsZero() && created.Before(end) {
				v["purchased_total"]++
			}
		}
	}
	var deliveries []financedomain.CDKDelivery
	_ = gormdb.DB.Where("deleted_at IS NULL").Find(&deliveries).Error
	for _, d := range deliveries {
		v := ensure(d.Plan)
		if d.DeliveryDate == date {
			v["delivered_day"] += float64(d.Quantity)
		}
		if d.DeliveryDate <= date {
			v["delivered_total"] += float64(d.Quantity)
		}
		var ids []string
		_ = json.Unmarshal([]byte(d.UpstreamCDKIDs), &ids)
		for _, id := range ids {
			if d.DeliveryDate <= date {
				idAssigned[id] = d.DeliveryDate
			}
			if d.DeliveryDate <= time.Now().In(loc).Format("2006-01-02") {
				delete(unused, id)
			}
		}
	}
	for id := range unused {
		if p := idPlan[id]; p != "" {
			ensure(p)["retained"]++
		}
	}
	redeemedToday := map[string]bool{}
	redeemedAt := map[string]time.Time{}
	for _, o := range s.CDKOrders {
		status := strings.ToLower(stringValue(o["status"]))
		if status != "completed" && status != "consumed" {
			continue
		}
		id := idString(o["cdk_id"])
		if automaticCDKIDs[id] || id == "" {
			continue
		}
		p := stringValue(o["plan"])
		if p == "" {
			p = idPlan[id]
		}
		if p == "" {
			continue
		}
		t := timeValue(firstNonNil(o["completed_at"], o["updated_at"], o["created_at"]))
		if !t.IsZero() {
			redeemedAt[id] = t
		}
		v := ensure(p)
		if dateOf(t, loc) == date {
			v["redeemed_day"]++
			redeemedToday[id] = true
			if assigned, ok := idAssigned[id]; ok {
				if assigned < date {
					v["redeemed_old_day"]++
				}
				if assigned == date {
					v["redeemed_same_day"]++
				}
			}
		}
		if _, ok := idAssigned[id]; ok && !t.IsZero() && t.Before(end) {
			v["outstanding"]--
		}
	}
	for _, v := range byPlan {
		v["outstanding"] += v["delivered_total"]
		v["new_outstanding_day"] = v["delivered_day"] - v["redeemed_same_day"]
		if v["new_outstanding_day"] < 0 {
			v["new_outstanding_day"] = 0
		}
	}
	plans := make([]string, 0, len(byPlan))
	for p := range byPlan {
		plans = append(plans, p)
	}
	sort.Strings(plans)
	rows := make([]any, 0, len(plans))
	for _, p := range plans {
		rows = append(rows, gin.H{"plan": p, "label": planLabel(p), "purchased_day": byPlan[p]["purchased_day"], "purchased_total": byPlan[p]["purchased_total"], "delivered_day": byPlan[p]["delivered_day"], "delivered_total": byPlan[p]["delivered_total"], "redeemed_day": byPlan[p]["redeemed_day"], "redeemed_old_day": byPlan[p]["redeemed_old_day"], "new_outstanding_day": byPlan[p]["new_outstanding_day"], "retained": byPlan[p]["retained"], "outstanding": byPlan[p]["outstanding"]})
	}
	result["plans"] = rows
	outstandingRows := make([]any, 0)
	for _, d := range deliveries {
		if d.DeliveryDate > date {
			continue
		}
		var ids []string
		_ = json.Unmarshal([]byte(d.UpstreamCDKIDs), &ids)
		redeemedCount := 0
		for _, id := range ids {
			if at, ok := redeemedAt[id]; ok && at.Before(end) {
				redeemedCount++
			}
		}
		outstandingCount := d.Quantity - redeemedCount
		if outstandingCount > 0 {
			outstandingRows = append(outstandingRows, gin.H{"delivery_date": d.DeliveryDate, "plan": d.Plan, "recipient": d.Recipient, "quantity": d.Quantity, "redeemed_quantity": redeemedCount, "outstanding_quantity": outstandingCount, "note": d.Note})
		}
	}
	result["outstanding_deliveries"] = outstandingRows
	var dailyDeliveries []financedomain.CDKDelivery
	_ = gormdb.DB.Where("deleted_at IS NULL AND delivery_date = ?", date).Order("id DESC").Find(&dailyDeliveries).Error
	result["deliveries"] = dailyDeliveries
	costDay, costTotal := map[string]float64{}, map[string]float64{}
	for _, cdk := range s.CDKs {
		created := timeValue(cdk["created_at"])
		fee := numberValue(cdk["fee_amount_minor"]) / 100
		currency := "USD"
		if fee != 0 {
			if !created.IsZero() && created.Before(end) {
				costTotal[currency] += fee
			}
			if dateOf(created, loc) == date {
				costDay[currency] += fee
			}
		}
	}
	for _, tx := range s.Transactions {
		if strings.ToUpper(stringValue(tx["status"])) != "COMPLETE" {
			continue
		}
		amount := numberValue(tx["settle_amount"])
		if amount == 0 {
			amount = numberValue(tx["auth_amount"])
		}
		t := timeValue(tx["auth_time"])
		if !t.IsZero() && t.Before(end) {
			costTotal["USD"] += amount
		}
		if dateOf(t, loc) == date {
			costDay["USD"] += amount
		}
	}
	fees := map[string]float64{}
	creditedDay, creditedTotal := 0.0, 0.0
	for _, log := range s.BalanceLogs {
		typ := stringValue(log["type"])
		amount := numberValue(log["amount"])
		t := timeValue(log["created_at"])
		if typ == "recharge" && amount > 0 {
			if !t.IsZero() && t.Before(end) {
				creditedTotal += amount
			}
			if dateOf(t, loc) == date {
				creditedDay += amount
			}
		}
		if typ == "card_recharge" {
			if match := balanceFeePattern.FindStringSubmatch(stringValue(log["remark"])); len(match) == 2 {
				fee, _ := strconv.ParseFloat(match[1], 64)
				fees["card_recharge_fee"] += fee
				if !t.IsZero() && t.Before(end) {
					costTotal["USD"] += fee
				}
				if dateOf(t, loc) == date {
					costDay["USD"] += fee
				}
			}
		}
		if !isFeeType(typ) {
			continue
		}
		if amount < 0 {
			amount = -amount
		}
		if dateOf(t, loc) == date {
			fees[typ] += amount
		}
		if !t.IsZero() && t.Before(end) {
			costTotal["USD"] += amount
		}
		if dateOf(t, loc) == date {
			costDay["USD"] += amount
		}
	}
	for _, card := range s.Cards {
		if numberValue(card["open_fee"]) <= 0 {
			continue
		}
		t := timeValue(card["created_at"])
		if !t.IsZero() && t.Before(end) {
			costTotal["USD"] += numberValue(card["open_fee"])
		}
		if dateOf(t, loc) == date {
			costDay["USD"] += numberValue(card["open_fee"])
		}
	}
	result["daily"] = gin.H{"redeemed_today": len(redeemedToday)}
	result["recharge_credited_day"] = creditedDay
	result["recharge_credited_total"] = creditedTotal
	result["expenses"] = fees
	result["costs_day"] = costDay
	result["costs_total"] = costTotal
	return result
}
func isFeeType(t string) bool {
	switch t {
	case "decline_fee", "reversal_fee", "refund_fee", "small_tx_fee", "merchant_violation_fee":
		return true
	}
	return false
}
func planLabel(p string) string {
	switch p {
	case "plus":
		return "ChatGPT Plus"
	case "pro_5x":
		return "ChatGPT Pro 5x"
	case "pro_20x":
		return "ChatGPT Pro 20x"
	case "pro_50x":
		return "ChatGPT Pro 50x"
	case "go":
		return "ChatGPT Go"
	}
	return p
}
func idString(v any) string {
	switch x := v.(type) {
	case string:
		return strings.TrimSpace(x)
	case float64:
		return strconv.FormatInt(int64(x), 10)
	case json.Number:
		return x.String()
	}
	return ""
}
func stringValue(v any) string {
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return ""
}
func numberValue(v any) float64 {
	switch x := v.(type) {
	case float64:
		return x
	case int:
		return float64(x)
	case int64:
		return float64(x)
	case string:
		f, _ := strconv.ParseFloat(strings.TrimSpace(x), 64)
		return f
	}
	return 0
}
func timeValue(v any) time.Time {
	s := stringValue(v)
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339Nano, "2006-01-02 15:04:05", "2006-01-02T15:04:05"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}
func dateOf(t time.Time, loc *time.Location) string {
	if t.IsZero() {
		return ""
	}
	return t.In(loc).Format("2006-01-02")
}
func firstNonNil(v ...any) any {
	for _, x := range v {
		if x != nil && x != "" {
			return x
		}
	}
	return nil
}
