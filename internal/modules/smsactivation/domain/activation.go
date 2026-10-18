package domain

import (
	"time"
)

const (
	StatusPending  = "pending"  // 等待收码
	StatusReceived = "received" // 已收到短信验证码
	StatusFinished = "finished" // 已结单
	StatusCanceled = "canceled" // 已取消（换号/退款）
	StatusTimeout  = "timeout"  // 超时未收码
)

// SmsActivation 短信接码记录
type SmsActivation struct {
	ID              uint       `gorm:"primaryKey" json:"id"`
	OrderID         uint       `gorm:"index;not null" json:"order_id"`
	OrderNo         string     `gorm:"index;not null;size:64" json:"order_no"`
	UserID          uint       `gorm:"index;default:0" json:"user_id"`
	Provider        string     `gorm:"size:32;not null;default:'5sim'" json:"-"`
	ProviderOrderID int64      `gorm:"index" json:"provider_order_id"`
	Phone           string     `gorm:"size:32" json:"phone"`
	Country         string     `gorm:"size:32" json:"country"`
	Operator        string     `gorm:"size:32" json:"operator"`
	Product         string     `gorm:"size:32;default:'openai'" json:"product"`
	Price           float64    `gorm:"type:decimal(10,4);default:0" json:"price"`
	Status          string     `gorm:"size:32;not null;default:'pending'" json:"status"`
	Code            string     `gorm:"size:32" json:"code"`
	FullSMS         string     `gorm:"type:text" json:"full_sms"`
	RetryCount      int        `gorm:"default:0" json:"retry_count"`
	ExpiresAt       *time.Time `json:"expires_at"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

func (SmsActivation) TableName() string {
	return "sms_activations"
}

type Repository interface {
	Create(activation *SmsActivation) error
	Update(activation *SmsActivation) error
	GetLatestByOrderNo(orderNo string) (*SmsActivation, error)
	ListByOrderNo(orderNo string) ([]SmsActivation, error)
}
