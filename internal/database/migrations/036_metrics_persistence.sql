-- +goose Up
CREATE TABLE persistent_counters (
    metric_name TEXT    NOT NULL,
    label_key   TEXT    NOT NULL DEFAULT '',
    value       BIGINT  NOT NULL DEFAULT 0,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    PRIMARY KEY (metric_name, label_key)
);

-- +goose Down
DROP TABLE IF EXISTS persistent_counters;
