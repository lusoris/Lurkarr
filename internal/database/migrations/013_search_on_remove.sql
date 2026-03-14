-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN search_on_remove BOOLEAN NOT NULL DEFAULT false;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN search_on_remove;
