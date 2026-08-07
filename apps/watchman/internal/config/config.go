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
	Interpreter      InterpreterConfig
	Sentry           SentryConfig
	Telegram         TelegramConfig
}

type InterpreterConfig struct {
	URL          string
	WorkerID     string
	Timeout      time.Duration
	PollInterval time.Duration
}

type SentryConfig struct {
	AuthToken   string
	BaseURL     string
	OrgSlug     string
	ProjectSlug string
}

type TelegramConfig struct {
	BotToken            string
	ChatID              string
	WorkerID            string
	Timeout             time.Duration
	PollInterval        time.Duration
	DeduplicationWindow time.Duration
	RateLimitWindow     time.Duration
	RateLimitPerApp     int
	InterpreterCooldown time.Duration
	AtalayaPublicURL    string
}

func (config TelegramConfig) Enabled() bool { return config.BotToken != "" && config.ChatID != "" }

func (config SentryConfig) Enabled() bool { return config.AuthToken != "" }

func Load() (Config, error) {
	cfg := Config{
		Environment:      envOrDefault("ATALAYA_ENV", "development"),
		HTTPAddress:      envOrDefault("HTTP_ADDRESS", ":8080"),
		DatabaseURL:      os.Getenv("DATABASE_URL"),
		ShutdownTimeout:  10 * time.Second,
		ReadinessTimeout: 2 * time.Second,
		Interpreter: InterpreterConfig{
			URL:      envOrDefault("INTERPRETER_URL", "http://interpreter:8000"),
			WorkerID: envOrDefault("INTERPRETER_WORKER_ID", "watchman-1"),
		},
		Sentry: SentryConfig{
			AuthToken:   os.Getenv("SENTRY_AUTH_TOKEN"),
			BaseURL:     envOrDefault("SENTRY_BASE_URL", "https://sentry.io"),
			OrgSlug:     os.Getenv("SENTRY_ORG_SLUG"),
			ProjectSlug: os.Getenv("SENTRY_PROJECT_SLUG"),
		},
		Telegram: TelegramConfig{
			BotToken:         envOrDefault("TELEGRAM_BOT_TOKEN", ""),
			ChatID:           envOrDefault("TELEGRAM_CHAT_ID", ""),
			WorkerID:         envOrDefault("TELEGRAM_WORKER_ID", "watchman-telegram-1"),
			AtalayaPublicURL: envOrDefault("ATALAYA_PUBLIC_URL", ""),
		},
	}
	var err error
	cfg.PollInterval, err = durationSeconds("POLL_INTERVAL_SECONDS", 120)
	if err != nil {
		return Config{}, err
	}
	cfg.Interpreter.Timeout, err = durationSeconds("INTERPRETER_TIMEOUT_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}
	cfg.Interpreter.PollInterval, err = durationSeconds("INTERPRETER_JOB_POLL_SECONDS", 2)
	if err != nil {
		return Config{}, err
	}
	cfg.Telegram.Timeout, err = durationSeconds("TELEGRAM_TIMEOUT_SECONDS", 10)
	if err != nil {
		return Config{}, err
	}
	cfg.Telegram.PollInterval, err = durationSeconds("TELEGRAM_JOB_POLL_SECONDS", 2)
	if err != nil {
		return Config{}, err
	}
	cfg.Telegram.DeduplicationWindow, err = durationSeconds("TELEGRAM_DEDUP_WINDOW_SECONDS", 900)
	if err != nil {
		return Config{}, err
	}
	cfg.Telegram.RateLimitWindow, err = durationSeconds("TELEGRAM_RATE_LIMIT_WINDOW_SECONDS", 600)
	if err != nil {
		return Config{}, err
	}
	cfg.Telegram.RateLimitPerApp, err = positiveInt("TELEGRAM_RATE_LIMIT_PER_APP", 10)
	if err != nil {
		return Config{}, err
	}
	cfg.Telegram.InterpreterCooldown, err = durationSeconds("TELEGRAM_INTERPRETER_ALERT_COOLDOWN_SECONDS", 1800)
	if err != nil {
		return Config{}, err
	}
	if (cfg.Telegram.BotToken == "") != (cfg.Telegram.ChatID == "") {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be configured together")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.Sentry.Enabled() && (cfg.Sentry.OrgSlug == "" || cfg.Sentry.ProjectSlug == "") {
		return Config{}, errors.New("SENTRY_ORG_SLUG and SENTRY_PROJECT_SLUG are required when SENTRY_AUTH_TOKEN is set")
	}

	return cfg, nil
}

func positiveInt(key string, fallback int) (int, error) {
	raw := envOrDefault(key, strconv.Itoa(fallback))
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return value, nil
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
