package application

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dujiao-next/internal/constants"
	orderapp "github.com/dujiao-next/internal/modules/order/application"
	paymentcontract "github.com/dujiao-next/internal/modules/payment/contract"
	paymentdomain "github.com/dujiao-next/internal/modules/payment/domain"
	"github.com/dujiao-next/internal/shared/jsonmap"
)

var manualCryptoTxHashPattern = regexp.MustCompile(`^(0x)?[0-9a-fA-F]{64}$`)

func normalizeManualCryptoTxHash(hash string) string {
	return strings.ToLower(strings.TrimSpace(hash))
}

// SubmitManualCryptoProof records a customer-supplied transaction hash for later human review.
// It never marks the payment successful and therefore cannot trigger fulfillment.
func (s *PaymentService) SubmitManualCryptoProof(paymentID uint, txHash string) (*paymentdomain.Payment, error) {
	txHash = normalizeManualCryptoTxHash(txHash)
	if paymentID == 0 || !manualCryptoTxHashPattern.MatchString(txHash) {
		return nil, ErrManualCryptoProofInvalid
	}
	var updated *paymentdomain.Payment
	err := s.paymentRepo.WithinTransaction(func(tx paymentcontract.Transaction) error {
		payments := tx.Payments()
		payment, err := payments.GetByIDForUpdate(paymentID)
		if err != nil {
			return ErrPaymentUpdateFailed
		}
		if payment == nil || payment.ProviderType != constants.PaymentProviderManualCrypto || payment.OrderID == 0 {
			return ErrManualCryptoProofInvalid
		}
		if payment.Status != constants.PaymentStatusPending && payment.Status != constants.PaymentStatusInitiated {
			return ErrManualCryptoProofInvalid
		}
		order, err := tx.Orders().GetByIDForUpdateWithChildren(payment.OrderID)
		if err != nil || order == nil {
			return orderapp.ErrOrderNotFound
		}
		if order.Status != constants.OrderStatusPendingPayment || order.PaidAt != nil || (order.ExpiresAt != nil && !order.ExpiresAt.After(time.Now())) {
			return orderapp.ErrOrderStatusInvalid
		}
		payload := payment.ProviderPayload
		if payload == nil {
			payload = jsonmap.JSON{}
		}
		proofStatus := strings.TrimSpace(asPayloadString(payload["manual_proof_status"]))
		if proofStatus == "submitted" {
			if normalizeManualCryptoTxHash(asPayloadString(payload["manual_tx_hash"])) == txHash {
				updated = payment
				return nil
			}
			return ErrManualCryptoProofAlreadySubmitted
		}
		claimed, err := tx.ClaimManualCryptoTxHash(txHash, payment.ID)
		if err != nil {
			return ErrPaymentUpdateFailed
		}
		if !claimed {
			return ErrManualCryptoProofDuplicate
		}
		payload["manual_tx_hash"] = txHash
		payload["manual_proof_status"] = "submitted"
		payload["manual_proof_submitted_at"] = time.Now().UTC().Format(time.RFC3339)
		delete(payload, "manual_review_note")
		payment.ProviderPayload = payload
		payment.ProviderRef = txHash
		payment.UpdatedAt = time.Now()
		if err := payments.Update(payment); err != nil {
			return ErrPaymentUpdateFailed
		}
		updated = payment
		return nil
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// ReviewManualCryptoPayment approves (and settles) or rejects a submitted on-chain proof.
func (s *PaymentService) ReviewManualCryptoPayment(paymentID, adminID uint, approve bool, note string) (*paymentdomain.Payment, error) {
	if paymentID == 0 || adminID == 0 || len([]rune(note)) > 500 {
		return nil, ErrManualCryptoReviewInvalid
	}
	payment, err := s.paymentRepo.GetByID(paymentID)
	if err != nil {
		return nil, ErrPaymentUpdateFailed
	}
	if payment == nil || payment.ProviderType != constants.PaymentProviderManualCrypto || payment.OrderID == 0 {
		return nil, ErrManualCryptoReviewInvalid
	}
	if payment.Status == constants.PaymentStatusSuccess {
		return payment, nil
	}
	if payment.Status != constants.PaymentStatusPending || payment.ProviderPayload == nil ||
		strings.TrimSpace(asPayloadString(payment.ProviderPayload["manual_proof_status"])) != "submitted" ||
		!manualCryptoTxHashPattern.MatchString(normalizeManualCryptoTxHash(asPayloadString(payment.ProviderPayload["manual_tx_hash"]))) {
		return nil, ErrManualCryptoReviewInvalid
	}
	txHash := normalizeManualCryptoTxHash(asPayloadString(payment.ProviderPayload["manual_tx_hash"]))
	reviewedAt := time.Now().UTC().Format(time.RFC3339)
	if !approve {
		returnVal := payment
		err := s.paymentRepo.WithinTransaction(func(tx paymentcontract.Transaction) error {
			locked, err := tx.Payments().GetByIDForUpdate(paymentID)
			if err != nil || locked == nil || locked.ProviderType != constants.PaymentProviderManualCrypto || locked.Status != constants.PaymentStatusPending {
				return ErrManualCryptoReviewInvalid
			}
			payload := locked.ProviderPayload
			if payload == nil || strings.TrimSpace(asPayloadString(payload["manual_proof_status"])) != "submitted" || normalizeManualCryptoTxHash(asPayloadString(payload["manual_tx_hash"])) != txHash {
				return ErrManualCryptoReviewInvalid
			}
			payload["manual_proof_status"] = "rejected"
			payload["manual_review_note"] = strings.TrimSpace(note)
			payload["manual_reviewed_by"] = adminID
			payload["manual_reviewed_at"] = reviewedAt
			locked.ProviderPayload = payload
			locked.UpdatedAt = time.Now()
			if err := tx.Payments().Update(locked); err != nil {
				return ErrPaymentUpdateFailed
			}
			returnVal = locked
			return nil
		})
		if err != nil {
			return nil, err
		}
		return returnVal, nil
	}

	return s.HandleCallback(PaymentCallbackInput{
		PaymentID:   payment.ID,
		OrderNo:     payment.GatewayOrderNo,
		ChannelID:   payment.ChannelID,
		Status:      constants.PaymentStatusSuccess,
		ProviderRef: txHash,
		Amount:      payment.Amount,
		Currency:    payment.Currency,
		PaidAt:      ptrTime(time.Now()),
		Payload: jsonmap.JSON{
			"manual_proof_status": "approved",
			"manual_review_note":  strings.TrimSpace(note),
			"manual_reviewed_by":  adminID,
			"manual_reviewed_at":  reviewedAt,
		},
	})
}

func asPayloadString(value interface{}) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(fmt.Sprint(value))
}

func ptrTime(value time.Time) *time.Time { return &value }
