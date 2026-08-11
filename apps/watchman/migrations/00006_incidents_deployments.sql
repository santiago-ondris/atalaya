-- +goose Up

CREATE TABLE incidents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id uuid NOT NULL REFERENCES applications(id),
    title text NOT NULL,
    status text NOT NULL DEFAULT 'investigating',
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    closed_at timestamptz,
    CONSTRAINT incidents_title_length CHECK (char_length(title) BETWEEN 1 AND 160),
    CONSTRAINT incidents_status CHECK (status IN ('investigating', 'resolved', 'noise')),
    CONSTRAINT incidents_closed_state CHECK (
        (status = 'investigating' AND closed_at IS NULL)
        OR (status IN ('resolved', 'noise') AND closed_at IS NOT NULL)
    )
);

CREATE INDEX incidents_application_created_idx ON incidents (application_id, created_at DESC);

CREATE TABLE incident_error_groups (
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    error_group_id uuid NOT NULL REFERENCES error_groups(id),
    added_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (incident_id, error_group_id)
);

CREATE INDEX incident_error_groups_group_idx ON incident_error_groups (error_group_id);

CREATE TABLE incident_entries (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents(id) ON DELETE CASCADE,
    kind text NOT NULL,
    body text,
    from_status text,
    to_status text,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT incident_entries_kind CHECK (kind IN ('note', 'status_change', 'group_added', 'group_removed')),
    CONSTRAINT incident_entries_body_length CHECK (body IS NULL OR char_length(body) BETWEEN 1 AND 5000)
);

CREATE INDEX incident_entries_incident_idx ON incident_entries (incident_id, created_at, id);

CREATE TABLE deployments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    application_id uuid NOT NULL REFERENCES applications(id),
    component text NOT NULL,
    environment text NOT NULL,
    provider text NOT NULL,
    external_id text NOT NULL,
    version text,
    commit_sha text,
    commit_url text,
    actor text,
    source_url text,
    deployed_at timestamptz NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT deployments_component_format CHECK (component ~ '^[a-z][a-z0-9_]*$'),
    CONSTRAINT deployments_provider CHECK (provider IN ('railway', 'github_actions', 'manual')),
    CONSTRAINT deployments_reference CHECK (version IS NOT NULL OR commit_sha IS NOT NULL),
    CONSTRAINT deployments_provider_identity UNIQUE (provider, external_id)
);

CREATE INDEX deployments_timeline_idx ON deployments (application_id, deployed_at DESC);

-- +goose Down

DROP TABLE deployments;
DROP TABLE incident_entries;
DROP TABLE incident_error_groups;
DROP TABLE incidents;
