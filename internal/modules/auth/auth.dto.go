package auth

import "github.com/google/uuid"

// @Description Information about authenticated user
// @Name UserAuthInfo
type UserAuthInfo struct {
	ID    uuid.UUID `json:"id"`
	Email string    `json:"email"`
	Role  string    `json:"role"`
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
