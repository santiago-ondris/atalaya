package meta

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Sender interface {
	Send(context.Context, string) (int64, int, error)
}
type Signal struct {
	Name       string     `json:"name"`
	Status     string     `json:"status"`
	Detail     string     `json:"detail,omitempty"`
	LastSeenAt *time.Time `json:"last_seen_at,omitempty"`
}
type Health struct {
	Status      string           `json:"status"`
	GeneratedAt time.Time        `json:"generated_at"`
	Signals     []Signal         `json:"signals"`
	Queues      map[string]int64 `json:"queues"`
}

func Heartbeat(ctx context.Context, pool *pgxpool.Pool, signal string) {
	write := func() {
		_, _ = pool.Exec(ctx, `INSERT INTO service_heartbeats(signal,last_seen_at)VALUES($1,now()) ON CONFLICT(signal)DO UPDATE SET last_seen_at=excluded.last_seen_at`, signal)
	}
	write()
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			write()
		}
	}
}

type Monitor struct {
	Pool         *pgxpool.Pool
	Sender       Sender
	PingURL      string
	PollInterval time.Duration
	Client       *http.Client
	Now          func() time.Time
}

func New(pool *pgxpool.Pool, sender Sender, pingURL string, pollInterval time.Duration) *Monitor {
	return &Monitor{Pool: pool, Sender: sender, PingURL: pingURL, PollInterval: pollInterval, Client: &http.Client{Timeout: 5 * time.Second}, Now: time.Now}
}
func (m *Monitor) Run(ctx context.Context) {
	m.cycle(ctx)
	t := time.NewTicker(time.Minute)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
			m.cycle(ctx)
		}
	}
}
func (m *Monitor) cycle(ctx context.Context) {
	health, err := m.Health(ctx)
	if err != nil {
		return
	}
	for _, s := range health.Signals {
		m.transition(ctx, s)
	}
	if m.PingURL != "" && m.Pool.Ping(ctx) == nil {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, m.PingURL, nil)
		if response, e := m.Client.Do(req); e == nil {
			response.Body.Close()
		}
	}
}
func (m *Monitor) transition(ctx context.Context, s Signal) {
	unhealthy := s.Status != "healthy"
	var changed bool
	err := m.Pool.QueryRow(ctx, `INSERT INTO health_signal_states(signal,unhealthy,detail,changed_at)VALUES($1,$2,$3,now()) ON CONFLICT(signal)DO UPDATE SET unhealthy=excluded.unhealthy,detail=excluded.detail,changed_at=CASE WHEN health_signal_states.unhealthy<>excluded.unhealthy THEN now() ELSE health_signal_states.changed_at END RETURNING (xmax=0 OR changed_at>now()-interval '2 seconds')`, s.Name, unhealthy, s.Detail).Scan(&changed)
	if err == nil && changed && m.Sender != nil {
		state := "recuperado"
		if unhealthy {
			state = "degradado"
		}
		_, _, _ = m.Sender.Send(ctx, fmt.Sprintf("Atalaya · salud interna\n%s: %s\n%s", s.Name, state, s.Detail))
	}
}

func (m *Monitor) Health(ctx context.Context) (Health, error) {
	now := m.Now().UTC()
	h := Health{Status: "healthy", GeneratedAt: now, Queues: map[string]int64{}, Signals: []Signal{}}
	rows, err := m.Pool.Query(ctx, `SELECT signal,last_seen_at FROM service_heartbeats ORDER BY signal`)
	if err != nil {
		return h, err
	}
	seen := map[string]bool{}
	for rows.Next() {
		var name string
		var at time.Time
		if err = rows.Scan(&name, &at); err != nil {
			rows.Close()
			return h, err
		}
		seen[name] = true
		status := "healthy"
		detail := ""
		if now.Sub(at) > 90*time.Second {
			status = "degraded"
			detail = "heartbeat vencido"
		}
		h.Signals = append(h.Signals, Signal{Name: name, Status: status, Detail: detail, LastSeenAt: &at})
	}
	rows.Close()
	for _, name := range []string{"interpreter_worker", "notification_worker", "report_worker", "health_checker", "internal_evaluator"} {
		if !seen[name] {
			h.Signals = append(h.Signals, Signal{Name: name, Status: "unknown", Detail: "sin heartbeat"})
		}
	}
	stale := int64(0)
	_ = m.Pool.QueryRow(ctx, `SELECT count(*) FROM integrations WHERE enabled AND (last_attempt_at IS NULL OR last_attempt_at < $1)`, now.Add(-3*m.PollInterval)).Scan(&stale)
	h.Queues["stale_pollers"] = stale
	for _, q := range []struct{ name, table string }{{"interpreter_pending", "interpretation_jobs"}, {"telegram_pending", "notification_jobs"}} {
		var count int64
		query := fmt.Sprintf("SELECT count(*) FROM %s WHERE status IN ('pending','processing')", q.table)
		if e := m.Pool.QueryRow(ctx, query).Scan(&count); e == nil {
			h.Queues[q.name] = count
		}
	}
	var oldInterpreter, failedInterpreter, oldTelegram int64
	_ = m.Pool.QueryRow(ctx, `SELECT count(*) FROM interpretation_jobs WHERE status='pending' AND created_at<$1`, now.Add(-5*time.Minute)).Scan(&oldInterpreter)
	_ = m.Pool.QueryRow(ctx, `SELECT count(*) FROM interpretation_jobs WHERE status='failed' AND updated_at>$1`, now.Add(-15*time.Minute)).Scan(&failedInterpreter)
	_ = m.Pool.QueryRow(ctx, `SELECT count(*) FROM notification_jobs WHERE status IN('pending','processing') AND created_at<$1`, now.Add(-5*time.Minute)).Scan(&oldTelegram)
	h.Signals = append(h.Signals, Signal{Name: "pollers", Status: state(stale > 0), Detail: detail(stale, "detenidos")}, Signal{Name: "interpreter_queue", Status: state(oldInterpreter > 0 || failedInterpreter >= 3 || h.Queues["interpreter_pending"] >= 20), Detail: fmt.Sprintf("pendientes=%d atrasados=%d fallos_15m=%d", h.Queues["interpreter_pending"], oldInterpreter, failedInterpreter)}, Signal{Name: "telegram_queue", Status: state(oldTelegram > 0 || h.Queues["telegram_pending"] >= 20), Detail: fmt.Sprintf("pendientes=%d atrasados=%d", h.Queues["telegram_pending"], oldTelegram)})
	for _, s := range h.Signals {
		if s.Status != "healthy" {
			h.Status = "degraded"
			break
		}
	}
	return h, nil
}
func state(b bool) string {
	if b {
		return "degraded"
	}
	return "healthy"
}
func detail(n int64, label string) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf("%d %s", n, strings.TrimSpace(label))
}
