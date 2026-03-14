-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN strike_queued BOOLEAN DEFAULT false;
ALTER TABLE queue_cleaner_settings ADD COLUMN max_strikes_queued INT DEFAULT 0;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN strike_queued;
ALTER TABLE queue_cleaner_settings DROP COLUMN max_strikes_queued;
