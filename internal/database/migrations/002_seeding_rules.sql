-- +goose Up

-- Seeding rule columns for torrent download management (Phase 7).
ALTER TABLE queue_cleaner_settings
    ADD COLUMN seeding_enabled           BOOLEAN DEFAULT false,
    ADD COLUMN seeding_max_ratio         DOUBLE PRECISION DEFAULT 0,
    ADD COLUMN seeding_max_hours         INT DEFAULT 0,
    ADD COLUMN seeding_mode              TEXT DEFAULT 'or' CHECK (seeding_mode IN ('and', 'or')),
    ADD COLUMN seeding_delete_files      BOOLEAN DEFAULT false,
    ADD COLUMN seeding_skip_private      BOOLEAN DEFAULT true;

-- Download client configuration per app type.
CREATE TABLE download_client_settings (
    app_type    TEXT PRIMARY KEY,
    client_type TEXT NOT NULL DEFAULT '',
    url         TEXT NOT NULL DEFAULT '',
    username    TEXT NOT NULL DEFAULT '',
    password    TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN DEFAULT false,
    timeout     INT DEFAULT 30
);

INSERT INTO download_client_settings (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros');

-- +goose Down

ALTER TABLE queue_cleaner_settings
    DROP COLUMN IF EXISTS seeding_enabled,
    DROP COLUMN IF EXISTS seeding_max_ratio,
    DROP COLUMN IF EXISTS seeding_max_hours,
    DROP COLUMN IF EXISTS seeding_mode,
    DROP COLUMN IF EXISTS seeding_delete_files,
    DROP COLUMN IF EXISTS seeding_skip_private;

DROP TABLE IF EXISTS download_client_settings;
