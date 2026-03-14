-- +goose Up
ALTER TABLE seerr_settings ADD COLUMN cleanup_enabled BOOLEAN NOT NULL DEFAULT false;
ALTER TABLE seerr_settings ADD COLUMN cleanup_after_days INTEGER NOT NULL DEFAULT 7;

-- +goose Down
ALTER TABLE seerr_settings DROP COLUMN cleanup_after_days;
ALTER TABLE seerr_settings DROP COLUMN cleanup_enabled;
