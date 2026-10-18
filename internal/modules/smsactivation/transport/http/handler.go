package http

import (
	"errors"
	"strings"

	"github.com/dujiao-next/internal/modules/smsactivation/application"
	"github.com/dujiao-next/internal/platform/http/ginutil"
	"github.com/dujiao-next/internal/platform/http/response"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	service *application.Service
}

func respondActivationError(c *gin.Context, err error) {
	key := "error.sms_activation_failed"
	switch {
	case errors.Is(err, application.ErrOrderNotFound), errors.Is(err, application.ErrUnauthorized):
		key = "error.guest_order_not_found"
	case errors.Is(err, application.ErrOrderNotPaid):
		key = "error.sms_order_not_paid"
	case errors.Is(err, application.ErrAlreadyCompleted):
		key = "error.sms_activation_already_completed"
	case errors.Is(err, application.ErrServiceDisabled):
		key = "error.sms_activation_disabled"
	}
	ginutil.RespondError(c, response.CodeBadRequest, key, err)
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

// UserAcquireNumber 前台登录用户获取号码
func (h *Handler) UserAcquireNumber(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok || uid == 0 {
		ginutil.RespondError(c, response.CodeUnauthorized, "error.unauthorized", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	activation, err := h.service.AcquireNumber(c.Request.Context(), orderNo, uid, "")
	if err != nil {
		respondActivationError(c, err)
		return
	}
	response.Success(c, activation)
}

// UserCheckStatus 前台登录用户查码状态
func (h *Handler) UserCheckStatus(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok || uid == 0 {
		ginutil.RespondError(c, response.CodeUnauthorized, "error.unauthorized", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	activation, err := h.service.CheckStatus(c.Request.Context(), orderNo, uid, "")
	if err != nil {
		respondActivationError(c, err)
		return
	}
	response.Success(c, activation)
}

// UserRetryNumber 前台登录用户更换号码
func (h *Handler) UserRetryNumber(c *gin.Context) {
	uid, ok := ginutil.GetUserID(c)
	if !ok || uid == 0 {
		ginutil.RespondError(c, response.CodeUnauthorized, "error.unauthorized", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	activation, err := h.service.RetryNumber(c.Request.Context(), orderNo, uid, "")
	if err != nil {
		respondActivationError(c, err)
		return
	}
	response.Success(c, activation)
}

// GuestAcquireNumber 游客获取号码
func (h *Handler) GuestAcquireNumber(c *gin.Context) {
	email, _, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	activation, err := h.service.AcquireNumber(c.Request.Context(), orderNo, 0, email)
	if err != nil {
		respondActivationError(c, err)
		return
	}
	response.Success(c, activation)
}

// GuestCheckStatus 游客查码状态
func (h *Handler) GuestCheckStatus(c *gin.Context) {
	email, _, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	activation, err := h.service.CheckStatus(c.Request.Context(), orderNo, 0, email)
	if err != nil {
		respondActivationError(c, err)
		return
	}
	response.Success(c, activation)
}

// GuestRetryNumber 游客更换号码
func (h *Handler) GuestRetryNumber(c *gin.Context) {
	email, _, ok := ginutil.GetGuestCredentials(c)
	if !ok || email == "" {
		ginutil.RespondError(c, response.CodeBadRequest, "error.guest_email_required", nil)
		return
	}
	orderNo := strings.TrimSpace(c.Param("order_no"))
	activation, err := h.service.RetryNumber(c.Request.Context(), orderNo, 0, email)
	if err != nil {
		respondActivationError(c, err)
		return
	}
	response.Success(c, activation)
}

// RegisterRoutes 注册路由
func RegisterRoutes(user gin.IRoutes, guest gin.IRoutes, h *Handler) {
	if user != nil {
		user.POST("/orders/:order_no/sms/acquire", h.UserAcquireNumber)
		user.GET("/orders/:order_no/sms/status", h.UserCheckStatus)
		user.POST("/orders/:order_no/sms/retry", h.UserRetryNumber)
	}
	if guest != nil {
		guest.POST("/orders/:order_no/sms/acquire", h.GuestAcquireNumber)
		guest.GET("/orders/:order_no/sms/status", h.GuestCheckStatus)
		guest.POST("/orders/:order_no/sms/retry", h.GuestRetryNumber)
	}
}
