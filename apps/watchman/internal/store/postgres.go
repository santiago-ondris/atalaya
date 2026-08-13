package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/santiago-ondris/atalaya/apps/watchman/internal/domain"
)

type Postgres struct{ pool *pgxpool.Pool }

func NewPostgres(pool *pgxpool.Pool) *Postgres { return &Postgres{pool: pool} }

func (store *Postgres) CreateSession(ctx context.Context, tokenHash []byte, expiresAt time.Time) error {
	_, err := store.pool.Exec(ctx, `INSERT INTO command_center_sessions (token_hash,expires_at) VALUES ($1,$2)`, tokenHash, expiresAt)
	return err
}

func (store *Postgres) SessionValid(ctx context.Context, tokenHash []byte, now time.Time) (bool, error) {
	var valid bool
	err := store.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM command_center_sessions
		WHERE token_hash=$1 AND revoked_at IS NULL AND expires_at>$2)`, tokenHash, now).Scan(&valid)
	if err == nil && valid {
		_, err = store.pool.Exec(ctx, `UPDATE command_center_sessions SET last_seen_at=$2 WHERE token_hash=$1`, tokenHash, now)
	}
	return valid, err
}

func (store *Postgres) RevokeSession(ctx context.Context, tokenHash []byte, now time.Time) error {
	_, err := store.pool.Exec(ctx, `UPDATE command_center_sessions SET revoked_at=$2 WHERE token_hash=$1 AND revoked_at IS NULL`, tokenHash, now)
	return err
}

func (store *Postgres) EnsureDailyReport(ctx context.Context, date, timezone string, start, end time.Time) (domain.DailyReport, bool, error) {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return domain.DailyReport{}, false, err
	}
	defer tx.Rollback(ctx)
	var report domain.DailyReport
	err = tx.QueryRow(ctx, `INSERT INTO daily_reports (report_date,timezone,period_start,period_end)
		VALUES ($1,$2,$3,$4) ON CONFLICT (report_date) DO UPDATE SET report_date=EXCLUDED.report_date
		RETURNING id::text,report_date::text,timezone,period_start,period_end,status,attempts,COALESCE(last_error,''),created_at,sent_at,expired_at`,
		date, timezone, start, end).Scan(&report.ID, &report.Date, &report.Timezone, &report.PeriodStart, &report.PeriodEnd,
		&report.Status, &report.Attempts, &report.LastError, &report.CreatedAt, &report.SentAt, &report.ExpiredAt)
	if err != nil {
		return domain.DailyReport{}, false, fmt.Errorf("ensure daily report: %w", err)
	}
	if report.Status == "collecting" {
		_, err = tx.Exec(ctx, `INSERT INTO daily_report_applications
			(report_id,application_id,activity_kind,activity_source,activity_status,error_count,occurrence_count,
			 critical_count,high_count,medium_count,low_count,actionable_count)
			SELECT $1,a.id,'sessions',CASE WHEN a.slug='notizap' THEN 'application_insights' ELSE 'sentry' END,'unavailable',
			COUNT(DISTINCT e.error_group_id),COUNT(e.id),
			COUNT(*) FILTER (WHERE x.severity='critical'),COUNT(*) FILTER (WHERE x.severity='high'),
			COUNT(*) FILTER (WHERE x.severity='medium'),COUNT(*) FILTER (WHERE x.severity='low'),
			COUNT(*) FILTER (WHERE x.actionable=true)
			FROM applications a
			LEFT JOIN integrations src ON src.application_id=a.id
			LEFT JOIN error_events e ON e.integration_id=src.id AND e.occurred_at >= $2 AND e.occurred_at < $3
			LEFT JOIN LATERAL (SELECT severity,actionable FROM interpretations WHERE error_event_id=e.id ORDER BY created_at DESC LIMIT 1) x ON true
			GROUP BY a.id,a.slug
			ON CONFLICT (report_id,application_id) DO UPDATE SET
			error_count=EXCLUDED.error_count,occurrence_count=EXCLUDED.occurrence_count,
			critical_count=EXCLUDED.critical_count,high_count=EXCLUDED.high_count,medium_count=EXCLUDED.medium_count,
			low_count=EXCLUDED.low_count,actionable_count=EXCLUDED.actionable_count`, report.ID, start, end)
		if err != nil {
			return domain.DailyReport{}, false, fmt.Errorf("snapshot daily errors: %w", err)
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.DailyReport{}, false, err
	}
	return report, report.Status == "collecting", nil
}

func (store *Postgres) SaveDailyActivity(ctx context.Context, reportID, application string, metric domain.ActivityMetric, cause error) error {
	status, message := "available", ""
	var count *int64 = &metric.Count
	if cause != nil {
		status, message, count = "unavailable", cause.Error(), nil
	}
	result, err := store.pool.Exec(ctx, `UPDATE daily_report_applications item SET
		activity_count=$3,activity_kind=COALESCE(NULLIF($4,''),activity_kind),activity_source=COALESCE(NULLIF($5,''),activity_source),
		activity_status=$6,activity_error=NULLIF($7,'') FROM applications a
		WHERE item.application_id=a.id AND item.report_id=$1 AND a.slug=$2`, reportID, application, count, metric.Kind, metric.Source, status, message)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return fmt.Errorf("daily report application %q not found", application)
	}
	return nil
}

func (store *Postgres) FinalizeDailyReport(ctx context.Context, reportID string) error {
	result, err := store.pool.Exec(ctx, `UPDATE daily_reports SET status='pending',next_attempt_at=now(),updated_at=now()
		WHERE id=$1 AND status='collecting' AND (SELECT count(*) FROM daily_report_applications WHERE report_id=$1)=4`, reportID)
	if err != nil {
		return err
	}
	if result.RowsAffected() != 1 {
		return errors.New("daily report is incomplete or no longer collecting")
	}
	return nil
}

func (store *Postgres) ClaimDailyReport(ctx context.Context, workerID string, now time.Time) (domain.DailyReport, error) {
	_, err := store.pool.Exec(ctx, `UPDATE daily_reports SET status='expired',expired_at=$1,updated_at=$1
		WHERE status IN ('collecting','pending','processing') AND period_end <= $1`, now)
	if err != nil {
		return domain.DailyReport{}, err
	}
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return domain.DailyReport{}, err
	}
	defer tx.Rollback(ctx)
	var report domain.DailyReport
	err = tx.QueryRow(ctx, `SELECT id::text,report_date::text,timezone,period_start,period_end,status,attempts+1,
		COALESCE(last_error,''),created_at,sent_at,expired_at FROM daily_reports
		WHERE period_end>$1 AND ((status='pending' AND next_attempt_at<=$1) OR (status='processing' AND updated_at<$1-interval '5 minutes'))
		ORDER BY next_attempt_at,created_at FOR UPDATE SKIP LOCKED LIMIT 1`, now).Scan(&report.ID, &report.Date, &report.Timezone,
		&report.PeriodStart, &report.PeriodEnd, &report.Status, &report.Attempts, &report.LastError, &report.CreatedAt, &report.SentAt, &report.ExpiredAt)
	if err != nil {
		return domain.DailyReport{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE daily_reports SET status='processing',attempts=$2,updated_at=$3 WHERE id=$1`, report.ID, report.Attempts, now); err != nil {
		return domain.DailyReport{}, err
	}
	report.Status = "processing"
	apps, err := loadDailyReportApplications(ctx, tx, report.ID)
	if err != nil {
		return domain.DailyReport{}, err
	}
	report.Applications = apps
	if err := tx.Commit(ctx); err != nil {
		return domain.DailyReport{}, err
	}
	return report, nil
}

