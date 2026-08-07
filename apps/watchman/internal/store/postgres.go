package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Postgres struct{ pool *pgxpool.Pool }

func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (store *Postgres) EnsureSentryIntegration(ctx context.Context, appSlug, organization, project string) (uuid.UUID, error) {
	identifiers, _ := json.Marshal(map[string]string{"organization": organization, "project": project})
	var id uuid.UUID
	err := store.pool.QueryRow(ctx, `
		INSERT INTO integrations (application_id, source, external_identifier)
		SELECT id, 'sentry', $2 FROM applications WHERE slug = $1
		ON CONFLICT (application_id, source) DO UPDATE SET external_identifier = EXCLUDED.external_identifier
		RETURNING id`, appSlug, identifiers).Scan(&id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ensure sentry integration: %w", err)
	}
	_, err = store.pool.Exec(ctx, `INSERT INTO source_checkpoints (integration_id) VALUES ($1) ON CONFLICT DO NOTHING`, id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("ensure sentry checkpoint: %w", err)
	}
	return id, nil
}

func (store *Postgres) Checkpoint(ctx context.Context, integrationID uuid.UUID) (domain.Cursor, error) {
	var raw []byte
	err := store.pool.QueryRow(ctx, `SELECT cursor_data FROM source_checkpoints WHERE integration_id = $1`, integrationID).Scan(&raw)
	if err != nil {
		return domain.Cursor{}, fmt.Errorf("load checkpoint: %w", err)
	}
	var cursor domain.Cursor
	if err := json.Unmarshal(raw, &cursor); err != nil {
		return domain.Cursor{}, fmt.Errorf("decode checkpoint: %w", err)
	}
	return cursor, nil
}

func (store *Postgres) RecordAttempt(ctx context.Context, integrationID uuid.UUID, pollErr error) error {
	var message *string
	if pollErr != nil {
		value := pollErr.Error()
		message = &value
	}
	_, err := store.pool.Exec(ctx, `UPDATE source_checkpoints SET last_attempt_at = now(), last_error = $2, updated_at = now() WHERE integration_id = $1`, integrationID, message)
	return err
}

func (store *Postgres) ImportBatch(ctx context.Context, integrationID uuid.UUID, batch domain.EventBatch) (int, error) {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return 0, fmt.Errorf("begin import: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `SELECT 1 FROM source_checkpoints WHERE integration_id = $1 FOR UPDATE`, integrationID); err != nil {
		return 0, fmt.Errorf("lock checkpoint: %w", err)
	}
	imported := 0
	for _, event := range batch.Events {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM error_events WHERE integration_id=$1 AND source_event_id=$2)`, integrationID, event.SourceEventID).Scan(&exists); err != nil {
			return 0, fmt.Errorf("check event idempotency: %w", err)
		}
		if exists {
			continue
		}
		var groupID uuid.UUID
		err := tx.QueryRow(ctx, `
			INSERT INTO error_groups (integration_id, environment, fingerprint, error_type, sample_message, first_seen_at, last_seen_at)
			VALUES ($1,$2,$3,$4,$5,$6,$6)
			ON CONFLICT (integration_id, environment, fingerprint) DO UPDATE SET
				last_seen_at=GREATEST(error_groups.last_seen_at, EXCLUDED.last_seen_at),
				first_seen_at=LEAST(error_groups.first_seen_at, EXCLUDED.first_seen_at),
				occurrence_count=error_groups.occurrence_count+1, updated_at=now()
			RETURNING id`, integrationID, event.Environment, event.Fingerprint, event.ErrorType,
			event.Message, event.OccurredAt).Scan(&groupID)
		if err != nil {
			return 0, fmt.Errorf("upsert error group: %w", err)
		}
		metadata, err := json.Marshal(event.Metadata)
		if err != nil {
			return 0, fmt.Errorf("encode event metadata: %w", err)
		}
		_, err = tx.Exec(ctx, `INSERT INTO error_events
			(integration_id,error_group_id,source_event_id,occurred_at,error_type,message,stack_trace,release,metadata)
			VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),$9)`, integrationID, groupID,
			event.SourceEventID, event.OccurredAt, event.ErrorType, event.Message, event.StackTrace, event.Release, metadata)
		if err != nil {
			return 0, fmt.Errorf("insert error event: %w", err)
		}
		imported++
	}
	cursorJSON, _ := json.Marshal(batch.NextCursor)
	_, err = tx.Exec(ctx, `UPDATE source_checkpoints SET cursor_data=$2,last_attempt_at=now(),last_success_at=now(),last_error=NULL,updated_at=now() WHERE integration_id=$1`, integrationID, cursorJSON)
	if err != nil {
		return 0, fmt.Errorf("advance checkpoint: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit import: %w", err)
	}
	return imported, nil
}

type EventSummary struct {
	ID            string    `json:"id"`
	SourceEventID string    `json:"source_event_id"`
	Application   string    `json:"application"`
	Environment   string    `json:"environment"`
	ErrorType     string    `json:"error_type"`
	Message       string    `json:"message"`
	OccurredAt    time.Time `json:"occurred_at"`
	IngestedAt    time.Time `json:"ingested_at"`
}

type EventDetail struct {
	EventSummary
	Fingerprint string         `json:"fingerprint"`
	StackTrace  string         `json:"stack_trace,omitempty"`
	Release     string         `json:"release,omitempty"`
	Metadata    map[string]any `json:"metadata"`
}

func (store *Postgres) ListEvents(ctx context.Context, limit int) ([]EventSummary, error) {
	rows, err := store.pool.Query(ctx, `SELECT e.id,e.source_event_id,a.slug,g.environment,e.error_type,e.message,e.occurred_at,e.ingested_at
		FROM error_events e JOIN error_groups g ON g.id=e.error_group_id JOIN integrations i ON i.id=e.integration_id JOIN applications a ON a.id=i.application_id
		ORDER BY e.occurred_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]EventSummary, 0)
	for rows.Next() {
		var item EventSummary
		if err := rows.Scan(&item.ID, &item.SourceEventID, &item.Application, &item.Environment, &item.ErrorType, &item.Message, &item.OccurredAt, &item.IngestedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Postgres) Event(ctx context.Context, id string) (EventDetail, error) {
	var item EventDetail
	var metadata []byte
	err := store.pool.QueryRow(ctx, `SELECT e.id,e.source_event_id,a.slug,g.environment,e.error_type,e.message,e.occurred_at,e.ingested_at,
		g.fingerprint,COALESCE(e.stack_trace,''),COALESCE(e.release,''),e.metadata
		FROM error_events e JOIN error_groups g ON g.id=e.error_group_id JOIN integrations i ON i.id=e.integration_id JOIN applications a ON a.id=i.application_id WHERE e.id=$1`, id).
		Scan(&item.ID, &item.SourceEventID, &item.Application, &item.Environment, &item.ErrorType, &item.Message, &item.OccurredAt, &item.IngestedAt, &item.Fingerprint, &item.StackTrace, &item.Release, &metadata)
	if err != nil {
		return EventDetail{}, err
	}
	if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
		return EventDetail{}, err
	}
	return item, nil
}
