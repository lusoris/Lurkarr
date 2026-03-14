package database

import (
	"context"
	"fmt"
)

// GetOIDCSettings returns the singleton OIDC settings row.
func (db *DB) GetOIDCSettings(ctx context.Context) (*OIDCSettings, error) {
	var s OIDCSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT enabled, issuer_url, client_id, client_secret, redirect_url, scopes,
		        auto_create, admin_group, created_at, updated_at
		 FROM oidc_settings WHERE id = 1`).
		Scan(&s.Enabled, &s.IssuerURL, &s.ClientID, &s.ClientSecret, &s.RedirectURL,
			&s.Scopes, &s.AutoCreate, &s.AdminGroup, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get oidc settings: %w", err)
	}
	return &s, nil
}

// UpdateOIDCSettings updates the singleton OIDC settings row.
func (db *DB) UpdateOIDCSettings(ctx context.Context, s *OIDCSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE oidc_settings
		 SET enabled = $1, issuer_url = $2, client_id = $3, client_secret = $4,
		     redirect_url = $5, scopes = $6, auto_create = $7, admin_group = $8,
		     updated_at = now()
		 WHERE id = 1`,
		s.Enabled, s.IssuerURL, s.ClientID, s.ClientSecret,
		s.RedirectURL, s.Scopes, s.AutoCreate, s.AdminGroup)
	if err != nil {
		return fmt.Errorf("update oidc settings: %w", err)
	}
	return nil
}
