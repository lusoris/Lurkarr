-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN deletion_detection_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE queue_cleaner_settings ADD COLUMN unmonitored_cleanup_enabled BOOLEAN NOT NULL DEFAULT FALSE;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN unmonitored_cleanup_enabled;
ALTER TABLE queue_cleaner_settings DROP COLUMN deletion_detection_enabled;
