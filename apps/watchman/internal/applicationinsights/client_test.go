package applicationinsights

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

func TestFetchEventsAuthenticatesQueriesAndNormalizes(t *testing.T) {
	var tokenCalls int
	var queryAuthorization, timespan, query string
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		switch {
		case strings.HasSuffix(request.URL.Path, "/oauth2/v2.0/token"):
			tokenCalls++
			body, _ := io.ReadAll(request.Body)
			if !strings.Contains(string(body), "scope=https%3A%2F%2Fapi.loganalytics.io%2F.default") || strings.Contains(string(body), "secret-value=") {
				t.Fatalf("unexpected token form %q", body)
			}
			return response(request, 200, `{"access_token":"access","expires_in":3600}`), nil
		case strings.HasSuffix(request.URL.Path, "/v1/workspaces/workspace-id/query"):
			queryAuthorization = request.Header.Get("Authorization")
			var payload map[string]string
			if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
				t.Fatal(err)
			}
			timespan = payload["timespan"]
			query = payload["query"]
			return response(request, 200, `{"tables":[{"columns":[
				{"name":"TimeGenerated"},{"name":"ItemId"},{"name":"OperationId"},{"name":"AppRoleName"},
				{"name":"AppVersion"},{"name":"ExceptionType"},{"name":"OuterMessage"},{"name":"InnermostMessage"},
				{"name":"Details"},{"name":"Properties"}],"rows":[
				["2026-08-10T14:30:00Z","item-1","operation-1","notizap-api","2.4.0","InvalidOperationException","outer","inner","stack",{"route":"/news"}]
			]}]}`), nil
		default:
			t.Fatalf("unexpected URL %s", request.URL)
			return nil, nil
		}
	})}
	started := time.Date(2026, 8, 10, 14, 0, 0, 0, time.UTC)
	client := NewClient("https://login.example", "https://logs.example", "tenant", "client", "secret-value",
		"workspace-id", "/subscriptions/sub/resourceGroups/rg/providers/microsoft.insights/components/notizap-insights", "production", started, 5*time.Minute, httpClient)
	batch, err := client.FetchEvents(context.Background(), domain.Cursor{Value: "2026-08-10T14:20:00Z"})
	if err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 1 || queryAuthorization != "Bearer access" {
		t.Fatalf("unexpected auth: calls=%d header=%q", tokenCalls, queryAuthorization)
	}
	if !strings.HasPrefix(timespan, "2026-08-10T14:15:00Z/") {
		t.Fatalf("expected overlap in timespan, got %q", timespan)
	}
	if !strings.Contains(query, `_ResourceId =~ "/subscriptions/sub/resourcegroups/rg/providers/microsoft.insights/components/notizap-insights"`) {
		t.Fatalf("expected resource-scoped query, got %q", query)
	}
	if len(batch.Events) != 1 {
		t.Fatalf("expected one event, got %d", len(batch.Events))
	}
	event := batch.Events[0]
	if event.SourceEventID != "item-1" || event.ErrorType != "InvalidOperationException" || event.Message != "inner" || event.Release != "2.4.0" {
		t.Fatalf("unexpected event: %#v", event)
	}
	if event.Environment != "production" || batch.NextCursor.Value == "" {
		t.Fatalf("unexpected checkpoint/environment: %#v", batch)
	}
	if _, err := client.FetchEvents(context.Background(), batch.NextCursor); err != nil {
		t.Fatal(err)
	}
	if tokenCalls != 1 {
		t.Fatalf("expected cached token, got %d token calls", tokenCalls)
	}
}

func TestNormalizeGeneratesStableIDWhenItemIDIsMissing(t *testing.T) {
	table := queryTable{Columns: []queryColumn{{Name: "TimeGenerated"}, {Name: "ExceptionType"}, {Name: "OuterMessage"}},
		Rows: [][]any{{"2026-08-10T14:30:00Z", "Error", "broken"}}}
	first, err := normalizeTables([]queryTable{table}, "production")
	if err != nil {
		t.Fatal(err)
	}
	second, err := normalizeTables([]queryTable{table}, "production")
	if err != nil {
		t.Fatal(err)
	}
	if first[0].SourceEventID == "" || first[0].SourceEventID != second[0].SourceEventID {
		t.Fatalf("expected stable generated ID")
	}
}

func TestProviderErrorsDoNotLeakSecrets(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		return response(request, 401, `{"error":"invalid_client","client_secret":"leaked"}`), nil
	})}
	client := NewClient("https://login.example", "https://logs.example", "tenant", "client", "secret", "workspace",
		"/subscriptions/sub/resourceGroups/rg/providers/microsoft.insights/components/notizap-insights", "production", time.Now(), time.Minute, httpClient)
	_, err := client.FetchEvents(context.Background(), domain.Cursor{})
	if err == nil || strings.Contains(err.Error(), "leaked") {
		t.Fatalf("expected sanitized error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
func response(request *http.Request, status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Status: http.StatusText(status), Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body)), Request: request}
}
