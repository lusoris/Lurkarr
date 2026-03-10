-- +goose Up

-- Prowlarr connection settings (singleton)
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

-- SABnzbd connection settings (singleton)
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

-- Add prowlarr to app_settings and hunt_stats for consistency
INSERT INTO app_settings (app_type) VALUES ('prowlarr') ON CONFLICT DO NOTHING;
INSERT INTO hunt_stats (app_type) VALUES ('prowlarr') ON CONFLICT DO NOTHING;

-- +goose Down
DELETE FROM hunt_stats WHERE app_type = 'prowlarr';
DELETE FROM app_settings WHERE app_type = 'prowlarr';
DROP TABLE IF EXISTS sabnzbd_settings;
DROP TABLE IF EXISTS prowlarr_settings;
