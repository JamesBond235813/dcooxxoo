package domain

import "time"

// ManualCryptoTxClaim makes an on-chain transaction hash globally single-use.
// Claims live as long as the payment history, including rejected proofs.
type ManualCryptoTxClaim struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TxHash    string    `gorm:"size:66;not null;uniqueIndex" json:"tx_hash"`
	PaymentID uint      `gorm:"not null;index" json:"payment_id"`
	CreatedAt time.Time `json:"created_at"`
}

func (ManualCryptoTxClaim) TableName() string { return "manual_crypto_tx_claims" }
