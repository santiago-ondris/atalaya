package store

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
)

var (
	ErrConflict = errors.New("operation conflicts with current state")
	ErrInvalid  = errors.New("invalid operation")
)

type ErrorGroupSummary struct {
	ID              string    `json:"id"`
	Application     string    `json:"application"`
	Component       string    `json:"component"`
	ErrorType       string    `json:"error_type"`
	Message         string    `json:"message"`
	OccurrenceCount int64     `json:"occurrence_count"`
	LastSeenAt      time.Time `json:"last_seen_at"`
}

type IncidentEntry struct {
	ID         string    `json:"id"`
	Kind       string    `json:"kind"`
	Body       string    `json:"body,omitempty"`
	FromStatus string    `json:"from_status,omitempty"`
	ToStatus   string    `json:"to_status,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type Incident struct {
	ID          string              `json:"id"`
	Application string              `json:"application"`
	Title       string              `json:"title"`
	Status      string              `json:"status"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
	ClosedAt    *time.Time          `json:"closed_at,omitempty"`
	Groups      []ErrorGroupSummary `json:"groups,omitempty"`
	Entries     []IncidentEntry     `json:"entries,omitempty"`
}

type IncidentFilter struct {
	Application, Status string
	Limit, Offset       int
}

type Deployment struct {
	ID          string    `json:"id"`
	Application string    `json:"application"`
	Component   string    `json:"component"`
	Environment string    `json:"environment"`
	Provider    string    `json:"provider"`
	ExternalID  string    `json:"external_id"`
	Version     string    `json:"version,omitempty"`
	CommitSHA   string    `json:"commit_sha,omitempty"`
	CommitURL   string    `json:"commit_url,omitempty"`
	Actor       string    `json:"actor,omitempty"`
	SourceURL   string    `json:"source_url,omitempty"`
	DeployedAt  time.Time `json:"deployed_at"`
	CreatedAt   time.Time `json:"created_at"`
}

type DeploymentInput struct {
	Application, Component, Environment, Provider, ExternalID string
	Version, CommitSHA, CommitURL, Actor, SourceURL           string
	DeployedAt                                                time.Time
}

type TimelineBucket struct {
	Start      time.Time `json:"start"`
	ErrorCount int64     `json:"error_count"`
}
type OperationsTimeline struct {
	Buckets     []TimelineBucket `json:"buckets"`
	Deployments []Deployment     `json:"deployments"`
	Incidents   []Incident       `json:"incidents"`
}

