-- +goose Up
ALTER TABLE general_settings ADD COLUMN csrf_key TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE general_settings DROP COLUMN csrf_key;
