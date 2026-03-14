-- +goose Up
ALTER TABLE queue_cleaner_settings
    ADD COLUMN ignored_download_clients TEXT NOT NULL DEFAULT '';

-- +goose Down
ALTER TABLE queue_cleaner_settings
    DROP COLUMN ignored_download_clients;
