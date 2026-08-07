package sentry

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

func TestFetchEventsNormalizesFixtureAndUsesCursor(t *testing.T) {
	var authorization, cursor string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		authorization = request.Header.Get("Authorization")
		cursor = request.URL.Query().Get("cursor")
		body := `[{
			"eventID":"fixture-event-001","dateCreated":"2026-08-06T14:30:00Z",
			"title":"TypeError","message":"","culprit":"loadArticle(articles.ts)",
			"platform":"javascript","environment":"production","release":{"version":"release-001"},
			"entries":[{"type":"exception","data":{"values":[{"type":"TypeError","value":"broken token=private-value",
			"stacktrace":{"frames":[{"filename":"articles.ts","function":"loadArticle","lineNo":42,"colNo":13}]}}]}}]
		}]`
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Header: http.Header{
			"Link": {`<https://sentry.example/api/0/events/?cursor=next-page>; rel="next"; results="true"`},
		}, Body: io.NopCloser(strings.NewReader(body)), Request: request}, nil
	})}

	client := NewClient("https://sentry.example", "org", "prensap", "test-token", httpClient)
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
	_, err := NewClient("https://sentry.example", "org", "project", "secret", httpClient).FetchEvents(context.Background(), domain.Cursor{})
	if err == nil || strings.Contains(err.Error(), "leaked-token") {
		t.Fatalf("expected sanitized error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
