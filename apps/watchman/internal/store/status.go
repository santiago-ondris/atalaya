package store

import (
	"context"
	"strings"
	"time"
)

func (s *Postgres) SetIncidentPublication(ctx context.Context, id string, published bool, title, message string) error {
	title, message = cleanPublicText(title), cleanPublicText(message)
	if published && (title == "" || message == "" || len(title) > 160 || len(message) > 2000) {
		return ErrInvalid
	}
	if !published {
		r, err := s.pool.Exec(ctx, `UPDATE incidents SET public_title=NULL,public_message=NULL,published_at=NULL,updated_at=now() WHERE id=$1`, id)
		if err == nil && r.RowsAffected() == 0 {
			return ErrInvalid
		}
		return err
	}
	r, err := s.pool.Exec(ctx, `UPDATE incidents SET public_title=$2,public_message=$3,published_at=COALESCE(published_at,now()),updated_at=now() WHERE id=$1 AND status<>'noise'`, id, title, message)
	if err == nil && r.RowsAffected() == 0 {
		return ErrInvalid
	}
	return err
}

func cleanPublicText(value string) string {
	value = strings.TrimSpace(value)
	value = strings.NewReplacer("<", "", ">", "", "\x00", "").Replace(value)
	return value
}

func (s *Postgres) PublicStatus(ctx context.Context, now time.Time) (PublicStatus, error) {
	result := PublicStatus{Applications: []PublicApplicationStatus{}, Incidents: []PublicIncident{}}
	rows, err := s.pool.Query(ctx, `WITH catalog(slug,name,ord) AS (VALUES ('farmami','Farmami',1),('wheels_house','Wheels House',2),('prensap','Prensap',3),('notizap','Notizap',4))
		SELECT c.slug,c.name,t.component,COALESCE(t.confirmed_status,'unknown'),t.last_checked_at
		FROM catalog c LEFT JOIN availability_targets t ON t.application=c.slug ORDER BY c.ord,t.component`)
	if err != nil {
		return result, err
	}
	byApp := map[string]int{}
	for rows.Next() {
		var slug, name string
		var component *string
		var status string
		var checked *time.Time
		if err = rows.Scan(&slug, &name, &component, &status, &checked); err != nil {
			rows.Close()
			return result, err
		}
		idx, ok := byApp[slug]
		if !ok {
			idx = len(result.Applications)
			byApp[slug] = idx
			result.Applications = append(result.Applications, PublicApplicationStatus{Slug: slug, DisplayName: name, Status: "unknown", Components: []PublicComponent{}})
		}
		if component != nil {
			result.Applications[idx].Components = append(result.Applications[idx].Components, PublicComponent{Name: *component, Status: status, LastCheckedAt: checked})
			if checked != nil && (result.Applications[idx].LastCheckedAt == nil || checked.After(*result.Applications[idx].LastCheckedAt)) {
				result.Applications[idx].LastCheckedAt = checked
			}
		}
	}
	rows.Close()
	if err = rows.Err(); err != nil {
		return result, err
	}
	for i := range result.Applications {
		up, down, unknown := 0, 0, 0
		for _, c := range result.Applications[i].Components {
			if c.Status == "up" {
				up++
			} else if c.Status == "down" {
				down++
			} else {
				unknown++
			}
		}
		if len(result.Applications[i].Components) < 2 || unknown > 0 {
			result.Applications[i].Status = "unknown"
		} else if up == 2 {
			result.Applications[i].Status = "operational"
		} else if down == 2 {
			result.Applications[i].Status = "major_outage"
		} else {
			result.Applications[i].Status = "degraded"
		}
		var oldest *time.Time
		var total, operational int64
		err = s.pool.QueryRow(ctx, `SELECT min(captured_at),count(*),count(*) FILTER(WHERE status='operational') FROM availability_snapshots WHERE application=$1 AND captured_at >= $2`, result.Applications[i].Slug, now.Add(-30*24*time.Hour)).Scan(&oldest, &total, &operational)
		if err != nil {
			return result, err
		}
		if oldest != nil && !oldest.After(now.Add(-30*24*time.Hour+2*time.Minute)) && total > 0 {
			value := float64(operational) * 100 / float64(total)
			result.Applications[i].Uptime30Days = &value
		}
	}
	rows, err = s.pool.Query(ctx, `SELECT n.id::text,a.slug,n.public_title,n.public_message,n.status,n.published_at,n.closed_at FROM incidents n JOIN applications a ON a.id=n.application_id WHERE n.published_at IS NOT NULL AND n.status<>'noise' AND (n.closed_at IS NULL OR n.closed_at >= $1) ORDER BY n.published_at DESC LIMIT 50`, now.Add(-30*24*time.Hour))
	if err != nil {
		return result, err
	}
	defer rows.Close()
	for rows.Next() {
		var n PublicIncident
		if err = rows.Scan(&n.ID, &n.Application, &n.Title, &n.Message, &n.Status, &n.PublishedAt, &n.ResolvedAt); err != nil {
			return result, err
		}
		result.Incidents = append(result.Incidents, n)
	}
	return result, rows.Err()
}
