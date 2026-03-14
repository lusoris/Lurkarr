-- +goose Up

CREATE TABLE notification_history (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    provider_id   UUID REFERENCES notification_providers(id) ON DELETE SET NULL,
    provider_type TEXT NOT NULL,
    provider_name TEXT NOT NULL DEFAULT '',
    event_type    TEXT NOT NULL,
    title         TEXT NOT NULL,
    message       TEXT NOT NULL DEFAULT '',
    app_type      TEXT NOT NULL DEFAULT '',
    instance      TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'sent',
    error         TEXT NOT NULL DEFAULT '',
    duration_ms   INT  NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_history_created  ON notification_history (created_at DESC);
CREATE INDEX idx_notification_history_status   ON notification_history (status);

-- +goose Down

DROP TABLE IF EXISTS notification_history;
