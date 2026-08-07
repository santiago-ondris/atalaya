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
