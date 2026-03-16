-- +goose Up
CREATE TABLE bazarr_settings (
    id         SERIAL PRIMARY KEY,
    url        TEXT NOT NULL DEFAULT '',
    api_key    TEXT NOT NULL DEFAULT '',
    enabled    BOOLEAN NOT NULL DEFAULT false,
    timeout    INTEGER NOT NULL DEFAULT 30,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO bazarr_settings (id) VALUES (1);

-- +goose Down
DROP TABLE IF EXISTS bazarr_settings;
