package interpreter

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

func TestClientSendsContractAndParsesInterpretation(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path != "/v1/interpretations" || request.Header.Get("X-Correlation-ID") == "" {
			t.Fatalf("unexpected request: %s", request.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		if body["prompt_version"] != "error-analysis-v1" {
			t.Fatalf("unexpected prompt version: %v", body["prompt_version"])
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader(`{"summary":"Resumen","explanation":"Explicación","severity":"medium","actionable":true,"suggested_actions":["Validar"],"model":"fixture/model","prompt_version":"error-analysis-v1","usage":{"input_tokens":10,"output_tokens":4,"total_tokens":14},"estimated_cost_usd":0.0001,"latency_ms":20}`)), Header: make(http.Header)}, nil
	})}

	client := NewClient("http://interpreter", time.Second, httpClient)
	result, err := client.Interpret(context.Background(), domain.InterpretationJob{
		EventID: "018f47a8-7b2a-7a68-aeb3-2fcb95ea1031", Source: "sentry", SourceEventID: "event-1",
		Application: "prensap", Environment: "production", OccurredAt: time.Now(), ErrorType: "TypeError",
		Message: "undefined", Fingerprint: "typeerror:load", PromptVersion: "error-analysis-v1", Metadata: map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Summary != "Resumen" || result.Usage.TotalTokens != 14 {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestClientClassifiesServerFailureAsRetryable(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusServiceUnavailable, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}

	_, err := NewClient("http://interpreter", time.Second, httpClient).Interpret(context.Background(), domain.InterpretationJob{})
	providerErr, ok := err.(*Error)
	if !ok || !providerErr.Retryable {
		t.Fatalf("expected retryable provider error, got %v", err)
	}
}

func TestClientClassifiesInvalidProviderResponseAsPermanent(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(_ *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})}

	_, err := NewClient("http://interpreter", time.Second, httpClient).Interpret(context.Background(), domain.InterpretationJob{})
	providerErr, ok := err.(*Error)
	if !ok || providerErr.Retryable {
		t.Fatalf("expected permanent provider error, got %v", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (function roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return function(request)
}