func loadDailyReportApplications(ctx context.Context, queryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, reportID string) ([]domain.DailyReportApplication, error) {
	rows, err := queryer.Query(ctx, `SELECT a.slug,a.display_name,r.activity_count,r.activity_kind,r.activity_source,r.activity_status,
		COALESCE(r.activity_error,''),r.error_count,r.occurrence_count,r.critical_count,r.high_count,r.medium_count,r.low_count,r.actionable_count
		FROM daily_report_applications r JOIN applications a ON a.id=r.application_id WHERE r.report_id=$1 ORDER BY a.display_name`, reportID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var apps []domain.DailyReportApplication
	for rows.Next() {
		var app domain.DailyReportApplication
		var critical, high, medium, low int64
		if err := rows.Scan(&app.Application, &app.DisplayName, &app.ActivityCount, &app.ActivityKind, &app.ActivitySource, &app.ActivityStatus,
			&app.ActivityError, &app.ErrorCount, &app.OccurrenceCount, &critical, &high, &medium, &low, &app.ActionableCount); err != nil {
			return nil, err
		}
		app.SeverityCounts = map[string]int64{"critical": critical, "high": high, "medium": medium, "low": low}
		apps = append(apps, app)
	}
	return apps, rows.Err()
}

func (store *Postgres) CompleteDailyReport(ctx context.Context, report domain.DailyReport, started time.Time, result domain.DeliveryResult) error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO daily_report_delivery_attempts
		(report_id,attempt_number,started_at,outcome,http_status,telegram_message_id) VALUES ($1,$2,$3,'sent',$4,$5)`,
		report.ID, report.Attempts, started, result.HTTPStatus, result.MessageID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE daily_reports SET status='sent',sent_at=now(),telegram_message_id=$2,last_error=NULL,updated_at=now() WHERE id=$1`, report.ID, result.MessageID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (store *Postgres) FailDailyReport(ctx context.Context, report domain.DailyReport, started time.Time, cause error, retryable bool, httpStatus int) error {
	now := time.Now()
	status, outcome, next, expiredAt := dailyReportFailureState(now, report, retryable)
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO daily_report_delivery_attempts
		(report_id,attempt_number,started_at,outcome,http_status,error_class,error_message) VALUES ($1,$2,$3,$4,NULLIF($5,0),$6,$7)`,
		report.ID, report.Attempts, started, outcome, httpStatus, fmt.Sprintf("%T", cause), cause.Error()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE daily_reports SET status=$2,next_attempt_at=$3,expired_at=$4,last_error=$5,updated_at=now() WHERE id=$1`,
		report.ID, status, next, expiredAt, cause.Error()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func dailyReportFailureState(now time.Time, report domain.DailyReport, retryable bool) (string, string, time.Time, *time.Time) {
	next := now.Add(30 * time.Minute)
	if retryable && report.Attempts < 8 && next.Before(report.PeriodEnd) {
		return "pending", "retryable_failure", next, nil
	}
	return "expired", "permanent_failure", next, &now
}

func (store *Postgres) ListDailyReports(ctx context.Context, limit int) ([]domain.DailyReport, error) {
	rows, err := store.pool.Query(ctx, `SELECT id::text,report_date::text,timezone,period_start,period_end,status,attempts,
		COALESCE(last_error,''),created_at,sent_at,expired_at FROM daily_reports ORDER BY report_date DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var reports []domain.DailyReport
	for rows.Next() {
		var report domain.DailyReport
		if err := rows.Scan(&report.ID, &report.Date, &report.Timezone, &report.PeriodStart, &report.PeriodEnd, &report.Status, &report.Attempts,
			&report.LastError, &report.CreatedAt, &report.SentAt, &report.ExpiredAt); err != nil {
			return nil, err
		}
		report.Applications, err = loadDailyReportApplications(ctx, store.pool, report.ID)
		if err != nil {
			return nil, err
		}
		reports = append(reports, report)
	}
	return reports, rows.Err()
}

type SentryIntegrationSpec struct {
	Application, Component, DisplayName, Organization, Project string
	Environments                                               []string
	AlertPolicy                                                domain.AlertPolicy
}

type ApplicationInsightsIntegrationSpec struct {
	Application, Component, DisplayName, WorkspaceID, ResourceID, Environment string
	AlertPolicy                                                               domain.AlertPolicy
}

type IntegrationRuntime struct {
	ID                  uuid.UUID
	MonitoringStartedAt time.Time
}

func (store *Postgres) EnsureApplicationInsightsIntegration(ctx context.Context, spec ApplicationInsightsIntegrationSpec) (IntegrationRuntime, error) {
	identifiers, _ := json.Marshal(map[string]string{"workspace_id": spec.WorkspaceID, "resource_id": spec.ResourceID})
	policy, _ := json.Marshal(spec.AlertPolicy)
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("begin ensure Application Insights integration: %w", err)
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE applications SET alert_policy=$2 WHERE slug=$1`, spec.Application, policy)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("update application alert policy: %w", err)
	}
	if result.RowsAffected() != 1 {
		return IntegrationRuntime{}, fmt.Errorf("unknown application %q", spec.Application)
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO integrations (application_id,source,component,display_name,external_identifier,environment_filters)
		SELECT id,'application_insights',$2,$3,$4,ARRAY[$5] FROM applications WHERE slug=$1
		ON CONFLICT (application_id,source,component) DO UPDATE SET display_name=EXCLUDED.display_name,
		external_identifier=EXCLUDED.external_identifier,environment_filters=EXCLUDED.environment_filters,enabled=true RETURNING id`,
		spec.Application, spec.Component, spec.DisplayName, identifiers, spec.Environment).Scan(&id)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("ensure Application Insights integration: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO source_checkpoints (integration_id,monitoring_started_at) VALUES ($1,now()) ON CONFLICT DO NOTHING`, id)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("ensure Application Insights checkpoint: %w", err)
	}
	var startedAt time.Time
	if err := tx.QueryRow(ctx, `SELECT monitoring_started_at FROM source_checkpoints WHERE integration_id=$1`, id).Scan(&startedAt); err != nil {
		return IntegrationRuntime{}, fmt.Errorf("load monitoring start: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return IntegrationRuntime{}, fmt.Errorf("commit Application Insights integration: %w", err)
	}
	return IntegrationRuntime{ID: id, MonitoringStartedAt: startedAt}, nil
}

func (store *Postgres) EnsureSentryIntegration(ctx context.Context, spec SentryIntegrationSpec) (IntegrationRuntime, error) {
	identifiers, _ := json.Marshal(map[string]string{"organization": spec.Organization, "project": spec.Project})
	policy, _ := json.Marshal(spec.AlertPolicy)
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("begin ensure Sentry integration: %w", err)
	}
	defer tx.Rollback(ctx)
	result, err := tx.Exec(ctx, `UPDATE applications SET alert_policy=$2 WHERE slug=$1`, spec.Application, policy)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("update application alert policy: %w", err)
	}
	if result.RowsAffected() != 1 {
		return IntegrationRuntime{}, fmt.Errorf("unknown application %q", spec.Application)
	}
	var id uuid.UUID
	err = tx.QueryRow(ctx, `
		INSERT INTO integrations (application_id, source, component, display_name, external_identifier, environment_filters)
		SELECT id, 'sentry', $2, $3, $4, $5 FROM applications WHERE slug = $1
		ON CONFLICT (application_id, source, component) DO UPDATE SET
			display_name=EXCLUDED.display_name, external_identifier=EXCLUDED.external_identifier,
			environment_filters=EXCLUDED.environment_filters, enabled=true
		RETURNING id`, spec.Application, spec.Component, spec.DisplayName, identifiers, spec.Environments).Scan(&id)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("ensure sentry integration: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO source_checkpoints (integration_id,monitoring_started_at) VALUES ($1,now()) ON CONFLICT DO NOTHING`, id)
	if err != nil {
		return IntegrationRuntime{}, fmt.Errorf("ensure sentry checkpoint: %w", err)
	}
	var startedAt time.Time
	if err := tx.QueryRow(ctx, `SELECT monitoring_started_at FROM source_checkpoints WHERE integration_id=$1`, id).Scan(&startedAt); err != nil {
		return IntegrationRuntime{}, fmt.Errorf("load monitoring start: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return IntegrationRuntime{}, fmt.Errorf("commit Sentry integration: %w", err)
	}
	return IntegrationRuntime{ID: id, MonitoringStartedAt: startedAt}, nil
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
		var eventID uuid.UUID
		err = tx.QueryRow(ctx, `INSERT INTO error_events
			(integration_id,error_group_id,source_event_id,occurred_at,error_type,message,stack_trace,release,metadata)
			VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,''),NULLIF($8,''),$9) RETURNING id`, integrationID, groupID,
			event.SourceEventID, event.OccurredAt, event.ErrorType, event.Message, event.StackTrace, event.Release, metadata).Scan(&eventID)
		if err != nil {
			return 0, fmt.Errorf("insert error event: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO interpretation_jobs (error_event_id, prompt_version) VALUES ($1, 'error-analysis-v1') ON CONFLICT DO NOTHING`, eventID); err != nil {
			return 0, fmt.Errorf("enqueue interpretation: %w", err)
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

func (store *Postgres) ClaimInterpretationJob(ctx context.Context, workerID string) (domain.InterpretationJob, error) {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return domain.InterpretationJob{}, err
	}
	defer tx.Rollback(ctx)
	var job domain.InterpretationJob
	var metadata, alertPolicy []byte
	err = tx.QueryRow(ctx, `
		SELECT j.id,e.id,i.source,e.source_event_id,a.slug,g.environment,e.occurred_at,e.error_type,e.message,
		       COALESCE(e.stack_trace,''),COALESCE(e.release,''),g.fingerprint,e.metadata,a.alert_policy,j.prompt_version,j.attempts+1,j.max_attempts
		FROM interpretation_jobs j
		JOIN error_events e ON e.id=j.error_event_id
		JOIN error_groups g ON g.id=e.error_group_id
		JOIN integrations i ON i.id=e.integration_id
		JOIN applications a ON a.id=i.application_id
		WHERE (j.status='pending' AND j.available_at <= now())
		   OR (j.status='processing' AND j.locked_at < now() - interval '5 minutes')
		ORDER BY j.available_at,j.created_at FOR UPDATE OF j SKIP LOCKED LIMIT 1`).Scan(
		&job.ID, &job.EventID, &job.Source, &job.SourceEventID, &job.Application, &job.Environment,
		&job.OccurredAt, &job.ErrorType, &job.Message, &job.StackTrace, &job.Release, &job.Fingerprint,
		&metadata, &alertPolicy, &job.PromptVersion, &job.Attempts, &job.MaxAttempts)
	if err != nil {
		return domain.InterpretationJob{}, err
	}
	if err := json.Unmarshal(metadata, &job.Metadata); err != nil {
		return domain.InterpretationJob{}, err
	}
	if err := json.Unmarshal(alertPolicy, &job.AlertPolicy); err != nil {
		return domain.InterpretationJob{}, err
	}
	if job.Attempts > job.MaxAttempts {
		job.Attempts = job.MaxAttempts
	}
	if _, err := tx.Exec(ctx, `UPDATE interpretation_jobs SET status='processing',attempts=$3,locked_at=now(),locked_by=$2,updated_at=now() WHERE id=$1`, job.ID, workerID, job.Attempts); err != nil {
		return domain.InterpretationJob{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.InterpretationJob{}, err
	}
	return job, nil
}

func (store *Postgres) CompleteInterpretation(ctx context.Context, job domain.InterpretationJob, result domain.Interpretation) error {
	actions, err := json.Marshal(result.SuggestedActions)
	if err != nil {
		return err
	}
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var interpretationID uuid.UUID
	err = tx.QueryRow(ctx, `INSERT INTO interpretations
		(job_id,error_event_id,summary,explanation,severity,actionable,suggested_actions,model,prompt_version,input_tokens,output_tokens,total_tokens,estimated_cost_usd,latency_ms)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14) RETURNING id`, job.ID, job.EventID, result.Summary,
		result.Explanation, result.Severity, result.Actionable, actions, result.Model, result.PromptVersion,
		result.Usage.InputTokens, result.Usage.OutputTokens, result.Usage.TotalTokens, result.EstimatedCostUSD, result.LatencyMS).Scan(&interpretationID)
	if err != nil {
		return fmt.Errorf("insert interpretation: %w", err)
	}
	deduplicationWindow := time.Duration(job.AlertPolicy.DeduplicationWindowSeconds) * time.Second
	if err := enqueueEventAlert(ctx, tx, job, interpretationID, deduplicationWindow, job.AlertPolicy.Eligible(result)); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE interpretation_jobs SET status='completed',completed_at=now(),locked_at=NULL,locked_by=NULL,last_error=NULL,updated_at=now() WHERE id=$1`, job.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func enqueueEventAlert(ctx context.Context, tx pgx.Tx, job domain.InterpretationJob, interpretationID uuid.UUID, window time.Duration, eligible bool) error {
	var groupID uuid.UUID
	if err := tx.QueryRow(ctx, `SELECT error_group_id FROM error_events WHERE id=$1`, job.EventID).Scan(&groupID); err != nil {
		return fmt.Errorf("load alert group: %w", err)
	}
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1,0))`, groupID.String()); err != nil {
		return fmt.Errorf("lock alert group: %w", err)
	}
	var windowID uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM alert_windows
		WHERE error_group_id=$1 AND closes_at > now()
		ORDER BY closes_at DESC LIMIT 1 FOR UPDATE`, groupID).Scan(&windowID)
	if errors.Is(err, pgx.ErrNoRows) {
		if !eligible {
			return nil
		}
		err = tx.QueryRow(ctx, `INSERT INTO alert_windows
			(error_group_id,first_interpretation_id,closes_at,first_occurred_at,last_occurred_at)
			VALUES ($1,$2,now()+$3::interval,$4,$4) RETURNING id`,
			groupID, interpretationID, window.String(), job.OccurredAt).Scan(&windowID)
		if err != nil {
			return fmt.Errorf("create alert window: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO notification_jobs (kind,alert_window_id) VALUES ('event_alert',$1)`, windowID); err != nil {
			return fmt.Errorf("enqueue event alert: %w", err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO notification_jobs (kind,alert_window_id,available_at)
			SELECT 'group_summary',id,closes_at FROM alert_windows WHERE id=$1`, windowID); err != nil {
			return fmt.Errorf("enqueue group summary: %w", err)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("find alert window: %w", err)
	}
	_, err = tx.Exec(ctx, `UPDATE alert_windows SET occurrence_count=occurrence_count+1,
		last_occurred_at=GREATEST(last_occurred_at,$2),updated_at=now() WHERE id=$1`, windowID, job.OccurredAt)
	if err != nil {
		return fmt.Errorf("update alert window: %w", err)
	}
	return nil
}

func (store *Postgres) FailInterpretationJob(ctx context.Context, job domain.InterpretationJob, cause error, retryable bool, alertCooldown time.Duration) error {
	status := "failed"
	availableAt := time.Now()
	if retryable && job.Attempts < job.MaxAttempts {
		status = "pending"
		delay := time.Minute * time.Duration(1<<(job.Attempts-1))
		availableAt = availableAt.Add(delay)
	}
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `UPDATE interpretation_jobs SET status=$2,available_at=$3,locked_at=NULL,locked_by=NULL,last_error=$4,updated_at=now() WHERE id=$1`, job.ID, status, availableAt, cause.Error()); err != nil {
		return err
	}
	if status == "failed" {
		bucket := time.Now().UTC().Truncate(alertCooldown).Format(time.RFC3339)
		key := "interpreter-degraded:" + bucket
		if _, err := tx.Exec(ctx, `INSERT INTO notification_jobs (kind,deduplication_key)
			VALUES ('interpreter_degraded',$1) ON CONFLICT DO NOTHING`, key); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (store *Postgres) ClaimNotificationJob(ctx context.Context, workerID string) (domain.NotificationJob, error) {
	if _, err := store.pool.Exec(ctx, `UPDATE notification_jobs j SET status='skipped',completed_at=now(),updated_at=now()
		FROM alert_windows w WHERE j.alert_window_id=w.id AND j.kind='group_summary' AND j.status='pending'
		AND j.available_at <= now() AND w.occurrence_count=1`); err != nil {
		return domain.NotificationJob{}, err
	}
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return domain.NotificationJob{}, err
	}
	defer tx.Rollback(ctx)
	var job domain.NotificationJob
	var actions []byte
	err = tx.QueryRow(ctx, `SELECT j.id,COALESCE(e.id::text,''),j.kind,COALESCE(a.slug,'atalaya'),COALESCE(src.component,''),COALESCE(src.source,'internal'),
		COALESCE(e.source_event_id,''),COALESCE(g.environment,''),COALESCE(g.error_type,''),COALESCE(e.release,''),
		COALESCE(i.summary,'Interpreter degradado'),COALESCE(i.explanation,'Una interpretación agotó sus reintentos. Los eventos siguen almacenándose.'),
		COALESCE(i.severity,'high'),COALESCE(i.actionable,true),COALESCE(i.suggested_actions,'["Revisar los logs del interpreter y OpenRouter"]'::jsonb),
		COALESCE(w.occurrence_count,1),COALESCE(w.first_occurred_at,j.created_at),COALESCE(w.last_occurred_at,j.created_at),j.attempts+1,j.max_attempts
		FROM notification_jobs j
		LEFT JOIN alert_windows w ON w.id=j.alert_window_id
		LEFT JOIN interpretations i ON i.id=w.first_interpretation_id
		LEFT JOIN error_events e ON e.id=i.error_event_id
		LEFT JOIN error_groups g ON g.id=e.error_group_id
		LEFT JOIN integrations src ON src.id=g.integration_id
		LEFT JOIN applications a ON a.id=src.application_id
		WHERE ((j.status='pending' AND j.available_at <= now()) OR (j.status='processing' AND j.locked_at < now()-interval '5 minutes'))
		AND (j.kind <> 'event_alert' OR (SELECT count(*) FROM notification_delivery_attempts da
			JOIN notification_jobs sent_job ON sent_job.id=da.notification_job_id
			JOIN alert_windows sent_window ON sent_window.id=sent_job.alert_window_id
			JOIN error_groups sent_group ON sent_group.id=sent_window.error_group_id
			JOIN integrations sent_source ON sent_source.id=sent_group.integration_id
			WHERE da.outcome='sent' AND sent_job.kind='event_alert' AND sent_source.application_id=src.application_id
			AND da.finished_at > now()-make_interval(secs => COALESCE((a.alert_policy->>'rate_limit_window_seconds')::int,600)))
			< COALESCE((a.alert_policy->>'rate_limit_count')::int,10))
		ORDER BY j.available_at,j.created_at FOR UPDATE OF j SKIP LOCKED LIMIT 1`).Scan(
		&job.ID, &job.EventID, &job.Kind, &job.Application, &job.Component, &job.Source, &job.SourceEventID, &job.Environment, &job.ErrorType, &job.Release,
		&job.Summary, &job.Explanation, &job.Severity, &job.Actionable, &actions, &job.OccurrenceCount,
		&job.FirstOccurredAt, &job.LastOccurredAt, &job.Attempts, &job.MaxAttempts)
	if err != nil {
		return domain.NotificationJob{}, err
	}
	if err := json.Unmarshal(actions, &job.SuggestedActions); err != nil {
		return domain.NotificationJob{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE notification_jobs SET status='processing',attempts=$3,locked_at=now(),locked_by=$2,updated_at=now() WHERE id=$1`, job.ID, workerID, job.Attempts); err != nil {
		return domain.NotificationJob{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.NotificationJob{}, err
	}
	return job, nil
}

func (store *Postgres) CompleteNotification(ctx context.Context, job domain.NotificationJob, started time.Time, result domain.DeliveryResult) error {
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO notification_delivery_attempts
		(notification_job_id,attempt_number,started_at,outcome,http_status,telegram_message_id)
		VALUES ($1,$2,$3,'sent',$4,$5)`, job.ID, job.Attempts, started, result.HTTPStatus, result.MessageID); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE notification_jobs SET status='completed',completed_at=now(),locked_at=NULL,locked_by=NULL,last_error=NULL,updated_at=now() WHERE id=$1`, job.ID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (store *Postgres) FailNotification(ctx context.Context, job domain.NotificationJob, started time.Time, cause error, retryable bool, httpStatus int) error {
	status, outcome := "failed", "permanent_failure"
	availableAt := time.Now()
	if retryable && job.Attempts < job.MaxAttempts {
		status, outcome = "pending", "retryable_failure"
		availableAt = availableAt.Add(time.Minute * time.Duration(1<<(job.Attempts-1)))
	}
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, `INSERT INTO notification_delivery_attempts
		(notification_job_id,attempt_number,started_at,outcome,http_status,error_class,error_message)
		VALUES ($1,$2,$3,$4,NULLIF($5,0),$6,$7)`, job.ID, job.Attempts, started, outcome, httpStatus, fmt.Sprintf("%T", cause), cause.Error()); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE notification_jobs SET status=$2,available_at=$3,locked_at=NULL,locked_by=NULL,last_error=$4,updated_at=now() WHERE id=$1`, job.ID, status, availableAt, cause.Error()); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type EventSummary struct {
	ID              string    `json:"id"`
	ErrorGroupID    string    `json:"error_group_id"`
	SourceEventID   string    `json:"source_event_id"`
	Application     string    `json:"application"`
	Component       string    `json:"component"`
	Environment     string    `json:"environment"`
	ErrorType       string    `json:"error_type"`
	Message         string    `json:"message"`
	OccurredAt      time.Time `json:"occurred_at"`
	IngestedAt      time.Time `json:"ingested_at"`
	Severity        string    `json:"severity"`
	State           string    `json:"state"`
	OccurrenceCount int64     `json:"occurrence_count"`
}

type EventFilter struct {
	Limit       int
	Offset      int
	Application string
	Component   string
	Severity    string
	State       string
	Since       *time.Time
}

type EventPage struct {
	Items []EventSummary `json:"items"`
	Total int            `json:"total"`
}

type IntegrationStatus struct {
	ID                  string     `json:"id"`
	Application         string     `json:"application"`
	Component           string     `json:"component"`
	DisplayName         string     `json:"display_name"`
	Source              string     `json:"source"`
	Project             string     `json:"project"`
	Enabled             bool       `json:"enabled"`
	Environments        []string   `json:"environments"`
	MonitoringStartedAt time.Time  `json:"monitoring_started_at"`
	LastAttemptAt       *time.Time `json:"last_attempt_at,omitempty"`
	LastSuccessAt       *time.Time `json:"last_success_at,omitempty"`
	LastError           *string    `json:"last_error,omitempty"`
	Status              string     `json:"status"`
}

type EventDetail struct {
	EventSummary
	Fingerprint    string                 `json:"fingerprint"`
	StackTrace     string                 `json:"stack_trace,omitempty"`
	Release        string                 `json:"release,omitempty"`
	Metadata       map[string]any         `json:"metadata"`
	Interpretation *domain.Interpretation `json:"interpretation,omitempty"`
	Occurrences    []EventOccurrence      `json:"occurrences"`
}

type EventOccurrence struct {
	ID            string    `json:"id"`
	SourceEventID string    `json:"source_event_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	Message       string    `json:"message"`
}

func (store *Postgres) ListEvents(ctx context.Context, filter EventFilter) ([]EventSummary, error) {
	page, err := store.ListEventPage(ctx, filter)
	return page.Items, err
}

func (store *Postgres) ListEventPage(ctx context.Context, filter EventFilter) (EventPage, error) {
	where := `($3='' OR a.slug=$3) AND ($4='' OR i.component=$4)
		AND ($5='' OR COALESCE(latest.severity,'pending')=$5)
		AND ($6='' OR CASE WHEN latest.severity IS NULL THEN 'pending' WHEN latest.actionable THEN 'actionable' ELSE 'noise' END=$6)
		AND ($7::timestamptz IS NULL OR e.occurred_at >= $7)`
	joins := `FROM error_events e JOIN error_groups g ON g.id=e.error_group_id JOIN integrations i ON i.id=e.integration_id JOIN applications a ON a.id=i.application_id
		LEFT JOIN LATERAL (SELECT severity,actionable FROM interpretations it WHERE it.error_event_id=e.id ORDER BY it.created_at DESC LIMIT 1) latest ON true`
	query := `SELECT e.id,g.id,e.source_event_id,a.slug,i.component,g.environment,e.error_type,e.message,e.occurred_at,e.ingested_at,
		COALESCE(latest.severity,'pending'),CASE WHEN latest.severity IS NULL THEN 'pending' WHEN latest.actionable THEN 'actionable' ELSE 'noise' END,g.occurrence_count ` + joins + ` WHERE ` + where + `
		ORDER BY e.occurred_at DESC LIMIT $1 OFFSET $2`
	rows, err := store.pool.Query(ctx, query, filter.Limit, filter.Offset, filter.Application, filter.Component, filter.Severity, filter.State, filter.Since)
	if err != nil {
		return EventPage{}, err
	}
	defer rows.Close()
	items := make([]EventSummary, 0)
	for rows.Next() {
		var item EventSummary
		if err := rows.Scan(&item.ID, &item.ErrorGroupID, &item.SourceEventID, &item.Application, &item.Component, &item.Environment, &item.ErrorType, &item.Message, &item.OccurredAt, &item.IngestedAt, &item.Severity, &item.State, &item.OccurrenceCount); err != nil {
			return EventPage{}, err
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return EventPage{}, err
	}
	var total int
	countWhere := `($1='' OR a.slug=$1) AND ($2='' OR i.component=$2)
		AND ($3='' OR COALESCE(latest.severity,'pending')=$3)
		AND ($4='' OR CASE WHEN latest.severity IS NULL THEN 'pending' WHEN latest.actionable THEN 'actionable' ELSE 'noise' END=$4)
		AND ($5::timestamptz IS NULL OR e.occurred_at >= $5)`
	countQuery := `SELECT count(*) ` + joins + ` WHERE ` + countWhere
	if err := store.pool.QueryRow(ctx, countQuery, filter.Application, filter.Component, filter.Severity, filter.State, filter.Since).Scan(&total); err != nil {
		return EventPage{}, err
	}
	return EventPage{Items: items, Total: total}, nil
}

func (store *Postgres) Event(ctx context.Context, id string) (EventDetail, error) {
	var item EventDetail
	var metadata []byte
	var groupID string
	err := store.pool.QueryRow(ctx, `SELECT e.id,g.id,e.source_event_id,a.slug,i.component,g.environment,e.error_type,e.message,e.occurred_at,e.ingested_at,
		g.fingerprint,COALESCE(e.stack_trace,''),COALESCE(e.release,''),e.metadata,g.id
		FROM error_events e JOIN error_groups g ON g.id=e.error_group_id JOIN integrations i ON i.id=e.integration_id JOIN applications a ON a.id=i.application_id WHERE e.id=$1`, id).
		Scan(&item.ID, &item.ErrorGroupID, &item.SourceEventID, &item.Application, &item.Component, &item.Environment, &item.ErrorType, &item.Message, &item.OccurredAt, &item.IngestedAt, &item.Fingerprint, &item.StackTrace, &item.Release, &metadata, &groupID)
	if err != nil {
		return EventDetail{}, err
	}
	if err := json.Unmarshal(metadata, &item.Metadata); err != nil {
		return EventDetail{}, err
	}
	rows, err := store.pool.Query(ctx, `SELECT id,source_event_id,occurred_at,message FROM error_events WHERE error_group_id=$1 ORDER BY occurred_at DESC LIMIT 50`, groupID)
	if err != nil {
		return EventDetail{}, err
	}
	defer rows.Close()
	item.Occurrences = make([]EventOccurrence, 0)
	for rows.Next() {
		var occurrence EventOccurrence
		if err := rows.Scan(&occurrence.ID, &occurrence.SourceEventID, &occurrence.OccurredAt, &occurrence.Message); err != nil {
			return EventDetail{}, err
		}
		item.Occurrences = append(item.Occurrences, occurrence)
	}
	if err := rows.Err(); err != nil {
		return EventDetail{}, err
	}
	var interpretation domain.Interpretation
	var actions []byte
	err = store.pool.QueryRow(ctx, `SELECT summary,explanation,severity,actionable,suggested_actions,model,prompt_version,
		input_tokens,output_tokens,total_tokens,estimated_cost_usd,latency_ms
		FROM interpretations WHERE error_event_id=$1 ORDER BY created_at DESC LIMIT 1`, id).Scan(
		&interpretation.Summary, &interpretation.Explanation, &interpretation.Severity, &interpretation.Actionable,
		&actions, &interpretation.Model, &interpretation.PromptVersion, &interpretation.Usage.InputTokens,
		&interpretation.Usage.OutputTokens, &interpretation.Usage.TotalTokens, &interpretation.EstimatedCostUSD, &interpretation.LatencyMS)
	if err == nil {
		if err := json.Unmarshal(actions, &interpretation.SuggestedActions); err != nil {
			return EventDetail{}, err
		}
		item.Interpretation = &interpretation
	} else if err != pgx.ErrNoRows {
		return EventDetail{}, err
	}
	return item, nil
}

func (store *Postgres) ListIntegrations(ctx context.Context) ([]IntegrationStatus, error) {
	rows, err := store.pool.Query(ctx, `SELECT i.id,a.slug,i.component,i.display_name,i.source,
		COALESCE(i.external_identifier->>'project',''),(a.enabled AND i.enabled),i.environment_filters,
		c.monitoring_started_at,c.last_attempt_at,c.last_success_at,c.last_error
		FROM integrations i JOIN applications a ON a.id=i.application_id
		JOIN source_checkpoints c ON c.integration_id=i.id
		ORDER BY a.slug,i.component`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]IntegrationStatus, 0)
	for rows.Next() {
		var item IntegrationStatus
		if err := rows.Scan(&item.ID, &item.Application, &item.Component, &item.DisplayName, &item.Source,
			&item.Project, &item.Enabled, &item.Environments, &item.MonitoringStartedAt,
			&item.LastAttemptAt, &item.LastSuccessAt, &item.LastError); err != nil {
			return nil, err
		}
		switch {
		case !item.Enabled:
			item.Status = "disabled"
		case item.LastAttemptAt == nil:
			item.Status = "never_run"
		case item.LastError != nil:
			item.Status = "error"
		default:
			item.Status = "ok"
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Postgres) GetCostSummary(ctx context.Context, monthlyBudgetUSD float64) (domain.CostSummary, error) {
	var summary domain.CostSummary
	summary.MonthlyBudgetUSD = monthlyBudgetUSD

	// Global metrics
	err := store.pool.QueryRow(ctx, `
		SELECT
			COALESCE(SUM(estimated_cost_usd), 0),
			COALESCE(SUM(total_tokens), 0),
			COALESCE(SUM(input_tokens), 0),
			COALESCE(SUM(output_tokens), 0),
			COUNT(*),
			COALESCE(AVG(latency_ms), 0)::bigint
		FROM interpretations
	`).Scan(&summary.TotalCostUSD, &summary.TotalTokens, &summary.InputTokens, &summary.OutputTokens, &summary.TotalRequests, &summary.AverageLatencyMS)
	if err != nil {
		return summary, fmt.Errorf("query total cost metrics: %w", err)
	}

	// Monthly metrics (current month)
	err = store.pool.QueryRow(ctx, `
		SELECT COALESCE(SUM(estimated_cost_usd), 0)
		FROM interpretations
		WHERE created_at >= date_trunc('month', NOW())
	`).Scan(&summary.MonthlyCostUSD)
	if err != nil {
		return summary, fmt.Errorf("query monthly cost metrics: %w", err)
	}

	if summary.MonthlyBudgetUSD > 0 {
		summary.BudgetUsedPercent = (summary.MonthlyCostUSD / summary.MonthlyBudgetUSD) * 100.0
	}

	// By Application
	appRows, err := store.pool.Query(ctx, `
		SELECT
			a.slug,
			COALESCE(SUM(i.total_tokens), 0),
			COALESCE(SUM(i.estimated_cost_usd), 0),
			COUNT(i.id)
		FROM applications a
		LEFT JOIN integrations src ON src.application_id = a.id
		LEFT JOIN error_events e ON e.integration_id = src.id
		LEFT JOIN interpretations i ON i.error_event_id = e.id
		GROUP BY a.slug
		ORDER BY COALESCE(SUM(i.estimated_cost_usd), 0) DESC
	`)
	if err != nil {
		return summary, fmt.Errorf("query cost by app: %w", err)
	}
	defer appRows.Close()

	summary.ByApplication = make([]domain.ApplicationCostBreakdown, 0)
	for appRows.Next() {
		var item domain.ApplicationCostBreakdown
		if err := appRows.Scan(&item.Application, &item.TotalTokens, &item.EstimatedCostUSD, &item.RequestCount); err != nil {
			return summary, err
		}
		summary.ByApplication = append(summary.ByApplication, item)
	}

	// By Model
	modelRows, err := store.pool.Query(ctx, `
		SELECT
			model,
			COALESCE(SUM(total_tokens), 0),
			COALESCE(SUM(estimated_cost_usd), 0),
			COUNT(id)
		FROM interpretations
		WHERE model IS NOT NULL AND model != ''
		GROUP BY model
		ORDER BY COALESCE(SUM(estimated_cost_usd), 0) DESC
	`)
	if err != nil {
		return summary, fmt.Errorf("query cost by model: %w", err)
	}
	defer modelRows.Close()

	summary.ByModel = make([]domain.ModelCostBreakdown, 0)
	for modelRows.Next() {
		var item domain.ModelCostBreakdown
		if err := modelRows.Scan(&item.Model, &item.TotalTokens, &item.EstimatedCostUSD, &item.RequestCount); err != nil {
			return summary, err
		}
		summary.ByModel = append(summary.ByModel, item)
	}

	return summary, nil
}

func (store *Postgres) PurgeOldEvents(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		return 0, nil
	}
	res, err := store.pool.Exec(ctx, `
		DELETE FROM error_events
		WHERE occurred_at < NOW() - ($1 * INTERVAL '1 day')
		  AND error_group_id NOT IN (SELECT DISTINCT error_group_id FROM incident_error_groups)
	`, retentionDays)
	if err != nil {
		return 0, fmt.Errorf("purge old events: %w", err)
	}
	return res.RowsAffected(), nil
}
