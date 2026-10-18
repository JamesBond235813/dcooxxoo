package gormstore

import (
	cryptoRand "crypto/rand"
	"errors"
	"math/big"
	"strings"
	"time"

	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"
	resellercontract "github.com/dujiao-next/internal/modules/reseller/contract"
	resellerdomain "github.com/dujiao-next/internal/modules/reseller/domain"

	"gorm.io/gorm"
)

func generateRandomAgentCode() string {
	const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789" // 避免容易混淆的 0, O, 1, I
	result := make([]byte, 6)
	charsetLen := big.NewInt(int64(len(charset)))
	for i := range result {
		n, err := cryptoRand.Int(cryptoRand.Reader, charsetLen)
		if err != nil {
			result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
		} else {
			result[i] = charset[n.Int64()]
		}
	}
	return string(result)
}

func (r *Store) isAgentMemberLevelID(levelID uint) (bool, error) {
	if levelID == 0 {
		return false, nil
	}
	var count int64
	err := r.db.Table("member_levels").Where("id = ? AND slug = ? AND is_active = ? AND deleted_at IS NULL", levelID, "agent", true).Count(&count).Error
	return count > 0, err
}

// CreateProfile 创建分销商资料。
func (r *Store) CreateProfile(profile *resellerdomain.Profile) error {
	if profile == nil {
		return errors.New("reseller profile is nil")
	}
	return r.db.Create(profile).Error
}

// GetProfileByID 按 ID 获取分销商资料。
func (r *Store) GetProfileByID(id uint) (*resellerdomain.Profile, error) {
	if id == 0 {
		return nil, nil
	}
	var profile resellerdomain.Profile
	if err := r.db.Preload("User", "deleted_at IS NULL").
		Where("reseller_profiles.id = ? AND reseller_profiles.deleted_at IS NULL", id).
		First(&profile).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &profile, nil
}

// GetProfileByUserID 按用户 ID 获取分销商资料。
func (r *Store) GetProfileByUserID(userID uint) (*resellerdomain.Profile, error) {
	if userID == 0 {
		return nil, nil
	}
	var profile resellerdomain.Profile
	err := r.db.Preload("User", "deleted_at IS NULL").
		Where("user_id = ? AND reseller_profiles.deleted_at IS NULL", userID).
		First(&profile).Error
	if err == nil {
		isAgent := false
		if profile.User != nil {
			isAgent, err = r.isAgentMemberLevelID(profile.User.MemberLevelID)
			if err != nil {
				return nil, err
			}
		}
		if isAgent {
			needsSave := false
			if profile.Status != resellerdomain.ProfileStatusActive {
				profile.Status = resellerdomain.ProfileStatusActive
				needsSave = true
			}
			if profile.AgentCode == "" || strings.HasPrefix(profile.AgentCode, "AGENT") {
				profile.AgentCode = generateRandomAgentCode()
				needsSave = true
			}
			if needsSave {
				_ = r.db.Save(&profile)
			}
		}
		return &profile, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	var user userdomain.User
	if uErr := r.db.Where("id = ? AND deleted_at IS NULL", userID).First(&user).Error; uErr == nil {
		isAgent, levelErr := r.isAgentMemberLevelID(user.MemberLevelID)
		if levelErr != nil {
			return nil, levelErr
		}
		if !isAgent {
			return nil, nil
		}
		now := time.Now()
		newProfile := resellerdomain.Profile{
			UserID:           user.ID,
			Status:           resellerdomain.ProfileStatusActive,
			AgentCode:        generateRandomAgentCode(),
			SettlementStatus: resellerdomain.SettlementStatusNormal,
			CreatedAt:        now,
			UpdatedAt:        now,
		}
		if cErr := r.db.Create(&newProfile).Error; cErr == nil {
			newProfile.User = &user
			return &newProfile, nil
		} else if !errors.Is(cErr, gorm.ErrDuplicatedKey) {
			return nil, cErr
		}
	} else if !errors.Is(uErr, gorm.ErrRecordNotFound) {
		return nil, uErr
	}

	return nil, nil
}

// UpdateProfile 更新分销商资料。
func (r *Store) UpdateProfile(profile *resellerdomain.Profile) error {
	if profile == nil || profile.ID == 0 {
		return errors.New("invalid reseller profile")
	}
	return r.db.Model(&resellerdomain.Profile{}).
		Where("id = ? AND deleted_at IS NULL", profile.ID).
		Select("*").
		Updates(profile).Error
}

// ListProfiles 分页列出分销商资料。
func (r *Store) ListProfiles(filter resellercontract.ProfileListFilter) ([]resellerdomain.Profile, int64, error) {
	rows := make([]resellerdomain.Profile, 0)
	query := r.db.Model(&resellerdomain.Profile{}).
		Preload("User", "deleted_at IS NULL").
		Where("reseller_profiles.deleted_at IS NULL")
	if filter.UserID > 0 {
		query = query.Where("reseller_profiles.user_id = ?", filter.UserID)
	}
	if status := strings.TrimSpace(filter.Status); status != "" {
		query = query.Where("reseller_profiles.status = ?", status)
	}
	if settlement := strings.TrimSpace(filter.SettlementStatus); settlement != "" {
		query = query.Where("reseller_profiles.settlement_status = ?", settlement)
	}
	if keyword := strings.TrimSpace(filter.Keyword); keyword != "" {
		like := "%" + strings.ToLower(keyword) + "%"
		query = query.Joins("LEFT JOIN users ON users.id = reseller_profiles.user_id").
			Where("LOWER(users.email) LIKE ? OR LOWER(users.display_name) LIKE ? OR CAST(reseller_profiles.id AS TEXT) = ?", like, like, keyword)
	}
	if filter.CreatedFrom != nil {
		query = query.Where("reseller_profiles.created_at >= ?", *filter.CreatedFrom)
	}
	if filter.CreatedTo != nil {
		query = query.Where("reseller_profiles.created_at <= ?", *filter.CreatedTo)
	}
	var total int64
	if err := query.Session(&gorm.Session{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := applyPagination(query.Session(&gorm.Session{}), filter.Page, filter.PageSize).
		Order("reseller_profiles.id DESC").
		Find(&rows).Error; err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// IsActiveRelatedAccount 判断用户是否为分销商已启用的关联账号。
func (r *Store) IsActiveRelatedAccount(resellerID uint, userID uint) (bool, error) {
	if resellerID == 0 || userID == 0 {
		return false, nil
	}
	var count int64
	if err := r.db.Model(&resellerdomain.RelatedAccount{}).
		Where("reseller_id = ? AND user_id = ? AND status = ? AND deleted_at IS NULL", resellerID, userID, resellerdomain.RelatedAccountStatusActive).
		Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
