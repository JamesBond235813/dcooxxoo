package gormstore

import (
	"errors"
	"strings"
	"time"

	resellerdomain "github.com/dujiao-next/internal/modules/reseller/domain"
	"gorm.io/gorm"
)

var (
	ErrInviteCodeNotFound  = errors.New("invite code not found")
	ErrInviteCodeInvalid   = errors.New("invite code invalid")
	ErrInviteCodeExhausted = errors.New("invite code exhausted")
	ErrInviteCodeExpired   = errors.New("invite code expired")
)

// CreateInviteCode 创建代理邀请码。
func (s *Store) CreateInviteCode(item *resellerdomain.AgentInviteCode) error {
	if s == nil || s.db == nil || item == nil {
		return errors.New("store or item is nil")
	}
	item.Code = strings.ToUpper(strings.TrimSpace(item.Code))
	if item.Code == "" {
		return errors.New("invite code cannot be empty")
	}
	if item.Status == "" {
		item.Status = "active"
	}
	// Select all fields so an explicit MaxUses=0 (unlimited) is not replaced
	// with the schema default of 1 by GORM's zero-value default handling.
	return s.db.Select("*").Create(item).Error
}

// ListInviteCodes 分页查询邀请码列表。
func (s *Store) ListInviteCodes(page, pageSize int, status string) ([]resellerdomain.AgentInviteCode, int64, error) {
	if s == nil || s.db == nil {
		return nil, 0, errors.New("store is nil")
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	query := s.db.Model(&resellerdomain.AgentInviteCode{})
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var items []resellerdomain.AgentInviteCode
	err := query.Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&items).Error
	return items, total, err
}

// GetInviteCode 根据代码查询邀请码。
func (s *Store) GetInviteCode(code string) (*resellerdomain.AgentInviteCode, error) {
	if s == nil || s.db == nil {
		return nil, errors.New("store is nil")
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return nil, ErrInviteCodeNotFound
	}

	var item resellerdomain.AgentInviteCode
	if err := s.db.Where("code = ?", code).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrInviteCodeNotFound
		}
		return nil, err
	}
	return &item, nil
}

// VerifyInviteCode 校验邀请码是否有效（不进行核销）。
func (s *Store) VerifyInviteCode(code string) error {
	if s == nil || s.db == nil {
		return errors.New("store is nil")
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ErrInviteCodeNotFound
	}

	var item resellerdomain.AgentInviteCode
	if err := s.db.Where("code = ?", code).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInviteCodeNotFound
		}
		return err
	}

	if item.Status != "active" {
		return ErrInviteCodeInvalid
	}

	if item.ExpiresAt != nil && time.Now().After(*item.ExpiresAt) {
		return ErrInviteCodeExpired
	}

	if item.MaxUses > 0 && item.UsedCount >= item.MaxUses {
		return ErrInviteCodeExhausted
	}

	return nil
}

// UseInviteCode 验证并核销/使用邀请码一次。
func (s *Store) UseInviteCode(code string) error {
	if s == nil || s.db == nil {
		return errors.New("store is nil")
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ErrInviteCodeNotFound
	}

	now := time.Now()
	result := s.db.Model(&resellerdomain.AgentInviteCode{}).
		Where("code = ? AND status = ? AND (expires_at IS NULL OR expires_at > ?) AND (max_uses = 0 OR used_count < max_uses)", code, "active", now).
		Updates(map[string]interface{}{
			"used_count": gorm.Expr("used_count + 1"),
			"status":     gorm.Expr("CASE WHEN max_uses > 0 AND used_count + 1 >= max_uses THEN ? ELSE status END", "exhausted"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 1 {
		return nil
	}
	return s.inviteCodeStateError(code, now)
}

// ReleaseInviteCode compensates a reservation if account creation fails.
func (s *Store) ReleaseInviteCode(code string) error {
	if s == nil || s.db == nil {
		return errors.New("store is nil")
	}
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return ErrInviteCodeNotFound
	}
	result := s.db.Model(&resellerdomain.AgentInviteCode{}).
		Where("code = ? AND used_count > 0", code).
		Updates(map[string]interface{}{
			"used_count": gorm.Expr("used_count - 1"),
			"status":     gorm.Expr("CASE WHEN status = ? AND (max_uses = 0 OR used_count - 1 < max_uses) THEN ? ELSE status END", "exhausted", "active"),
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInviteCodeNotFound
	}
	return nil
}

func (s *Store) inviteCodeStateError(code string, now time.Time) error {
	var item resellerdomain.AgentInviteCode
	if err := s.db.Where("code = ?", code).First(&item).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrInviteCodeNotFound
		}
		return err
	}
	if item.Status != "active" {
		return ErrInviteCodeInvalid
	}
	if item.ExpiresAt != nil && !item.ExpiresAt.After(now) {
		return ErrInviteCodeExpired
	}
	if item.MaxUses > 0 && item.UsedCount >= item.MaxUses {
		return ErrInviteCodeExhausted
	}
	return ErrInviteCodeInvalid
}

// DeleteInviteCode 删除邀请码。
func (s *Store) DeleteInviteCode(id uint) error {
	if s == nil || s.db == nil || id == 0 {
		return errors.New("store is nil or invalid id")
	}
	return s.db.Delete(&resellerdomain.AgentInviteCode{}, id).Error
}

// UpdateInviteCodeMemo 更新邀请码备注说明。
func (s *Store) UpdateInviteCodeMemo(id uint, memo string) error {
	if s == nil || s.db == nil || id == 0 {
		return errors.New("store is nil or invalid id")
	}
	result := s.db.Model(&resellerdomain.AgentInviteCode{}).Where("id = ?", id).Update("memo", strings.TrimSpace(memo))
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		var count int64
		if err := s.db.Model(&resellerdomain.AgentInviteCode{}).Where("id = ?", id).Count(&count).Error; err != nil {
			return err
		}
		if count == 0 {
			return ErrInviteCodeNotFound
		}
	}
	return nil
}

// ToggleInviteCode 切换邀请码状态（active <-> disabled）。
func (s *Store) ToggleInviteCode(id uint) error {
	if s == nil || s.db == nil || id == 0 {
		return errors.New("store is nil or invalid id")
	}
	var item resellerdomain.AgentInviteCode
	if err := s.db.First(&item, id).Error; err != nil {
		return err
	}
	if item.Status == "active" {
		item.Status = "disabled"
	} else if item.Status == "disabled" {
		if item.MaxUses > 0 && item.UsedCount >= item.MaxUses {
			item.Status = "exhausted"
		} else {
			item.Status = "active"
		}
	}
	return s.db.Save(&item).Error
}
