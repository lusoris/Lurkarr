-- +goose Up
CREATE TABLE IF NOT EXISTS seerr_settings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url         TEXT NOT NULL DEFAULT '',
    api_key     TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT false,
    sync_interval_minutes INTEGER NOT NULL DEFAULT 30,
    auto_approve BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Seed a single settings row.
INSERT INTO seerr_settings (id) VALUES (gen_random_uuid())
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS seerr_settings;
