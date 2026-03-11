-- +goose Up

-- ── Users & Sessions ──────────────────────────────────────────────────────────

CREATE TABLE users (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username        TEXT UNIQUE NOT NULL,
    password        TEXT NOT NULL,
    totp_secret     TEXT,
    auth_provider   TEXT NOT NULL DEFAULT 'local',
    external_id     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ DEFAULT now(),
    updated_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_users_external ON users(auth_provider, external_id) WHERE external_id != '';

CREATE TABLE sessions (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID REFERENCES users(id) ON DELETE CASCADE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_sessions_expires ON sessions(expires_at);

-- ── App Instances & Settings ──────────────────────────────────────────────────

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
    lurk_missing_count      INT DEFAULT 1,
    lurk_upgrade_count      INT DEFAULT 0,
    lurk_missing_mode       TEXT DEFAULT 'items',
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

-- ── Lurking Engine ────────────────────────────────────────────────────────────

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

CREATE TABLE lurk_history (
    id              BIGSERIAL PRIMARY KEY,
    app_type        TEXT NOT NULL,
    instance_id     UUID REFERENCES app_instances(id) ON DELETE SET NULL,
    instance_name   TEXT NOT NULL,
    media_id        INT NOT NULL,
    media_title     TEXT NOT NULL,
    operation       TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_lurk_history_app ON lurk_history(app_type, created_at DESC);
CREATE INDEX idx_lurk_history_search ON lurk_history USING gin(to_tsvector('english', media_title));

CREATE TABLE lurk_stats (
    app_type    TEXT NOT NULL,
    instance_id UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    lurked      BIGINT DEFAULT 0,
    upgraded    BIGINT DEFAULT 0,
    updated_at  TIMESTAMPTZ DEFAULT now(),
    PRIMARY KEY (app_type, instance_id)
);

CREATE TABLE hourly_caps (
    app_type    TEXT NOT NULL,
    instance_id UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    hour_bucket TIMESTAMPTZ NOT NULL,
    api_hits    INT DEFAULT 0,
    PRIMARY KEY (app_type, instance_id, hour_bucket)
);

-- ── Scheduling ────────────────────────────────────────────────────────────────

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

-- ── Logging ───────────────────────────────────────────────────────────────────

CREATE TABLE logs (
    id          BIGSERIAL PRIMARY KEY,
    app_type    TEXT NOT NULL,
    level       TEXT NOT NULL,
    message     TEXT NOT NULL,
    created_at  TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_logs_app_created ON logs(app_type, created_at DESC);
CREATE INDEX idx_logs_created ON logs(created_at DESC);

-- ── Prowlarr & SABnzbd ───────────────────────────────────────────────────────

CREATE TABLE prowlarr_settings (
    id          INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    url         TEXT NOT NULL DEFAULT '',
    api_key     TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN DEFAULT false,
    sync_indexers BOOLEAN DEFAULT true,
    timeout     INT DEFAULT 30,
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);
INSERT INTO prowlarr_settings (id) VALUES (1);

CREATE TABLE sabnzbd_settings (
    id          INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    url         TEXT NOT NULL DEFAULT '',
    api_key     TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN DEFAULT false,
    timeout     INT DEFAULT 30,
    category    TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ DEFAULT now(),
    updated_at  TIMESTAMPTZ DEFAULT now()
);
INSERT INTO sabnzbd_settings (id) VALUES (1);

-- ── Queue Cleaner ─────────────────────────────────────────────────────────────

CREATE TABLE queue_cleaner_settings (
    app_type                        TEXT PRIMARY KEY,
    enabled                         BOOLEAN DEFAULT false,
    stalled_threshold_minutes       INT DEFAULT 30,
    slow_threshold_bytes_per_sec    BIGINT DEFAULT 0,
    max_strikes                     INT DEFAULT 3,
    strike_window_hours             INT DEFAULT 24,
    check_interval_seconds          INT DEFAULT 300,
    remove_from_client              BOOLEAN DEFAULT false,
    blocklist_on_remove             BOOLEAN DEFAULT true,
    strike_public                   BOOLEAN DEFAULT true,
    strike_private                  BOOLEAN DEFAULT false,
    slow_ignore_above_bytes         BIGINT DEFAULT 0,
    failed_import_remove            BOOLEAN DEFAULT false,
    failed_import_blocklist         BOOLEAN DEFAULT true,
    metadata_stuck_minutes          INT DEFAULT 0
);

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

-- ── Notifications ─────────────────────────────────────────────────────────────

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

-- ── Seerr ─────────────────────────────────────────────

CREATE TABLE seerr_settings (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    url         TEXT NOT NULL DEFAULT '',
    api_key     TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN NOT NULL DEFAULT false,
    sync_interval_minutes INTEGER NOT NULL DEFAULT 30,
    auto_approve BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
INSERT INTO seerr_settings (id) VALUES (gen_random_uuid());

-- ── Seed Data ─────────────────────────────────────────────────────────────────

INSERT INTO app_settings (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros'), ('prowlarr');

INSERT INTO queue_cleaner_settings (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

INSERT INTO scoring_profiles (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

-- +goose Down
DROP TABLE IF EXISTS seerr_settings;
DROP TABLE IF EXISTS notification_providers;
DROP TABLE IF EXISTS blocklist_log;
DROP TABLE IF EXISTS scoring_profiles;
DROP TABLE IF EXISTS auto_import_log;
DROP TABLE IF EXISTS queue_strikes;
DROP TABLE IF EXISTS queue_cleaner_settings;
DROP TABLE IF EXISTS sabnzbd_settings;
DROP TABLE IF EXISTS prowlarr_settings;
DROP TABLE IF EXISTS logs;
DROP TABLE IF EXISTS schedule_executions;
DROP TABLE IF EXISTS schedules;
DROP TABLE IF EXISTS hourly_caps;
DROP TABLE IF EXISTS lurk_stats;
DROP TABLE IF EXISTS lurk_history;
DROP TABLE IF EXISTS state_resets;
DROP TABLE IF EXISTS processed_items;
DROP TABLE IF EXISTS general_settings;
DROP TABLE IF EXISTS app_settings;
DROP TABLE IF EXISTS app_instances;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;
