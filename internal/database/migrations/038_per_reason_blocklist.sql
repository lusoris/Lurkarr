-- +goose Up
ALTER TABLE queue_cleaner_settings
    ADD COLUMN blocklist_stalled    BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN blocklist_slow       BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN blocklist_metadata   BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN blocklist_duplicate  BOOLEAN NOT NULL DEFAULT TRUE,
    ADD COLUMN blocklist_unregistered BOOLEAN NOT NULL DEFAULT TRUE;

-- Initialise from the existing global flag so behaviour is unchanged.
UPDATE queue_cleaner_settings SET
    blocklist_stalled      = blocklist_on_remove,
    blocklist_slow         = blocklist_on_remove,
    blocklist_metadata     = blocklist_on_remove,
    blocklist_duplicate    = blocklist_on_remove,
    blocklist_unregistered = blocklist_on_remove;

-- +goose Down
ALTER TABLE queue_cleaner_settings
    DROP COLUMN blocklist_stalled,
    DROP COLUMN blocklist_slow,
    DROP COLUMN blocklist_metadata,
    DROP COLUMN blocklist_duplicate,
    DROP COLUMN blocklist_unregistered;
