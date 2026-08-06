package httpserver

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
)

const correlationIDHeader = "X-Correlation-ID"

type correlationIDContextKey struct{}

type Database interface {
	Ping(context.Context) error
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

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           correlation(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
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
