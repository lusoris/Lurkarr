-- +goose Up

CREATE TABLE notification_providers (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type        TEXT NOT NULL,
    name        TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT false,
    config      JSONB NOT NULL DEFAULT '{}',
    events      TEXT[] NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_notification_providers_type ON notification_providers (type);
CREATE INDEX idx_notification_providers_enabled ON notification_providers (enabled);

-- +goose Down

DROP TABLE IF EXISTS notification_providers;
