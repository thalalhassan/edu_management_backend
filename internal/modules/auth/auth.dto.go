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

// SignInRequest represents the signin request payload
type SignInRequest struct {
	Email    string `json:"email" validate:"required,email" example:"user@example.com"`
	Password string `json:"password" validate:"required,min=6" example:"password123"`
}

// @Description Response after successful authentication
// @Name AuthResponse
type AuthResponse struct {
	User        *UserAuthInfo `json:"user"`
	AccessToken string        `json:"accessToken"`
	TokenType   string        `json:"tokenType"`
	ExpiresIn   int64         `json:"expiresIn"`                 // in seconds
	IsNewUser   bool          `json:"isNewUser" default:"false"` // indicates if the user was newly created during sign-in
}

// TokenClaims represents the JWT token claims
type TokenClaims struct {
	UserID string `json:"userId"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
	Role   string `json:"role"`
}
