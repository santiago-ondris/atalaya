package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
)

const correlationIDHeader = "X-Correlation-ID"

type correlationIDContextKey struct{}

type Database interface {
	Ping(context.Context) error
	ListEvents(context.Context, store.EventFilter) ([]store.EventSummary, error)
	ListIntegrations(context.Context) ([]store.IntegrationStatus, error)
	Event(context.Context, string) (store.EventDetail, error)
}

type Server struct {
	httpServer       *http.Server
	database         Database
	logger           *slog.Logger
	readinessTimeout time.Duration
}

func New(address string, database Database, logger *slog.Logger, readinessTimeout time.Duration) *Server {
	server := &Server{database: database, logger: logger, readinessTimeout: readinessTimeout}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("GET /ready", server.ready)
	mux.HandleFunc("GET /internal/events", server.listEvents)
	mux.HandleFunc("GET /internal/events/{id}", server.eventDetail)
	mux.HandleFunc("GET /internal/integrations", server.listIntegrations)

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           correlation(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
}

func (s *Server) listEvents(w http.ResponseWriter, request *http.Request) {
	limit := 50
	if raw := request.URL.Query().Get("limit"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 || parsed > 200 {
			writeJSON(w, http.StatusBadRequest, map[string]string{"error": "limit must be between 1 and 200"})
			return
		}
		limit = parsed
	}
	application, component := request.URL.Query().Get("application"), request.URL.Query().Get("component")
	if (application != "" && !slugPattern.MatchString(application)) || (component != "" && !slugPattern.MatchString(component)) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "application and component must be valid slugs"})
		return
	}
	events, err := s.database.ListEvents(request.Context(), store.EventFilter{Limit: limit, Application: application, Component: component})
	if err != nil {
		s.logger.Error("failed to list events", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list events"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"events": events})
}

var slugPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*$`)

func (s *Server) listIntegrations(w http.ResponseWriter, request *http.Request) {
	integrations, err := s.database.ListIntegrations(request.Context())
	if err != nil {
		s.logger.Error("failed to list integrations", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to list integrations"})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"integrations": integrations})
}

func (s *Server) eventDetail(w http.ResponseWriter, request *http.Request) {
	id := request.PathValue("id")
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid event id"})
		return
	}
	event, err := s.database.Event(request.Context(), id)
	if errors.Is(err, pgx.ErrNoRows) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "event not found"})
		return
	}
	if err != nil {
		s.logger.Error("failed to load event", "error", err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "failed to load event"})
		return
	}
	writeJSON(w, http.StatusOK, event)
}

func (s *Server) ListenAndServe() error { return s.httpServer.ListenAndServe() }

func (s *Server) Shutdown(ctx context.Context) error { return s.httpServer.Shutdown(ctx) }

func (s *Server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) ready(w http.ResponseWriter, request *http.Request) {
	ctx, cancel := context.WithTimeout(request.Context(), s.readinessTimeout)
	defer cancel()

	if err := s.database.Ping(ctx); err != nil {
		s.logger.Error("readiness check failed", "component", "postgres", "error", err)
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{
			"status": "unavailable", "dependency": "postgres",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		slog.Error("failed to encode HTTP response", "error", err)
	}
}

func correlation(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		correlationID := validCorrelationID(r.Header.Get(correlationIDHeader))
		w.Header().Set(correlationIDHeader, correlationID)
		ctx := context.WithValue(r.Context(), correlationIDContextKey{}, correlationID)
		next.ServeHTTP(w, r.WithContext(ctx))
		logger.Info("HTTP request", "method", r.Method, "path", r.URL.Path,
			"correlation_id", correlationID,
			"duration_ms", time.Since(startedAt).Milliseconds())
	})
}

func validCorrelationID(value string) string {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return uuid.NewString()
	}
	return parsed.String()
}

func CorrelationIDFromContext(ctx context.Context) string {
	correlationID, _ := ctx.Value(correlationIDContextKey{}).(string)
	return correlationID
}
