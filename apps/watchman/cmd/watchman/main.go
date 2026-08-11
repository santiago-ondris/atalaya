package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	applicationinsights "github.com/santiago-ondris/atalaya/apps/watchman/internal/applicationinsights"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/auth"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/config"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/database"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/httpserver"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/interpreter"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/notification"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/poller"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/reporting"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/sentry"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/telegram"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	cfg, err := config.Load()
	if err != nil {
		logger.Error("invalid configuration", "error", err)
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	pool, err := database.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to connect to postgres", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	postgresStore := store.NewPostgres(pool)
	activitySources := map[string][]reporting.ActivitySource{}
	interpreterClient := interpreter.NewClient(cfg.Interpreter.URL, cfg.Interpreter.Timeout, nil)
	go interpreter.NewWorker(postgresStore, interpreterClient, cfg.Interpreter.WorkerID, cfg.Interpreter.PollInterval,
		cfg.Telegram.InterpreterCooldown, logger).Run(ctx)
	logger.Info("interpretation worker enabled", "interpreter_url", cfg.Interpreter.URL)
	var telegramClient *telegram.Client
	if cfg.Telegram.Enabled() {
		telegramClient = telegram.NewClient(cfg.Telegram.BotToken, cfg.Telegram.ChatID, cfg.Telegram.Timeout, nil)
		links := notification.Links{AtalayaBaseURL: cfg.Telegram.AtalayaPublicURL,
			SentryBaseURL: cfg.Sentry.BaseURL, SentryOrganization: cfg.Sentry.OrgSlug}
		go notification.NewWorker(postgresStore, telegramClient, cfg.Telegram.WorkerID,
			cfg.Telegram.PollInterval, links, logger).Run(ctx)
		logger.Info("Telegram notification worker enabled")
	} else {
		logger.Warn("Telegram notifications disabled", "reason", "TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are not configured")
	}
	if cfg.Sentry.Enabled() {
		enabled := 0
		requestGate := sentry.NewRequestGate()
		for _, application := range cfg.Sentry.Applications {
			for _, integration := range application.Integrations {
				runtime, ensureErr := postgresStore.EnsureSentryIntegration(ctx, store.SentryIntegrationSpec{
					Application: application.Slug, Component: integration.Component, DisplayName: integration.DisplayName,
					Organization: cfg.Sentry.OrgSlug, Project: integration.Project, Environments: integration.Environments,
					AlertPolicy: application.AlertPolicy,
				})
				if ensureErr != nil {
					logger.Error("failed to configure Sentry integration", "application", application.Slug,
						"component", integration.Component, "project", integration.Project, "error", ensureErr)
					continue
				}
				sentryClient := sentry.NewClient(cfg.Sentry.BaseURL, cfg.Sentry.OrgSlug, integration.Project,
					cfg.Sentry.AuthToken, integration.Environments, runtime.MonitoringStartedAt, requestGate, nil)
				go poller.New(sentryClient, postgresStore, runtime.ID, application.Slug, integration.Component, "sentry",
					cfg.PollInterval, logger).Run(ctx)
				activitySources[application.Slug] = append(activitySources[application.Slug], sentryClient)
				enabled++
				logger.Info("Sentry polling enabled", "application", application.Slug, "component", integration.Component,
					"project", integration.Project, "environments", integration.Environments, "interval", cfg.PollInterval)
			}
		}
		if enabled == 0 {
			logger.Warn("no Sentry pollers could be configured")
		}
	} else {
		logger.Warn("Sentry polling disabled", "reason", "SENTRY_AUTH_TOKEN is not configured")
	}
	if cfg.ApplicationInsights.Enabled() {
		azure := cfg.ApplicationInsights
		runtime, ensureErr := postgresStore.EnsureApplicationInsightsIntegration(ctx, store.ApplicationInsightsIntegrationSpec{
			Application: "notizap", Component: azure.Component, DisplayName: azure.DisplayName,
			WorkspaceID: azure.WorkspaceID, ResourceID: azure.ResourceID, Environment: azure.Environment,
			AlertPolicy: domainDefaultPolicy(cfg),
		})
		if ensureErr != nil {
			logger.Error("failed to configure Application Insights integration", "error", ensureErr)
		} else {
			client := applicationinsights.NewClient(azure.TokenURL, azure.LogsURL, azure.TenantID, azure.ClientID,
				azure.ClientSecret, azure.WorkspaceID, azure.ResourceID, azure.Environment, runtime.MonitoringStartedAt, azure.Overlap, nil)
			go poller.New(client, postgresStore, runtime.ID, "notizap", azure.Component, "application_insights", cfg.PollInterval, logger).Run(ctx)
			activitySources["notizap"] = append(activitySources["notizap"], client)
			logger.Info("Application Insights polling enabled", "application", "notizap", "component", azure.Component,
				"workspace_id", azure.WorkspaceID, "interval", cfg.PollInterval, "overlap", azure.Overlap)
		}
	} else {
		logger.Warn("Application Insights polling disabled", "reason", "Azure credentials and workspace are not configured")
	}

	reportSources := make([]reporting.Source, 0, len(activitySources))
	for application, providers := range activitySources {
		sourceName := "sentry"
		if application == "notizap" {
			sourceName = "application_insights"
		}
		reportSources = append(reportSources, reporting.Source{Application: application, Provider: reporting.SumSources{
			Sources: providers, Kind: "sessions", Source: sourceName,
		}})
	}
	reportScheduler, scheduleErr := reporting.NewScheduler(postgresStore, reportSources, cfg.Reporting.SchedulerInterval, logger)
	if scheduleErr != nil {
		logger.Error("failed to configure daily report scheduler", "error", scheduleErr)
	} else {
		go reportScheduler.Run(ctx)
		logger.Info("daily report scheduler enabled", "timezone", reporting.Timezone, "interval", cfg.Reporting.SchedulerInterval)
	}
	if telegramClient != nil {
		go reporting.NewWorker(postgresStore, telegramClient, cfg.Reporting.WorkerID, cfg.Telegram.PollInterval, logger).Run(ctx)
		logger.Info("daily report delivery worker enabled")
	}

	authService := auth.New(postgresStore, cfg.Auth.PasswordHash, cfg.Auth.SessionDuration)
	server := httpserver.New(cfg.HTTPAddress, databaseWithQueries{Pool: pool, Postgres: postgresStore}, authService, logger, cfg.ReadinessTimeout, cfg.Auth.CookieSecure)
	serverErrors := make(chan error, 1)
	go func() {
		logger.Info("watchman listening", "address", cfg.HTTPAddress, "environment", cfg.Environment)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		logger.Info("shutdown signal received")
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error("HTTP server stopped unexpectedly", "error", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("graceful shutdown failed", "error", err)
	}
}

func domainDefaultPolicy(cfg config.Config) domain.AlertPolicy {
	return domain.AlertPolicy{Enabled: true, AlwaysAlertSeverities: []string{"critical", "high"},
		ActionableAlertSeverities: []string{"medium"}, DeduplicationWindowSeconds: int(cfg.Telegram.DeduplicationWindow.Seconds()),
		RateLimitWindowSeconds: int(cfg.Telegram.RateLimitWindow.Seconds()), RateLimitCount: cfg.Telegram.RateLimitPerApp}
}

type databaseWithQueries struct {
	*pgxpool.Pool
	*store.Postgres
}
