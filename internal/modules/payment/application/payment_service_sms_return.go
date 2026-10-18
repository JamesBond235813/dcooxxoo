package application

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"time"

	"github.com/dujiao-next/internal/constants"
)

// ResolveSMSReturn 用短期随机凭据恢复支付宝回跳时丢失的游客订单邮箱。
// 未支付时仅返回 ready=false，不透露订单信息。
func (s *PaymentService) ResolveSMSReturn(paymentID uint, ticket string) (orderNo, email string, ready bool, err error) {
	if len(ticket) != 64 || paymentID == 0 {
		return "", "", false, ErrPaymentNotFound
	}
	if _, err := hex.DecodeString(ticket); err != nil {
		return "", "", false, ErrPaymentNotFound
	}
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil || payment == nil || payment.ProviderPayload == nil {
		return "", "", false, ErrPaymentNotFound
	}
	storedHash, _ := payment.ProviderPayload["sms_return_ticket_hash"].(string)
	expiryText, _ := payment.ProviderPayload["sms_return_ticket_expires"].(string)
	expiry, parseErr := time.Parse(time.RFC3339, expiryText)
	if parseErr != nil || !time.Now().Before(expiry) {
		return "", "", false, ErrPaymentNotFound
	}
	hash := sha256.Sum256([]byte(ticket))
	stored, decodeErr := hex.DecodeString(storedHash)
	if decodeErr != nil || len(stored) != len(hash) || subtle.ConstantTimeCompare(hash[:], stored) != 1 {
		return "", "", false, ErrPaymentNotFound
	}
	if payment.Status != constants.PaymentStatusSuccess {
		return "", "", false, nil
	}
	order, err := s.orderRepo.GetByID(payment.OrderID)
	if err != nil || order == nil || order.GuestEmail == "" {
		return "", "", false, ErrPaymentNotFound
	}
	return order.OrderNo, order.GuestEmail, true, nil
}
