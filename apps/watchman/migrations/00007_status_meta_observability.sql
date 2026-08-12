-- +goose Up

ALTER TABLE incidents
    ADD COLUMN public_title text,
    ADD COLUMN public_message text,
    ADD COLUMN published_at timestamptz;

ALTER TABLE incidents ADD CONSTRAINT incidents_publication_complete CHECK (
    (published_at IS NULL AND public_title IS NULL AND public_message IS NULL)
    OR (published_at IS NOT NULL
        AND char_length(public_title) BETWEEN 1 AND 160
        AND char_length(public_message) BETWEEN 1 AND 2000)
);

CREATE TABLE availability_targets (
    id text PRIMARY KEY,
    application text NOT NULL,
    display_name text NOT NULL,
    component text NOT NULL,
    url text NOT NULL,
    enabled boolean NOT NULL DEFAULT true,
    consecutive_successes integer NOT NULL DEFAULT 0,
    consecutive_failures integer NOT NULL DEFAULT 0,
    confirmed_status text NOT NULL DEFAULT 'unknown',
    status_changed_at timestamptz,
    last_checked_at timestamptz,
    CONSTRAINT availability_target_status CHECK (confirmed_status IN ('unknown','up','down')),
    CONSTRAINT availability_target_component CHECK (component IN ('frontend','backend'))
);

CREATE TABLE availability_probes (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    target_id text NOT NULL REFERENCES availability_targets(id) ON DELETE CASCADE,
    checked_at timestamptz NOT NULL,
    succeeded boolean NOT NULL,
    latency_ms integer NOT NULL,
    http_status integer,
    error text
);
CREATE INDEX availability_probes_target_time_idx ON availability_probes(target_id, checked_at DESC);

CREATE TABLE availability_snapshots (
    id bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    application text NOT NULL,
    status text NOT NULL,
    captured_at timestamptz NOT NULL,
    CONSTRAINT availability_snapshot_status CHECK (status IN ('unknown','operational','degraded','major_outage'))
);
CREATE INDEX availability_snapshots_app_time_idx ON availability_snapshots(application, captured_at DESC);

CREATE TABLE service_heartbeats (
    signal text PRIMARY KEY,
    last_seen_at timestamptz NOT NULL,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE health_signal_states (
    signal text PRIMARY KEY,
    unhealthy boolean NOT NULL,
    detail text NOT NULL DEFAULT '',
    changed_at timestamptz NOT NULL,
    alerted_at timestamptz
);

-- +goose Down
DROP TABLE health_signal_states;
DROP TABLE service_heartbeats;
DROP TABLE availability_snapshots;
DROP TABLE availability_probes;
DROP TABLE availability_targets;
ALTER TABLE incidents DROP CONSTRAINT incidents_publication_complete;
ALTER TABLE incidents DROP COLUMN published_at, DROP COLUMN public_message, DROP COLUMN public_title;
