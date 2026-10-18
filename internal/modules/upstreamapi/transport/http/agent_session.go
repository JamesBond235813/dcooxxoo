package upstreamhttp

import (
	"errors"
	"net/http"

	"github.com/dujiao-next/internal/constants"
	resellercontract "github.com/dujiao-next/internal/modules/reseller/contract"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

type agentSessionRequest struct {
	Session string `json:"session" binding:"required"`
}

// agentOrderForCredential limits Session operations to orders created with this API credential.
func (h *Handler) agentOrderForCredential(c *gin.Context, requireFullWalletPayment bool) (uint, string, bool) {
	userID := getUpstreamUserID(c)
	credentialID := getUpstreamCredentialID(c)
	if userID == 0 || credentialID == 0 {
		errorResponse(c, http.StatusUnauthorized, "unauthorized", "invalid credentials")
		return 0, "", false
	}
	orderID, err := ginutil.ParseParamUint(c, "id")
	if err != nil {
		errorResponse(c, http.StatusBadRequest, "bad_request", "invalid order id")
		return 0, "", false
	}
	if h.AgentRefs == nil {
		errorResponse(c, http.StatusServiceUnavailable, "agent_unavailable", "agent API is unavailable")
		return 0, "", false
	}
	ref, err := h.AgentRefs.GetByOrderID(orderID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get order")
		return 0, "", false
	}
	if ref == nil || ref.ApiCredentialID != credentialID {
		errorResponse(c, http.StatusNotFound, "order_not_found", "order not found")
		return 0, "", false
	}
	if h.AgentProfiles == nil {
		errorResponse(c, http.StatusServiceUnavailable, "agent_unavailable", "agent API is unavailable")
		return 0, "", false
	}
	isWholesale, err := h.AgentProfiles.IsAgentWholesaleOrder(orderID, userID)
	if err != nil {
		errorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get order")
		return 0, "", false
	}
	if !isWholesale {
		errorResponse(c, http.StatusNotFound, "order_not_found", "order not found")
		return 0, "", false
	}
	order, err := h.Orders.GetOrderByUser(orderID, userID)
	if err != nil && !errors.Is(err, ErrOrderNotFound) {
		errorResponse(c, http.StatusInternalServerError, "internal_error", "failed to get order")
		return 0, "", false
	}
	if order == nil {
		errorResponse(c, http.StatusNotFound, "order_not_found", "order not found")
		return 0, "", false
	}
	if requireFullWalletPayment && (order.PaidAt == nil ||
		(order.Status != constants.OrderStatusPaid && order.Status != constants.OrderStatusFulfilling) ||
		order.TotalAmount.Decimal.LessThanOrEqual(decimal.Zero) ||
		order.WalletPaidAmount.Decimal.LessThan(order.TotalAmount.Decimal)) {
		errorResponse(c, http.StatusConflict, "payment_not_confirmed", "full wallet payment is required before session submission")
		return 0, "", false
	}
	return userID, order.OrderNo, true
}

func (h *Handler) GetAgentOrder(c *gin.Context) {
	if _, _, ok := h.agentOrderForCredential(c, false); ok {
		h.GetOrder(c)
	}
}

func (h *Handler) PreflightAgentSession(c *gin.Context) {
	if h.SessionOrders == nil {
		errorResponse(c, http.StatusServiceUnavailable, "session_unavailable", "session API is unavailable")
		return
	}
	userID, orderNo, ok := h.agentOrderForCredential(c, true)
	if !ok {
		return
	}
	var req agentSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "bad_request", "session is required")
		return
	}
	result, err := h.SessionOrders.PreflightSessionForUser(resellercontract.MainTenantContext(c.Request.Host), orderNo, userID, req.Session)
	if err != nil {
		respondAgentSessionError(c, err, "session_preflight_failed")
		return
	}
	successResponse(c, gin.H{"ok": true, "account": result})
}

func (h *Handler) SubmitAgentSession(c *gin.Context) {
	if h.SessionOrders == nil {
		errorResponse(c, http.StatusServiceUnavailable, "session_unavailable", "session API is unavailable")
		return
	}
	userID, orderNo, ok := h.agentOrderForCredential(c, true)
	if !ok {
		return
	}
	var req agentSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errorResponse(c, http.StatusBadRequest, "bad_request", "session is required")
		return
	}
	order, err := h.SessionOrders.SubmitSessionForUser(resellercontract.MainTenantContext(c.Request.Host), orderNo, userID, req.Session)
	if err != nil {
		respondAgentSessionError(c, err, "session_submission_failed")
		return
	}
	successResponse(c, gin.H{"ok": true, "order_id": order.ID, "order_no": order.OrderNo, "status": order.Status})
}

func respondAgentSessionError(c *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, ErrSessionInvalid):
		errorResponse(c, http.StatusBadRequest, "session_invalid", "session format is invalid")
	case errors.Is(err, ErrSessionNotAllowed):
		errorResponse(c, http.StatusConflict, "session_not_allowed", "order must be paid before session submission")
	case errors.Is(err, ErrSessionNotRequired):
		errorResponse(c, http.StatusConflict, "session_not_required", "this order does not require a session")
	case errors.Is(err, ErrSessionMultipleItems):
		errorResponse(c, http.StatusConflict, "session_multiple_items", "submit session for one order item at a time")
	default:
		errorResponse(c, http.StatusUnprocessableEntity, fallback, "session could not be processed; check account access and retry")
	}
}
