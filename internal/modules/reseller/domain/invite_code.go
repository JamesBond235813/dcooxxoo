package domain

import (
	"time"
)

// AgentInviteCode 代理邀请码。
type AgentInviteCode struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	Code      string     `gorm:"type:varchar(64);uniqueIndex;not null" json:"code"`
	Memo      string     `gorm:"type:varchar(255)" json:"memo"`
	MaxUses   int        `gorm:"not null" json:"max_uses"` // 0 为无限制，1 为单次使用
	UsedCount int        `gorm:"not null;default:0" json:"used_count"`
	Status    string     `gorm:"type:varchar(32);index;not null;default:'active'" json:"status"` // active, disabled, exhausted
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	CreatedAt time.Time  `gorm:"index" json:"created_at"`
	UpdatedAt time.Time  `gorm:"index" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" json:"-"`
}

func (AgentInviteCode) TableName() string { return "agent_invite_codes" }
