-- +goose Up
ALTER TABLE general_settings RENAME COLUMN min_download_queue_size TO max_download_queue_size;
ALTER TABLE general_settings ALTER COLUMN max_download_queue_size SET DEFAULT 0;
UPDATE general_settings SET max_download_queue_size = 0 WHERE max_download_queue_size = -1;

-- +goose Down
UPDATE general_settings SET max_download_queue_size = -1 WHERE max_download_queue_size = 0;
ALTER TABLE general_settings ALTER COLUMN max_download_queue_size SET DEFAULT -1;
ALTER TABLE general_settings RENAME COLUMN max_download_queue_size TO min_download_queue_size;
