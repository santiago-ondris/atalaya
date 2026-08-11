package applicationinsights

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

const (
	maxResponseBytes = 10 << 20
	logsScope        = "https://api.loganalytics.io/.default"
)

// exceptionsQuery deliberately projects a stable provider-neutral result shape.
// Provider column names remain isolated in this package.
const exceptionsQueryTemplate = `AppExceptions
| where _ResourceId =~ %q
| project TimeGenerated,
          ItemId=tostring(column_ifexists("ItemId", "")),
          OperationId=tostring(column_ifexists("OperationId", "")),
          AppRoleName=tostring(column_ifexists("AppRoleName", "")),
          AppVersion=tostring(column_ifexists("AppVersion", "")),
          ExceptionType=tostring(column_ifexists("ExceptionType", "Exception")),
          OuterMessage=tostring(column_ifexists("OuterMessage", "")),
          InnermostMessage=tostring(column_ifexists("InnermostMessage", "")),
          Details=tostring(column_ifexists("Details", "")),
          Properties=tostring(column_ifexists("Properties", ""))
| order by TimeGenerated asc`

type Client struct {
	tokenBaseURL, logsBaseURL, tenantID, clientID, clientSecret string
	workspaceID, resourceID, environment                        string
	monitoringStartedAt                                         time.Time
	overlap                                                     time.Duration
	httpClient                                                  *http.Client
	mu                                                          sync.Mutex
	token                                                       string
	tokenExpiresAt                                              time.Time
}

func NewClient(tokenBaseURL, logsBaseURL, tenantID, clientID, clientSecret, workspaceID, resourceID, environment string,
	monitoringStartedAt time.Time, overlap time.Duration, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{tokenBaseURL: strings.TrimRight(tokenBaseURL, "/"), logsBaseURL: strings.TrimRight(logsBaseURL, "/"),
		tenantID: tenantID, clientID: clientID, clientSecret: clientSecret, workspaceID: workspaceID, resourceID: strings.ToLower(resourceID),
		environment: environment, monitoringStartedAt: monitoringStartedAt.UTC(), overlap: overlap, httpClient: httpClient}
}

