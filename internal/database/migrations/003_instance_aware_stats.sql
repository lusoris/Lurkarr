-- +goose Up

-- Rebuild hunt_stats to be per-instance instead of per-app-type.
-- Drop existing aggregate rows (early dev, no production data to preserve).
DELETE FROM hunt_stats;

ALTER TABLE hunt_stats
    DROP CONSTRAINT hunt_stats_pkey,
    ADD COLUMN instance_id UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    ADD PRIMARY KEY (app_type, instance_id);

-- Rebuild hourly_caps to be per-instance.
DELETE FROM hourly_caps;

ALTER TABLE hourly_caps
    DROP CONSTRAINT hourly_caps_pkey,
    ADD COLUMN instance_id UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    ADD PRIMARY KEY (app_type, instance_id, hour_bucket);

-- +goose Down
DELETE FROM hourly_caps;
ALTER TABLE hourly_caps
    DROP CONSTRAINT hourly_caps_pkey,
    DROP COLUMN instance_id,
    ADD PRIMARY KEY (app_type, hour_bucket);

DELETE FROM hunt_stats;
ALTER TABLE hunt_stats
    DROP CONSTRAINT hunt_stats_pkey,
    DROP COLUMN instance_id,
    ADD PRIMARY KEY (app_type);

-- Re-seed default hunt stats
INSERT INTO hunt_stats (app_type) VALUES
    ('sonarr'), ('radarr'), ('lidarr'), ('readarr'), ('whisparr'), ('eros'), ('prowlarr');
