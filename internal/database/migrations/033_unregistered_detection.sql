-- +goose Up
ALTER TABLE queue_cleaner_settings
    ADD COLUMN unregistered_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN max_strikes_unregistered INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE queue_cleaner_settings
    DROP COLUMN unregistered_enabled,
    DROP COLUMN max_strikes_unregistered;
