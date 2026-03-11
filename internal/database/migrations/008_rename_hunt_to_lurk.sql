-- +goose Up
-- Rename hunt_history → lurk_history
ALTER TABLE hunt_history RENAME TO lurk_history;
ALTER INDEX idx_hunt_history_app RENAME TO idx_lurk_history_app;
ALTER INDEX idx_hunt_history_search RENAME TO idx_lurk_history_search;

-- Rename hunt_stats → lurk_stats
ALTER TABLE hunt_stats RENAME TO lurk_stats;

-- Rename columns in app_settings
ALTER TABLE app_settings RENAME COLUMN hunt_missing_count TO lurk_missing_count;
ALTER TABLE app_settings RENAME COLUMN hunt_upgrade_count TO lurk_upgrade_count;
ALTER TABLE app_settings RENAME COLUMN hunt_missing_mode TO lurk_missing_mode;

-- Rename column in lurk_stats
ALTER TABLE lurk_stats RENAME COLUMN hunted TO lurked;

-- +goose Down
ALTER TABLE lurk_stats RENAME COLUMN lurked TO hunted;

ALTER TABLE app_settings RENAME COLUMN lurk_missing_mode TO hunt_missing_mode;
ALTER TABLE app_settings RENAME COLUMN lurk_upgrade_count TO hunt_upgrade_count;
ALTER TABLE app_settings RENAME COLUMN lurk_missing_count TO hunt_missing_count;

ALTER TABLE lurk_stats RENAME TO hunt_stats;

ALTER INDEX idx_lurk_history_search RENAME TO idx_hunt_history_search;
ALTER INDEX idx_lurk_history_app RENAME TO idx_hunt_history_app;
ALTER TABLE lurk_history RENAME TO hunt_history;
