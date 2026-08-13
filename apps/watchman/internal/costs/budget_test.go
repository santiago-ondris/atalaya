package costs

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type stubCostStore struct {
	summary domain.CostSummary
	err     error
}

func (s *stubCostStore) GetCostSummary(ctx context.Context, monthlyBudgetUSD float64) (domain.CostSummary, error) {
	s.summary.MonthlyBudgetUSD = monthlyBudgetUSD
	if monthlyBudgetUSD > 0 {
		s.summary.BudgetUsedPercent = (s.summary.MonthlyCostUSD / monthlyBudgetUSD) * 100.0
	}
	return s.summary, s.err
}

type stubTelegramSender struct {
	sentMessages []string
}

func (s *stubTelegramSender) Send(ctx context.Context, text string) (int64, int, error) {
	s.sentMessages = append(s.sentMessages, text)
	return 100, 200, nil
}

func TestBudgetMonitorAlertsWhenOverThreshold(t *testing.T) {
	store := &stubCostStore{
		summary: domain.CostSummary{
			MonthlyCostUSD: 4.50,
			TotalCostUSD:   12.00,
			TotalTokens:    50000,
			TotalRequests:  15,
		},
	}
	sender := &stubTelegramSender{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	monitor := NewBudgetMonitor(store, sender, 5.0, logger)

	summary, err := monitor.CheckBudget(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.BudgetUsedPercent != 90.0 {
		t.Fatalf("expected 90%% budget used, got %.1f%%", summary.BudgetUsedPercent)
	}

	if len(sender.sentMessages) != 1 {
		t.Fatalf("expected 1 Telegram alert sent, got %d", len(sender.sentMessages))
	}
}

func TestBudgetMonitorDoesNotAlertWhenUnderThreshold(t *testing.T) {
	store := &stubCostStore{
		summary: domain.CostSummary{
			MonthlyCostUSD: 2.00,
		},
	}
	sender := &stubTelegramSender{}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	monitor := NewBudgetMonitor(store, sender, 5.0, logger)

	summary, err := monitor.CheckBudget(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if summary.BudgetUsedPercent != 40.0 {
		t.Fatalf("expected 40%% budget used, got %.1f%%", summary.BudgetUsedPercent)
	}

	if len(sender.sentMessages) != 0 {
		t.Fatalf("expected 0 Telegram alerts sent, got %d", len(sender.sentMessages))
	}
}
