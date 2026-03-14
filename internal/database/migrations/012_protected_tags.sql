-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN protected_tags TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN protected_tags;
