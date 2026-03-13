-- +goose Up

-- ── Recovery Codes (stored as bcrypt hashes) ─────────────────────────────────
ALTER TABLE users ADD COLUMN recovery_codes TEXT[] DEFAULT '{}';

-- ── Session metadata for management UI ───────────────────────────────────────
ALTER TABLE sessions ADD COLUMN ip_address TEXT NOT NULL DEFAULT '';
ALTER TABLE sessions ADD COLUMN user_agent TEXT NOT NULL DEFAULT '';

-- ── WebAuthn Credentials ─────────────────────────────────────────────────────
CREATE TABLE webauthn_credentials (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id         UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name            TEXT NOT NULL DEFAULT 'Security Key',
    credential_id   BYTEA NOT NULL UNIQUE,
    public_key      BYTEA NOT NULL,
    attestation_type TEXT NOT NULL DEFAULT '',
    sign_count      BIGINT NOT NULL DEFAULT 0,
    created_at      TIMESTAMPTZ DEFAULT now()
);
CREATE INDEX idx_webauthn_user ON webauthn_credentials(user_id);

-- +goose Down
DROP TABLE IF EXISTS webauthn_credentials;
ALTER TABLE sessions DROP COLUMN IF EXISTS user_agent;
ALTER TABLE sessions DROP COLUMN IF EXISTS ip_address;
ALTER TABLE users DROP COLUMN IF EXISTS recovery_codes;
