-- +goose Up

-- ── Download Client Instances (multi-instance, replaces singleton sabnzbd_settings) ──

CREATE TABLE download_client_instances (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        TEXT NOT NULL,
    client_type TEXT NOT NULL,  -- qbittorrent, transmission, deluge, sabnzbd, nzbget
    url         TEXT NOT NULL DEFAULT '',
    api_key     TEXT NOT NULL DEFAULT '',  -- for sabnzbd/nzbget
    username    TEXT NOT NULL DEFAULT '',  -- for torrent clients
    password    TEXT NOT NULL DEFAULT '',  -- for torrent clients
    category    TEXT NOT NULL DEFAULT '',
    enabled     BOOLEAN DEFAULT true,
    timeout     INT DEFAULT 30,
    created_at  TIMESTAMPTZ DEFAULT now(),
    UNIQUE(name)
);

-- Migrate existing SABnzbd singleton settings into the new table (if configured).
INSERT INTO download_client_instances (name, client_type, url, api_key, category, enabled, timeout)
SELECT 'SABnzbd', 'sabnzbd', url, api_key, category, enabled, timeout
FROM sabnzbd_settings
WHERE id = 1 AND url != '';

-- Migrate existing per-app download_client_settings into instances (if configured).
INSERT INTO download_client_instances (name, client_type, url, username, password, enabled, timeout)
SELECT
    CASE client_type
        WHEN 'qbittorrent' THEN 'qBittorrent (' || app_type || ')'
        WHEN 'transmission' THEN 'Transmission (' || app_type || ')'
        WHEN 'deluge' THEN 'Deluge (' || app_type || ')'
        WHEN 'sabnzbd' THEN 'SABnzbd (' || app_type || ')'
        WHEN 'nzbget' THEN 'NZBGet (' || app_type || ')'
        ELSE client_type || ' (' || app_type || ')'
    END,
    client_type, url, username, password, enabled, timeout
FROM download_client_settings
WHERE url != '' AND enabled = true
ON CONFLICT (name) DO NOTHING;

-- +goose Down

DROP TABLE IF EXISTS download_client_instances;
