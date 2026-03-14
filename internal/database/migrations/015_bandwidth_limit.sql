-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN bandwidth_limit_bytes_per_sec BIGINT NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN bandwidth_limit_bytes_per_sec;
