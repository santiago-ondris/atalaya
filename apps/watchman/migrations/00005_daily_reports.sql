-- +goose Up

CREATE TABLE daily_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    report_date date NOT NULL UNIQUE,
    timezone text NOT NULL,
    period_start timestamptz NOT NULL,
    period_end timestamptz NOT NULL,
    status text NOT NULL DEFAULT 'collecting',
    attempts integer NOT NULL DEFAULT 0,
    max_attempts integer NOT NULL DEFAULT 8,
    next_attempt_at timestamptz NOT NULL DEFAULT now(),
    telegram_message_id bigint,
    last_error text,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    sent_at timestamptz,
    expired_at timestamptz,
    CONSTRAINT daily_reports_period_order CHECK (period_end > period_start),
    CONSTRAINT daily_reports_status CHECK (status IN ('collecting', 'pending', 'processing', 'sent', 'expired')),
    CONSTRAINT daily_reports_attempts_valid CHECK (attempts >= 0 AND max_attempts > 0 AND attempts <= max_attempts),
    CONSTRAINT daily_reports_terminal_time CHECK (
        (status = 'sent' AND sent_at IS NOT NULL AND expired_at IS NULL)
        OR (status = 'expired' AND expired_at IS NOT NULL AND sent_at IS NULL)
        OR (status NOT IN ('sent', 'expired') AND sent_at IS NULL AND expired_at IS NULL)
    )
);

CREATE INDEX daily_reports_claim_idx
    ON daily_reports (next_attempt_at, created_at)
    WHERE status IN ('pending', 'processing');

CREATE TABLE daily_report_applications (
    report_id uuid NOT NULL REFERENCES daily_reports(id) ON DELETE CASCADE,
    application_id uuid NOT NULL REFERENCES applications(id),
    activity_count bigint,
    activity_kind text NOT NULL,
    activity_source text NOT NULL,
    activity_status text NOT NULL,
    activity_error text,
    error_count bigint NOT NULL,
    occurrence_count bigint NOT NULL,
    critical_count bigint NOT NULL,
    high_count bigint NOT NULL,
    medium_count bigint NOT NULL,
    low_count bigint NOT NULL,
    actionable_count bigint NOT NULL,
    PRIMARY KEY (report_id, application_id),
    CONSTRAINT daily_report_activity_kind CHECK (activity_kind IN ('sessions', 'page_views')),
    CONSTRAINT daily_report_activity_status CHECK (activity_status IN ('available', 'unavailable')),
    CONSTRAINT daily_report_counts_nonnegative CHECK (
        (activity_count IS NULL OR activity_count >= 0)
        AND error_count >= 0 AND occurrence_count >= 0
        AND critical_count >= 0 AND high_count >= 0
        AND medium_count >= 0 AND low_count >= 0 AND actionable_count >= 0
    )
);

CREATE TABLE daily_report_delivery_attempts (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    report_id uuid NOT NULL REFERENCES daily_reports(id) ON DELETE CASCADE,
    attempt_number integer NOT NULL,
    started_at timestamptz NOT NULL,
    finished_at timestamptz NOT NULL DEFAULT now(),
    outcome text NOT NULL,
    http_status integer,
    telegram_message_id bigint,
    error_class text,
    error_message text,
    CONSTRAINT daily_report_delivery_outcome CHECK (outcome IN ('sent', 'retryable_failure', 'permanent_failure')),
    CONSTRAINT daily_report_delivery_attempt_unique UNIQUE (report_id, attempt_number)
);

CREATE INDEX daily_report_delivery_attempts_report_idx
    ON daily_report_delivery_attempts (report_id, attempt_number DESC);

-- +goose Down

DROP TABLE daily_report_delivery_attempts;
DROP TABLE daily_report_applications;
DROP TABLE daily_reports;
