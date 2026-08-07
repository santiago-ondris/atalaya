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
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/config"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/database"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/httpserver"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/poller"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/sentry"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
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
	if cfg.Sentry.Enabled() {
		integrationID, err := postgresStore.EnsureSentryIntegration(ctx, "prensap", cfg.Sentry.OrgSlug, cfg.Sentry.ProjectSlug)
		if err != nil {
			logger.Error("failed to configure Prensap Sentry integration", "error", err)
			os.Exit(1)
		}
		sentryClient := sentry.NewClient(cfg.Sentry.BaseURL, cfg.Sentry.OrgSlug, cfg.Sentry.ProjectSlug, cfg.Sentry.AuthToken, nil)
		go poller.New(sentryClient, postgresStore, integrationID, cfg.PollInterval, logger).Run(ctx)
		logger.Info("Prensap Sentry polling enabled", "interval", cfg.PollInterval)
	} else {
		logger.Warn("Prensap Sentry polling disabled", "reason", "SENTRY_AUTH_TOKEN is not configured")
	}

	server := httpserver.New(cfg.HTTPAddress, databaseWithQueries{Pool: pool, Postgres: postgresStore}, logger, cfg.ReadinessTimeout)
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

type databaseWithQueries struct {
	*pgxpool.Pool
	*store.Postgres
}
