package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Config struct {
	Environment             string
	HTTPAddress             string
	DatabaseURL             string
	ShutdownTimeout         time.Duration
	ReadinessTimeout        time.Duration
	PollInterval            time.Duration
	Interpreter             InterpreterConfig
	Sentry                  SentryConfig
	ApplicationInsights     ApplicationInsightsConfig
	Telegram                TelegramConfig
	Auth                    AuthConfig
	Reporting               ReportingConfig
	Deployments             DeploymentConfig
	AvailabilityCatalogPath string
	HealthchecksPingURL     string
	LLMMonthlyBudgetUSD     float64
	EventRetentionDays      int
}

type DeploymentConfig struct{ IngestToken, RailwayToken string }

type ReportingConfig struct {
	SchedulerInterval time.Duration
	WorkerID          string
}

type AuthConfig struct {
	PasswordHash    string
	SessionDuration time.Duration
	CookieSecure    bool
}

type ApplicationInsightsConfig struct {
	TenantID, ClientID, ClientSecret, WorkspaceID, ResourceID string
	TokenURL, LogsURL                                         string
	Component, DisplayName, Environment                       string
	Overlap                                                   time.Duration
}

func (config ApplicationInsightsConfig) Enabled() bool {
	return config.TenantID != "" && config.ClientID != "" && config.ClientSecret != "" && config.WorkspaceID != "" && config.ResourceID != ""
}

type InterpreterConfig struct {
	URL          string
	WorkerID     string
	Timeout      time.Duration
	PollInterval time.Duration
}

type SentryConfig struct {
	AuthToken    string
	BaseURL      string
	OrgSlug      string
	CatalogPath  string
	Applications []SentryApplication
}

type SentryApplication struct {
	Slug         string
	AlertPolicy  domain.AlertPolicy
	Integrations []SentryIntegration
}

type SentryIntegration struct {
	Component    string   `json:"component"`
	DisplayName  string   `json:"display_name"`
	Project      string   `json:"project"`
	Environments []string `json:"environments"`
}

type sentryCatalog struct {
	Applications []sentryCatalogApplication `json:"applications"`
}

type sentryCatalogApplication struct {
	Slug         string              `json:"slug"`
	AlertPolicy  alertPolicyOverride `json:"alert_policy"`
	Integrations []SentryIntegration `json:"integrations"`
}

