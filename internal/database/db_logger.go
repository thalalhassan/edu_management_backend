package database

import (
	"context"
	"time"

	appLogger "github.com/thalalhassan/edu_management/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

// contextKey is an unexported type to avoid context key collisions across packages.
type contextKey string

const loggerContextKey contextKey = "logger"

// GormLogger bridges GORM's logger.Interface to your custom appLogger.Logger.
type GormLogger struct {
	log   appLogger.Logger
	level logger.LogLevel
}

// NewGormLogger returns a GORM-compatible logger wrapping your appLogger.
// Level defaults to logger.Info; callers can change it via LogMode.
func NewGormLogger(log appLogger.Logger) logger.Interface {
	return &GormLogger{
		log:   log,
		level: logger.Info,
	}
}

// LogMode returns a new GormLogger at the requested level (does not mutate receiver).
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{
		log:   l.log,
		level: level,
	}
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Info {
		l.log.Info(msg, zap.Any("data", data))
	}
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Warn {
		l.log.Warn(msg, zap.Any("data", data))
	}
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.level >= logger.Error {
		l.log.Error(msg, zap.Any("data", data))
	}
}

// Trace logs every SQL query executed by GORM.
// It promotes slow queries to Warn and failed queries to Error.
func (l *GormLogger) Trace(
	ctx context.Context,
	begin time.Time,
	fc func() (string, int64),
	err error,
) {
	if l.level == logger.Silent {
		return
	}

	// Use a request-scoped logger when the middleware has injected one into ctx;
	// otherwise fall back to the base logger.  We use a local variable so the
	// receiver is never mutated.
	activeLog := l.log
	if ctxLogger, ok := ctx.Value(loggerContextKey).(appLogger.Logger); ok {
		activeLog = ctxLogger
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	fields := []zap.Field{
		zap.Duration("duration", elapsed),
		zap.Int64("rows", rows),
		zap.String("sql", sql),
	}

	switch {
	case err != nil && l.level >= logger.Error:
		activeLog.Error("db query failed", append(fields, zap.Error(err))...)

	case elapsed > 200*time.Millisecond && l.level >= logger.Warn:
		activeLog.Warn("db slow query", fields...)

	case l.level >= logger.Info:
		activeLog.Info("db query", fields...)
	}
}
