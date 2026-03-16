-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN custom_unregistered_keywords TEXT NOT NULL DEFAULT '';
ALTER TABLE queue_cleaner_settings ADD COLUMN custom_mismatch_keywords TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS custom_mismatch_keywords;
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS custom_unregistered_keywords;
