-- +goose Up
CREATE TABLE IF NOT EXISTS search_cooldowns (
    app_type TEXT NOT NULL,
    instance_id UUID NOT NULL,
    media_id INT NOT NULL,
    searched_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (app_type, instance_id, media_id)
);

ALTER TABLE queue_cleaner_settings ADD COLUMN search_cooldown_hours INT DEFAULT 0;

-- +goose Down
DROP TABLE IF EXISTS search_cooldowns;
ALTER TABLE queue_cleaner_settings DROP COLUMN search_cooldown_hours;
