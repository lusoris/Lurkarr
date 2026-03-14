-- +goose Up
ALTER TABLE queue_cleaner_settings
    ADD COLUMN recycle_bin_enabled BOOLEAN NOT NULL DEFAULT false,
    ADD COLUMN recycle_bin_path TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings
    DROP COLUMN recycle_bin_enabled,
    DROP COLUMN recycle_bin_path;
