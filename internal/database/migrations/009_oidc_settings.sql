-- +goose Up

CREATE TABLE oidc_settings (
    id              INTEGER PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    enabled         BOOLEAN NOT NULL DEFAULT false,
    issuer_url      TEXT NOT NULL DEFAULT '',
    client_id       TEXT NOT NULL DEFAULT '',
    client_secret   TEXT NOT NULL DEFAULT '',
    redirect_url    TEXT NOT NULL DEFAULT '',
    scopes          TEXT NOT NULL DEFAULT 'openid,profile,email',
    auto_create     BOOLEAN NOT NULL DEFAULT true,
    admin_group     TEXT NOT NULL DEFAULT '',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO oidc_settings (id) VALUES (1);

-- +goose Down
DROP TABLE IF EXISTS oidc_settings;
