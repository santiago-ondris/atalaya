package httpserver

import (
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/store"
)

func decodeBody(w http.ResponseWriter, r *http.Request, target any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid request"})
		return false
	}
	return true
}

func operationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		writeJSON(w, 404, map[string]string{"error": "resource not found"})
	case errors.Is(err, store.ErrConflict):
		writeJSON(w, 409, map[string]string{"error": err.Error()})
	case errors.Is(err, store.ErrInvalid):
		writeJSON(w, 400, map[string]string{"error": "invalid operation"})
	default:
		writeJSON(w, 500, map[string]string{"error": "operation failed"})
	}
}

func (s *Server) listErrorGroups(w http.ResponseWriter, r *http.Request) {
	limit := 25
	if raw := r.URL.Query().Get("limit"); raw != "" {
		parsed, e := strconv.Atoi(raw)
		if e != nil || parsed < 1 || parsed > 100 {
			writeJSON(w, 400, map[string]string{"error": "invalid limit"})
			return
		}
		limit = parsed
	}
	items, err := s.database.ListErrorGroups(r.Context(), r.URL.Query().Get("application"), strings.TrimSpace(r.URL.Query().Get("q")), limit)
	if err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"groups": items})
}

func (s *Server) createIncident(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Title    string   `json:"title"`
		GroupIDs []string `json:"error_group_ids"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	for _, id := range body.GroupIDs {
		if _, e := uuid.Parse(id); e != nil {
			writeJSON(w, 400, map[string]string{"error": "invalid error group id"})
			return
		}
	}
	item, err := s.database.CreateIncident(r.Context(), body.Title, body.GroupIDs)
	if err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 201, item)
}

func (s *Server) listIncidents(w http.ResponseWriter, r *http.Request) {
	limit, offset := 50, 0
	var err error
	if raw := r.URL.Query().Get("limit"); raw != "" {
		limit, err = strconv.Atoi(raw)
	}
	if err != nil || limit < 1 || limit > 100 {
		writeJSON(w, 400, map[string]string{"error": "invalid limit"})
		return
	}
	if raw := r.URL.Query().Get("offset"); raw != "" {
		offset, err = strconv.Atoi(raw)
	}
	if err != nil || offset < 0 {
		writeJSON(w, 400, map[string]string{"error": "invalid offset"})
		return
	}
	status := r.URL.Query().Get("status")
	if !store.ValidateIncidentStatus(status) {
		writeJSON(w, 400, map[string]string{"error": "invalid status"})
		return
	}
	items, err := s.database.ListIncidents(r.Context(), store.IncidentFilter{Application: r.URL.Query().Get("application"), Status: status, Limit: limit, Offset: offset})
	if err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"incidents": items})
}

func validPathUUID(w http.ResponseWriter, r *http.Request, name string) (string, bool) {
	id := r.PathValue(name)
	if _, err := uuid.Parse(id); err != nil {
		writeJSON(w, 400, map[string]string{"error": "invalid id"})
		return "", false
	}
	return id, true
}
func (s *Server) incidentDetail(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathUUID(w, r, "id")
	if !ok {
		return
	}
	item, err := s.database.Incident(r.Context(), id)
	if err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 200, item)
}
func (s *Server) addIncidentNote(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathUUID(w, r, "id")
	if !ok {
		return
	}
	var body struct {
		Body string `json:"body"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if err := s.database.AddIncidentNote(r.Context(), id, body.Body); err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 201, map[string]bool{"created": true})
}
func (s *Server) changeIncidentStatus(w http.ResponseWriter, r *http.Request) {
	id, ok := validPathUUID(w, r, "id")
	if !ok {
		return
	}
	var body struct {
		Status string `json:"status"`
		Note   string `json:"note"`
	}
	if !decodeBody(w, r, &body) {
		return
	}
	if err := s.database.ChangeIncidentStatus(r.Context(), id, body.Status, body.Note); err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 200, map[string]bool{"updated": true})
}
func (s *Server) addIncidentGroup(w http.ResponseWriter, r *http.Request) {
	s.changeIncidentGroup(w, r, true)
}
func (s *Server) removeIncidentGroup(w http.ResponseWriter, r *http.Request) {
	s.changeIncidentGroup(w, r, false)
}
func (s *Server) changeIncidentGroup(w http.ResponseWriter, r *http.Request, add bool) {
	id, ok := validPathUUID(w, r, "id")
	if !ok {
		return
	}
	gid, ok := validPathUUID(w, r, "groupId")
	if !ok {
		return
	}
	var err error
	if add {
		err = s.database.AddIncidentGroup(r.Context(), id, gid)
	} else {
		err = s.database.RemoveIncidentGroup(r.Context(), id, gid)
	}
	if err != nil {
		operationError(w, err)
		return
	}
	w.WriteHeader(204)
}

type deploymentRequest struct {
	Application string    `json:"application"`
	Component   string    `json:"component"`
	Environment string    `json:"environment"`
	ExternalID  string    `json:"external_id"`
	Version     string    `json:"version"`
	CommitSHA   string    `json:"commit_sha"`
	CommitURL   string    `json:"commit_url"`
	Actor       string    `json:"actor"`
	SourceURL   string    `json:"source_url"`
	DeployedAt  time.Time `json:"deployed_at"`
}