type alertPolicyOverride struct {
	Enabled                    *bool    `json:"enabled"`
	AlwaysAlertSeverities      []string `json:"always_alert_severities"`
	ActionableAlertSeverities  []string `json:"actionable_alert_severities"`
	DeduplicationWindowSeconds *int     `json:"deduplication_window_seconds"`
	RateLimitWindowSeconds     *int     `json:"rate_limit_window_seconds"`
	RateLimitCount             *int     `json:"rate_limit_count"`
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
			CatalogPath: envOrDefault("SENTRY_CATALOG_PATH", "config/sentry-integrations.json"),
		},
		ApplicationInsights: ApplicationInsightsConfig{
			TenantID: os.Getenv("AZURE_TENANT_ID"), ClientID: os.Getenv("AZURE_CLIENT_ID"),
			ClientSecret: os.Getenv("AZURE_CLIENT_SECRET"), WorkspaceID: os.Getenv("AZURE_LOG_ANALYTICS_WORKSPACE_ID"),
			ResourceID: os.Getenv("AZURE_APPLICATION_INSIGHTS_RESOURCE_ID"),
			TokenURL:   envOrDefault("AZURE_TOKEN_BASE_URL", "https://login.microsoftonline.com"),
			LogsURL:    envOrDefault("AZURE_LOGS_BASE_URL", "https://api.loganalytics.azure.com"),
			Component:  envOrDefault("NOTIZAP_COMPONENT", "backend"), DisplayName: envOrDefault("NOTIZAP_DISPLAY_NAME", "Backend"),
			Environment: envOrDefault("NOTIZAP_ENVIRONMENT", "production"),
		},
		Telegram: TelegramConfig{
			BotToken:         envOrDefault("TELEGRAM_BOT_TOKEN", ""),
			ChatID:           envOrDefault("TELEGRAM_CHAT_ID", ""),
			WorkerID:         envOrDefault("TELEGRAM_WORKER_ID", "watchman-telegram-1"),
			AtalayaPublicURL: envOrDefault("ATALAYA_PUBLIC_URL", ""),
		},
		Auth: AuthConfig{
			PasswordHash: os.Getenv("ATALAYA_ADMIN_PASSWORD_HASH"),
			CookieSecure: envOrDefault("ATALAYA_COOKIE_SECURE", "true") == "true",
		},
		Reporting:               ReportingConfig{WorkerID: envOrDefault("DAILY_REPORT_WORKER_ID", "watchman-daily-report-1")},
		Deployments:             DeploymentConfig{IngestToken: os.Getenv("DEPLOYMENT_INGEST_TOKEN"), RailwayToken: os.Getenv("RAILWAY_WEBHOOK_TOKEN")},
		AvailabilityCatalogPath: envOrDefault("AVAILABILITY_CATALOG_PATH", "config/availability-targets.json"),
		HealthchecksPingURL:     os.Getenv("HEALTHCHECKS_PING_URL"),
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
	cfg.ApplicationInsights.Overlap, err = durationSeconds("AZURE_QUERY_OVERLAP_SECONDS", 300)
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
	cfg.Auth.SessionDuration, err = durationSeconds("ATALAYA_SESSION_DURATION_SECONDS", 86400)
	if err != nil {
		return Config{}, err
	}
	cfg.Reporting.SchedulerInterval, err = durationSeconds("DAILY_REPORT_SCHEDULER_SECONDS", 30)
	if err != nil {
		return Config{}, err
	}
	cfg.LLMMonthlyBudgetUSD, err = floatFromEnv("LLM_MONTHLY_BUDGET_USD", 5.0)
	if err != nil {
		return Config{}, err
	}
	cfg.EventRetentionDays, err = positiveInt("EVENT_RETENTION_DAYS", 90)
	if err != nil {
		return Config{}, err
	}

	if cfg.Auth.PasswordHash == "" {
		return Config{}, errors.New("ATALAYA_ADMIN_PASSWORD_HASH is required")
	}
	if (cfg.Telegram.BotToken == "") != (cfg.Telegram.ChatID == "") {
		return Config{}, errors.New("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be configured together")
	}

	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL is required")
	}
	if cfg.Sentry.Enabled() && cfg.Sentry.OrgSlug == "" {
		return Config{}, errors.New("SENTRY_ORG_SLUG is required when SENTRY_AUTH_TOKEN is set")
	}
	azureValues := []string{cfg.ApplicationInsights.TenantID, cfg.ApplicationInsights.ClientID, cfg.ApplicationInsights.ClientSecret, cfg.ApplicationInsights.WorkspaceID, cfg.ApplicationInsights.ResourceID}
	configuredAzureValues := 0
	for _, value := range azureValues {
		if value != "" {
			configuredAzureValues++
		}
	}
	if configuredAzureValues != 0 && configuredAzureValues != len(azureValues) {
		return Config{}, errors.New("Azure credentials, workspace ID and Application Insights resource ID must be configured together")
	}
	if cfg.ApplicationInsights.Enabled() && !strings.HasPrefix(strings.ToLower(cfg.ApplicationInsights.ResourceID), "/subscriptions/") {
		return Config{}, errors.New("AZURE_APPLICATION_INSIGHTS_RESOURCE_ID must be an Azure resource ID")
	}
	defaultPolicy := domain.AlertPolicy{
		Enabled: true, AlwaysAlertSeverities: []string{"critical", "high"}, ActionableAlertSeverities: []string{"medium"},
		DeduplicationWindowSeconds: int(cfg.Telegram.DeduplicationWindow.Seconds()),
		RateLimitWindowSeconds:     int(cfg.Telegram.RateLimitWindow.Seconds()), RateLimitCount: cfg.Telegram.RateLimitPerApp,
	}
	catalog, err := loadSentryCatalog(cfg.Sentry.CatalogPath, defaultPolicy)
	if err != nil {
		return Config{}, err
	}
	cfg.Sentry.Applications = catalog

	return cfg, nil
}

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func loadSentryCatalog(path string, defaults domain.AlertPolicy) ([]SentryApplication, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read Sentry catalog: %w", err)
	}
	var catalog sentryCatalog
	if err := json.Unmarshal(raw, &catalog); err != nil {
		return nil, fmt.Errorf("decode Sentry catalog: %w", err)
	}
	if len(catalog.Applications) == 0 {
		return nil, errors.New("Sentry catalog must contain at least one application")
	}
	applications := make([]SentryApplication, 0, len(catalog.Applications))
	seenApplications := map[string]bool{}
	for _, item := range catalog.Applications {
		if !slugPattern.MatchString(item.Slug) || seenApplications[item.Slug] {
			return nil, fmt.Errorf("invalid or duplicate Sentry application slug %q", item.Slug)
		}
		seenApplications[item.Slug] = true
		policy := mergePolicy(defaults, item.AlertPolicy)
		if err := validatePolicy(policy); err != nil {
			return nil, fmt.Errorf("invalid alert policy for %s: %w", item.Slug, err)
		}
		if len(item.Integrations) == 0 {
			return nil, fmt.Errorf("Sentry application %s has no integrations", item.Slug)
		}
		seenComponents := map[string]bool{}
		for _, integration := range item.Integrations {
			if !slugPattern.MatchString(integration.Component) || seenComponents[integration.Component] {
				return nil, fmt.Errorf("invalid or duplicate component %q for %s", integration.Component, item.Slug)
			}
			seenComponents[integration.Component] = true
			if integration.DisplayName == "" || integration.Project == "" || len(integration.Environments) == 0 {
				return nil, fmt.Errorf("component %s/%s requires display_name, project and environments", item.Slug, integration.Component)
			}
			for _, environment := range integration.Environments {
				if environment == "" {
					return nil, fmt.Errorf("component %s/%s contains an empty environment", item.Slug, integration.Component)
				}
			}
		}
		applications = append(applications, SentryApplication{Slug: item.Slug, AlertPolicy: policy, Integrations: item.Integrations})
	}
	return applications, nil
}

