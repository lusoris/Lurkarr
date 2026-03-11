-- +goose Up

-- Community blocklist sources (configurable URLs).
CREATE TABLE blocklist_sources (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name                TEXT NOT NULL DEFAULT '',
    url                 TEXT NOT NULL,
    enabled             BOOLEAN DEFAULT true,
    sync_interval_hours INT DEFAULT 24,
    last_synced_at      TIMESTAMPTZ,
    etag                TEXT DEFAULT '',
    created_at          TIMESTAMPTZ DEFAULT NOW()
);

-- Blocklist rules (from community sources or manual).
CREATE TABLE blocklist_rules (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    source_id    UUID REFERENCES blocklist_sources(id) ON DELETE CASCADE,
    pattern      TEXT NOT NULL,
    pattern_type TEXT NOT NULL DEFAULT 'release_group'
                    CHECK (pattern_type IN ('release_group', 'title_contains', 'title_regex', 'indexer')),
    reason       TEXT DEFAULT '',
    enabled      BOOLEAN DEFAULT true,
    created_at   TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_blocklist_rules_enabled ON blocklist_rules (enabled) WHERE enabled = true;
CREATE INDEX idx_blocklist_rules_source ON blocklist_rules (source_id);

-- +goose Down

DROP TABLE IF EXISTS blocklist_rules;
DROP TABLE IF EXISTS blocklist_sources;
