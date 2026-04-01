package config

import (
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	App      AppConfig
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type AppConfig struct {
	Name string
	Env  string // dev, staging, prod
}

type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
}

type JWTConfig struct {
	Secret     string
	Expiration time.Duration
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "my-service")
	v.SetDefault("app.env", "development")

	v.SetDefault("server.port", 8080)
	v.SetDefault("server.readtimeout", "5s")
	v.SetDefault("server.writetimeout", "10s")

	v.SetDefault("database.maxconnections", 10)

	v.SetDefault("database.url", "postgresql://user:password@localhost/dbname")
	v.SetDefault("database.maxopenconns", 10)
	v.SetDefault("database.maxidleconns", 5)
	v.SetDefault("database.connmaxlifetime", "5m")
	v.SetDefault("database.connmaxidletime", "2m")

	v.SetDefault("jwt.expiration", "24h")
}

func LoadConfig(path string) (*Config, error) {

	vipConfig := viper.New()
	vipConfig.SetConfigFile(path)

	vipConfig.SetEnvPrefix("MAD") // every env var must start with MAD_
	vipConfig.AutomaticEnv()

	// Replace dots with underscore
	vipConfig.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	setDefaults(vipConfig)

	if err := vipConfig.ReadInConfig(); err != nil {
		return nil, err
	}

	// settings := vipConfig.AllSettings()
	// fmt.Printf("All Config: %+v\n", settings)

	var config Config
	if err := vipConfig.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}
