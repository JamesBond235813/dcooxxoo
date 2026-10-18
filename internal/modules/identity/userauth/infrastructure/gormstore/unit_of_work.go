package gormstore

import (
	"context"
	"errors"
	"strings"
	"time"

	externalidentitydomain "github.com/dujiao-next/internal/modules/identity/externalidentity/domain"
	userdomain "github.com/dujiao-next/internal/modules/identity/user/domain"
	userauthapp "github.com/dujiao-next/internal/modules/identity/userauth/application"

	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UnitOfWork implements the cross-aggregate authentication transaction port.
type UnitOfWork struct {
	db *gorm.DB
}

// New constructs an authentication unit of work.
func New(db *gorm.DB) *UnitOfWork {
	return &UnitOfWork{db: db}
}

// WithinTransaction runs all authentication persistence operations on one
// connection and transaction.
func (u *UnitOfWork) WithinTransaction(ctx context.Context, fn func(userauthapp.AuthTransaction) error) error {
	if u == nil || u.db == nil {
		return errors.New("user auth unit of work is not configured")
	}
	if fn == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}
	const maxAttempts = 3
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		err = u.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
			return fn(&transaction{db: tx})
		})
		if err == nil || !isRetryableTransactionError(err) || attempt == maxAttempts-1 {
			return err
		}
		timer := time.NewTimer(time.Duration(attempt+1) * 5 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return err
}

func isRetryableTransactionError(err error) bool {
	if err == nil {
		return false
	}
	var postgresErr *pgconn.PgError
	if errors.As(err, &postgresErr) {
		return postgresErr.Code == "40001" || postgresErr.Code == "40P01"
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "database is locked") ||
		strings.Contains(message, "database table is locked") ||
		strings.Contains(message, "database is deadlocked")
}

type transaction struct {
	db *gorm.DB
}

func (t *transaction) GetUserByEmail(email string) (*userdomain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, nil
	}
	var user userdomain.User
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("email = ? AND deleted_at IS NULL", email).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (t *transaction) GetUserByIDForUpdate(userID uint) (*userdomain.User, error) {
	if userID == 0 {
		return nil, nil
	}
	var user userdomain.User
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND deleted_at IS NULL", userID).
		First(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (t *transaction) CreateUser(user *userdomain.User) error {
	if user == nil {
		return nil
	}
	return t.db.Create(user).Error
}

func (t *transaction) UpdateUser(user *userdomain.User) error {
	if user == nil || user.ID == 0 {
		return nil
	}
	return t.db.Model(&userdomain.User{}).
		Where("id = ? AND deleted_at IS NULL", user.ID).
		Select("*").
		Omit("id", "deleted_at").
		Updates(user).Error
}

func (t *transaction) ConsumeAgentInviteCode(code string, now time.Time) error {
	code = strings.ToUpper(strings.TrimSpace(code))
	if code == "" {
		return userauthapp.ErrAgentInviteCodeNotFound
	}
	result := t.db.Table("agent_invite_codes").
		Where("code = ? AND status = ? AND deleted_at IS NULL AND (expires_at IS NULL OR expires_at > ?) AND (max_uses = 0 OR used_count < max_uses)", code, "active", now).
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
	var row struct {
		Status    string
		MaxUses   int
		UsedCount int
		ExpiresAt *time.Time
	}
	if err := t.db.Table("agent_invite_codes").Select("status, max_uses, used_count, expires_at").Where("code = ? AND deleted_at IS NULL", code).First(&row).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return userauthapp.ErrAgentInviteCodeNotFound
		}
		return err
	}
	if row.ExpiresAt != nil && !row.ExpiresAt.After(now) {
		return userauthapp.ErrAgentInviteCodeExpired
	}
	if row.MaxUses > 0 && row.UsedCount >= row.MaxUses {
		return userauthapp.ErrAgentInviteCodeExhausted
	}
	return userauthapp.ErrAgentInviteCodeInvalid
}

func (t *transaction) GetIdentityByProviderUserID(provider, providerUserID string) (*externalidentitydomain.Identity, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	providerUserID = strings.TrimSpace(providerUserID)
	if provider == "" || providerUserID == "" {
		return nil, nil
	}
	var identity externalidentitydomain.Identity
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("provider = ? AND provider_user_id = ?", provider, providerUserID).
		First(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (t *transaction) GetIdentityByUserProvider(userID uint, provider string) (*externalidentitydomain.Identity, error) {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if userID == 0 || provider == "" {
		return nil, nil
	}
	var identity externalidentitydomain.Identity
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ? AND provider = ?", userID, provider).
		Order("id ASC").
		First(&identity).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &identity, nil
}

func (t *transaction) ListIdentitiesByUserID(userID uint) ([]externalidentitydomain.Identity, error) {
	if userID == 0 {
		return []externalidentitydomain.Identity{}, nil
	}
	var identities []externalidentitydomain.Identity
	err := t.db.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("user_id = ?", userID).
		Order("id ASC").
		Find(&identities).Error
	return identities, err
}

func (t *transaction) CreateIdentity(identity *externalidentitydomain.Identity) error {
	if identity == nil {
		return nil
	}
	return t.db.Create(identity).Error
}

func (t *transaction) UpdateIdentity(identity *externalidentitydomain.Identity) error {
	if identity == nil || identity.ID == 0 {
		return nil
	}
	return t.db.Save(identity).Error
}

func (t *transaction) DeleteIdentityByID(identityID uint) error {
	if identityID == 0 {
		return nil
	}
	return t.db.Delete(&externalidentitydomain.Identity{}, identityID).Error
}

var _ userauthapp.AuthUnitOfWork = (*UnitOfWork)(nil)
var _ userauthapp.AuthTransaction = (*transaction)(nil)
