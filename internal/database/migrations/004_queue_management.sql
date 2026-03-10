-- +goose Up

-- Queue cleaner settings (one row per app type, like app_settings)
CREATE TABLE queue_cleaner_settings (
    app_type                    TEXT PRIMARY KEY,
    enabled                     BOOLEAN DEFAULT false,
    stalled_threshold_minutes   INT DEFAULT 30,
    slow_threshold_bytes_per_sec BIGINT DEFAULT 0,
    max_strikes                 INT DEFAULT 3,
    strike_window_hours         INT DEFAULT 24,
    check_interval_seconds      INT DEFAULT 300,
    remove_from_client          BOOLEAN DEFAULT false,
    blocklist_on_remove         BOOLEAN DEFAULT true
);

-- Seed defaults for all huntable app types
INSERT INTO queue_cleaner_settings (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

-- Per-download strike tracking for stalled/slow detection
CREATE TABLE queue_strikes (
    id          BIGSERIAL PRIMARY KEY,
    app_type    TEXT NOT NULL,
    instance_id UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    download_id TEXT NOT NULL,
    title       TEXT NOT NULL,
    reason      TEXT NOT NULL,
    struck_at   TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_queue_strikes_lookup ON queue_strikes(app_type, instance_id, download_id, struck_at);

-- Auto-import action log
CREATE TABLE auto_import_log (
    id             BIGSERIAL PRIMARY KEY,
    app_type       TEXT NOT NULL,
    instance_id    UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    media_id       INT NOT NULL,
    media_title    TEXT NOT NULL,
    queue_item_id  INT NOT NULL,
    action         TEXT NOT NULL,
    reason         TEXT NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_auto_import_log_app ON auto_import_log(app_type, created_at DESC);

-- Scoring profiles for queue deduplication
CREATE TABLE scoring_profiles (
    id                      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_type                TEXT NOT NULL UNIQUE,
    name                    TEXT NOT NULL DEFAULT 'Default',
    strategy                TEXT NOT NULL DEFAULT 'highest',
    adequate_threshold      INT DEFAULT 0,
    prefer_higher_quality   BOOLEAN DEFAULT true,
    prefer_larger_size      BOOLEAN DEFAULT false,
    prefer_indexer_flags    BOOLEAN DEFAULT false,
    custom_format_weight    INT DEFAULT 100,
    size_weight             INT DEFAULT 10,
    age_weight              INT DEFAULT 5,
    seeders_weight          INT DEFAULT 5,
    created_at              TIMESTAMPTZ DEFAULT now()
);

-- Seed default scoring profiles
INSERT INTO scoring_profiles (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

-- Blocklist action log
CREATE TABLE blocklist_log (
    id              BIGSERIAL PRIMARY KEY,
    app_type        TEXT NOT NULL,
    instance_id     UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    download_id     TEXT NOT NULL,
    title           TEXT NOT NULL,
    reason          TEXT NOT NULL,
    blocklisted_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_blocklist_log_app ON blocklist_log(app_type, blocklisted_at DESC);

-- +goose Down
DROP TABLE IF EXISTS blocklist_log;
DROP TABLE IF EXISTS scoring_profiles;
DROP TABLE IF EXISTS auto_import_log;
DROP TABLE IF EXISTS queue_strikes;
DROP TABLE IF EXISTS queue_cleaner_settings;
