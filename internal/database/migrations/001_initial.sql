-- +goose Up
CREATE TABLE users (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username    TEXT UNIQUE NOT NULL,
    password    TEXT NOT NULL,
    totp_secret TEXT,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

CREATE TABLE app_instances (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_type   TEXT NOT NULL,
    name       TEXT NOT NULL,
    api_url    TEXT NOT NULL,
    api_key    TEXT NOT NULL,
    enabled    BOOLEAN DEFAULT true,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(app_type, name)
);

CREATE TABLE app_settings (
    app_type                TEXT PRIMARY KEY,
    hunt_missing_count      INT DEFAULT 1,
    hunt_upgrade_count      INT DEFAULT 0,
    hunt_missing_mode       TEXT DEFAULT 'items',
    upgrade_mode            TEXT DEFAULT 'items',
    sleep_duration          INT DEFAULT 900,
    monitored_only          BOOLEAN DEFAULT true,
    skip_future             BOOLEAN DEFAULT true,
    hourly_cap              INT DEFAULT 20,
    random_selection        BOOLEAN DEFAULT true,
    debug_mode              BOOLEAN DEFAULT false,
    extra_settings          JSONB DEFAULT '{}'::jsonb
);

CREATE TABLE general_settings (
    id                          INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    secret_key                  TEXT NOT NULL,
    proxy_auth_bypass           BOOLEAN DEFAULT false,
    ssl_verify                  BOOLEAN DEFAULT true,
    api_timeout                 INT DEFAULT 120,
    stateful_reset_hours        INT DEFAULT 168,
    command_wait_delay          INT DEFAULT 1,
    command_wait_attempts       INT DEFAULT 600,
    min_download_queue_size     INT DEFAULT -1
);

CREATE TABLE processed_items (
    id              BIGSERIAL PRIMARY KEY,
    app_type        TEXT NOT NULL,
    instance_id     UUID REFERENCES app_instances(id) ON DELETE CASCADE,
    media_id        INT NOT NULL,
    operation       TEXT NOT NULL,
    processed_at    TIMESTAMPTZ DEFAULT now(),
    UNIQUE(app_type, instance_id, media_id, operation)
);
CREATE INDEX idx_processed_items_lookup ON processed_items(app_type, instance_id, operation);

CREATE TABLE state_resets (
    app_type    TEXT NOT NULL,
    instance_id UUID REFERENCES app_instances(id) ON DELETE CASCADE,
    last_reset  TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (app_type, instance_id)
);

CREATE TABLE hunt_history (
    id              BIGSERIAL PRIMARY KEY,
    app_type        TEXT NOT NULL,
    instance_id     UUID REFERENCES app_instances(id) ON DELETE SET NULL,
    instance_name   TEXT NOT NULL,
    media_id        INT NOT NULL,
    media_title     TEXT NOT NULL,
    operation       TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_hunt_history_app ON hunt_history(app_type, created_at DESC);
CREATE INDEX idx_hunt_history_search ON hunt_history USING gin(to_tsvector('english', media_title));

CREATE TABLE hunt_stats (
    app_type    TEXT PRIMARY KEY,
    hunted      BIGINT DEFAULT 0,
    upgraded    BIGINT DEFAULT 0,
    updated_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE hourly_caps (
    app_type    TEXT NOT NULL,
    hour_bucket TIMESTAMPTZ NOT NULL,
    api_hits    INT DEFAULT 0,
    PRIMARY KEY (app_type, hour_bucket)
);

CREATE TABLE schedules (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_type    TEXT NOT NULL,
    action      TEXT NOT NULL,
    days        TEXT[] DEFAULT '{}',
    hour        INT NOT NULL,
    minute      INT NOT NULL,
    enabled     BOOLEAN DEFAULT true,
    created_at  TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE schedule_executions (
    id          BIGSERIAL PRIMARY KEY,
    schedule_id UUID REFERENCES schedules(id) ON DELETE CASCADE,
    executed_at TIMESTAMPTZ DEFAULT now(),
    result      TEXT
);

CREATE TABLE logs (
    id          BIGSERIAL PRIMARY KEY,
    app_type    TEXT NOT NULL,
    level       TEXT NOT NULL,
    message     TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_logs_app_created ON logs(app_type, created_at DESC);
CREATE INDEX idx_logs_created ON logs(created_at DESC);

-- Seed default app settings for all app types
INSERT INTO app_settings (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

-- Seed default hunt stats
INSERT INTO hunt_stats (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

-- +goose Down
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS schedule_executions;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS hourly_caps;
DROP TABLE IF EXISTS hunt_stats;
DROP TABLE IF EXISTS hunt_history;
DROP TABLE IF EXISTS state_resets;
DROP TABLE IF EXISTS processed_items;
DROP TABLE IF EXISTS general_settings;
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS app_instances;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
