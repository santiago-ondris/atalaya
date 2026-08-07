package sentry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

const maxResponseBytes = 10 << 20

type Client struct {
	baseURL, organization, project, token string
	httpClient                            *http.Client
}

func NewClient(baseURL, organization, project, token string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{strings.TrimRight(baseURL, "/"), organization, project, token, httpClient}
}

func (client *Client) FetchEvents(ctx context.Context, cursor domain.Cursor) (domain.EventBatch, error) {
	endpoint := fmt.Sprintf("%s/api/0/projects/%s/%s/events/", client.baseURL,
		url.PathEscape(client.organization), url.PathEscape(client.project))
	query := url.Values{"full": {"true"}, "query": {"event.type:error"}}
	if cursor.Value != "" {
		query.Set("cursor", cursor.Value)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint+"?"+query.Encode(), nil)
	if err != nil {
		return domain.EventBatch{}, fmt.Errorf("build sentry request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+client.token)
	req.Header.Set("Accept", "application/json")

	response, err := client.httpClient.Do(req)
	if err != nil {
		return domain.EventBatch{}, fmt.Errorf("request sentry events: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return domain.EventBatch{}, fmt.Errorf("sentry returned %s: %s", response.Status, sanitizeText(string(body)))
	}

	var payload []apiEvent
	decoder := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes))
	if err := decoder.Decode(&payload); err != nil {
		return domain.EventBatch{}, fmt.Errorf("decode sentry events: %w", err)
	}
	events := make([]domain.Event, 0, len(payload))
	for _, item := range payload {
		event, err := normalize(item)
		if err != nil {
			return domain.EventBatch{}, err
		}
		events = append(events, event)
	}
	return domain.EventBatch{Events: events, NextCursor: domain.Cursor{Value: nextCursor(response.Header.Get("Link"))}}, nil
}

type apiEvent struct {
	EventID     string          `json:"eventID"`
	DateCreated time.Time       `json:"dateCreated"`
	Title       string          `json:"title"`
	Message     string          `json:"message"`
	Culprit     string          `json:"culprit"`
	Environment string          `json:"environment"`
	Platform    string          `json:"platform"`
	Release     json.RawMessage `json:"release"`
	Entries     []apiEntry      `json:"entries"`
	Tags        []apiTag        `json:"tags"`
}

type apiEntry struct {
	Type string `json:"type"`
	Data struct {
		Values []apiException `json:"values"`
	} `json:"data"`
}
type apiException struct {
	Type       string `json:"type"`
	Value      string `json:"value"`
	Stacktrace struct {
		Frames []apiFrame `json:"frames"`
	} `json:"stacktrace"`
}
type apiFrame struct {
	Filename string `json:"filename"`
	Function string `json:"function"`
	LineNo   int    `json:"lineNo"`
	ColNo    int    `json:"colNo"`
	InApp    bool   `json:"inApp"`
}
type apiTag struct{ Key, Value string }

func normalize(item apiEvent) (domain.Event, error) {
	if item.EventID == "" || item.DateCreated.IsZero() {
		return domain.Event{}, errors.New("sentry event is missing eventID or dateCreated")
	}
	errorType, message, stack := exceptionDetails(item)
	if errorType == "" {
		errorType = "Error"
	}
	if message == "" {
		message = item.Message
	}
	if message == "" {
		message = item.Title
	}
	message = sanitizeText(message)
	environment := item.Environment
	if environment == "" {
		environment = tagValue(item.Tags, "environment", "unknown")
	}
	culprit := sanitizeText(item.Culprit)
	fingerprint := errorType + ":" + culprit
	if item.Culprit == "" {
		fingerprint = errorType + ":" + message
	}
	return domain.Event{
		SourceEventID: item.EventID, Environment: environment, Fingerprint: fingerprint,
		ErrorType: errorType, Message: message, StackTrace: sanitizeText(stack),
		Release: releaseVersion(item.Release), OccurredAt: item.DateCreated,
		Metadata: map[string]any{"platform": item.Platform, "culprit": culprit},
	}, nil
}

func exceptionDetails(item apiEvent) (string, string, string) {
	for _, entry := range item.Entries {
		if entry.Type != "exception" || len(entry.Data.Values) == 0 {
			continue
		}
		exception := entry.Data.Values[len(entry.Data.Values)-1]
		lines := make([]string, 0, len(exception.Stacktrace.Frames))
		for _, frame := range exception.Stacktrace.Frames {
			lines = append(lines, fmt.Sprintf("%s:%d:%d in %s", frame.Filename, frame.LineNo, frame.ColNo, frame.Function))
		}
		return exception.Type, exception.Value, strings.Join(lines, "\n")
	}
	return "", "", ""
}

func releaseVersion(raw json.RawMessage) string {
	var value string
	if json.Unmarshal(raw, &value) == nil {
		return value
	}
	var object struct {
		Version string `json:"version"`
	}
	_ = json.Unmarshal(raw, &object)
	return object.Version
}

func tagValue(tags []apiTag, key, fallback string) string {
	for _, tag := range tags {
		if tag.Key == key && tag.Value != "" {
			return tag.Value
		}
	}
	return fallback
}

func nextCursor(link string) string {
	for _, part := range strings.Split(link, ",") {
		if !strings.Contains(part, `rel="next"`) || !strings.Contains(part, `results="true"`) {
			continue
		}
		start, end := strings.Index(part, "<"), strings.Index(part, ">")
		if start < 0 || end <= start {
			continue
		}
		parsed, err := url.Parse(part[start+1 : end])
		if err == nil {
			return parsed.Query().Get("cursor")
		}
	}
	return ""
}
