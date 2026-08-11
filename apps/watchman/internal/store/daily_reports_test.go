package store

import (
	"testing"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

func TestDailyReportRetryIsThirtyMinutesInsideSameDay(t *testing.T) {
	now := time.Date(2026, 8, 10, 20, 0, 0, 0, time.FixedZone("ART", -3*60*60))
	report := domain.DailyReport{Attempts: 1, PeriodEnd: time.Date(2026, 8, 11, 0, 0, 0, 0, now.Location())}
	status, outcome, next, expiredAt := dailyReportFailureState(now, report, true)
	if status != "pending" || outcome != "retryable_failure" || !next.Equal(now.Add(30*time.Minute)) || expiredAt != nil {
		t.Fatalf("unexpected retry state: %s %s %s %v", status, outcome, next, expiredAt)
	}
}

func TestDailyReportDoesNotRetryAcrossMidnight(t *testing.T) {
	location := time.FixedZone("ART", -3*60*60)
	now := time.Date(2026, 8, 10, 23, 45, 0, 0, location)
	report := domain.DailyReport{Attempts: 2, PeriodEnd: time.Date(2026, 8, 11, 0, 0, 0, 0, location)}
	status, outcome, _, expiredAt := dailyReportFailureState(now, report, true)
	if status != "expired" || outcome != "permanent_failure" || expiredAt == nil {
		t.Fatalf("unexpected expiry state: %s %s %v", status, outcome, expiredAt)
	}
}
