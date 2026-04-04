package config

import (
	"time"

	"github.com/caarlos0/env/v11"
	"github.com/joho/godotenv"
)

type Config struct {
	App      AppConfig      `envPrefix:"EDU_APP_"`
	Server   ServerConfig   `envPrefix:"EDU_SERVER_"`
	Database DatabaseConfig `envPrefix:"EDU_DATABASE_"`
	JWT      JWTConfig      `envPrefix:"EDU_JWT_"`
}

type AppConfig struct {
	Name string `env:"NAME" envDefault:"MyGoApp"`
	Env  string `env:"ENV" envDefault:"development"` // development, staging, production
}

type ServerConfig struct {
	Port         int           `env:"PORT" envDefault:"8080"`
	ReadTimeout  time.Duration `env:"READ_TIMEOUT" envDefault:"5s"`
	WriteTimeout time.Duration `env:"WRITE_TIMEOUT" envDefault:"10s"`
}

type DatabaseConfig struct {
	URL             string        `env:"URL" envDefault:"postgresql://postgres:postgres@localhost:5432/edu_db"`
	MaxOpenConns    int           `env:"MAX_OPEN_CONNS" envDefault:"10"`
	MaxIdleConns    int           `env:"MAX_IDLE_CONNS" envDefault:"5"`
	ConnMaxLifetime time.Duration `env:"CONN_MAX_LIFETIME" envDefault:"1h"`
	ConnMaxIdleTime time.Duration `env:"CONN_MAX_IDLE_TIME" envDefault:"30m"`
}

type JWTConfig struct {
	Secret     string        `env:"SECRET" envDefault:"10"`
	Expiration time.Duration `env:"EXPIRATION" envDefault:"72h"`
}

func LoadConfig(path string) (*Config, error) {
	// Ignore error if .env is missing (common in CI/CD or production environments)
	_ = godotenv.Load(path)

	cfg := Config{}

	// env.Parse will populate defaults if the ENV var is missing or empty
	if err := env.Parse(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
