package availability

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	Interval = time.Minute
	Timeout  = 5 * time.Second
)

type Target struct {
	ID, Application, DisplayName, Component, URL string
}

type catalogTarget struct {
	ID          string `json:"id"`
	Application string `json:"application"`
	DisplayName string `json:"display_name"`
	Component   string `json:"component"`
	URLEnv      string `json:"url_env"`
}

type catalog struct {
	Targets []catalogTarget `json:"targets"`
}

func LoadCatalog(path string) ([]Target, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c catalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	result := make([]Target, 0, len(c.Targets))
	for _, item := range c.Targets {
		t := Target{ID: item.ID, Application: item.Application, DisplayName: item.DisplayName, Component: item.Component, URL: strings.TrimSpace(os.Getenv(item.URLEnv))}
		if seen[t.ID] || t.ID == "" || t.Application == "" || (t.Component != "frontend" && t.Component != "backend") {
			return nil, fmt.Errorf("invalid availability target %q", t.ID)
		}
		if t.URL != "" {
			u, parseErr := url.ParseRequestURI(t.URL)
			if parseErr != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
				return nil, fmt.Errorf("invalid URL in %s", item.URLEnv)
			}
		}
		seen[t.ID] = true
		result = append(result, t)
	}
	if len(result) != 8 {
		return nil, fmt.Errorf("availability catalog must contain eight targets")
	}
	return result, nil
}

type Sender interface {
	Send(context.Context, string) (int64, int, error)
}

type Checker struct {
	pool    *pgxpool.Pool
	targets []Target
	client  *http.Client
	sender  Sender
	now     func() time.Time
}

func New(pool *pgxpool.Pool, targets []Target, sender Sender) *Checker {
	return &Checker{pool: pool, targets: targets, sender: sender, client: &http.Client{Timeout: Timeout}, now: time.Now}
}

func (c *Checker) Run(ctx context.Context) {
	c.Cycle(ctx)
	ticker := time.NewTicker(Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			c.Cycle(ctx)
		}
	}
}

func (c *Checker) Cycle(ctx context.Context) {
	apps := map[string]bool{}
	for _, target := range c.targets {
		apps[target.Application] = true
		_ = c.check(ctx, target)
	}
	now := c.now().UTC()
	for app := range apps {
		_, _ = c.pool.Exec(ctx, `INSERT INTO availability_snapshots(application,status,captured_at)
			SELECT $1, CASE WHEN count(*) FILTER(WHERE confirmed_status='unknown')>0 THEN 'unknown'
			WHEN count(*) FILTER(WHERE confirmed_status='up')=2 THEN 'operational'
			WHEN count(*) FILTER(WHERE confirmed_status='down')=2 THEN 'major_outage' ELSE 'degraded' END,$2
			FROM availability_targets WHERE application=$1`, app, now)
	}
}

func (c *Checker) check(ctx context.Context, target Target) error {
	now := c.now().UTC()
	enabled := target.URL != ""
	_, err := c.pool.Exec(ctx, `INSERT INTO availability_targets(id,application,display_name,component,url,enabled)
		VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT(id) DO UPDATE SET application=excluded.application,display_name=excluded.display_name,component=excluded.component,url=excluded.url,enabled=excluded.enabled`,
		target.ID, target.Application, target.DisplayName, target.Component, target.URL, enabled)
	if err != nil || !enabled {
		return err
	}
	started := c.now()
	request, _ := http.NewRequestWithContext(ctx, http.MethodGet, target.URL, nil)
	response, probeErr := c.client.Do(request)
	latency := int(c.now().Sub(started).Milliseconds())
	status := 0
	success := false
	errorText := ""
	if probeErr != nil {
		errorText = sanitizeError(probeErr)
	} else {
		status = response.StatusCode
		success = status >= 200 && status < 300
		if !success {
			errorText = fmt.Sprintf("HTTP %d", status)
		}
		response.Body.Close()
	}
	var oldStatus, newStatus string
	var changedAt *time.Time
	err = c.pool.QueryRow(ctx, `WITH previous AS (SELECT confirmed_status,status_changed_at FROM availability_targets WHERE id=$1 FOR UPDATE),
		updated AS (UPDATE availability_targets SET consecutive_successes=CASE WHEN $3 THEN consecutive_successes+1 ELSE 0 END,
		consecutive_failures=CASE WHEN $3 THEN 0 ELSE consecutive_failures+1 END,last_checked_at=$2,
		confirmed_status=CASE WHEN $3 AND consecutive_successes+1>=2 THEN 'up' WHEN NOT $3 AND consecutive_failures+1>=2 THEN 'down' ELSE confirmed_status END,
		status_changed_at=CASE WHEN ($3 AND consecutive_successes+1>=2 AND confirmed_status<>'up') OR (NOT $3 AND consecutive_failures+1>=2 AND confirmed_status<>'down') THEN $2 ELSE status_changed_at END
		WHERE id=$1 RETURNING confirmed_status,status_changed_at)
		SELECT previous.confirmed_status,updated.confirmed_status,previous.status_changed_at FROM previous,updated`, target.ID, now, success).Scan(&oldStatus, &newStatus, &changedAt)
	if err != nil {
		return err
	}
	var httpStatus any
	if status != 0 {
		httpStatus = status
	}
	_, err = c.pool.Exec(ctx, `INSERT INTO availability_probes(target_id,checked_at,succeeded,latency_ms,http_status,error)VALUES($1,$2,$3,$4,$5,NULLIF($6,''))`, target.ID, now, success, latency, httpStatus, errorText)
	if err == nil && oldStatus != newStatus && (newStatus == "down" || oldStatus != "unknown") && c.sender != nil {
		duration := "duración desconocida"
		if changedAt != nil {
			duration = now.Sub(*changedAt).Round(time.Second).String()
		}
		_, _, _ = c.sender.Send(ctx, fmt.Sprintf("Atalaya · disponibilidad\n%s / %s: %s\nDuración del estado anterior: %s", target.DisplayName, target.Component, newStatus, duration))
	}
	return err
}

func sanitizeError(err error) string {
	if err == nil {
		return ""
	}
	text := strings.ToLower(err.Error())
	for _, kind := range []string{"timeout", "deadline exceeded", "connection refused", "no such host", "tls"} {
		if strings.Contains(text, kind) {
			return kind
		}
	}
	return "request failed"
}

func Aggregate(componentStatuses []string) string {
	if len(componentStatuses) == 0 {
		return "unknown"
	}
	sort.Strings(componentStatuses)
	for _, status := range componentStatuses {
		if status == "unknown" {
			return "unknown"
		}
	}
	up := 0
	for _, status := range componentStatuses {
		if status == "up" {
			up++
		}
	}
	if up == len(componentStatuses) {
		return "operational"
	}
	if up == 0 {
		return "major_outage"
	}
	return "degraded"
}