func (body deploymentRequest) input(provider string) store.DeploymentInput {
	return store.DeploymentInput{Application: body.Application, Component: body.Component, Environment: body.Environment, Provider: provider, ExternalID: body.ExternalID, Version: body.Version, CommitSHA: body.CommitSHA, CommitURL: body.CommitURL, Actor: body.Actor, SourceURL: body.SourceURL, DeployedAt: body.DeployedAt}
}
func (s *Server) saveDeployment(w http.ResponseWriter, r *http.Request, in store.DeploymentInput) {
	item, created, err := s.database.SaveDeployment(r.Context(), in)
	if err != nil {
		operationError(w, err)
		return
	}
	status := 200
	if created {
		status = 201
	}
	writeJSON(w, status, item)
}
func (s *Server) createManualDeployment(w http.ResponseWriter, r *http.Request) {
	var body deploymentRequest
	if !decodeBody(w, r, &body) {
		return
	}
	body.Environment = "production"
	body.ExternalID = "manual:" + uuid.NewString()
	s.saveDeployment(w, r, body.input("manual"))
}
func (s *Server) ingestDeployment(w http.ResponseWriter, r *http.Request) {
	if s.deployToken == "" || !secureEqual(strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "), s.deployToken) {
		writeJSON(w, 401, map[string]string{"error": "invalid deployment token"})
		return
	}
	var body deploymentRequest
	if !decodeBody(w, r, &body) {
		return
	}
	s.saveDeployment(w, r, body.input("github_actions"))
}
func secureEqual(got, want string) bool {
	return len(got) == len(want) && subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func (s *Server) ingestRailwayDeployment(w http.ResponseWriter, r *http.Request) {
	if s.railwayToken == "" || !secureEqual(r.PathValue("secret"), s.railwayToken) {
		writeJSON(w, 401, map[string]string{"error": "invalid railway token"})
		return
	}
	var payload struct {
		Type      string                                                               `json:"type"`
		Timestamp time.Time                                                            `json:"timestamp"`
		Details   struct{ ID, Status, CommitHash, CommitAuthor, CommitMessage string } `json:"details"`
		Resource  struct {
			Deployment struct {
				ID string `json:"id"`
			} `json:"deployment"`
			Environment struct {
				Name string `json:"name"`
			} `json:"environment"`
		} `json:"resource"`
	}
	if !decodeBody(w, r, &payload) {
		return
	}
	if !strings.EqualFold(payload.Details.Status, "success") && !strings.Contains(strings.ToLower(payload.Type), "success") {
		w.WriteHeader(204)
		return
	}
	if !strings.EqualFold(payload.Resource.Environment.Name, "production") {
		w.WriteHeader(204)
		return
	}
	external := payload.Resource.Deployment.ID
	if external == "" {
		external = payload.Details.ID
	}
	s.saveDeployment(w, r, store.DeploymentInput{Application: r.PathValue("application"), Component: r.PathValue("component"), Environment: "production", Provider: "railway", ExternalID: external, Version: payload.Details.CommitMessage, CommitSHA: payload.Details.CommitHash, Actor: payload.Details.CommitAuthor, DeployedAt: payload.Timestamp})
}

func (s *Server) listDeployments(w http.ResponseWriter, r *http.Request) {
	from, to, ok := timelineRange(w, r)
	if !ok {
		return
	}
	items, err := s.database.ListDeployments(r.Context(), r.URL.Query().Get("application"), r.URL.Query().Get("component"), from, to)
	if err != nil {
		operationError(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"deployments": items})
}
func timelineRange(w http.ResponseWriter, r *http.Request) (time.Time, time.Time, bool) {
	period := r.URL.Query().Get("period")
	if period == "" {
		period = "168h"
	}
	duration, err := time.ParseDuration(period)
	if err != nil || !(duration == 24*time.Hour || duration == 168*time.Hour || duration == 720*time.Hour) {
		writeJSON(w, 400, map[string]string{"error": "period must be 24h, 168h or 720h"})
		return time.Time{}, time.Time{}, false
	}
	to := time.Now().UTC().Truncate(time.Hour)
	return to.Add(-duration), to, true
}
func (s *Server) operationsTimeline(w http.ResponseWriter, r *http.Request) {
	application := r.URL.Query().Get("application")
	if !slugPattern.MatchString(application) {
		writeJSON(w, 400, map[string]string{"error": "application is required"})
		return
	}
	from, to, ok := timelineRange(w, r)
	if !ok {
		return
	}
	bucket := time.Hour
	if to.Sub(from) > 168*time.Hour {
		bucket = 24 * time.Hour
	}
	item, err := s.database.OperationsTimeline(r.Context(), application, r.URL.Query().Get("component"), from, to, bucket)
	if err != nil {
		s.logger.Error("failed to load operations timeline", "error", err)
		operationError(w, fmt.Errorf("timeline: %w", err))
		return
	}
	writeJSON(w, 200, map[string]any{"application": application, "from": from, "to": to, "bucket_seconds": int(bucket.Seconds()), "timeline": item})
}
