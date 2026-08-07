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
