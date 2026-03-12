-- +goose Up
ALTER TABLE queue_cleaner_settings
    ADD COLUMN cross_arr_sync BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE queue_cleaner_settings
    DROP COLUMN IF EXISTS cross_arr_sync;