func (s *Postgres) ListErrorGroups(ctx context.Context, application, query string, limit int) ([]ErrorGroupSummary, error) {
	rows, err := s.pool.Query(ctx, `SELECT g.id::text,a.slug,i.component,g.error_type,g.sample_message,g.occurrence_count,g.last_seen_at
		FROM error_groups g JOIN integrations i ON i.id=g.integration_id JOIN applications a ON a.id=i.application_id
		WHERE ($1='' OR a.slug=$1) AND ($2='' OR g.error_type ILIKE '%'||$2||'%' OR g.sample_message ILIKE '%'||$2||'%')
		ORDER BY g.last_seen_at DESC LIMIT $3`, application, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []ErrorGroupSummary{}
	for rows.Next() {
		var item ErrorGroupSummary
		if err := rows.Scan(&item.ID, &item.Application, &item.Component, &item.ErrorType, &item.Message, &item.OccurrenceCount, &item.LastSeenAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Postgres) CreateIncident(ctx context.Context, title string, groupIDs []string) (Incident, error) {
	title = strings.TrimSpace(title)
	if len(title) < 1 || len(title) > 160 || len(groupIDs) == 0 {
		return Incident{}, ErrInvalid
	}
	groupIDs = append([]string(nil), groupIDs...)
	sort.Strings(groupIDs)
	for index := 1; index < len(groupIDs); index++ {
		if groupIDs[index] == groupIDs[index-1] {
			return Incident{}, ErrInvalid
		}
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return Incident{}, err
	}
	defer tx.Rollback(ctx)
	var applicationID string
	for _, groupID := range groupIDs {
		if _, err = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, groupID); err != nil {
			return Incident{}, err
		}
		var candidate string
		err = tx.QueryRow(ctx, `SELECT i.application_id::text FROM error_groups g JOIN integrations i ON i.id=g.integration_id WHERE g.id=$1`, groupID).Scan(&candidate)
		if err != nil {
			return Incident{}, err
		}
		if applicationID != "" && applicationID != candidate {
			return Incident{}, ErrInvalid
		}
		applicationID = candidate
		var occupied bool
		if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM incident_error_groups x JOIN incidents n ON n.id=x.incident_id WHERE x.error_group_id=$1 AND n.status='investigating')`, groupID).Scan(&occupied); err != nil {
			return Incident{}, err
		}
		if occupied {
			return Incident{}, ErrConflict
		}
	}
	var id string
	if err = tx.QueryRow(ctx, `INSERT INTO incidents(application_id,title) VALUES($1,$2) RETURNING id::text`, applicationID, title).Scan(&id); err != nil {
		return Incident{}, err
	}
	for _, groupID := range groupIDs {
		if _, err = tx.Exec(ctx, `INSERT INTO incident_error_groups(incident_id,error_group_id) VALUES($1,$2)`, id, groupID); err != nil {
			return Incident{}, err
		}
		if _, err = tx.Exec(ctx, `INSERT INTO incident_entries(incident_id,kind,body) VALUES($1,'group_added',$2)`, id, groupID); err != nil {
			return Incident{}, err
		}
	}
	if err = tx.Commit(ctx); err != nil {
		return Incident{}, err
	}
	return s.Incident(ctx, id)
}

func (s *Postgres) ListIncidents(ctx context.Context, f IncidentFilter) ([]Incident, error) {
	rows, err := s.pool.Query(ctx, `SELECT n.id::text,a.slug,n.title,n.status,n.created_at,n.updated_at,n.closed_at FROM incidents n JOIN applications a ON a.id=n.application_id WHERE($1=''OR a.slug=$1)AND($2=''OR n.status=$2)ORDER BY n.created_at DESC LIMIT $3 OFFSET $4`, f.Application, f.Status, f.Limit, f.Offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Incident{}
	for rows.Next() {
		var n Incident
		if err := rows.Scan(&n.ID, &n.Application, &n.Title, &n.Status, &n.CreatedAt, &n.UpdatedAt, &n.ClosedAt); err != nil {
			return nil, err
		}
		items = append(items, n)
	}
	return items, rows.Err()
}

func (s *Postgres) Incident(ctx context.Context, id string) (Incident, error) {
	var n Incident
	err := s.pool.QueryRow(ctx, `SELECT n.id::text,a.slug,n.title,n.status,n.created_at,n.updated_at,n.closed_at FROM incidents n JOIN applications a ON a.id=n.application_id WHERE n.id=$1`, id).Scan(&n.ID, &n.Application, &n.Title, &n.Status, &n.CreatedAt, &n.UpdatedAt, &n.ClosedAt)
	if err != nil {
		return n, err
	}
	rows, err := s.pool.Query(ctx, `SELECT g.id::text,a.slug,i.component,g.error_type,g.sample_message,g.occurrence_count,g.last_seen_at FROM incident_error_groups x JOIN error_groups g ON g.id=x.error_group_id JOIN integrations i ON i.id=g.integration_id JOIN applications a ON a.id=i.application_id WHERE x.incident_id=$1 ORDER BY x.added_at`, id)
	if err != nil {
		return n, err
	}
	for rows.Next() {
		var g ErrorGroupSummary
		if err := rows.Scan(&g.ID, &g.Application, &g.Component, &g.ErrorType, &g.Message, &g.OccurrenceCount, &g.LastSeenAt); err != nil {
			rows.Close()
			return n, err
		}
		n.Groups = append(n.Groups, g)
	}
	rows.Close()
	rows, err = s.pool.Query(ctx, `SELECT id::text,kind,COALESCE(body,''),COALESCE(from_status,''),COALESCE(to_status,''),created_at FROM incident_entries WHERE incident_id=$1 ORDER BY created_at,id`, id)
	if err != nil {
		return n, err
	}
	defer rows.Close()
	for rows.Next() {
		var e IncidentEntry
		if err := rows.Scan(&e.ID, &e.Kind, &e.Body, &e.FromStatus, &e.ToStatus, &e.CreatedAt); err != nil {
			return n, err
		}
		n.Entries = append(n.Entries, e)
	}
	return n, rows.Err()
}

func (s *Postgres) AddIncidentNote(ctx context.Context, id, body string) error {
	body = strings.TrimSpace(body)
	if len(body) < 1 || len(body) > 5000 {
		return ErrInvalid
	}
	r, err := s.pool.Exec(ctx, `INSERT INTO incident_entries(incident_id,kind,body) SELECT id,'note',$2 FROM incidents WHERE id=$1`, id, body)
	if err == nil && r.RowsAffected() != 1 {
		return pgx.ErrNoRows
	}
	return err
}

func (s *Postgres) ChangeIncidentStatus(ctx context.Context, id, status, note string) error {
	note = strings.TrimSpace(note)
	if note == "" || len(note) > 5000 || !(status == "investigating" || status == "resolved" || status == "noise") {
		return ErrInvalid
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var current string
	if err = tx.QueryRow(ctx, `SELECT status FROM incidents WHERE id=$1 FOR UPDATE`, id).Scan(&current); err != nil {
		return err
	}
	if current == status {
		return ErrConflict
	}
	if current == "investigating" && status == "investigating" || current != "investigating" && status != "investigating" {
		return ErrInvalid
	}
	if status == "investigating" {
		rows, e := tx.Query(ctx, `SELECT error_group_id::text FROM incident_error_groups WHERE incident_id=$1`, id)
		if e != nil {
			return e
		}
		groupIDs := []string{}
		for rows.Next() {
			var gid string
			if e = rows.Scan(&gid); e != nil {
				rows.Close()
				return e
			}
			groupIDs = append(groupIDs, gid)
		}
		if e = rows.Err(); e != nil {
			rows.Close()
			return e
		}
		rows.Close()
		sort.Strings(groupIDs)
		for _, gid := range groupIDs {
			_, _ = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, gid)
			var occupied bool
			if e = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM incident_error_groups x JOIN incidents n ON n.id=x.incident_id WHERE x.error_group_id=$1 AND n.status='investigating' AND n.id<>$2)`, gid, id).Scan(&occupied); e != nil {
				return e
			}
			if occupied {
				return ErrConflict
			}
		}
	}
	_, err = tx.Exec(ctx, `UPDATE incidents SET status=$2,closed_at=CASE WHEN $2='investigating' THEN NULL ELSE now() END,updated_at=now() WHERE id=$1`, id, status)
	if err == nil {
		_, err = tx.Exec(ctx, `INSERT INTO incident_entries(incident_id,kind,body,from_status,to_status)VALUES($1,'status_change',$2,$3,$4)`, id, note, current, status)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Postgres) AddIncidentGroup(ctx context.Context, id, groupID string) error {
	return s.changeIncidentGroup(ctx, id, groupID, true)
}
func (s *Postgres) RemoveIncidentGroup(ctx context.Context, id, groupID string) error {
	return s.changeIncidentGroup(ctx, id, groupID, false)
}
func (s *Postgres) changeIncidentGroup(ctx context.Context, id, groupID string, add bool) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	_, _ = tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, groupID)
	var sameApp bool
	err = tx.QueryRow(ctx, `SELECT n.status='investigating' AND n.application_id=i.application_id FROM incidents n,error_groups g JOIN integrations i ON i.id=g.integration_id WHERE n.id=$1 AND g.id=$2`, id, groupID).Scan(&sameApp)
	if err != nil {
		return err
	}
	if !sameApp {
		return ErrInvalid
	}
	if add {
		var occupied bool
		_ = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM incident_error_groups x JOIN incidents n ON n.id=x.incident_id WHERE x.error_group_id=$1 AND n.status='investigating')`, groupID).Scan(&occupied)
		if occupied {
			return ErrConflict
		}
		_, err = tx.Exec(ctx, `INSERT INTO incident_error_groups VALUES($1,$2,now())`, id, groupID)
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO incident_entries(incident_id,kind,body)VALUES($1,'group_added',$2)`, id, groupID)
		}
	} else {
		var count int
		_ = tx.QueryRow(ctx, `SELECT count(*) FROM incident_error_groups WHERE incident_id=$1`, id).Scan(&count)
		if count <= 1 {
			return ErrInvalid
		}
		r, e := tx.Exec(ctx, `DELETE FROM incident_error_groups WHERE incident_id=$1 AND error_group_id=$2`, id, groupID)
		err = e
		if err == nil && r.RowsAffected() != 1 {
			return pgx.ErrNoRows
		}
		if err == nil {
			_, err = tx.Exec(ctx, `INSERT INTO incident_entries(incident_id,kind,body)VALUES($1,'group_removed',$2)`, id, groupID)
		}
	}
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `UPDATE incidents SET updated_at=now() WHERE id=$1`, id)
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Postgres) SaveDeployment(ctx context.Context, in DeploymentInput) (Deployment, bool, error) {
	if in.DeployedAt.IsZero() {
		in.DeployedAt = time.Now()
	}
	if in.Application == "" || in.Component == "" || in.Environment != "production" || in.ExternalID == "" || in.Version == "" && in.CommitSHA == "" {
		return Deployment{}, false, ErrInvalid
	}
	var d Deployment
	var created bool
	err := s.pool.QueryRow(ctx, `WITH inserted AS(INSERT INTO deployments(application_id,component,environment,provider,external_id,version,commit_sha,commit_url,actor,source_url,deployed_at)SELECT id,$2,$3,$4,$5,NULLIF($6,''),NULLIF($7,''),NULLIF($8,''),NULLIF($9,''),NULLIF($10,''),$11 FROM applications WHERE slug=$1 ON CONFLICT(provider,external_id)DO NOTHING RETURNING *) SELECT id::text,$1,component,environment,provider,external_id,COALESCE(version,''),COALESCE(commit_sha,''),COALESCE(commit_url,''),COALESCE(actor,''),COALESCE(source_url,''),deployed_at,created_at,true FROM inserted UNION ALL SELECT d.id::text,a.slug,d.component,d.environment,d.provider,d.external_id,COALESCE(d.version,''),COALESCE(d.commit_sha,''),COALESCE(d.commit_url,''),COALESCE(d.actor,''),COALESCE(d.source_url,''),d.deployed_at,d.created_at,false FROM deployments d JOIN applications a ON a.id=d.application_id WHERE d.provider=$4 AND d.external_id=$5 LIMIT 1`, in.Application, in.Component, in.Environment, in.Provider, in.ExternalID, in.Version, in.CommitSHA, in.CommitURL, in.Actor, in.SourceURL, in.DeployedAt).Scan(&d.ID, &d.Application, &d.Component, &d.Environment, &d.Provider, &d.ExternalID, &d.Version, &d.CommitSHA, &d.CommitURL, &d.Actor, &d.SourceURL, &d.DeployedAt, &d.CreatedAt, &created)
	if err == nil && !created && (d.Application != in.Application || d.Component != in.Component || d.Environment != in.Environment || d.Version != in.Version || d.CommitSHA != in.CommitSHA) {
		return Deployment{}, false, ErrConflict
	}
	return d, created, err
}

