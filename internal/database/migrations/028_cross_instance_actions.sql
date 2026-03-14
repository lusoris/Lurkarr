-- +goose Up
CREATE TABLE cross_instance_actions (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    group_id    UUID NOT NULL REFERENCES instance_groups(id) ON DELETE CASCADE,
    external_id TEXT NOT NULL,
    title       TEXT NOT NULL,
    action      TEXT NOT NULL,  -- 'auto_approved', 'declined_duplicate', 'skipped_higher_exists'
    reason      TEXT NOT NULL DEFAULT '',
    seerr_request_id INTEGER,
    source_instance_id UUID REFERENCES app_instances(id) ON DELETE SET NULL,
    target_instance_id UUID REFERENCES app_instances(id) ON DELETE SET NULL,
    executed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_cross_instance_actions_group ON cross_instance_actions(group_id);
CREATE INDEX idx_cross_instance_actions_executed ON cross_instance_actions(executed_at DESC);

-- +goose Down
DROP TABLE IF EXISTS cross_instance_actions;
