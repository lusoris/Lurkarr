-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN ignore_above_bytes BIGINT DEFAULT 0;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN ignore_above_bytes;
