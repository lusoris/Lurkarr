-- +goose Up
CREATE TABLE split_season_rules (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id      UUID NOT NULL REFERENCES instance_groups(id) ON DELETE CASCADE,
    external_id   TEXT NOT NULL,
    title         TEXT NOT NULL DEFAULT '',
    instance_id   UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    season_from   INT NOT NULL,
    season_to     INT,
    created_at    TIMESTAMPTZ DEFAULT now(),
    UNIQUE(group_id, external_id, season_from)
);
CREATE INDEX idx_split_season_rules_group ON split_season_rules(group_id);
CREATE INDEX idx_split_season_rules_ext ON split_season_rules(external_id);

-- +goose Down
DROP TABLE IF EXISTS split_season_rules;
