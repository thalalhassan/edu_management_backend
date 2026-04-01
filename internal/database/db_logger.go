package database

import (
	"context"
	"log"
	"time"

	"gorm.io/gorm/logger"
)

type GormLogger struct{}

func NewGormLogger() logger.Interface {
	return &GormLogger{}
}

func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return l
}

func (l *GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	log.Printf("[GORM][INFO] "+msg, data...)
}

func (l *GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	log.Printf("[GORM][WARN] "+msg, data...)
}

func (l *GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	log.Printf("[GORM][ERROR] "+msg, data...)
}

func (l *GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	if err != nil {
		log.Printf("[GORM][ERROR] %s | %v | rows=%d | %s", elapsed, err, rows, sql)
		return
	}

	// slow query threshold (tune this)
	if elapsed > 200*time.Millisecond {
		log.Printf("[GORM][SLOW] %s | rows=%d | %s", elapsed, rows, sql)
	}
}
