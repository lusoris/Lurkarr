-- +goose Up
-- Make the external user lookup truly unique so INSERT ... ON CONFLICT works.
DROP INDEX IF EXISTS idx_users_external;
CREATE UNIQUE INDEX idx_users_external ON users(auth_provider, external_id) WHERE external_id != '';

-- +goose Down
DROP INDEX IF EXISTS idx_users_external;
CREATE INDEX idx_users_external ON users(auth_provider, external_id) WHERE external_id != '';
