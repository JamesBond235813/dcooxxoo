package application

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	"github.com/dujiao-next/internal/logger"
	"github.com/dujiao-next/internal/modules/smsactivation/domain"
)

var (
	ErrOrderNotFound    = errors.New("order not found")
	ErrOrderNotPaid     = errors.New("order not paid")
	ErrNotSmsOrder      = errors.New("not an sms activation order")
	ErrUnauthorized     = errors.New("unauthorized order access")
	ErrAlreadyCompleted = errors.New("sms activation already completed")
	ErrServiceDisabled  = errors.New("sms activation service is disabled")
)

type Config struct {
	Enabled   bool
	Countries []string
	Product   string
	MaxPrice  float64
}

type Service struct {
	orders    domain.OrderReader
	repo      domain.Repository
	provider  domain.Provider
	deliverer domain.Deliverer
	config    Config
}

func NewService(orders domain.OrderReader, repo domain.Repository, provider domain.Provider, deliverer domain.Deliverer, cfg Config) *Service {
	if len(cfg.Countries) == 0 {
		cfg.Countries = []string{"any"}
	}
	if cfg.Product == "" {
		cfg.Product = "openai"
	}
	return &Service{
		orders:    orders,
		repo:      repo,
		provider:  provider,
		deliverer: deliverer,
		config:    cfg,
	}
}

// verifyAndGetOrder 校验并获取订单
func (s *Service) verifyAndGetOrder(orderNo string, userID uint, guestEmail string) (*domain.Order, error) {
	order, err := s.orders.GetSmsOrderByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}

	// 权限检查：若已登录则匹配 UserID；若是游客则匹配 GuestEmail
	if userID > 0 {
		if order.UserID != userID {
			return nil, ErrUnauthorized
		}
	} else if guestEmail != "" {
		if !strings.EqualFold(strings.TrimSpace(order.GuestEmail), strings.TrimSpace(guestEmail)) {
			return nil, ErrUnauthorized
		}
	}

	// 状态检查：必须是已付款状态
	validStatuses := map[string]bool{
		constants.OrderStatusPaid:               true,
		constants.OrderStatusFulfilling:         true,
		constants.OrderStatusPartiallyDelivered: true,
		constants.OrderStatusDelivered:          true,
		constants.OrderStatusCompleted:          true,
	}
	if !validStatuses[order.Status] {
		return nil, ErrOrderNotPaid
	}

	// 商品检查：必须包含接码商品
	if !s.isSmsOrder(order) {
		return nil, ErrNotSmsOrder
	}

	return order, nil
}

func (s *Service) isSmsOrder(order *domain.Order) bool {
	for _, item := range order.Items {
		// 1. 检查 schema 中是否有 sms_activation 标识
		if schemaType, ok := item.Schema["type"].(string); ok && schemaType == "sms_activation" {
			return true
		}
		if service, ok := item.Schema["service"].(string); ok && service == "openai" {
			return true
		}
		// 2. 检查标题是否包含接码/收码
		for _, v := range item.TitleJSON {
			if str, ok := v.(string); ok {
				if strings.Contains(str, "接码") || strings.Contains(str, "收码") || strings.Contains(strings.ToLower(str), "sms") {
					return true
				}
			}
		}
	}
	return false
}

