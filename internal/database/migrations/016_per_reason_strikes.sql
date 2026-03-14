-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN max_strikes_stalled INT DEFAULT 0;
ALTER TABLE queue_cleaner_settings ADD COLUMN max_strikes_slow INT DEFAULT 0;
ALTER TABLE queue_cleaner_settings ADD COLUMN max_strikes_metadata INT DEFAULT 0;
ALTER TABLE queue_cleaner_settings ADD COLUMN max_strikes_paused INT DEFAULT 0;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN max_strikes_stalled;
ALTER TABLE queue_cleaner_settings DROP COLUMN max_strikes_slow;
ALTER TABLE queue_cleaner_settings DROP COLUMN max_strikes_metadata;
ALTER TABLE queue_cleaner_settings DROP COLUMN max_strikes_paused;
