-- +goose Up
CREATE TABLE seeding_rule_groups (
    id              SERIAL PRIMARY KEY,
    name            TEXT NOT NULL,
    priority        INTEGER NOT NULL DEFAULT 0,
    match_type      TEXT NOT NULL CHECK (match_type IN ('tracker', 'category', 'tag')),
    match_pattern   TEXT NOT NULL,
    max_ratio       DOUBLE PRECISION NOT NULL DEFAULT 0,
    max_hours       INTEGER NOT NULL DEFAULT 0,
    seeding_mode    TEXT NOT NULL DEFAULT 'or' CHECK (seeding_mode IN ('and', 'or')),
    skip_removal    BOOLEAN NOT NULL DEFAULT FALSE,
    delete_files    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_seeding_rule_groups_priority ON seeding_rule_groups (priority DESC);

-- +goose Down
DROP TABLE IF EXISTS seeding_rule_groups;
