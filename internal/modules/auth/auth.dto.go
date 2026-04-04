package auth

import (
	"github.com/thalalhassan/edu_management/internal/database"
)

// @Description Information about authenticated user
// @Name UserAuthInfo
type UserAuthInfo struct {
	ID    string            `json:"id"`
	Email string            `json:"email"`
	Role  database.UserRole `json:"role"`
}

// TokenClaims represents the JWT token claims
type TokenClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
	Role   string `json:"role"`
}

// ──────────────────────────────────────────────────────────────
// LOGIN
// ──────────────────────────────────────────────────────────────

type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserAuthInfo `json:"user"`
}

// RefreshRequest exchanges a refresh token for a new access token.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type RefreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"` // rotated
}
