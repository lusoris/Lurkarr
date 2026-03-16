-- +goose Up
ALTER TABLE scoring_profiles ADD COLUMN hdr_weight INTEGER NOT NULL DEFAULT 20;
ALTER TABLE scoring_profiles ADD COLUMN audio_weight INTEGER NOT NULL DEFAULT 15;

-- +goose Down
ALTER TABLE scoring_profiles DROP COLUMN IF EXISTS audio_weight;
ALTER TABLE scoring_profiles DROP COLUMN IF EXISTS hdr_weight;
