-- +goose Up

CREATE TABLE alert_windows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    error_group_id uuid NOT NULL REFERENCES error_groups(id) ON DELETE CASCADE,
    first_interpretation_id uuid NOT NULL REFERENCES interpretations(id),
    opened_at timestamptz NOT NULL DEFAULT now(),
    closes_at timestamptz NOT NULL,
    first_occurred_at timestamptz NOT NULL,
    last_occurred_at timestamptz NOT NULL,
    occurrence_count integer NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT alert_windows_time_order CHECK (closes_at > opened_at),
    CONSTRAINT alert_windows_occurrence_count_positive CHECK (occurrence_count > 0),
    CONSTRAINT alert_windows_occurrence_order CHECK (last_occurred_at >= first_occurred_at)
);

CREATE INDEX alert_windows_active_idx
    ON alert_windows (error_group_id, closes_at DESC);

CREATE TABLE notification_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    kind text NOT NULL,
    alert_window_id uuid REFERENCES alert_windows(id) ON DELETE CASCADE,
    deduplication_key text,
    status text NOT NULL DEFAULT 'pending',
    attempts integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 5,
    available_at timestamptz NOT NULL DEFAULT now(),
    locked_at timestamptz,
    locked_by text,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    completed_at timestamptz,
    CONSTRAINT notification_jobs_kind
        CHECK (kind IN ('event_alert', 'group_summary', 'interpreter_degraded')),
    CONSTRAINT notification_jobs_status
        CHECK (status IN ('pending', 'processing', 'completed', 'failed', 'skipped')),
    CONSTRAINT notification_jobs_attempts_valid
        CHECK (attempts >= 0 AND max_attempts > 0 AND attempts <= max_attempts),
    CONSTRAINT notification_jobs_target
        CHECK (
            (kind IN ('event_alert', 'group_summary') AND alert_window_id IS NOT NULL)
            OR (kind = 'interpreter_degraded' AND alert_window_id IS NULL AND deduplication_key IS NOT NULL)
        )
);

CREATE UNIQUE INDEX notification_jobs_window_kind_unique
    ON notification_jobs (alert_window_id, kind)
    WHERE alert_window_id IS NOT NULL;

CREATE UNIQUE INDEX notification_jobs_deduplication_key_unique
    ON notification_jobs (deduplication_key)
    WHERE deduplication_key IS NOT NULL;

CREATE INDEX notification_jobs_claim_idx
    ON notification_jobs (available_at, created_at)
    WHERE status = 'pending';

CREATE TABLE notification_delivery_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    notification_job_id uuid NOT NULL REFERENCES notification_jobs(id) ON DELETE CASCADE,
    attempt_number integer NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL DEFAULT now(),
    outcome text NOT NULL,
    http_status integer,
    telegram_message_id bigint,
    error_class text,
    error_message text,
    CONSTRAINT notification_delivery_attempts_outcome
        CHECK (outcome IN ('sent', 'retryable_failure', 'permanent_failure')),
    CONSTRAINT notification_delivery_attempts_number_positive CHECK (attempt_number > 0),
    CONSTRAINT notification_delivery_attempts_time_order CHECK (finished_at >= started_at),
    CONSTRAINT notification_delivery_attempts_job_number_unique
        UNIQUE (notification_job_id, attempt_number)
);

CREATE INDEX notification_delivery_attempts_job_idx
    ON notification_delivery_attempts (notification_job_id, attempt_number DESC);

-- +goose Down

DROP TABLE notification_delivery_attempts;
DROP TABLE notification_jobs;
DROP TABLE alert_windows;
