package config

import (
	"errors"
	"os"
	"time"
)

type Config struct {
	Environment      string
	HTTPAddress      string
	DatabaseURL      string
	ShutdownTimeout  time.Duration
	ReadinessTimeout time.Duration
}

func Load() (Config, error) {
	cfg := Config{
		Environment:      envOrDefault("ATALAYA_ENV", "development"),
		HTTPAddress:      envOrDefault("HTTP_ADDRESS", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		ShutdownTimeout:  10 * time.Second,
		ReadinessTimeout: 2 * time.Second,
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}

	return cfg, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
