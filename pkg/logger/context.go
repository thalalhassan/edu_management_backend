package logger

import "context"

// contextKey is a private type for context keys
type contextKey string

const loggerKey contextKey = "logger"

// FromContext retrieves the logger from the context
// Returns the logger if found, otherwise returns nil
func FromContext(ctx context.Context) Logger {
	if logger, ok := ctx.Value(loggerKey).(Logger); ok {
		return logger
	}
	return nil
}

// WithContext stores the logger in the context
// Returns a new context with the logger embedded
func WithContext(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, loggerKey, logger)
}
