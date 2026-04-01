package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/thalalhassan/edu_management/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DB struct {
	Gorm *gorm.DB // manage ORM features, migrations, and complex queries
	PSQL *sql.DB  // manage db connection pool and health checks
}

// New initializes DB with retry + pooling + validation
func New(ctx context.Context, cfg *config.Config) (*DB, error) {
	var gormDB *gorm.DB
	var err error

	gormDB, err = gorm.Open(postgres.Open(cfg.Database.URL), &gorm.Config{
		Logger: NewGormLogger(), // custom logger (see below)
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect DB: %w", err)
	}

	// ---- GET SQL DB ----
	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, err
	}

	// ---- CONNECTION POOL TUNING ----
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.Database.ConnMaxIdleTime)

	// ---- INITIAL PING ----
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

// Health check (used in readiness probes)
func (db *DB) Health(ctx context.Context) error {
	return db.PSQL.PingContext(ctx)
}

// Graceful shutdown
func (db *DB) Close() error {
	return db.PSQL.Close()
}
