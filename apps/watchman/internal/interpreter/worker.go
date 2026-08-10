package interpreter

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Store interface {
	ClaimInterpretationJob(context.Context, string) (domain.InterpretationJob, error)
	CompleteInterpretation(context.Context, domain.InterpretationJob, domain.Interpretation) error
	FailInterpretationJob(context.Context, domain.InterpretationJob, error, bool, time.Duration) error
}

type Interpreter interface {
	Interpret(context.Context, domain.InterpretationJob) (domain.Interpretation, error)
}

type Worker struct {
	store         Store
	client        Interpreter
	id            string
	interval      time.Duration
	logger        *slog.Logger
	alertCooldown time.Duration
}

func NewWorker(store Store, client Interpreter, id string, interval, alertCooldown time.Duration, logger *slog.Logger) *Worker {
	return &Worker{store: store, client: client, id: id, interval: interval, alertCooldown: alertCooldown, logger: logger}
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
	job, err := worker.store.ClaimInterpretationJob(ctx, worker.id)
	if errors.Is(err, pgx.ErrNoRows) {
		return
	}
	if err != nil {
		worker.logger.Error("failed to claim interpretation job", "error", err)
		return
	}
	result, err := worker.client.Interpret(ctx, job)
	if err != nil {
		var providerErr *Error
		retryable := errors.As(err, &providerErr) && providerErr.Retryable
		if failErr := worker.store.FailInterpretationJob(ctx, job, err, retryable, worker.alertCooldown); failErr != nil {
			worker.logger.Error("failed to update interpretation job", "job_id", job.ID, "error", failErr)
		}
		worker.logger.Warn("interpretation failed", "job_id", job.ID, "attempt", job.Attempts, "retryable", retryable, "error", err)
		return
	}
	if err := worker.store.CompleteInterpretation(ctx, job, result); err != nil {
		worker.logger.Error("failed to persist interpretation", "job_id", job.ID, "error", err)
		if failErr := worker.store.FailInterpretationJob(ctx, job, err, true, worker.alertCooldown); failErr != nil {
			worker.logger.Error("failed to release interpretation job", "job_id", job.ID, "error", failErr)
		}
		return
	}
	worker.logger.Info("interpretation completed", "job_id", job.ID, "model", result.Model, "tokens", result.Usage.TotalTokens)
}
