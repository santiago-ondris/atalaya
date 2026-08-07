package poller

import (
	"context"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Store interface {
	Checkpoint(context.Context, uuid.UUID) (domain.Cursor, error)
	ImportBatch(context.Context, uuid.UUID, domain.EventBatch) (int, error)
	RecordAttempt(context.Context, uuid.UUID, error) error
}

type Poller struct {
	source        domain.ErrorSource
	store         Store
	integrationID uuid.UUID
	interval      time.Duration
	logger        *slog.Logger
}

func New(source domain.ErrorSource, store Store, integrationID uuid.UUID, interval time.Duration, logger *slog.Logger) *Poller {
	return &Poller{source: source, store: store, integrationID: integrationID, interval: interval, logger: logger}
}

func (poller *Poller) Run(ctx context.Context) {
	poller.poll(ctx)
	ticker := time.NewTicker(poller.interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			poller.poll(ctx)
		}
	}
}

func (poller *Poller) poll(ctx context.Context) {
	cursor, err := poller.store.Checkpoint(ctx, poller.integrationID)
	if err != nil {
		poller.failed(ctx, "load checkpoint", err)
		return
	}
	batch, err := poller.source.FetchEvents(ctx, cursor)
	if err != nil {
		poller.failed(ctx, "fetch sentry events", err)
		return
	}
	imported, err := poller.store.ImportBatch(ctx, poller.integrationID, batch)
	if err != nil {
		poller.failed(ctx, "persist sentry events", err)
		return
	}
	poller.logger.Info("sentry poll completed", "received", len(batch.Events), "imported", imported)
}

func (poller *Poller) failed(ctx context.Context, operation string, err error) {
	poller.logger.Error("sentry poll failed", "operation", operation, "error", err)
	if recordErr := poller.store.RecordAttempt(ctx, poller.integrationID, err); recordErr != nil {
		poller.logger.Error("failed to record poll attempt", "error", recordErr)
	}
}
