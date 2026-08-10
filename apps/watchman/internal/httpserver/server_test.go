package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/auth"
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

type recordingDatabase struct {
	stubDatabase
	filter store.EventFilter
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
