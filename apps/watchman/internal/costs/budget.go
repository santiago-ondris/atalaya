package costs

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Store interface {
	GetCostSummary(ctx context.Context, monthlyBudgetUSD float64) (domain.CostSummary, error)
}

type TelegramSender interface {
	Send(ctx context.Context, text string) (int64, int, error)
}

type BudgetMonitor struct {
	store            Store
	sender           TelegramSender
	monthlyBudgetUSD float64
	logger           *slog.Logger
	mu               sync.Mutex
	lastAlertAt      time.Time
}

func NewBudgetMonitor(store Store, sender TelegramSender, monthlyBudgetUSD float64, logger *slog.Logger) *BudgetMonitor {
	return &BudgetMonitor{
		store:            store,
		sender:           sender,
		monthlyBudgetUSD: monthlyBudgetUSD,
		logger:           logger.With("component", "budget_monitor"),
	}
}

func (bm *BudgetMonitor) CheckBudget(ctx context.Context) (domain.CostSummary, error) {
	summary, err := bm.store.GetCostSummary(ctx, bm.monthlyBudgetUSD)
	if err != nil {
		return summary, err
	}

	if bm.monthlyBudgetUSD > 0 && summary.BudgetUsedPercent >= 80.0 {
		bm.maybeAlert(ctx, summary)
	}

	return summary, nil
}

func (bm *BudgetMonitor) maybeAlert(ctx context.Context, summary domain.CostSummary) {
	if bm.sender == nil {
		return
	}
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.lastAlertAt.IsZero() && time.Since(bm.lastAlertAt) < 24*time.Hour {
		return
	}

	msg := fmt.Sprintf("⚠️ <b>Alerta de Presupuesto LLM — Atalaya</b>\n\n"+
		"Gasto mensual: <b>$%.4f USD</b> / $%.2f USD (%.1f%% utilizado)\n"+
		"Tokens totales: %d\nSolitidudes: %d\n\n"+
		"<i>Sugerencia: Revisar el desglose de consumo por aplicación en Command Center.</i>",
		summary.MonthlyCostUSD, summary.MonthlyBudgetUSD, summary.BudgetUsedPercent,
		summary.TotalTokens, summary.TotalRequests)

	_, _, err := bm.sender.Send(ctx, msg)
	if err != nil {
		bm.logger.Error("failed to send budget alert Telegram message", "error", err)
	} else {
		bm.lastAlertAt = time.Now()
		bm.logger.Info("sent budget alert Telegram message", "used_percent", summary.BudgetUsedPercent)
	}
}
