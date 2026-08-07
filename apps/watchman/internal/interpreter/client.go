package interpreter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Error struct {
	StatusCode int
	Retryable  bool
}

func (err *Error) Error() string { return fmt.Sprintf("interpreter returned HTTP %d", err.StatusCode) }

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string, timeout time.Duration, client *http.Client) *Client {
	if client == nil {
		client = &http.Client{Timeout: timeout}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: client}
}

func (client *Client) Interpret(ctx context.Context, job domain.InterpretationJob) (domain.Interpretation, error) {
	payload := map[string]any{
		"prompt_version": job.PromptVersion,
		"event": map[string]any{
			"id": job.EventID, "source": job.Source, "source_event_id": job.SourceEventID,
			"application": job.Application, "environment": job.Environment, "occurred_at": job.OccurredAt,
			"error_type": job.ErrorType, "message": job.Message, "stack_trace": nullable(job.StackTrace),
			"release": nullable(job.Release), "fingerprint": job.Fingerprint, "metadata": job.Metadata,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return domain.Interpretation{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, client.baseURL+"/v1/interpretations", bytes.NewReader(body))
	if err != nil {
		return domain.Interpretation{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-ID", uuid.NewString())
	response, err := client.http.Do(req)
	if err != nil {
		return domain.Interpretation{}, &Error{StatusCode: 0, Retryable: true}
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
		retryable := response.StatusCode == 408 || response.StatusCode == 429 || response.StatusCode == 503 || response.StatusCode == 504
		return domain.Interpretation{}, &Error{StatusCode: response.StatusCode, Retryable: retryable}
	}
	var result domain.Interpretation
	decoder := json.NewDecoder(io.LimitReader(response.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result); err != nil {
		return domain.Interpretation{}, fmt.Errorf("decode interpretation: %w", err)
	}
	if result.Summary == "" || result.Explanation == "" || result.Model == "" || result.PromptVersion == "" || result.Usage.TotalTokens != result.Usage.InputTokens+result.Usage.OutputTokens {
		return domain.Interpretation{}, fmt.Errorf("invalid interpretation response")
	}
	return result, nil
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}
