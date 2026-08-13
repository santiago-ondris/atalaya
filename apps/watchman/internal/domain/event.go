package domain

import (
	"context"
	"time"
)

type Cursor struct {
	Value string `json:"value,omitempty"`
}

type Event struct {
	SourceEventID string
	Environment   string
	Fingerprint   string
	ErrorType     string
	Message       string
	StackTrace    string
	Release       string
	OccurredAt    time.Time
	Metadata      map[string]any
}

type EventBatch struct {
	Events     []Event
	NextCursor Cursor
}

type ErrorSource interface {
	FetchEvents(ctx context.Context, cursor Cursor) (EventBatch, error)
}

type AlertPolicy struct {
	Enabled                    bool     `json:"enabled"`
	AlwaysAlertSeverities      []string `json:"always_alert_severities"`
	ActionableAlertSeverities  []string `json:"actionable_alert_severities"`
	DeduplicationWindowSeconds int      `json:"deduplication_window_seconds"`
	RateLimitWindowSeconds     int      `json:"rate_limit_window_seconds"`
	RateLimitCount             int      `json:"rate_limit_count"`
}

func (policy AlertPolicy) Eligible(interpretation Interpretation) bool {
	if !policy.Enabled {
		return false
	}
	for _, severity := range policy.AlwaysAlertSeverities {
		if interpretation.Severity == severity {
			return true
		}
	}
	if interpretation.Actionable {
		for _, severity := range policy.ActionableAlertSeverities {
			if interpretation.Severity == severity {
				return true
			}
		}
	}
	return false
}

type InterpretationJob struct {
	ID            string
	EventID       string
	Source        string
	SourceEventID string
	Application   string
	Environment   string
	OccurredAt    time.Time
	ErrorType     string
	Message       string
	StackTrace    string
	Release       string
	Fingerprint   string
	Metadata      map[string]any
	PromptVersion string
	AlertPolicy   AlertPolicy
	Attempts      int
	MaxAttempts   int
}

type TokenUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

type Interpretation struct {
	Summary          string     `json:"summary"`
	Explanation      string     `json:"explanation"`
	Severity         string     `json:"severity"`
	Actionable       bool       `json:"actionable"`
	SuggestedActions []string   `json:"suggested_actions"`
	Model            string     `json:"model"`
	PromptVersion    string     `json:"prompt_version"`
	Usage            TokenUsage `json:"usage"`
	EstimatedCostUSD *float64   `json:"estimated_cost_usd"`
	LatencyMS        int        `json:"latency_ms"`
}

type NotificationJob struct {
	ID               string
	EventID          string
	Kind             string
	Application      string
	Component        string
	Source           string
	SourceEventID    string
	Environment      string
	ErrorType        string
	Release          string
	Summary          string
	Explanation      string
	Severity         string
	Actionable       bool
	SuggestedActions []string
	OccurrenceCount  int
	FirstOccurredAt  time.Time
	LastOccurredAt   time.Time
	Attempts         int
	MaxAttempts      int
}

type DeliveryResult struct {
	MessageID  int64
	HTTPStatus int
}

type ActivityMetric struct {
	Count  int64
	Kind   string
	Source string
}

type DailyReportApplication struct {
	Application     string           `json:"application"`
	DisplayName     string           `json:"display_name"`
	ActivityCount   *int64           `json:"activity_count,omitempty"`
	ActivityKind    string           `json:"activity_kind"`
	ActivitySource  string           `json:"activity_source"`
	ActivityStatus  string           `json:"activity_status"`
	ActivityError   string           `json:"activity_error,omitempty"`
	ErrorCount      int64            `json:"error_count"`
	OccurrenceCount int64            `json:"occurrence_count"`
	SeverityCounts  map[string]int64 `json:"severity_counts"`
	ActionableCount int64            `json:"actionable_count"`
}

type DailyReport struct {
	ID           string                   `json:"id"`
	Date         string                   `json:"date"`
	Timezone     string                   `json:"timezone"`
	PeriodStart  time.Time                `json:"period_start"`
	PeriodEnd    time.Time                `json:"period_end"`
	Status       string                   `json:"status"`
	Attempts     int                      `json:"attempts"`
	LastError    string                   `json:"last_error,omitempty"`
	CreatedAt    time.Time                `json:"created_at"`
	SentAt       *time.Time               `json:"sent_at,omitempty"`
	ExpiredAt    *time.Time               `json:"expired_at,omitempty"`
	Applications []DailyReportApplication `json:"applications"`
}

type ApplicationCostBreakdown struct {
	Application      string  `json:"application"`
	TotalTokens      int64   `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	RequestCount     int64   `json:"request_count"`
}

type ModelCostBreakdown struct {
	Model            string  `json:"model"`
	TotalTokens      int64   `json:"total_tokens"`
	EstimatedCostUSD float64 `json:"estimated_cost_usd"`
	RequestCount     int64   `json:"request_count"`
}

type CostSummary struct {
	TotalCostUSD      float64                    `json:"total_cost_usd"`
	MonthlyCostUSD    float64                    `json:"monthly_cost_usd"`
	MonthlyBudgetUSD  float64                    `json:"monthly_budget_usd"`
	BudgetUsedPercent float64                    `json:"budget_used_percent"`
	TotalTokens       int64                      `json:"total_tokens"`
	InputTokens       int64                      `json:"input_tokens"`
	OutputTokens      int64                      `json:"output_tokens"`
	TotalRequests     int64                      `json:"total_requests"`
	AverageLatencyMS  int64                      `json:"average_latency_ms"`
	ByApplication     []ApplicationCostBreakdown `json:"by_application"`
	ByModel           []ModelCostBreakdown       `json:"by_model"`
}

