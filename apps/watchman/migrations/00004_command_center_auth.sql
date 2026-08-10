-- +goose Up

CREATE TABLE command_center_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    token_hash bytea NOT NULL UNIQUE,
    created_at timestamptz NOT NULL DEFAULT now(),
    expires_at timestamptz NOT NULL,
    last_seen_at timestamptz NOT NULL DEFAULT now(),
    revoked_at timestamptz,
    CONSTRAINT command_center_sessions_expiry CHECK (expires_at > created_at)
);

CREATE INDEX command_center_sessions_active_idx
    ON command_center_sessions (expires_at)
    WHERE revoked_at IS NULL;

-- +goose Down

DROP TABLE command_center_sessions;
