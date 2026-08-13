-- +goose Up

CREATE INDEX IF NOT EXISTS error_groups_fingerprint_idx ON error_groups (fingerprint);
CREATE INDEX IF NOT EXISTS error_groups_env_idx ON error_groups (environment);
CREATE INDEX IF NOT EXISTS interpretations_created_idx ON interpretations (created_at DESC);
CREATE INDEX IF NOT EXISTS interpretations_model_idx ON interpretations (model);

-- +goose Down

DROP INDEX IF EXISTS interpretations_model_idx;
DROP INDEX IF EXISTS interpretations_created_idx;
DROP INDEX IF EXISTS error_groups_env_idx;
DROP INDEX IF EXISTS error_groups_fingerprint_idx;
