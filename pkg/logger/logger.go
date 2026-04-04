package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is an interface for logging operations
// This allows for dependency injection and better testability
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	With(fields ...zap.Field) Logger
	Sugar() *zap.SugaredLogger
	Sync() error
	WithRequestID(requestID string) Logger
	WithUser(userID string, email string) Logger
	WithUserID(userID string) Logger
	WithMethod(method string) Logger
	WithPath(path string) Logger
	WithStatusCode(statusCode int) Logger
	WithDuration(duration interface{}) Logger
	WithRequest(requestID string, method string, path string) Logger
	WithContext(fields map[string]interface{}) Logger
	WithError(err error) Logger
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

// zapLogger is a concrete implementation of the Logger interface using zap
type zapLogger struct {
	logger *zap.Logger
}

// New creates a new logger instance based on the environment
func New(appEnv string) (Logger, error) {
	var config zap.Config

	switch appEnv {
	case "development":
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		config.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
	case "staging":
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.InfoLevel)
	default: // production
		config = zap.NewProductionConfig()
		config.Level = zap.NewAtomicLevelAt(zapcore.WarnLevel)
	}

	zapL, err := config.Build()
	if err != nil {
		return nil, err
	}

	return &zapLogger{logger: zapL}, nil
}

// Debug logs a message at DebugLevel
func (l *zapLogger) Debug(msg string, fields ...zap.Field) {
	l.logger.Debug(msg, fields...)
}

// Info logs a message at InfoLevel
func (l *zapLogger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

// Warn logs a message at WarnLevel
func (l *zapLogger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

// Error logs a message at ErrorLevel
func (l *zapLogger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

// Fatal logs a message at FatalLevel and exits
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

// Panic logs a message at PanicLevel and panics
func (l *zapLogger) Panic(msg string, fields ...zap.Field) {
	l.logger.Panic(msg, fields...)
}

// With returns a logger with additional context fields
func (l *zapLogger) With(fields ...zap.Field) Logger {
	return &zapLogger{logger: l.logger.With(fields...)}
}

// Sugar returns a sugared logger for more convenient logging
func (l *zapLogger) Sugar() *zap.SugaredLogger {
	return l.logger.Sugar()
}

// Sync flushes any buffered log entries
func (l *zapLogger) Sync() error {
	return l.logger.Sync()
}

// WithError returns a logger with an error field
func (l *zapLogger) WithError(err error) Logger {
	return &zapLogger{logger: l.logger.With(zap.Error(err))}
}

// WithField returns a logger with a single field
func (l *zapLogger) WithField(key string, value interface{}) Logger {
	return &zapLogger{logger: l.logger.With(zap.Any(key, value))}
}

// WithFields returns a logger with multiple fields
func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &zapLogger{logger: l.logger.With(zapFields...)}
}

// WithRequestID adds request ID to logger context
func (l *zapLogger) WithRequestID(requestID string) Logger {
	return &zapLogger{logger: l.logger.With(zap.String("request_id", requestID))}
}

// WithUser adds user information to logger context
func (l *zapLogger) WithUser(userID string, email string) Logger {
	return &zapLogger{logger: l.logger.With(
		zap.String("user_id", userID),
		zap.String("user_email", email),
	)}
}

// WithUserID adds user ID to logger context
func (l *zapLogger) WithUserID(userID string) Logger {
	return &zapLogger{logger: l.logger.With(zap.String("user_id", userID))}
}

// WithMethod adds HTTP method to logger context
func (l *zapLogger) WithMethod(method string) Logger {
	return &zapLogger{logger: l.logger.With(zap.String("method", method))}
}

// WithPath adds HTTP path to logger context
func (l *zapLogger) WithPath(path string) Logger {
	return &zapLogger{logger: l.logger.With(zap.String("path", path))}
}

// WithStatusCode adds HTTP status code to logger context
func (l *zapLogger) WithStatusCode(statusCode int) Logger {
	return &zapLogger{logger: l.logger.With(zap.Int("status_code", statusCode))}
}

// WithDuration adds operation duration to logger context
func (l *zapLogger) WithDuration(duration interface{}) Logger {
	return &zapLogger{logger: l.logger.With(zap.Any("duration", duration))}
}

// WithRequest adds HTTP request context information
func (l *zapLogger) WithRequest(requestID string, method string, path string) Logger {
	return &zapLogger{logger: l.logger.With(
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path),
	)}
}

// WithContext adds multiple context fields at once
func (l *zapLogger) WithContext(fields map[string]interface{}) Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}
	return &zapLogger{logger: l.logger.With(zapFields...)}
}
