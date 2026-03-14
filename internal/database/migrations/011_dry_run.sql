-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN dry_run BOOLEAN NOT NULL DEFAULT true;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN dry_run;
