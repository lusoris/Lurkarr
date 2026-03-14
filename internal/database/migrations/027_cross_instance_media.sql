-- +goose Up
ALTER TABLE instance_groups ADD COLUMN mode TEXT NOT NULL DEFAULT 'quality_hierarchy';
ALTER TABLE instance_group_members ADD COLUMN is_independent BOOLEAN NOT NULL DEFAULT FALSE;

CREATE TABLE cross_instance_media (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES instance_groups(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    title       TEXT NOT NULL DEFAULT '',
    detected_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE cross_instance_presence (
    media_id    UUID NOT NULL REFERENCES cross_instance_media(id) ON DELETE CASCADE,
    instance_id UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    monitored   BOOLEAN NOT NULL DEFAULT TRUE,
    has_file    BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (media_id, instance_id)
);

CREATE INDEX idx_cross_instance_media_group ON cross_instance_media(group_id);
CREATE UNIQUE INDEX idx_cross_instance_media_ext ON cross_instance_media(group_id, external_id);

-- +goose Down
DROP TABLE IF EXISTS cross_instance_presence;
DROP TABLE IF EXISTS cross_instance_media;
ALTER TABLE instance_group_members DROP COLUMN IF EXISTS is_independent;
ALTER TABLE instance_groups DROP COLUMN IF EXISTS mode;
