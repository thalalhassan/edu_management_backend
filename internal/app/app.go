package app

import (
	"context"
	"fmt"

	"github.com/thalalhassan/edu_management/internal/config"
	"github.com/thalalhassan/edu_management/internal/database"
)

type App struct {
	Config *config.Config
	DB     *database.DB
}

func NewApp(ctx context.Context) *App {

	config, err := config.LoadConfig(".env")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	db, err := database.New(ctx, config)
	if err != nil {
		panic(fmt.Errorf("failed to connect to database: %w", err))
	}

	return &App{
		Config: config,
		DB:     db,
	}
}