// AcquireNumber 获取（或返回已有有效）手机号
func (s *Service) AcquireNumber(ctx context.Context, orderNo string, userID uint, guestEmail string) (*domain.SmsActivation, error) {
	if !s.config.Enabled {
		return nil, ErrServiceDisabled
	}
	order, err := s.verifyAndGetOrder(orderNo, userID, guestEmail)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetLatestByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}

	// 如果已有记录且已收到验证码，直接返回
	if existing != nil && (existing.Status == domain.StatusReceived || existing.Status == domain.StatusFinished) {
		return existing, nil
	}

	// 如果已有待接收的号码且尚未过期，直接复用当前号码，避免重复花钱
	if existing != nil && existing.Status == domain.StatusPending && existing.ExpiresAt != nil && existing.ExpiresAt.After(time.Now()) {
		return existing, nil
	}

	// 调用 5sim 自动从备选池申请号码
	fivesimOrder, err := s.provider.AutoBuyFromPool(ctx, s.config.Countries, s.config.Product, s.config.MaxPrice, nil)
	if err != nil {
		return nil, fmt.Errorf("5sim acquire number failed: %w", err)
	}

	var expiresAt *time.Time
	if fivesimOrder.Expires != "" {
		if t, err := time.Parse(time.RFC3339Nano, fivesimOrder.Expires); err == nil {
			expiresAt = &t
		} else if t, err := time.Parse(time.RFC3339, fivesimOrder.Expires); err == nil {
			expiresAt = &t
		}
	}
	if expiresAt == nil {
		defaultExp := time.Now().Add(15 * time.Minute)
		expiresAt = &defaultExp
	}

	retryCount := 0
	if existing != nil {
		retryCount = existing.RetryCount
	}

	activation := &domain.SmsActivation{
		OrderID:         order.ID,
		OrderNo:         orderNo,
		UserID:          order.UserID,
		Provider:        "5sim",
		ProviderOrderID: fivesimOrder.ID,
		Phone:           fivesimOrder.Phone,
		Country:         fivesimOrder.Country,
		Operator:        fivesimOrder.Operator,
		Product:         fivesimOrder.Product,
		Price:           fivesimOrder.Price,
		Status:          domain.StatusPending,
		RetryCount:      retryCount,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Create(activation); err != nil {
		return nil, err
	}
	return activation, nil
}

// CheckStatus 轮询或检查当前接码状态
func (s *Service) CheckStatus(ctx context.Context, orderNo string, userID uint, guestEmail string) (*domain.SmsActivation, error) {
	order, err := s.verifyAndGetOrder(orderNo, userID, guestEmail)
	if err != nil {
		return nil, err
	}

	activation, err := s.repo.GetLatestByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	if activation == nil {
		return nil, nil
	}

	// 如果已经是完成或结束态，直接返回
	if activation.Status == domain.StatusReceived || activation.Status == domain.StatusFinished {
		return activation, nil
	}
	if activation.Status == domain.StatusCanceled || activation.Status == domain.StatusTimeout {
		return activation, nil
	}

	// 检查是否已超时
	if activation.ExpiresAt != nil && activation.ExpiresAt.Before(time.Now()) {
		activation.Status = domain.StatusTimeout
		activation.UpdatedAt = time.Now()
		_ = s.repo.Update(activation)
		return activation, nil
	}

	// 调用 5sim 查询订单
	orderInfo, err := s.provider.CheckOrder(ctx, activation.ProviderOrderID)
	if err != nil {
		logger.Warnw("5sim_check_order_error", "order_id", activation.ProviderOrderID, "error", err)
		return activation, nil
	}

	switch strings.ToUpper(orderInfo.Status) {
	case "RECEIVED":
		if len(orderInfo.SMS) > 0 {
			latestSMS := orderInfo.SMS[len(orderInfo.SMS)-1]
			activation.Code = latestSMS.Code
			activation.FullSMS = latestSMS.Text
			activation.Status = domain.StatusReceived
			activation.UpdatedAt = time.Now()

			// 向 5sim 提交 finish 结单
			if finishErr := s.provider.FinishOrder(ctx, activation.ProviderOrderID); finishErr != nil {
				logger.Warnw("5sim_finish_order_warning", "order_id", activation.ProviderOrderID, "error", finishErr)
			} else {
				activation.Status = domain.StatusFinished
			}

			_ = s.repo.Update(activation)

			// 履约交付订单
			s.fulfillOrder(order, activation)
		}
	case "CANCELED":
		activation.Status = domain.StatusCanceled
		activation.UpdatedAt = time.Now()
		_ = s.repo.Update(activation)
	case "TIMEOUT":
		activation.Status = domain.StatusTimeout
		activation.UpdatedAt = time.Now()
		_ = s.repo.Update(activation)
	}

	return activation, nil
}

