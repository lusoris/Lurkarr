-- +goose Up
ALTER TABLE general_settings ADD COLUMN auto_import_interval_minutes INT NOT NULL DEFAULT 5;

-- +goose Down
ALTER TABLE general_settings DROP COLUMN auto_import_interval_minutes;
