-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN keep_archives BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS keep_archives;
