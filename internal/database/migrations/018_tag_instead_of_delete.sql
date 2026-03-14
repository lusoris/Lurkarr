-- +goose Up
ALTER TABLE queue_cleaner_settings ADD COLUMN tag_instead_of_delete BOOLEAN DEFAULT false;
ALTER TABLE queue_cleaner_settings ADD COLUMN obsolete_tag_label TEXT DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings DROP COLUMN tag_instead_of_delete;
ALTER TABLE queue_cleaner_settings DROP COLUMN obsolete_tag_label;
