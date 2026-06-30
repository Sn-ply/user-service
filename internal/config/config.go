package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	JWT      JWTConfig
}

type ServerConfig struct {
	Port string
}

type DatabaseConfig struct {
	URL string
}

type JWTConfig struct {
	Secret              string
	AccessTokenMinutes  int
	RefreshTokenDays    int
}

func Load() (*Config, error) {
	viper.SetDefault("SERVER_PORT", "8081")
	viper.SetDefault("JWT_ACCESS_TOKEN_MINUTES", 15)
	viper.SetDefault("JWT_REFRESH_TOKEN_DAYS", 7)

	viper.AutomaticEnv()

	cfg := &Config{
		Server: ServerConfig{
			Port: viper.GetString("SERVER_PORT"),
		},
		Database: DatabaseConfig{
			URL: viper.GetString("DATABASE_URL"),
		},
		JWT: JWTConfig{
			Secret:             viper.GetString("JWT_SECRET"),
			AccessTokenMinutes: viper.GetInt("JWT_ACCESS_TOKEN_MINUTES"),
			RefreshTokenDays:   viper.GetInt("JWT_REFRESH_TOKEN_DAYS"),
		},
	}

	if cfg.Database.URL == "" {
		cfg.Database.URL = "postgres://snaply:snaply_secret@localhost:5432/users?sslmode=disable"
	}
	if cfg.JWT.Secret == "" {
		cfg.JWT.Secret = "dev_secret_change_in_production"
	}

	return cfg, nil
}
