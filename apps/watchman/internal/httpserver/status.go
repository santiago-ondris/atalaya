package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
)

type statusDatabase interface {
	SetIncidentPublication(context.Context, string, bool, string, string) error
	PublicStatus(context.Context, time.Time) (store.PublicStatus, error)
}

func (s *Server) publicStatus(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	database, ok := s.database.(statusDatabase)
	if !ok {
		writeJSON(w, 503, map[string]string{"error": "status unavailable"})
		return
	}
	status, err := database.PublicStatus(r.Context(), now)
	if err != nil {
		s.logger.Error("failed to load public status", "error", err)
		writeJSON(w, 500, map[string]string{"error": "status unavailable"})
		return
	}
	global := "operational"
	for _, app := range status.Applications {
		if app.Status == "major_outage" {
			global = "major_outage"
			break
		}
		if app.Status == "unknown" && global != "degraded" {
			global = "unknown"
		}
		if app.Status == "degraded" {
			global = "degraded"
		}
	}
	w.Header().Set("Cache-Control", "public, max-age=30, stale-while-revalidate=30")
	writeJSON(w, 200, map[string]any{"status": global, "generated_at": now, "applications": status.Applications, "incidents": status.Incidents})
}

func (s *Server) changeIncidentPublication(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Published bool   `json:"published"`
		Title     string `json:"title"`
		Message   string `json:"message"`
	}
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&body) != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid request"})
		return
	}
	database, ok := s.database.(statusDatabase)
	if !ok {
		writeJSON(w, 503, map[string]string{"error": "publication unavailable"})
		return
	}
	if err := database.SetIncidentPublication(r.Context(), r.PathValue("id"), body.Published, body.Title, body.Message); err != nil {
		operationError(w, err)
		return
	}
	incident, err := s.database.Incident(r.Context(), r.PathValue("id"))
	if err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 200, incident)
}
