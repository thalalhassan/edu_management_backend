package auth

import (
	"context"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/apperrors"
	"github.com/thalalhassan/edu_management/internal/database"
	"gorm.io/gorm"
)

type Repository interface {
	GetUserByEmail(ctx context.Context, email string) (*database.User, error)

	// Refresh tokens
	SaveRefreshToken(ctx context.Context, token *database.UserRefreshToken) error
	GetRefreshToken(ctx context.Context, rawToken string) (*database.UserRefreshToken, error)
	RevokeRefreshToken(ctx context.Context, rawToken string) error
	RevokeAllRefreshTokensForUser(ctx context.Context, userID uuid.UUID) error
}

type repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repository{db: db}
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (*database.User, error) {
	var user database.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	// che

	return &user, nil
}

// ──────────────────────────────────────────────────────────────
// Refresh tokens
// ──────────────────────────────────────────────────────────────

func (r *repository) SaveRefreshToken(ctx context.Context, token *database.UserRefreshToken) error {
	return r.db.WithContext(ctx).Create(token).Error
}

func (r *repository) GetRefreshToken(ctx context.Context, rawToken string) (*database.UserRefreshToken, error) {
	var t database.UserRefreshToken
	err := r.db.WithContext(ctx).
		Where("token = ? AND revoked = false", rawToken).
		First(&t).Error
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *repository) RevokeRefreshToken(ctx context.Context, rawToken string) error {
	return r.db.WithContext(ctx).
		Model(&database.UserRefreshToken{}).
		Where("token = ?", rawToken).
		Update("revoked", true).Error
}

func (r *repository) RevokeAllRefreshTokensForUser(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Model(&database.UserRefreshToken{}).
		Where("user_id = ?", userID).
		Update("revoked", true).Error
}
