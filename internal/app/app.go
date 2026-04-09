package app

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/database"
	"github.com/thalalhassan/edu_management/internal/shared/validation"
	"github.com/thalalhassan/edu_management/pkg/logger"
)

type App struct {
	Config *config.Config
	DB     *database.DB
	Logger logger.Logger
}

func NewApp(ctx context.Context) (*App, error) {

	config, err := config.LoadConfig(".env")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Initialize logger based on environment
	appLogger, err := logger.New(config.App.Env)
	if err != nil {
		return nil, err
	}

	db, err := database.New(ctx, config, appLogger)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	validation.InitValidator()

	return &App{
		Config: config,
		DB:     db,
		Logger: appLogger,
	}, nil
}
