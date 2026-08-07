package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Environment      string
	HTTPAddress      string
	DatabaseURL      string
	ShutdownTimeout  time.Duration
	ReadinessTimeout time.Duration
	PollInterval     time.Duration
	Sentry           SentryConfig
}

type SentryConfig struct {
	AuthToken   string
	BaseURL     string
	OrgSlug     string
	ProjectSlug string
}

func (config SentryConfig) Enabled() bool { return config.AuthToken != "" }

func Load() (Config, error) {
	cfg := Config{
		Environment:      envOrDefault("ATALAYA_ENV", "development"),
		HTTPAddress:      envOrDefault("HTTP_ADDRESS", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		ShutdownTimeout:  10 * time.Second,
		ReadinessTimeout: 2 * time.Second,
		Sentry: SentryConfig{
			AuthToken:   os.Getenv("SENTRY_AUTH_TOKEN"),
			BaseURL:     envOrDefault("SENTRY_BASE_URL", "https://sentry.io"),
			OrgSlug:     os.Getenv("SENTRY_ORG_SLUG"),
			ProjectSlug: os.Getenv("SENTRY_PROJECT_SLUG"),
		},
	}
	var err error
	cfg.PollInterval, err = durationSeconds("POLL_INTERVAL_SECONDS", 120)
	if err != nil {
		return Config{}, err
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.Sentry.Enabled() && (cfg.Sentry.OrgSlug == "" || cfg.Sentry.ProjectSlug == "") {
		return Config{}, errors.New("SENTRY_ORG_SLUG and SENTRY_PROJECT_SLUG are required when SENTRY_AUTH_TOKEN is set")
	}

	return cfg, nil
}

func durationSeconds(key string, fallback int) (time.Duration, error) {
	raw := envOrDefault(key, strconv.Itoa(fallback))
	seconds, err := strconv.Atoi(raw)
	if err != nil || seconds <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return time.Duration(seconds) * time.Second, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
