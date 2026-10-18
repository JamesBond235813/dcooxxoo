package orderhttp

import (
	"errors"
	"strings"

	orderdomain "github.com/dujiao-next/internal/modules/order/domain"

	orderpresenter "github.com/dujiao-next/internal/modules/order/transport/presenter"
	reseller "github.com/dujiao-next/internal/modules/reseller/contract"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"

	"github.com/gin-gonic/gin"
)

// GuestOrderQuery 游客订单只读端口。
type GuestOrderQuery interface {
	ListOrdersByGuestForTenant(tenant reseller.TenantContext, email, password string, page, pageSize int) ([]orderdomain.Order, int64, error)
	GetOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*orderdomain.Order, error)
	GetAnyOrderByGuestOrderNoForTenant(tenant reseller.TenantContext, orderNo, email, password string) (*orderdomain.Order, error)
	SubmitSessionForGuest(tenant reseller.TenantContext, orderNo, email, password, session string) (*orderdomain.Order, error)
	PreflightSessionForGuest(tenant reseller.TenantContext, orderNo, email, password, session string) (map[string]interface{}, error)
}

func (h *GuestHandler) SubmitSession(c *gin.Context) {
	email, password, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	var req struct {
		Session string `json:"session"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.session_submission_failed", err)
		return
	}
	order, err := h.orders.SubmitSessionForGuest(tenantFromRequest(c), strings.TrimSpace(c.Param("order_no")), email, password, req.Session)
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.session_submission_failed", err)
		return
	}
	response.Success(c, orderpresenter.NewOrderDetailTruncated(order))
}

func (h *GuestHandler) PreflightSession(c *gin.Context) {
	email, password, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	var req struct {
		Session string `json:"session"`
	}
	if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Session) == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.session_submission_failed", err)
		return
	}
	result, err := h.orders.PreflightSessionForGuest(tenantFromRequest(c), strings.TrimSpace(c.Param("order_no")), email, password, req.Session)
	if err != nil {
		ginutil.RespondError(c, response.CodeBadRequest, "error.session_submission_failed", err)
		return
	}
	response.Success(c, result)
}

// GuestHandler 处理前台游客订单只读 HTTP。
type GuestHandler struct {
	orders   GuestOrderQuery
	payments PaymentChannelPolicy
	refunds  RefundRecordDirectory
	progress OrderProgressReader
}

func NewGuestHandler(orders GuestOrderQuery, payments PaymentChannelPolicy, refunds RefundRecordDirectory, progress ...OrderProgressReader) *GuestHandler {
	if orders == nil {
		panic("order guest handler: orders is nil")
	}
	h := &GuestHandler{orders: orders, payments: payments, refunds: refunds}
	if len(progress) > 0 {
		h.progress = progress[0]
	}
	return h
}

// ListGuestOrders 获取游客订单列表
func (h *GuestHandler) ListGuestOrders(c *gin.Context) {
	email, password, ok := ginutil.GetGuestCredentials(c)
	orderNo := strings.TrimSpace(c.Query("order_no"))
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	if orderNo != "" {
		order, err := h.orders.GetOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
		if err != nil {
			if errors.Is(err, ErrGuestOrderNotFound) {
				pagination := response.Pagination{
					Page:      1,
					PageSize:  1,
					Total:     0,
					TotalPage: 1,
				}
				response.SuccessWithPage(c, []orderdomain.Order{}, pagination)
				return
			}
			ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
			return
		}
		pagination := response.Pagination{
			Page:      1,
			PageSize:  1,
			Total:     1,
			TotalPage: 1,
		}
		response.SuccessWithPage(c, orderpresenter.NewOrderSummaryList([]orderdomain.Order{*order}), pagination)
		return
	}

	page, pageSize := ginutil.ParsePagination(c)

	orders, total, err := h.orders.ListOrdersByGuestForTenant(tenantFromRequest(c), email, password, page, pageSize)
	if err != nil {
		ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	pagination := response.BuildPagination(page, pageSize, total)
	response.SuccessWithPage(c, orderpresenter.NewOrderSummaryList(orders), pagination)
}

// GetGuestOrderByOrderNo 按订单号获取游客订单详情
func (h *GuestHandler) GetGuestOrderByOrderNo(c *gin.Context) {
	email, password, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return
	}
	order, err := h.orders.GetOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
	if err != nil {
		if errors.Is(err, ErrGuestOrderNotFound) {
			ginutil.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
			return
		}
		ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	orderDetail := orderpresenter.NewOrderDetailTruncated(order)
	enrichOrderWithAllowedChannels(h.payments, order, &orderDetail)
	enrichOrderWithRefundRecords(h.refunds, order, &orderDetail)
	if c.Query("progress") == "1" {
		enrichOrderWithProgress(h.progress, &orderDetail)
	}
	response.Success(c, orderDetail)
}

// DownloadGuestFulfillment 下载订单交付内容（游客）
// 支持父订单或子订单的 order_no
func (h *GuestHandler) DownloadGuestFulfillment(c *gin.Context) {
	email, password, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	if orderNo == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.order_item_invalid", nil)
		return
	}
	order, err := h.orders.GetAnyOrderByGuestOrderNoForTenant(tenantFromRequest(c), orderNo, email, password)
	if err != nil {
		if errors.Is(err, ErrGuestOrderNotFound) {
			ginutil.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
			return
		}
		ginutil.RespondError(c, response.CodeInternal, "error.order_fetch_failed", err)
		return
	}
	if order == nil {
		ginutil.RespondError(c, response.CodeNotFound, "error.guest_order_not_found", nil)
		return
	}
	respondFulfillmentDownload(c, order)
}
