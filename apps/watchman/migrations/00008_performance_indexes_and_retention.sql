-- +goose Up

CREATE INDEX IF NOT EXISTS events_app_occurred_idx ON events (application, occurred_at DESC);
CREATE INDEX IF NOT EXISTS events_fingerprint_idx ON events (fingerprint);
CREATE INDEX IF NOT EXISTS interpretations_created_idx ON interpretations (created_at DESC);
CREATE INDEX IF NOT EXISTS deployments_app_time_idx ON deployments (application, deployed_at DESC);
CREATE INDEX IF NOT EXISTS error_groups_app_last_seen_idx ON error_groups (application, last_seen_at DESC);
CREATE INDEX IF NOT EXISTS job_queue_status_scheduled_idx ON job_queue (status, scheduled_at);

-- +goose Down

DROP INDEX IF EXISTS job_queue_status_scheduled_idx;
DROP INDEX IF EXISTS error_groups_app_last_seen_idx;
DROP INDEX IF EXISTS deployments_app_time_idx;
DROP INDEX IF EXISTS interpretations_created_idx;
DROP INDEX IF EXISTS events_fingerprint_idx;
DROP INDEX IF EXISTS events_app_occurred_idx;
