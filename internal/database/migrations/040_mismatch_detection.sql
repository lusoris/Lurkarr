-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN mismatch_enabled BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE queue_cleaner_settings ADD COLUMN max_strikes_mismatch INTEGER NOT NULL DEFAULT 0;
ALTER TABLE queue_cleaner_settings ADD COLUMN blocklist_mismatch BOOLEAN NOT NULL DEFAULT TRUE;

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS blocklist_mismatch;
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS max_strikes_mismatch;
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS mismatch_enabled;