func mergePolicy(policy domain.AlertPolicy, override alertPolicyOverride) domain.AlertPolicy {
	if override.Enabled != nil {
		policy.Enabled = *override.Enabled
	}
	if override.AlwaysAlertSeverities != nil {
		policy.AlwaysAlertSeverities = override.AlwaysAlertSeverities
	}
	if override.ActionableAlertSeverities != nil {
		policy.ActionableAlertSeverities = override.ActionableAlertSeverities
	}
	if override.DeduplicationWindowSeconds != nil {
		policy.DeduplicationWindowSeconds = *override.DeduplicationWindowSeconds
	}
	if override.RateLimitWindowSeconds != nil {
		policy.RateLimitWindowSeconds = *override.RateLimitWindowSeconds
	}
	if override.RateLimitCount != nil {
		policy.RateLimitCount = *override.RateLimitCount
	}
	return policy
}

func validatePolicy(policy domain.AlertPolicy) error {
	if policy.DeduplicationWindowSeconds <= 0 || policy.RateLimitWindowSeconds <= 0 || policy.RateLimitCount <= 0 {
		return errors.New("windows and rate limit must be positive")
	}
	valid := map[string]bool{"critical": true, "high": true, "medium": true, "low": true}
	for _, severities := range [][]string{policy.AlwaysAlertSeverities, policy.ActionableAlertSeverities} {
		for _, severity := range severities {
			if !valid[severity] {
				return fmt.Errorf("unknown severity %q", severity)
			}
		}
	}
	return nil
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

func floatFromEnv(key string, fallback float64) (float64, error) {
	raw := os.Getenv(key)
	if raw == "" {
		return fallback, nil
	}
	val, err := strconv.ParseFloat(raw, 64)
	if err != nil || val <= 0 {
		return 0, fmt.Errorf("%s must be a positive float", key)
	}
	return val, nil
}

