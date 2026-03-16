-- +goose Up

CREATE INDEX IF NOT EXISTS idx_sessions_user_id
    ON sessions (user_id);

CREATE INDEX IF NOT EXISTS idx_schedule_executions_executed_at
    ON schedule_executions (executed_at DESC);

CREATE INDEX IF NOT EXISTS idx_notification_history_provider_type
    ON notification_history (provider_type);

-- +goose Down

DROP INDEX IF EXISTS idx_notification_history_provider_type;
DROP INDEX IF EXISTS idx_schedule_executions_executed_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
