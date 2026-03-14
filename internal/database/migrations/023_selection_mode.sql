-- +goose Up
-- Convert boolean random_selection to text selection_mode.
-- true → 'random', false → 'oldest' (the previous non-random behaviour picked
-- items in API default order which was effectively by date ascending for most apps).
ALTER TABLE app_settings ADD COLUMN selection_mode TEXT DEFAULT 'random';
UPDATE app_settings SET selection_mode = CASE WHEN random_selection THEN 'random' ELSE 'oldest' END;
ALTER TABLE app_settings DROP COLUMN random_selection;

-- +goose Down
ALTER TABLE app_settings ADD COLUMN random_selection BOOLEAN DEFAULT true;
UPDATE app_settings SET random_selection = (selection_mode = 'random');
ALTER TABLE app_settings DROP COLUMN selection_mode;
