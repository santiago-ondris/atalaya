package reporting

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/telegram"
)

type schedulerStore struct {
	ensured   int
	saved     int
	finalized int
	report    domain.DailyReport
}

func (store *schedulerStore) EnsureDailyReport(context.Context, string, string, time.Time, time.Time) (domain.DailyReport, bool, error) {
	store.ensured++
	return store.report, true, nil
}
func (store *schedulerStore) SaveDailyActivity(context.Context, string, string, domain.ActivityMetric, error) error {
	store.saved++
	return nil
}
func (store *schedulerStore) FinalizeDailyReport(context.Context, string) error {
	store.finalized++
	return nil
}
func (*schedulerStore) ClaimDailyReport(context.Context, string, time.Time) (domain.DailyReport, error) {
	return domain.DailyReport{}, errors.New("unused")
}
func (*schedulerStore) CompleteDailyReport(context.Context, domain.DailyReport, time.Time, domain.DeliveryResult) error {
	return nil
}
func (*schedulerStore) FailDailyReport(context.Context, domain.DailyReport, time.Time, error, bool, int) error {
	return nil
}

type activitySource struct{}

func (activitySource) FetchActivity(context.Context, time.Time, time.Time) (domain.ActivityMetric, error) {
	return domain.ActivityMetric{Count: 3, Kind: "sessions", Source: "fixture"}, nil
}

type failingDeliveryStore struct {
	report    domain.DailyReport
	failed    bool
	retryable bool
}

func (store *failingDeliveryStore) EnsureDailyReport(context.Context, string, string, time.Time, time.Time) (domain.DailyReport, bool, error) {
	return domain.DailyReport{}, false, errors.New("unused")
}
func (*failingDeliveryStore) SaveDailyActivity(context.Context, string, string, domain.ActivityMetric, error) error {
	return errors.New("unused")
}
func (*failingDeliveryStore) FinalizeDailyReport(context.Context, string) error {
	return errors.New("unused")
}
func (store *failingDeliveryStore) ClaimDailyReport(context.Context, string, time.Time) (domain.DailyReport, error) {
	return store.report, nil
}
func (*failingDeliveryStore) CompleteDailyReport(context.Context, domain.DailyReport, time.Time, domain.DeliveryResult) error {
	return errors.New("unexpected completion")
}
func (store *failingDeliveryStore) FailDailyReport(_ context.Context, _ domain.DailyReport, _ time.Time, _ error, retryable bool, _ int) error {
	store.failed, store.retryable = true, retryable
	return nil
}

type failingSender struct{}

func (failingSender) Send(context.Context, string) (int64, int, error) {
	return 0, 503, &telegram.Error{StatusCode: 503, Retryable: true, Message: "fixture unavailable"}
}

func TestSchedulerWaitsUntil20Argentina(t *testing.T) {
	store := &schedulerStore{report: domain.DailyReport{ID: "report"}}
	scheduler, err := NewScheduler(store, []Source{{Application: "notizap", Provider: activitySource{}}}, time.Minute, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	scheduler.RunOnce(context.Background(), time.Date(2026, 8, 10, 22, 59, 0, 0, time.UTC)) // 19:59 ARG
	if store.ensured != 0 {
		t.Fatal("report created before 20:00 ARG")
	}
	scheduler.RunOnce(context.Background(), time.Date(2026, 8, 10, 23, 0, 0, 0, time.UTC))
	if store.ensured != 1 || store.saved != 1 || store.finalized != 1 {
		t.Fatalf("unexpected calls: %#v", store)
	}
}

func TestFormatMarksUnavailableActivity(t *testing.T) {
	report := domain.DailyReport{Date: "2026-08-10", Applications: []domain.DailyReportApplication{{DisplayName: "Notizap", ActivityStatus: "unavailable", SeverityCounts: map[string]int64{}}}}
	message := Format(report)
	if message == "" || !containsText(message, "no disponible") {
		t.Fatalf("unexpected message: %s", message)
	}
}

func TestWorkerPersistsRetryableTelegramFailure(t *testing.T) {
	store := &failingDeliveryStore{report: domain.DailyReport{ID: "report", Date: "2026-08-10", Attempts: 1}}
	worker := NewWorker(store, failingSender{}, "test-worker", time.Second, slog.New(slog.NewTextHandler(io.Discard, nil)))
	worker.processOne(context.Background())
	if !store.failed || !store.retryable {
		t.Fatalf("expected retryable failure, got failed=%t retryable=%t", store.failed, store.retryable)
	}
}

func containsText(value, fragment string) bool {
	for i := 0; i+len(fragment) <= len(value); i++ {
		if value[i:i+len(fragment)] == fragment {
			return true
		}
	}
	return false
}
