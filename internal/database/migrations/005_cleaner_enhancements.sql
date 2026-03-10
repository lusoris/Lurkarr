-- +goose Up

ALTER TABLE queue_cleaner_settings
    ADD COLUMN strike_public BOOLEAN DEFAULT true,
    ADD COLUMN strike_private BOOLEAN DEFAULT false,
    ADD COLUMN slow_ignore_above_bytes BIGINT DEFAULT 0,
    ADD COLUMN failed_import_remove BOOLEAN DEFAULT false,
    ADD COLUMN failed_import_blocklist BOOLEAN DEFAULT true,
    ADD COLUMN metadata_stuck_minutes INT DEFAULT 0;

-- +goose Down

ALTER TABLE queue_cleaner_settings
    DROP COLUMN IF EXISTS strike_public,
    DROP COLUMN IF EXISTS strike_private,
    DROP COLUMN IF EXISTS slow_ignore_above_bytes,
    DROP COLUMN IF EXISTS failed_import_remove,
    DROP COLUMN IF EXISTS failed_import_blocklist,
    DROP COLUMN IF EXISTS metadata_stuck_minutes;
