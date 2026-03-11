-- +goose Up

-- Seeding rule columns for torrent download management (Phase 7).
ALTER TABLE queue_cleaner_settings
    ADD COLUMN seeding_enabled           BOOLEAN DEFAULT false,
    ADD COLUMN seeding_max_ratio         DOUBLE PRECISION DEFAULT 0,
    ADD COLUMN seeding_max_hours         INT DEFAULT 0,
    ADD COLUMN seeding_mode              TEXT DEFAULT 'or' CHECK (seeding_mode IN ('and', 'or')),
    ADD COLUMN seeding_delete_files      BOOLEAN DEFAULT false,
    ADD COLUMN seeding_skip_private      BOOLEAN DEFAULT true;

-- Orphan download cleanup columns.
ALTER TABLE queue_cleaner_settings
    ADD COLUMN orphan_enabled            BOOLEAN DEFAULT false,
    ADD COLUMN orphan_grace_minutes      INT DEFAULT 120,
    ADD COLUMN orphan_delete_files       BOOLEAN DEFAULT false,
    ADD COLUMN orphan_excluded_categories TEXT DEFAULT '';

-- Hardlink protection: skip file deletion when files have hardlinks.
ALTER TABLE queue_cleaner_settings
    ADD COLUMN hardlink_protection       BOOLEAN DEFAULT false;

-- Cross-seed awareness: skip removal when multiple torrents share the same content.
ALTER TABLE queue_cleaner_settings
    ADD COLUMN skip_cross_seeds          BOOLEAN DEFAULT true;

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
    DROP COLUMN IF EXISTS seeding_skip_private,
    DROP COLUMN IF EXISTS orphan_enabled,
    DROP COLUMN IF EXISTS orphan_grace_minutes,
    DROP COLUMN IF EXISTS orphan_delete_files,
    DROP COLUMN IF EXISTS orphan_excluded_categories,
    DROP COLUMN IF EXISTS hardlink_protection,
    DROP COLUMN IF EXISTS skip_cross_seeds;

DROP TABLE IF EXISTS download_client_settings;
