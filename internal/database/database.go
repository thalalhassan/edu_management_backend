package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/pkg/logger"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Gorm *gorm.DB // ORM features, migrations, complex queries
	PSQL *sql.DB  // connection-pool management and health checks
}

// New initialises a DB with connection-pool tuning and an initial ping.
func New(ctx context.Context, cfg *config.Config, zapLogger logger.Logger) (*DB, error) {
	gormDB, err := gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		// NewGormLogger now takes a single argument; the log level can be
		// changed later via gormDB.Logger = gormDB.Logger.LogMode(…).
		Logger: NewGormLogger(zapLogger),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect DB: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	// ── Connection-pool tuning ────────────────────────────────────────────────
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// ── Initial ping ──────────────────────────────────────────────────────────
	dbCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(dbCtx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}

	return &DB{
		Gorm: gormDB,
		PSQL: sqlDB,
	}, nil
}

// Health is used by readiness probes.
func (db *DB) Health(ctx context.Context) error {
	return db.PSQL.PingContext(ctx)
}

// Close performs a graceful shutdown of the connection pool.
func (db *DB) Close() error {
	return db.PSQL.Close()
}