func (client *Client) FetchEvents(ctx context.Context, cursor domain.Cursor) (domain.EventBatch, error) {
	end := time.Now().UTC()
	start := client.monitoringStartedAt
	if cursor.Value != "" {
		checkpoint, err := time.Parse(time.RFC3339Nano, cursor.Value)
		if err != nil {
			return domain.EventBatch{}, fmt.Errorf("decode Application Insights checkpoint: %w", err)
		}
		start = checkpoint.Add(-client.overlap)
		if start.Before(client.monitoringStartedAt) {
			start = client.monitoringStartedAt
		}
	}
	token, err := client.accessToken(ctx)
	if err != nil {
		return domain.EventBatch{}, err
	}
	query := fmt.Sprintf(exceptionsQueryTemplate, client.resourceID)
	payload, _ := json.Marshal(map[string]string{"query": query, "timespan": start.Format(time.RFC3339Nano) + "/" + end.Format(time.RFC3339Nano)})
	endpoint := fmt.Sprintf("%s/v1/workspaces/%s/query", client.logsBaseURL, url.PathEscape(client.workspaceID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return domain.EventBatch{}, fmt.Errorf("build Azure Logs request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(req)
	if err != nil {
		return domain.EventBatch{}, fmt.Errorf("query Azure Logs: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return domain.EventBatch{}, fmt.Errorf("Azure Logs returned %s: %s", response.Status, sanitize(string(body)))
	}
	var result queryResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(&result); err != nil {
		return domain.EventBatch{}, fmt.Errorf("decode Azure Logs response: %w", err)
	}
	events, err := normalizeTables(result.Tables, client.environment)
	if err != nil {
		return domain.EventBatch{}, err
	}
	return domain.EventBatch{Events: events, NextCursor: domain.Cursor{Value: end.Format(time.RFC3339Nano)}}, nil
}

func (client *Client) FetchActivity(ctx context.Context, start, end time.Time) (domain.ActivityMetric, error) {
	token, err := client.accessToken(ctx)
	if err != nil {
		return domain.ActivityMetric{}, err
	}
	query := fmt.Sprintf(`AppPageViews
| where _ResourceId =~ %q
| where AppRoleName == "notizap-frontend"
| where TimeGenerated >= datetime(%s) and TimeGenerated < datetime(%s)
| summarize Sessions=dcountif(SessionId, isnotempty(SessionId))`, client.resourceID,
		start.UTC().Format(time.RFC3339), end.UTC().Format(time.RFC3339))
	payload, _ := json.Marshal(map[string]string{"query": query})
	endpoint := fmt.Sprintf("%s/v1/workspaces/%s/query", client.logsBaseURL, url.PathEscape(client.workspaceID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return domain.ActivityMetric{}, fmt.Errorf("build Azure activity request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	response, err := client.httpClient.Do(req)
	if err != nil {
		return domain.ActivityMetric{}, fmt.Errorf("query Azure activity: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return domain.ActivityMetric{}, fmt.Errorf("Azure activity returned %s: %s", response.Status, sanitize(string(body)))
	}
	var result queryResponse
	if err := json.NewDecoder(io.LimitReader(response.Body, maxResponseBytes)).Decode(&result); err != nil {
		return domain.ActivityMetric{}, fmt.Errorf("decode Azure activity: %w", err)
	}
	if len(result.Tables) != 1 || len(result.Tables[0].Rows) != 1 || len(result.Tables[0].Rows[0]) != 1 {
		return domain.ActivityMetric{}, errors.New("Azure activity result has an unexpected shape")
	}
	value, ok := result.Tables[0].Rows[0][0].(float64)
	if !ok {
		return domain.ActivityMetric{}, errors.New("Azure activity session count is not numeric")
	}
	return domain.ActivityMetric{Count: int64(value), Kind: "sessions", Source: "application_insights"}, nil
}

func (client *Client) accessToken(ctx context.Context) (string, error) {
	client.mu.Lock()
	defer client.mu.Unlock()
	if client.token != "" && time.Now().Add(time.Minute).Before(client.tokenExpiresAt) {
		return client.token, nil
	}
	form := url.Values{"client_id": {client.clientID}, "client_secret": {client.clientSecret}, "scope": {logsScope}, "grant_type": {"client_credentials"}}
	endpoint := fmt.Sprintf("%s/%s/oauth2/v2.0/token", client.tokenBaseURL, url.PathEscape(client.tenantID))
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("build Azure token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request Azure access token: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return "", fmt.Errorf("Azure identity returned %s: %s", response.Status, sanitize(string(body)))
	}
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 1<<20)).Decode(&result); err != nil {
		return "", fmt.Errorf("decode Azure access token: %w", err)
	}
	if result.AccessToken == "" {
		return "", errors.New("Azure identity response omitted access_token")
	}
	client.token, client.tokenExpiresAt = result.AccessToken, time.Now().Add(time.Duration(result.ExpiresIn)*time.Second)
	return client.token, nil
}

type queryResponse struct {
	Tables []queryTable `json:"tables"`
}
type queryTable struct {
	Columns []queryColumn `json:"columns"`
	Rows    [][]any       `json:"rows"`
}
type queryColumn struct {
	Name string `json:"name"`
}

func normalizeTables(tables []queryTable, environment string) ([]domain.Event, error) {
	var events []domain.Event
	for _, table := range tables {
		indexes := map[string]int{}
		for index, column := range table.Columns {
			indexes[column.Name] = index
		}
		if _, ok := indexes["TimeGenerated"]; !ok {
			return nil, errors.New("Azure Logs result is missing TimeGenerated")
		}
		for _, row := range table.Rows {
			occurredAt, err := time.Parse(time.RFC3339Nano, stringValue(row, indexes, "TimeGenerated"))
			if err != nil {
				return nil, fmt.Errorf("decode Azure exception timestamp: %w", err)
			}
			errorType := firstNonEmpty(stringValue(row, indexes, "ExceptionType"), "Exception")
			message := firstNonEmpty(stringValue(row, indexes, "InnermostMessage"), stringValue(row, indexes, "OuterMessage"), errorType)
			operationID, itemID := stringValue(row, indexes, "OperationId"), stringValue(row, indexes, "ItemId")
			if itemID == "" {
				itemID = deterministicID(occurredAt.Format(time.RFC3339Nano), operationID, errorType, message)
			}
			role := stringValue(row, indexes, "AppRoleName")
			fingerprint := errorType + ":" + firstNonEmpty(role, operationID, message)
			events = append(events, domain.Event{SourceEventID: itemID, Environment: environment, Fingerprint: sanitize(fingerprint),
				ErrorType: sanitize(errorType), Message: sanitize(message), StackTrace: sanitize(stringValue(row, indexes, "Details")),
				Release: sanitize(stringValue(row, indexes, "AppVersion")), OccurredAt: occurredAt,
				Metadata: map[string]any{"operation_id": sanitize(operationID), "app_role_name": sanitize(role), "properties": sanitize(stringValue(row, indexes, "Properties"))}})
		}
	}
	return events, nil
}

func stringValue(row []any, indexes map[string]int, name string) string {
	index, ok := indexes[name]
	if !ok || index >= len(row) || row[index] == nil {
		return ""
	}
	if value, ok := row[index].(string); ok {
		return value
	}
	raw, _ := json.Marshal(row[index])
	return string(raw)
}
func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
func deterministicID(parts ...string) string {
	sum := sha256.Sum256([]byte(strings.Join(parts, "\x00")))
	return "generated-" + hex.EncodeToString(sum[:])
}
func sanitize(value string) string {
	replacements := []string{"client_secret", "authorization", "access_token", "password", "token"}
	lower := strings.ToLower(value)
	for _, key := range replacements {
		if strings.Contains(lower, key) {
			return "[REDACTED PROVIDER RESPONSE]"
		}
	}
	if len(value) > 32000 {
		return value[:32000] + "…[TRUNCATED]"
	}
	return value
}