func (s *Postgres) ListDeployments(ctx context.Context, application, component string, from, to time.Time) ([]Deployment, error) {
	rows, err := s.pool.Query(ctx, `SELECT d.id::text,a.slug,d.component,d.environment,d.provider,d.external_id,COALESCE(d.version,''),COALESCE(d.commit_sha,''),COALESCE(d.commit_url,''),COALESCE(d.actor,''),COALESCE(d.source_url,''),d.deployed_at,d.created_at FROM deployments d JOIN applications a ON a.id=d.application_id WHERE a.slug=$1 AND($2=''OR d.component=$2)AND d.deployed_at>=$3 AND d.deployed_at<$4 ORDER BY d.deployed_at DESC`, application, component, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []Deployment{}
	for rows.Next() {
		var d Deployment
		if err := rows.Scan(&d.ID, &d.Application, &d.Component, &d.Environment, &d.Provider, &d.ExternalID, &d.Version, &d.CommitSHA, &d.CommitURL, &d.Actor, &d.SourceURL, &d.DeployedAt, &d.CreatedAt); err != nil {
			return nil, err
		}
		items = append(items, d)
	}
	return items, rows.Err()
}

func (s *Postgres) OperationsTimeline(ctx context.Context, application, component string, from, to time.Time, bucket time.Duration) (OperationsTimeline, error) {
	result := OperationsTimeline{Buckets: []TimelineBucket{}, Deployments: []Deployment{}, Incidents: []Incident{}}
	rows, err := s.pool.Query(ctx, `SELECT series,COUNT(e.id) FROM generate_series($3::timestamptz,$4::timestamptz-make_interval(secs=>$5),make_interval(secs=>$5))series LEFT JOIN integrations i ON i.application_id=(SELECT id FROM applications WHERE slug=$1) AND($2=''OR i.component=$2) LEFT JOIN error_events e ON e.integration_id=i.id AND e.occurred_at>=series AND e.occurred_at<series+make_interval(secs=>$5) GROUP BY series ORDER BY series`, application, component, from, to, int(bucket.Seconds()))
	if err != nil {
		return result, err
	}
	for rows.Next() {
		var b TimelineBucket
		if err = rows.Scan(&b.Start, &b.ErrorCount); err != nil {
			rows.Close()
			return result, err
		}
		result.Buckets = append(result.Buckets, b)
	}
	rows.Close()
	result.Deployments, err = s.ListDeployments(ctx, application, component, from, to)
	if err != nil {
		return result, err
	}
	result.Incidents, err = s.ListIncidents(ctx, IncidentFilter{Application: application, Limit: 100})
	return result, err
}

func ValidateIncidentStatus(status string) bool {
	return status == "" || status == "investigating" || status == "resolved" || status == "noise"
}
func ValidateProvider(provider string) bool {
	return provider == "railway" || provider == "github_actions" || provider == "manual"
}
func WrapConflict(message string) error { return fmt.Errorf("%w: %s", ErrConflict, message) }
