-- +goose Up
CREATE TABLE instance_groups (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    app_type   TEXT NOT NULL,
    name       TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT now(),
    UNIQUE(app_type, name)
);

CREATE TABLE instance_group_members (
    group_id     UUID NOT NULL REFERENCES instance_groups(id) ON DELETE CASCADE,
    instance_id  UUID NOT NULL REFERENCES app_instances(id) ON DELETE CASCADE,
    quality_rank INT  NOT NULL DEFAULT 0,
    PRIMARY KEY (group_id, instance_id),
    UNIQUE(instance_id)
);

-- +goose Down
DROP TABLE instance_group_members;
DROP TABLE instance_groups;
