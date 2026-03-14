-- +goose Up
ALTER TABLE queue_cleaner_settings
    ADD COLUMN recheck_paused_enabled BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE queue_cleaner_settings
    DROP COLUMN recheck_paused_enabled;
