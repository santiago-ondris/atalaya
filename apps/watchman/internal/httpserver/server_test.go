package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/auth"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
)

type stubDatabase struct{ err error }

func (database stubDatabase) Ping(context.Context) error { return database.err }
func (database stubDatabase) ListEvents(context.Context, store.EventFilter) ([]store.EventSummary, error) {
	return nil, database.err
}
func (database stubDatabase) ListEventPage(context.Context, store.EventFilter) (store.EventPage, error) {
	return store.EventPage{}, database.err
}
func (database stubDatabase) ListIntegrations(context.Context) ([]store.IntegrationStatus, error) {
	return nil, database.err
}
func (database stubDatabase) Event(context.Context, string) (store.EventDetail, error) {
	return store.EventDetail{}, pgx.ErrNoRows
}
func (database stubDatabase) ListDailyReports(context.Context, int) ([]domain.DailyReport, error) {
	return nil, database.err
}
func (database stubDatabase) ListErrorGroups(context.Context, string, string, int) ([]store.ErrorGroupSummary, error) {
	return nil, database.err
}
func (database stubDatabase) CreateIncident(context.Context, string, []string) (store.Incident, error) {
	return store.Incident{}, database.err
}
func (database stubDatabase) ListIncidents(context.Context, store.IncidentFilter) ([]store.Incident, error) {
	return nil, database.err
}
func (database stubDatabase) Incident(context.Context, string) (store.Incident, error) {
	return store.Incident{}, pgx.ErrNoRows
}
func (database stubDatabase) AddIncidentNote(context.Context, string, string) error {
	return database.err
}
func (database stubDatabase) ChangeIncidentStatus(context.Context, string, string, string) error {
	return database.err
}
func (database stubDatabase) AddIncidentGroup(context.Context, string, string) error {
	return database.err
}
func (database stubDatabase) RemoveIncidentGroup(context.Context, string, string) error {
	return database.err
}
func (database stubDatabase) SaveDeployment(context.Context, store.DeploymentInput) (store.Deployment, bool, error) {
	return store.Deployment{}, false, database.err
}
func (database stubDatabase) ListDeployments(context.Context, string, string, time.Time, time.Time) ([]store.Deployment, error) {
	return nil, database.err
}
func (database stubDatabase) OperationsTimeline(context.Context, string, string, time.Time, time.Time, time.Duration) (store.OperationsTimeline, error) {
	return store.OperationsTimeline{}, database.err
}

type recordingDatabase struct {
	stubDatabase
	filter store.EventFilter
}

type deploymentDatabase struct {
	stubDatabase
	input store.DeploymentInput
}

func (database *deploymentDatabase) SaveDeployment(_ context.Context, input store.DeploymentInput) (store.Deployment, bool, error) {
	database.input = input
	return store.Deployment{ID: "deployment-1", Application: input.Application}, true, nil
}

func (database *recordingDatabase) ListEvents(_ context.Context, filter store.EventFilter) ([]store.EventSummary, error) {
	database.filter = filter
	return nil, nil
}

func TestHealth(t *testing.T) {
	server := newTestServer(stubDatabase{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)

	server.httpServer.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, response.Code)
	}
	if response.Header().Get(correlationIDHeader) == "" {
		t.Fatal("expected a generated correlation ID")
	}
}

func TestPreservesValidCorrelationID(t *testing.T) {
	server := newTestServer(stubDatabase{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/health", nil)
	correlationID := "8f6f0961-6de7-47af-9ab7-0ad4b82e18d8"
	request.Header.Set(correlationIDHeader, correlationID)

	server.httpServer.Handler.ServeHTTP(response, request)

	if actual := response.Header().Get(correlationIDHeader); actual != correlationID {
		t.Fatalf("expected correlation ID %q, got %q", correlationID, actual)
	}
}

func TestReadyWhenDatabaseIsUnavailable(t *testing.T) {
	server := newTestServer(stubDatabase{err: errors.New("database unavailable")})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/ready", nil)

	server.httpServer.Handler.ServeHTTP(response, request)

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status %d, got %d", http.StatusServiceUnavailable, response.Code)
	}
}

func TestListEventsAcceptsApplicationAndComponentFilters(t *testing.T) {
	database := &recordingDatabase{}
	server := newTestServer(database)
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/internal/events?limit=25&application=farmami&component=frontend", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "test-token"})
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	if database.filter != (store.EventFilter{Limit: 25, Application: "farmami", Component: "frontend"}) {
		t.Fatalf("unexpected filter: %#v", database.filter)
	}
}

func TestListEventsRejectsInvalidFilter(t *testing.T) {
	server := newTestServer(stubDatabase{})
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/internal/events?application=not-valid", nil)
	request.AddCookie(&http.Cookie{Name: sessionCookie, Value: "test-token"})
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", response.Code)
	}
}

func TestDeploymentHookRequiresBearerToken(t *testing.T) {
	database := &deploymentDatabase{}
	server := newTestServer(database)
	server.ConfigureDeploymentHooks("expected-token", "")
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodPost, "/hooks/v1/deployments", strings.NewReader(`{"application":"prensap"}`))
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", response.Code)
	}
}

func TestDeploymentHookNormalizesGitHubAction(t *testing.T) {
	database := &deploymentDatabase{}
	server := newTestServer(database)
	server.ConfigureDeploymentHooks("expected-token", "")
	response := httptest.NewRecorder()
	body := `{"application":"prensap","component":"frontend","environment":"production","external_id":"run-1","commit_sha":"abc123","deployed_at":"2026-08-11T12:00:00Z"}`
	request := httptest.NewRequest(http.MethodPost, "/hooks/v1/deployments", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer expected-token")
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	if database.input.Provider != "github_actions" || database.input.ExternalID != "run-1" {
		t.Fatalf("unexpected deployment: %#v", database.input)
	}
}

func TestRailwayRedeployUsesDeploymentIDWhenCommitIsMissing(t *testing.T) {
	database := &deploymentDatabase{}
	server := newTestServer(database)
	server.ConfigureDeploymentHooks("", "railway-secret")
	response := httptest.NewRecorder()
	body := `{"type":"Deployment.deployed","timestamp":"2026-08-11T12:00:00Z","details":{"status":"SUCCESS"},"resource":{"environment":{"name":"production"},"deployment":{"id":"railway-deploy-1"}}}`
	request := httptest.NewRequest(http.MethodPost, "/hooks/v1/deployments/railway/prensap/backend?token=railway-secret", strings.NewReader(body))
	server.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", response.Code, response.Body.String())
	}
	if database.input.ExternalID != "railway-deploy-1" || database.input.Version != "railway-deployment-railway-deploy-1" {
		t.Fatalf("unexpected deployment fallback: %#v", database.input)
	}
}

func newTestServer(database Database) *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(":0", database, auth.New(testSessionStore{}, "$argon2id$v=19$m=65536,t=1,p=1$c2FsdA$WnJdFRKGLdUVZUvrlcIHxA", time.Hour), logger, time.Second, false)
}

type testSessionStore struct{}

func (testSessionStore) CreateSession(context.Context, []byte, time.Time) error { return nil }
func (testSessionStore) SessionValid(context.Context, []byte, time.Time) (bool, error) {
	return true, nil
}
func (testSessionStore) RevokeSession(context.Context, []byte, time.Time) error { return nil }
