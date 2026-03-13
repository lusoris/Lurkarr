-- +goose Up

ALTER TABLE webauthn_credentials ADD COLUMN transport TEXT[] DEFAULT '{}';
ALTER TABLE webauthn_credentials ADD COLUMN aaguid BYTEA DEFAULT '';

-- +goose Down
ALTER TABLE webauthn_credentials DROP COLUMN IF EXISTS aaguid;
ALTER TABLE webauthn_credentials DROP COLUMN IF EXISTS transport;
