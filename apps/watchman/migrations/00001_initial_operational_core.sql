-- +goose Up

CREATE TABLE applications (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    slug text NOT NULL UNIQUE,
    display_name text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT applications_slug_format
        CHECK (slug ~ '^[a-z][a-z0-9_]*$')
);

CREATE TABLE integrations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id uuid NOT NULL REFERENCES applications(id),
    source text NOT NULL,
    external_identifier jsonb NOT NULL DEFAULT '{}'::jsonb,
    enabled boolean NOT NULL DEFAULT true,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT integrations_source
        CHECK (source IN ('sentry', 'application_insights')),
    CONSTRAINT integrations_application_source_unique
        UNIQUE (application_id, source)
);

COMMENT ON COLUMN integrations.external_identifier IS
    'Identificadores no sensibles del proveedor; nunca tokens ni credenciales.';

CREATE TABLE source_checkpoints (
    integration_id uuid PRIMARY KEY REFERENCES integrations(id) ON DELETE CASCADE,
    cursor_data jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_attempt_at timestamptz,
    last_success_at timestamptz,
    last_error text,
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT source_checkpoints_success_after_attempt
        CHECK (
            last_success_at IS NULL
            OR last_attempt_at IS NULL
            OR last_success_at <= last_attempt_at
        )
);

CREATE TABLE error_groups (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id uuid NOT NULL REFERENCES integrations(id),
    environment text NOT NULL,
    fingerprint text NOT NULL,
    error_type text NOT NULL,
    sample_message text NOT NULL,
    first_seen_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL,
    occurrence_count bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT error_groups_occurrence_count_positive
        CHECK (occurrence_count > 0),
    CONSTRAINT error_groups_seen_order
        CHECK (last_seen_at >= first_seen_at),
    CONSTRAINT error_groups_identity_unique
        UNIQUE (integration_id, environment, fingerprint)
);

CREATE INDEX error_groups_recent_idx
    ON error_groups (integration_id, last_seen_at DESC);

CREATE TABLE error_events (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    integration_id uuid NOT NULL REFERENCES integrations(id),
    error_group_id uuid NOT NULL REFERENCES error_groups(id),
    source_event_id text NOT NULL,
    occurred_at timestamptz NOT NULL,
    ingested_at timestamptz NOT NULL DEFAULT now(),
    error_type text NOT NULL,
    message text NOT NULL,
    stack_trace text,
    release text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    CONSTRAINT error_events_source_identity_unique
        UNIQUE (integration_id, source_event_id)
);

CREATE INDEX error_events_group_occurred_idx
    ON error_events (error_group_id, occurred_at DESC);

CREATE INDEX error_events_occurred_idx
    ON error_events (occurred_at DESC);

CREATE TABLE interpretation_jobs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    error_event_id uuid NOT NULL REFERENCES error_events(id) ON DELETE CASCADE,
    prompt_version text NOT NULL,
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
    CONSTRAINT interpretation_jobs_status
        CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    CONSTRAINT interpretation_jobs_attempts_valid
        CHECK (attempts >= 0 AND max_attempts > 0 AND attempts <= max_attempts),
    CONSTRAINT interpretation_jobs_event_prompt_unique
        UNIQUE (error_event_id, prompt_version)
);

CREATE INDEX interpretation_jobs_claim_idx
    ON interpretation_jobs (available_at, created_at)
    WHERE status = 'pending';

CREATE TABLE interpretations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id uuid NOT NULL UNIQUE REFERENCES interpretation_jobs(id),
    error_event_id uuid NOT NULL REFERENCES error_events(id),
    summary text NOT NULL,
    explanation text NOT NULL,
    severity text NOT NULL,
    actionable boolean NOT NULL,
    suggested_actions jsonb NOT NULL DEFAULT '[]'::jsonb,
    model text NOT NULL,
    prompt_version text NOT NULL,
    input_tokens integer NOT NULL,
    output_tokens integer NOT NULL,
    total_tokens integer NOT NULL,
    estimated_cost_usd numeric(16, 10),
    latency_ms integer NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT interpretations_severity
        CHECK (severity IN ('critical', 'high', 'medium', 'low')),
    CONSTRAINT interpretations_token_counts
        CHECK (
            input_tokens >= 0
            AND output_tokens >= 0
            AND total_tokens = input_tokens + output_tokens
        ),
    CONSTRAINT interpretations_cost_nonnegative
        CHECK (estimated_cost_usd IS NULL OR estimated_cost_usd >= 0),
    CONSTRAINT interpretations_latency_nonnegative
        CHECK (latency_ms >= 0),
    CONSTRAINT interpretations_actions_array
        CHECK (jsonb_typeof(suggested_actions) = 'array')
);

INSERT INTO applications (slug, display_name)
VALUES
    ('farmami', 'Farmami'),
    ('wheels_house', 'Wheels House'),
    ('prensap', 'Prensap'),
    ('notizap', 'Notizap');

-- +goose Down

DROP TABLE interpretations;
DROP TABLE interpretation_jobs;
DROP TABLE error_events;
DROP TABLE error_groups;
DROP TABLE source_checkpoints;
DROP TABLE integrations;
DROP TABLE applications;

