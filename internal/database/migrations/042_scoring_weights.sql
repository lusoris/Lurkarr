-- +goose Up
ALTER TABLE scoring_profiles ADD COLUMN resolution_weight INTEGER NOT NULL DEFAULT 50;
ALTER TABLE scoring_profiles ADD COLUMN source_weight INTEGER NOT NULL DEFAULT 30;
ALTER TABLE scoring_profiles ADD COLUMN revision_bonus INTEGER NOT NULL DEFAULT 50;

-- +goose Down
ALTER TABLE scoring_profiles DROP COLUMN IF EXISTS revision_bonus;
ALTER TABLE scoring_profiles DROP COLUMN IF EXISTS source_weight;
ALTER TABLE scoring_profiles DROP COLUMN IF EXISTS resolution_weight;
