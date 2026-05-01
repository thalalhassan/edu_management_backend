package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/thalalhassan/edu_management/pkg/logger"
	"go.uber.org/zap"
)

func ZapLogger(baseLogger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Generate unique request ID for tracing
		requestID, err := uuid.NewV7()
		if err != nil {
			requestID = uuid.New()
		}

		// Create request-scoped logger with request context
		reqLogger := baseLogger.With(
			zap.String("request_id", requestID.String()),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("ip", c.ClientIP()),
		)

		// Store in Gin context for easy access in handlers
		c.Set("logger", reqLogger)

		// Also store in request context for accessing via context.Context
		newCtx := logger.WithContext(c.Request.Context(), reqLogger)
		c.Request = c.Request.WithContext(newCtx)

		// Process request
		c.Next()

		duration := time.Since(start)

		status := c.Writer.Status()
		query := c.Request.URL.RawQuery
		userAgent := c.Request.UserAgent()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()

		fields := []zap.Field{
			zap.Int("status", status),
			zap.String("query", query),
			zap.String("user_agent", userAgent),
			zap.Duration("latency", duration),
		}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		if status >= 500 {
			reqLogger.Error("server error", fields...)
		} else if status >= 400 {
			reqLogger.Warn("client error", fields...)
		} else {
			reqLogger.Info("request completed", fields...)
		}
	}
}