// RetryNumber 更换新号码（取消旧号退款并申请新号）
func (s *Service) RetryNumber(ctx context.Context, orderNo string, userID uint, guestEmail string) (*domain.SmsActivation, error) {
	if !s.config.Enabled {
		return nil, ErrServiceDisabled
	}
	order, err := s.verifyAndGetOrder(orderNo, userID, guestEmail)
	if err != nil {
		return nil, err
	}

	existing, err := s.repo.GetLatestByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}

	// 如果已经成功收码并结单，不允许再换号
	if existing != nil && (existing.Status == domain.StatusReceived || existing.Status == domain.StatusFinished) {
		return nil, ErrAlreadyCompleted
	}

	// 如果当前号处于 pending 状态，先调用 5sim cancel 释放退款
	if existing != nil && existing.Status == domain.StatusPending {
		if cancelErr := s.provider.CancelOrder(ctx, existing.ProviderOrderID); cancelErr != nil {
			return nil, fmt.Errorf("cancel previous number before retry: %w", cancelErr)
		}
		existing.Status = domain.StatusCanceled
		existing.UpdatedAt = time.Now()
		_ = s.repo.Update(existing)
	}

	// 重新申请新号码
	priorFailures := map[string]int{}
	previous, err := s.repo.ListByOrderNo(orderNo)
	if err != nil {
		return nil, err
	}
	for _, attempt := range previous {
		if attempt.Status != domain.StatusReceived && attempt.Status != domain.StatusFinished {
			priorFailures[attempt.Country+"/"+attempt.Operator]++
		}
	}
	fivesimOrder, err := s.provider.AutoBuyFromPool(ctx, s.config.Countries, s.config.Product, s.config.MaxPrice, priorFailures)
	if err != nil {
		return nil, fmt.Errorf("5sim acquire retry number failed: %w", err)
	}

	var expiresAt *time.Time
	if fivesimOrder.Expires != "" {
		if t, err := time.Parse(time.RFC3339Nano, fivesimOrder.Expires); err == nil {
			expiresAt = &t
		} else if t, err := time.Parse(time.RFC3339, fivesimOrder.Expires); err == nil {
			expiresAt = &t
		}
	}
	if expiresAt == nil {
		defaultExp := time.Now().Add(15 * time.Minute)
		expiresAt = &defaultExp
	}

	retryCount := 1
	if existing != nil {
		retryCount = existing.RetryCount + 1
	}

	newActivation := &domain.SmsActivation{
		OrderID:         order.ID,
		OrderNo:         orderNo,
		UserID:          order.UserID,
		Provider:        "5sim",
		ProviderOrderID: fivesimOrder.ID,
		Phone:           fivesimOrder.Phone,
		Country:         fivesimOrder.Country,
		Operator:        fivesimOrder.Operator,
		Product:         fivesimOrder.Product,
		Price:           fivesimOrder.Price,
		Status:          domain.StatusPending,
		RetryCount:      retryCount,
		ExpiresAt:       expiresAt,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.repo.Create(newActivation); err != nil {
		return nil, err
	}
	return newActivation, nil
}

// fulfillOrder 成功收码后完成订单履约交付
func (s *Service) fulfillOrder(order *domain.Order, activation *domain.SmsActivation) {
	now := time.Now()
	payload := fmt.Sprintf("手机号码: %s\n国家区号: %s\n验证码: %s\n短信内容: %s\n接收时间: %s",
		activation.Phone,
		activation.Country,
		activation.Code,
		activation.FullSMS,
		now.Format("2006-01-02 15:04:05"),
	)

	if s.deliverer == nil {
		return
	}
	if err := s.deliverer.DeliverSmsOrder(order.ID, payload, now); err != nil {
		logger.Warnw("sms_activation_fulfillment_failed", "order_id", order.ID, "error", err)
	}
}
