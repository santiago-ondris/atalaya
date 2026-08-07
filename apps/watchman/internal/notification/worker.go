package notification

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/telegram"
)

type Store interface {
	ClaimNotificationJob(context.Context, string, time.Duration, int) (domain.NotificationJob, error)
	CompleteNotification(context.Context, domain.NotificationJob, time.Time, domain.DeliveryResult) error
	FailNotification(context.Context, domain.NotificationJob, time.Time, error, bool, int) error
}
type Sender interface {
	Send(context.Context, string) (int64, int, error)
}

type Worker struct {
	store                Store
	sender               Sender
	id                   string
	interval, rateWindow time.Duration
	rateLimit            int
	links                Links
	logger               *slog.Logger
}

func NewWorker(store Store, sender Sender, id string, interval, rateWindow time.Duration, rateLimit int, links Links, logger *slog.Logger) *Worker {
	return &Worker{store: store, sender: sender, id: id, interval: interval, rateWindow: rateWindow, rateLimit: rateLimit, links: links, logger: logger}
}
func (worker *Worker) Run(ctx context.Context) {
	ticker := time.NewTicker(worker.interval)
	defer ticker.Stop()
	for {
		worker.processOne(ctx)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
func (worker *Worker) processOne(ctx context.Context) {
	job, err := worker.store.ClaimNotificationJob(ctx, worker.id, worker.rateWindow, worker.rateLimit)
	if errors.Is(err, pgx.ErrNoRows) {
		return
	}
	if err != nil {
		worker.logger.Error("failed to claim notification job", "error", err)
		return
	}
	started := time.Now()
	messageID, status, err := worker.sender.Send(ctx, Format(job, worker.links))
	if err != nil {
		var sendErr *telegram.Error
		retryable := errors.As(err, &sendErr) && sendErr.Retryable
		if persistErr := worker.store.FailNotification(ctx, job, started, err, retryable, status); persistErr != nil {
			worker.logger.Error("failed to record Telegram failure", "job_id", job.ID, "error", persistErr)
		}
		worker.logger.Warn("Telegram delivery failed", "job_id", job.ID, "attempt", job.Attempts, "retryable", retryable, "error", err)
		return
	}
	if err := worker.store.CompleteNotification(ctx, job, started, domain.DeliveryResult{MessageID: messageID, HTTPStatus: status}); err != nil {
		worker.logger.Error("failed to record Telegram delivery", "job_id", job.ID, "error", err)
		return
	}
	worker.logger.Info("Telegram notification delivered", "job_id", job.ID, "kind", job.Kind, "telegram_message_id", messageID)
}
