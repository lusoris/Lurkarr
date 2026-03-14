-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN max_searches_per_run INT DEFAULT 0;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN max_searches_per_run;
