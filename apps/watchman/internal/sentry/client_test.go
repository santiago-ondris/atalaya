package sentry

import (
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

func TestFetchEventsNormalizesFixtureAndUsesCursor(t *testing.T) {
	var authorization, cursor, requestQuery string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		authorization = request.Header.Get("Authorization")
		cursor = request.URL.Query().Get("cursor")
		requestQuery = request.URL.Query().Get("query")
		body := `[{"eventID":"message-001","type":"default","dateCreated":"2026-08-06T14:31:00Z","title":"message"},{
			"eventID":"fixture-event-001","dateCreated":"2026-08-06T14:30:00Z",
			"type":"error",
			"title":"TypeError","message":"","culprit":"loadArticle(articles.ts)",
			"platform":"javascript","environment":"production","release":{"version":"release-001"},
			"entries":[{"type":"exception","data":{"values":[{"type":"TypeError","value":"broken token=private-value",
			"stacktrace":{"frames":[{"filename":"articles.ts","function":"loadArticle","lineNo":42,"colNo":13}]}}]}}]
		}]`
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{
			"Link": {`<https://sentry.example/api/0/events/?cursor=next-page>; rel="next"; results="true"`},
		}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}

	client := NewClient("https://sentry.example", "org", "prensap", "test-token", []string{"production"}, time.Time{}, nil, httpClient)
	batch, err := client.FetchEvents(context.Background(), domain.Cursor{Value: "current-page"})
	if err != nil {
		t.Fatal(err)
	}
	if authorization != "Bearer test-token" {
		t.Fatalf("unexpected authorization header %q", authorization)
	}
	if cursor != "current-page" {
		t.Fatalf("expected cursor, got %q", cursor)
	}
	if got := requestQuery; got != "" {
		t.Fatalf("expected no unsupported Sentry search query, got %q", got)
	}
	if len(batch.Events) != 1 {
		t.Fatalf("expected one event, got %d", len(batch.Events))
	}
	event := batch.Events[0]
	if event.ErrorType != "TypeError" || event.Environment != "production" || event.Release != "release-001" {
		t.Fatalf("unexpected normalized event: %#v", event)
	}
	if strings.Contains(event.Message, "private-value") || !strings.Contains(event.Message, "[REDACTED]") {
		t.Fatalf("expected sensitive value to be redacted, got %q", event.Message)
	}
	if batch.NextCursor.Value != "next-page" {
		t.Fatalf("unexpected next cursor %q", batch.NextCursor.Value)
	}
}

func TestFetchEventsDoesNotLeakTokenInProviderError(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusUnauthorized, Status: "401 Unauthorized", Header: make(http.Header),
			Body: io.NopCloser(strings.NewReader(`authorization: Bearer leaked-token`)), Request: request}, nil
	})}
	_, err := NewClient("https://sentry.example", "org", "project", "secret", nil, time.Time{}, nil, httpClient).FetchEvents(context.Background(), domain.Cursor{})
	if err == nil || strings.Contains(err.Error(), "leaked-token") {
		t.Fatalf("expected sanitized error, got %v", err)
	}
}

func TestFetchEventsFiltersEnvironmentAndStopsAtMonitoringStart(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		body := `[
			{"eventID":"new-production","type":"error","dateCreated":"2026-08-07T12:01:00Z","title":"new","environment":"production"},
			{"eventID":"new-development","type":"error","dateCreated":"2026-08-07T12:01:00Z","title":"dev","environment":"development"},
			{"eventID":"old-production","type":"error","dateCreated":"2026-08-07T11:59:00Z","title":"old","environment":"production"}
		]`
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{
			"Link": {`<https://sentry.example/events/?cursor=older>; rel="next"; results="true"`},
		}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}
	startedAt := time.Date(2026, 8, 7, 12, 0, 0, 0, time.UTC)
	client := NewClient("https://sentry.example", "org", "project", "token", []string{"production"}, startedAt, nil, httpClient)
	batch, err := client.FetchEvents(context.Background(), domain.Cursor{})
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Events) != 1 || batch.Events[0].SourceEventID != "new-production" {
		t.Fatalf("unexpected filtered events: %#v", batch.Events)
	}
	if batch.NextCursor.Value != "" {
		t.Fatalf("expected pagination to stop at monitoring boundary, got %q", batch.NextCursor.Value)
	}
	if batch.Events[0].Metadata["sentry_project"] != "project" {
		t.Fatalf("expected project metadata, got %#v", batch.Events[0].Metadata)
	}
}

func TestSharedRequestGateSerializesSentryCalls(t *testing.T) {
	var active, maximum atomic.Int32
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		current := active.Add(1)
		for {
			seen := maximum.Load()
			if current <= seen || maximum.CompareAndSwap(seen, current) {
				break
			}
		}
		time.Sleep(20 * time.Millisecond)
		active.Add(-1)
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: make(http.Header),
			Body: io.NopCloser(strings.NewReader(`[]`)), Request: request}, nil
	})}
	gate := NewRequestGate()
	clients := []*Client{
		NewClient("https://sentry.example", "org", "one", "token", nil, time.Time{}, gate, httpClient),
		NewClient("https://sentry.example", "org", "two", "token", nil, time.Time{}, gate, httpClient),
	}
	var wait sync.WaitGroup
	wait.Add(len(clients))
	for _, client := range clients {
		go func() { defer wait.Done(); _, _ = client.FetchEvents(context.Background(), domain.Cursor{}) }()
	}
	wait.Wait()
	if maximum.Load() != 1 {
		t.Fatalf("expected one concurrent Sentry request, got %d", maximum.Load())
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
