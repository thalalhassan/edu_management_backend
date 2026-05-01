package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/modules/user"
	"github.com/thalalhassan/edu_management/pkg/crypto"
	"github.com/thalalhassan/edu_management/pkg/jwt"
)

type Service interface {
	Login(ctx context.Context, req LoginRequest) (*LoginResponse, error)
	RefreshToken(ctx context.Context, req RefreshRequest) (*RefreshResponse, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAllSessions(ctx context.Context, userID uuid.UUID) error
}

type service struct {
	userRepo  user.Repository
	repo      Repository
	jwtConfig *config.JWTConfig
}

func NewService(repo Repository, userRepo user.Repository, jwtConfig *config.JWTConfig) Service {
	return &service{repo: repo, userRepo: userRepo, jwtConfig: jwtConfig}
}

// ──────────────────────────────────────────────────────────────
// AUTH
// ──────────────────────────────────────────────────────────────

func (s *service) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {

	u, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("user.Service.Login.GetByEmail: invalid credentials")
	}
	if !u.IsActive {
		return nil, errors.New("user.Service.Login: account is deactivated")
	}
	if !crypto.CheckHash(req.Password, u.PasswordHash) {
		return nil, errors.New("user.Service.Login: invalid credentials")
	}

	accessToken, err := jwt.GenerateAccessToken(u.ID.String(), u.Role.Slug, string(u.Email), s.jwtConfig.Secret, s.jwtConfig.Expiration)
	if err != nil {
		return nil, fmt.Errorf("user.Service.Login.GenerateAccessToken: %w", err)
	}
	rawRefresh, expiresAt, err := jwt.GenerateRefreshToken(s.jwtConfig.RefreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("user.Service.Login.GenerateRefreshToken: %w", err)
	}

	tokenRecord := &database.UserRefreshToken{
		UserID:    u.ID,
		Token:     rawRefresh,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.SaveRefreshToken(ctx, tokenRecord); err != nil {
		return nil, fmt.Errorf("user.Service.Login.SaveRefreshToken: %w", err)
	}

	// Record last login (fire-and-forget style — don't fail login on this error)
	now := time.Now()
	u.LastLoginAt = &now
	_ = s.userRepo.UpdateUser(ctx, u.ID, u)

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		User:         UserAuthInfo{ID: u.ID, Email: u.Email, Role: u.Role.Slug},
	}, nil
}

func (s *service) RefreshToken(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	record, err := s.repo.GetRefreshToken(ctx, req.RefreshToken)
	if err != nil {
		return nil, errors.New("user.Service.RefreshToken: token not found or revoked")
	}
	if time.Now().After(record.ExpiresAt) {
		return nil, errors.New("user.Service.RefreshToken: token expired")
	}

	u, err := s.userRepo.GetUserByID(ctx, record.UserID)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RefreshToken.GetByID: %w", err)
	}

	// Rotate: revoke old, issue new
	if err := s.repo.RevokeRefreshToken(ctx, req.RefreshToken); err != nil {
		return nil, fmt.Errorf("user.Service.RefreshToken.Revoke: %w", err)
	}

	accessToken, err := jwt.GenerateAccessToken(u.ID.String(), u.Role.Slug, string(u.Email), s.jwtConfig.Secret, s.jwtConfig.Expiration)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RefreshToken.GenerateAccessToken: %w", err)
	}
	newRaw, expiresAt, err := jwt.GenerateRefreshToken(s.jwtConfig.RefreshExpiration)
	if err != nil {
		return nil, fmt.Errorf("user.Service.RefreshToken.GenerateRefreshToken: %w", err)
	}
	newRecord := &database.UserRefreshToken{
		UserID:    u.ID,
		Token:     newRaw,
		ExpiresAt: expiresAt,
	}
	if err := s.repo.SaveRefreshToken(ctx, newRecord); err != nil {
		return nil, fmt.Errorf("user.Service.RefreshToken.SaveRefreshToken: %w", err)
	}

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRaw,
	}, nil
}

func (s *service) Logout(ctx context.Context, refreshToken string) error {
	if err := s.repo.RevokeRefreshToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("user.Service.Logout: %w", err)
	}
	return nil
}

func (s *service) LogoutAllSessions(ctx context.Context, userID uuid.UUID) error {
	if err := s.repo.RevokeAllRefreshTokensForUser(ctx, userID); err != nil {
		return fmt.Errorf("user.Service.LogoutAllSessions: %w", err)
	}
	return nil
}
