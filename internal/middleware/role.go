package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/response"
)

// RoleCheckMiddleware verifies that the user has the required role
func RoleCheckMiddleware(requiredRoles ...database.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, err := GetRoleFromContext(c)
		if err != nil {
			response.AccessDenied(c, "unable to determine user role: ")
			return
		}

		// Check if user's role is in the list of required roles
		hasRequiredRole := false
		for _, requiredRole := range requiredRoles {
			if role == string(requiredRole) {
				hasRequiredRole = true
				break
			}
		}

		if !hasRequiredRole {
			response.AccessDenied(c, "insufficient permissions for this resource")
			return
		}

		c.Next()
	}
}

// IsAdmin is a convenience middleware to check if user is admin
func IsAdmin() gin.HandlerFunc {
	return RoleCheckMiddleware(database.UserRoleAdmin, database.UserRoleSuperAdmin)
}

// Allowed is a convenience middleware to check if user has any of the allowed roles
func IsAllowed(allowedRoles ...database.UserRole) gin.HandlerFunc {
	return RoleCheckMiddleware(allowedRoles...)
}
