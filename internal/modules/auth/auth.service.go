package auth

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/pkg/crypto"
)

type Service interface {
	SignIn(ctx context.Context, req *SignInRequest) (*AuthResponse, error)
}

type service struct {
	repo      Repository
	jwtConfig *config.JWTConfig
}

func NewService(repo Repository, cfg *config.JWTConfig) Service {
	return &service{repo: repo, jwtConfig: cfg}
}

func (s *service) SignIn(ctx context.Context, req *SignInRequest) (*AuthResponse, error) {

	user, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, constants.Errors.ErrUserNotFound
	}

	// check password
	if err := crypto.ComparePasswords(user.PasswordHash, req.Password); err != nil {
		return nil, constants.Errors.ErrInvalidCredentials
	}

	user.PasswordHash = "" // clear password hash before returning user info

	return s.buildAuthResponse(user, false)
}

func (s *service) buildAuthResponse(u *database.User, isNew bool) (*AuthResponse, error) {
	jwtExpiry := time.Now().Add(s.jwtConfig.Expiration)

	token, err := s.generateToken(u, jwtExpiry)
	if err != nil {
		return nil, err
	}

	return &AuthResponse{
		User: &UserAuthInfo{
			ID:    u.ID,
			Email: u.Email,
			Role:  u.Role,
		},
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int64(s.jwtConfig.Expiration.Hours()),
		IsNewUser:   isNew,
	}, nil
}

func (s *service) generateToken(u *database.User, expiration time.Time) (string, error) {
	claims := jwt.MapClaims{
		"userId": u.ID,
		"role":   u.Role,
		"exp":    expiration.Unix(),
		"iat":    time.Now().Unix(),
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString([]byte(s.jwtConfig.Secret))
	if err != nil {
		return "", err
	}

	return tokenStr, nil
}
