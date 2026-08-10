-- +goose Up

ALTER TABLE integrations DROP CONSTRAINT integrations_application_source_unique;

ALTER TABLE integrations
    ADD COLUMN component text,
    ADD COLUMN display_name text,
    ADD COLUMN environment_filters text[] NOT NULL DEFAULT ARRAY[]::text[];

UPDATE integrations
SET component = CASE WHEN external_identifier->>'project' LIKE '%-frontend' THEN 'frontend' ELSE 'backend' END,
    display_name = CASE WHEN external_identifier->>'project' LIKE '%-frontend' THEN 'Frontend' ELSE 'Backend' END,
    environment_filters = ARRAY['production'];

ALTER TABLE integrations
    ALTER COLUMN component SET NOT NULL,
    ALTER COLUMN display_name SET NOT NULL,
    ADD CONSTRAINT integrations_component_format CHECK (component ~ '^[a-z][a-z0-9_]*$'),
    ADD CONSTRAINT integrations_application_source_component_unique UNIQUE (application_id, source, component);

ALTER TABLE applications
    ADD COLUMN alert_policy jsonb NOT NULL DEFAULT '{
        "enabled": true,
        "always_alert_severities": ["critical", "high"],
        "actionable_alert_severities": ["medium"],
        "deduplication_window_seconds": 900,
        "rate_limit_window_seconds": 600,
        "rate_limit_count": 10
    }'::jsonb,
    ADD CONSTRAINT applications_alert_policy_object CHECK (jsonb_typeof(alert_policy) = 'object');

ALTER TABLE source_checkpoints ADD COLUMN monitoring_started_at timestamptz;

UPDATE source_checkpoints checkpoint
SET monitoring_started_at = integration.created_at
FROM integrations integration
WHERE integration.id = checkpoint.integration_id;

ALTER TABLE source_checkpoints ALTER COLUMN monitoring_started_at SET NOT NULL;

-- +goose Down

DELETE FROM integrations WHERE component <> 'backend';
ALTER TABLE source_checkpoints DROP COLUMN monitoring_started_at;
ALTER TABLE applications DROP CONSTRAINT applications_alert_policy_object, DROP COLUMN alert_policy;
ALTER TABLE integrations
    DROP CONSTRAINT integrations_application_source_component_unique,
    DROP CONSTRAINT integrations_component_format,
    DROP COLUMN environment_filters,
    DROP COLUMN display_name,
    DROP COLUMN component,
    ADD CONSTRAINT integrations_application_source_unique UNIQUE (application_id, source);
