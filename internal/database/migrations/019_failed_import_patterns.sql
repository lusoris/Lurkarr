-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN failed_import_patterns TEXT DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN failed_import_patterns;
