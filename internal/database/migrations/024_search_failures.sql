-- +goose Up

CREATE TABLE IF NOT EXISTS search_failures (
    app_type    TEXT NOT NULL,
    instance_id UUID NOT NULL,
    media_id    INT NOT NULL,
    fail_count  INT NOT NULL DEFAULT 1,
    last_failed TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (app_type, instance_id, media_id)
);

ALTER TABLE app_settings ADD COLUMN max_search_failures INT DEFAULT 0;
ALTER TABLE queue_cleaner_settings ADD COLUMN max_search_failures INT DEFAULT 0;

-- +goose Down

DROP TABLE IF EXISTS search_failures;
ALTER TABLE app_settings DROP COLUMN IF EXISTS max_search_failures;
ALTER TABLE queue_cleaner_settings DROP COLUMN IF EXISTS max_search_failures;
