package middleware

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/constants"
	"github.com/thalalhassan/edu_management/internal/shared/response"
	"github.com/thalalhassan/edu_management/pkg/jwt"
)

// AuthCheckMiddleware verifies JWT token and extracts user info
func AuthCheckMiddleware(jwtConfig *config.JWTConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "authorization header is required")
			return
		}

		// Extract token from Bearer scheme
		parts := strings.Split(authHeader, " ")

		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization header format")
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := jwt.ParseAccessToken(tokenString, jwtConfig.Secret)
		if err != nil {
			response.Unauthorized(c, "invalid token: "+err.Error())
			return
		}

		// Store claims in context for later use

		c.Set(constants.UserIDContextKey, claims.UserID)
		c.Set(constants.EmailContextKey, claims.Email)
		c.Set(constants.RoleContextKey, claims.Role)

		c.Next()
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c *gin.Context) (string, error) {
	return getStringFromContext(c, constants.UserIDContextKey)
}

// GetRoleFromContext extracts user role from context
func GetRoleFromContext(c *gin.Context) (string, error) {
	return getStringFromContext(c, constants.RoleContextKey)
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
