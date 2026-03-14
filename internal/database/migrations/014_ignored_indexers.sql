-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN ignored_indexers TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN ignored_indexers;
