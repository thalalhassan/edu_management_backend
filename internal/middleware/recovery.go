package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/thalalhassan/edu_management/internal/shared/response"
	"github.com/thalalhassan/edu_management/pkg/logger"
	"go.uber.org/zap"
)

// RecoveryWithResponse is a Gin middleware that recovers from panics,
// logs the error, and returns a standardized JSON error response using
// the shared ResponseWrapper format.
func RecoveryWithResponse() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Try to use request-scoped logger when available
				if l, ok := c.Get("logger"); ok {
					if log, ok := l.(logger.Logger); ok {
						log.Error("panic recovered",
							zap.Any("panic", r),
							zap.Stack("stacktrace"),
						)
					}
				}

				// Ensure we send a consistent JSON error shape
				if !c.IsAborted() {
					response.InternalError(
						c,
						"An unexpected error occurred. Please try again later.",
					)
				}
			}
		}()

		c.Next()
	}
}
