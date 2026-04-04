package middleware

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

type AuthServiceClaims struct {
	UserID   string
	Username string
	Mobile   string
	Role     string
	jwt.RegisteredClaims
}

// AuthCheckMiddleware verifies JWT token and extracts user info
func AuthCheckMiddleware(jwtConfig *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.UnAuthorized(c, "authorization header is required")
			return
		}

		// Extract token from Bearer scheme
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.UnAuthorized(c, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := ValidateToken(c.Request.Context(), tokenString, jwtConfig.Secret)
		if err != nil {
			response.UnAuthorized(c, "invalid token: "+err.Error())
			return
		}

		// Store claims in context for later use
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("mobile", claims.Mobile)
		c.Set("role", claims.Role)

		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c *gin.Context) (string, error) {
	return getStringFromContext(c, "userID")
}

// GetRoleFromContext extracts user role from context
func GetRoleFromContext(c *gin.Context) (string, error) {
	return getStringFromContext(c, "role")
}

// GetUsernameFromContext extracts username from context
func GetUsernameFromContext(c *gin.Context) (string, error) {
	return getStringFromContext(c, "username")
}

// GetMobileFromContext extracts mobile from context
func GetMobileFromContext(c *gin.Context) (string, error) {
	return getStringFromContext(c, "mobile")
}

// helper function to get string value from context
func getStringFromContext(c *gin.Context, key string) (string, error) {
	v, exists := c.Get(key)
	if !exists {
		return "", fmt.Errorf("%s not found in context", key)
	}
	s, ok := v.(string)
	if !ok {
		return "", fmt.Errorf("%s is not a string", key)
	}
	return s, nil
}

func ValidateToken(ctx context.Context, tokenString string, jwtSecret string) (AuthServiceClaims, error) {
	claims := &AuthServiceClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})
	if err != nil || !token.Valid {
		return AuthServiceClaims{}, errors.New("invalid token: " + err.Error())
	}

	return *claims, nil
}
