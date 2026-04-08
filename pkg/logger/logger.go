package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is an interface for logging operations.
// Using an interface allows dependency injection and easier testing.
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

// zapLogger is the concrete Logger backed by go.uber.org/zap.
type zapLogger struct {
	logger *zap.Logger
	level  zapcore.Level
}

// wrap is a convenience constructor that keeps level in sync on every derived logger.
func (l *zapLogger) wrap(z *zap.Logger) Logger {
	return &zapLogger{logger: z, level: l.level}
}

// New creates a logger tuned for the given environment.
//
//	"development" → coloured debug output
//	"staging"     → JSON, Info level
//	default       → JSON, Warn level  (production)
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

	return &zapLogger{logger: zapL, level: config.Level.Level()}, nil
}

// ── Core log methods ──────────────────────────────────────────────────────────

func (l *zapLogger) Debug(msg string, fields ...zap.Field) { l.logger.Debug(msg, fields...) }
func (l *zapLogger) Info(msg string, fields ...zap.Field)  { l.logger.Info(msg, fields...) }
func (l *zapLogger) Warn(msg string, fields ...zap.Field)  { l.logger.Warn(msg, fields...) }
func (l *zapLogger) Error(msg string, fields ...zap.Field) { l.logger.Error(msg, fields...) }
func (l *zapLogger) Fatal(msg string, fields ...zap.Field) { l.logger.Fatal(msg, fields...) }
func (l *zapLogger) Panic(msg string, fields ...zap.Field) { l.logger.Panic(msg, fields...) }

// ── Utility ───────────────────────────────────────────────────────────────────

func (l *zapLogger) With(fields ...zap.Field) Logger { return l.wrap(l.logger.With(fields...)) }
func (l *zapLogger) Sugar() *zap.SugaredLogger       { return l.logger.Sugar() }
func (l *zapLogger) Sync() error                     { return l.logger.Sync() }

// ── Structured context helpers ────────────────────────────────────────────────

func (l *zapLogger) WithError(err error) Logger {
	return l.wrap(l.logger.With(zap.Error(err)))
}

func (l *zapLogger) WithField(key string, value interface{}) Logger {
	return l.wrap(l.logger.With(zap.Any(key, value)))
}

func (l *zapLogger) WithFields(fields map[string]interface{}) Logger {
	return l.wrap(l.logger.With(mapToFields(fields)...))
}

func (l *zapLogger) WithRequestID(requestID string) Logger {
	return l.wrap(l.logger.With(zap.String("request_id", requestID)))
}

func (l *zapLogger) WithUser(userID string, email string) Logger {
	return l.wrap(l.logger.With(
		zap.String("user_id", userID),
		zap.String("user_email", email),
	))
}

func (l *zapLogger) WithUserID(userID string) Logger {
	return l.wrap(l.logger.With(zap.String("user_id", userID)))
}

func (l *zapLogger) WithMethod(method string) Logger {
	return l.wrap(l.logger.With(zap.String("method", method)))
}

func (l *zapLogger) WithPath(path string) Logger {
	return l.wrap(l.logger.With(zap.String("path", path)))
}

func (l *zapLogger) WithStatusCode(statusCode int) Logger {
	return l.wrap(l.logger.With(zap.Int("status_code", statusCode)))
}

func (l *zapLogger) WithDuration(duration interface{}) Logger {
	return l.wrap(l.logger.With(zap.Any("duration", duration)))
}

func (l *zapLogger) WithRequest(requestID string, method string, path string) Logger {
	return l.wrap(l.logger.With(
		zap.String("request_id", requestID),
		zap.String("method", method),
		zap.String("path", path),
	))
}

func (l *zapLogger) WithContext(fields map[string]interface{}) Logger {
	return l.wrap(l.logger.With(mapToFields(fields)...))
}

// ── Helpers ───────────────────────────────────────────────────────────────────

// mapToFields converts a string-keyed map to a slice of zap.Field values.
func mapToFields(m map[string]interface{}) []zap.Field {
	fields := make([]zap.Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
