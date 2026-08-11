package reporting

import (
	"context"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"strings"
	"time"
	_ "time/tzdata"

	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/telegram"
)

const Timezone = "America/Argentina/Buenos_Aires"

type ActivitySource interface {
	FetchActivity(context.Context, time.Time, time.Time) (domain.ActivityMetric, error)
}

type Source struct {
	Application string
	Provider    ActivitySource
}

type SumSources struct {
	Sources []ActivitySource
	Kind    string
	Source  string
}

func (sum SumSources) FetchActivity(ctx context.Context, start, end time.Time) (domain.ActivityMetric, error) {
	metric := domain.ActivityMetric{Kind: sum.Kind, Source: sum.Source}
	for _, source := range sum.Sources {
		item, err := source.FetchActivity(ctx, start, end)
		if err != nil {
			return metric, err
		}
		metric.Count += item.Count
	}
	return metric, nil
}

type Store interface {
	EnsureDailyReport(context.Context, string, string, time.Time, time.Time) (domain.DailyReport, bool, error)
	SaveDailyActivity(context.Context, string, string, domain.ActivityMetric, error) error
	FinalizeDailyReport(context.Context, string) error
	ClaimDailyReport(context.Context, string, time.Time) (domain.DailyReport, error)
	CompleteDailyReport(context.Context, domain.DailyReport, time.Time, domain.DeliveryResult) error
	FailDailyReport(context.Context, domain.DailyReport, time.Time, error, bool, int) error
}

type Sender interface {
	Send(context.Context, string) (int64, int, error)
}

type Scheduler struct {
	store    Store
	sources  []Source
	location *time.Location
	interval time.Duration
	logger   *slog.Logger
}

func NewScheduler(store Store, sources []Source, interval time.Duration, logger *slog.Logger) (*Scheduler, error) {
	location, err := time.LoadLocation(Timezone)
	if err != nil {
		return nil, fmt.Errorf("load report timezone: %w", err)
	}
	return &Scheduler{store: store, sources: sources, location: location, interval: interval, logger: logger}, nil
}

func (scheduler *Scheduler) Run(ctx context.Context) {
	ticker := time.NewTicker(scheduler.interval)
	defer ticker.Stop()
	for {
		scheduler.RunOnce(ctx, time.Now())
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (scheduler *Scheduler) RunOnce(ctx context.Context, now time.Time) {
	local := now.In(scheduler.location)
	if local.Hour() < 20 {
		return
	}
	start := time.Date(local.Year(), local.Month(), local.Day(), 0, 0, 0, 0, scheduler.location)
	end := start.AddDate(0, 0, 1)
	report, collect, err := scheduler.store.EnsureDailyReport(ctx, start.Format(time.DateOnly), Timezone, start, end)
	if err != nil {
		scheduler.logger.Error("failed to ensure daily report", "date", start.Format(time.DateOnly), "error", err)
		return
	}
	if !collect {
		return
	}
	for _, source := range scheduler.sources {
		metric, fetchErr := source.Provider.FetchActivity(ctx, start, now)
		if err := scheduler.store.SaveDailyActivity(ctx, report.ID, source.Application, metric, fetchErr); err != nil {
			scheduler.logger.Error("failed to save daily activity", "application", source.Application, "error", err)
			return
		}
		if fetchErr != nil {
			scheduler.logger.Warn("daily activity unavailable", "application", source.Application, "error", fetchErr)
		}
	}
	if err := scheduler.store.FinalizeDailyReport(ctx, report.ID); err != nil {
		scheduler.logger.Error("failed to finalize daily report", "report_id", report.ID, "error", err)
	}
}

type Worker struct {
	store    Store
	sender   Sender
	id       string
	interval time.Duration
	logger   *slog.Logger
}

func NewWorker(store Store, sender Sender, id string, interval time.Duration, logger *slog.Logger) *Worker {
	return &Worker{store: store, sender: sender, id: id, interval: interval, logger: logger}
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
	started := time.Now()
	report, err := worker.store.ClaimDailyReport(ctx, worker.id, started)
	if errors.Is(err, pgx.ErrNoRows) {
		return
	}
	if err != nil {
		worker.logger.Error("failed to claim daily report", "error", err)
		return
	}
	messageID, status, err := worker.sender.Send(ctx, Format(report))
	if err != nil {
		var sendErr *telegram.Error
		retryable := errors.As(err, &sendErr) && sendErr.Retryable
		if persistErr := worker.store.FailDailyReport(ctx, report, started, err, retryable, status); persistErr != nil {
			worker.logger.Error("failed to record daily report failure", "report_id", report.ID, "error", persistErr)
		}
		return
	}
	if err := worker.store.CompleteDailyReport(ctx, report, started, domain.DeliveryResult{MessageID: messageID, HTTPStatus: status}); err != nil {
		worker.logger.Error("failed to complete daily report", "report_id", report.ID, "error", err)
	}
}

func Format(report domain.DailyReport) string {
	lines := []string{"⚓ <b>Atalaya · Reporte diario</b>", html.EscapeString(report.Date) + " · 20:00 ARG", ""}
	var totalErrors, totalOccurrences int64
	for _, app := range report.Applications {
		totalErrors += app.ErrorCount
		totalOccurrences += app.OccurrenceCount
		activity := "no disponible"
		if app.ActivityCount != nil {
			activity = fmt.Sprintf("%d sesiones", *app.ActivityCount)
		}
		lines = append(lines,
			"<b>"+html.EscapeString(app.DisplayName)+"</b>",
			fmt.Sprintf("Actividad: %s · Errores: %d · Ocurrencias: %d", activity, app.ErrorCount, app.OccurrenceCount),
			fmt.Sprintf("Severidad: 🔴 %d · 🟠 %d · 🟡 %d · 🟢 %d · Accionables: %d",
				app.SeverityCounts["critical"], app.SeverityCounts["high"], app.SeverityCounts["medium"], app.SeverityCounts["low"], app.ActionableCount), "")
	}
	lines = append(lines, fmt.Sprintf("<b>Total</b>: %d errores · %d ocurrencias", totalErrors, totalOccurrences))
	return strings.Join(lines, "\n")
}
