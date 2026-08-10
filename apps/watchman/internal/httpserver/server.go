package httpserver

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/auth"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
)

const correlationIDHeader = "X-Correlation-ID"

type correlationIDContextKey struct{}

type Database interface {
	Ping(context.Context) error
	ListEvents(context.Context, store.EventFilter) ([]store.EventSummary, error)
	ListEventPage(context.Context, store.EventFilter) (store.EventPage, error)
	ListIntegrations(context.Context) ([]store.IntegrationStatus, error)
	Event(context.Context, string) (store.EventDetail, error)
}

type Server struct {
	httpServer       *http.Server
	database         Database
	logger           *slog.Logger
	readinessTimeout time.Duration
	auth             *auth.Service
	cookieSecure     bool
}

func New(address string, database Database, authService *auth.Service, logger *slog.Logger, readinessTimeout time.Duration, cookieSecure bool) *Server {
	server := &Server{database: database, auth: authService, logger: logger, readinessTimeout: readinessTimeout, cookieSecure: cookieSecure}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", server.health)
	mux.HandleFunc("GET /ready", server.ready)
	mux.Handle("GET /internal/events", server.private(http.HandlerFunc(server.listEvents)))
	mux.Handle("GET /internal/events/{id}", server.private(http.HandlerFunc(server.eventDetail)))
	mux.Handle("GET /internal/integrations", server.private(http.HandlerFunc(server.listIntegrations)))
	mux.HandleFunc("POST /api/v1/session", server.login)
	mux.HandleFunc("DELETE /api/v1/session", server.logout)
	mux.Handle("GET /api/v1/session", server.private(http.HandlerFunc(server.session)))
	mux.Handle("GET /api/v1/overview", server.private(http.HandlerFunc(server.overview)))
	mux.Handle("GET /api/v1/events", server.private(http.HandlerFunc(server.listPublicEvents)))
	mux.Handle("GET /api/v1/events/{id}", server.private(http.HandlerFunc(server.eventDetail)))
	mux.Handle("GET /api/v1/integrations", server.private(http.HandlerFunc(server.listIntegrations)))

	server.httpServer = &http.Server{
		Addr:              address,
		Handler:           correlation(logger, mux),
		ReadHeaderTimeout: 5 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server
}

const sessionCookie = "atalaya_session"

func (s *Server) login(w http.ResponseWriter, request *http.Request) {
	var body struct {
		Password string `json:"password"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, request.Body, 4096))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&body); err != nil || body.Password == "" {
		writeJSON(w, 400, map[string]string{"error": "invalid request"})
		return
	}
	token, expiresAt, err := s.auth.Login(request.Context(), body.Password, time.Now())
	if errors.Is(err, auth.ErrInvalidCredentials) {
		time.Sleep(250 * time.Millisecond)
		writeJSON(w, 401, map[string]string{"error": "invalid credentials"})
		return
	}
	if err != nil {
		s.logger.Error("failed to create session", "error", err)
		writeJSON(w, 500, map[string]string{"error": "failed to create session"})
		return
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: token, Path: "/", Expires: expiresAt, HttpOnly: true, Secure: s.cookieSecure, SameSite: http.SameSiteLaxMode})
	writeJSON(w, 200, map[string]any{"authenticated": true, "expires_at": expiresAt})
}

func (s *Server) logout(w http.ResponseWriter, request *http.Request) {
	cookie, _ := request.Cookie(sessionCookie)
	if cookie != nil {
		_ = s.auth.Logout(request.Context(), cookie.Value, time.Now())
	}
	http.SetCookie(w, &http.Cookie{Name: sessionCookie, Value: "", Path: "/", MaxAge: -1, HttpOnly: true, Secure: s.cookieSecure, SameSite: http.SameSiteLaxMode})
	w.WriteHeader(204)
}
func (s *Server) session(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]bool{"authenticated": true})
}
func (s *Server) private(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(sessionCookie)
		if err != nil {
			writeJSON(w, 401, map[string]string{"error": "authentication required"})
			return
		}
		valid, err := s.auth.Valid(r.Context(), cookie.Value, time.Now())
		if err != nil {
			writeJSON(w, 500, map[string]string{"error": "session validation failed"})
			return
		}
		if !valid {
			writeJSON(w, 401, map[string]string{"error": "authentication required"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) overview(w http.ResponseWriter, r *http.Request) {
	integrations, err := s.database.ListIntegrations(r.Context())
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to load overview"})
		return
	}
	page, err := s.database.ListEventPage(r.Context(), store.EventFilter{Limit: 5})
	if err != nil {
		writeJSON(w, 500, map[string]string{"error": "failed to load overview"})
		return
	}
	type appStatus struct {
		Slug          string     `json:"slug"`
		DisplayName   string     `json:"display_name"`
		Status        string     `json:"status"`
		LastSuccessAt *time.Time `json:"last_success_at,omitempty"`
	}
	names := []appStatus{{Slug: "farmami", DisplayName: "Farmami"}, {Slug: "wheels_house", DisplayName: "Wheels House"}, {Slug: "prensap", DisplayName: "Prensap"}, {Slug: "notizap", DisplayName: "Notizap"}}
	for index := range names {
		names[index].Status = "unconfigured"
		for _, item := range integrations {
			if item.Application == names[index].Slug {
				if names[index].Status == "unconfigured" || item.Status != "healthy" {
					names[index].Status = item.Status
				}
				if item.LastSuccessAt != nil && (names[index].LastSuccessAt == nil || item.LastSuccessAt.After(*names[index].LastSuccessAt)) {
					names[index].LastSuccessAt = item.LastSuccessAt
				}
			}
		}
	}
	writeJSON(w, 200, map[string]any{"applications": names, "integrations": integrations, "recent_events": page.Items, "event_total": page.Total, "generated_at": time.Now()})
}

func (s *Server) listPublicEvents(w http.ResponseWriter, r *http.Request) {
	limit, offset := 25, 0
	var err error
	q := r.URL.Query()
	if raw := q.Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
		if err != nil || limit < 1 || limit > 100 {
			writeJSON(w, 400, map[string]string{"error": "invalid limit"})
			return
		}
	}
	if raw := q.Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
		if err != nil || offset < 0 {
			writeJSON(w, 400, map[string]string{"error": "invalid offset"})
			return
		}
	}
	filter := store.EventFilter{Limit: limit, Offset: offset, Application: q.Get("application"), Severity: q.Get("severity"), State: q.Get("state")}
	if filter.Application != "" && !slugPattern.MatchString(filter.Application) {
		writeJSON(w, 400, map[string]string{"error": "invalid application"})
		return
	}
	if filter.Severity != "" && !contains([]string{"critical", "high", "medium", "low", "pending"}, filter.Severity) {
		writeJSON(w, 400, map[string]string{"error": "invalid severity"})
		return
	}
	if filter.State != "" && !contains([]string{"actionable", "noise", "pending"}, filter.State) {
		writeJSON(w, 400, map[string]string{"error": "invalid state"})
		return
	}
	if period := q.Get("period"); period != "" {
		d, e := time.ParseDuration(period)
		if e != nil || d <= 0 || d > 2160*time.Hour {
			writeJSON(w, 400, map[string]string{"error": "invalid period"})
			return
		}
		since := time.Now().Add(-d)
		filter.Since = &since
	}
	page, err := s.database.ListEventPage(r.Context(), filter)
	if err != nil {
		s.logger.Error("failed to list events", "error", err)
		writeJSON(w, 500, map[string]string{"error": "failed to list events"})
		return
	}
	writeJSON(w, 200, map[string]any{"events": page.Items, "total": page.Total, "limit": limit, "offset": offset})
}
func contains(values []string, value string) bool {
	for _, v := range values {
		if strings.EqualFold(v, value) {
			return true
		}
	}
	return false
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
