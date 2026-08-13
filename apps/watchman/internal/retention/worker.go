package retention

import (
	"context"
	"log/slog"
	"time"
)

type Store interface {
	PurgeOldEvents(ctx context.Context, retentionDays int) (int64, error)
}

type Worker struct {
	store         Store
	retentionDays int
	interval      time.Duration
	logger        *slog.Logger
}

func NewWorker(store Store, retentionDays int, interval time.Duration, logger *slog.Logger) *Worker {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	return &Worker{
		store:         store,
		retentionDays: retentionDays,
		interval:      interval,
		logger:        logger.With("component", "retention_worker"),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if w.retentionDays <= 0 {
		w.logger.Info("event retention worker disabled")
		return
	}
	w.logger.Info("starting event retention worker", "retention_days", w.retentionDays, "interval", w.interval)

	// Run initial cleanup
	w.purge(ctx)

	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("event retention worker stopped")
			return
		case <-ticker.C:
			w.purge(ctx)
		}
	}
}

func (w *Worker) purge(ctx context.Context) {
	purged, err := w.store.PurgeOldEvents(ctx, w.retentionDays)
	if err != nil {
		w.logger.Error("failed to purge old events", "error", err)
		return
	}
	if purged > 0 {
		w.logger.Info("purged old events", "count", purged, "retention_days", w.retentionDays)
	}
}
