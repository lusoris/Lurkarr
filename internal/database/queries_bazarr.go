package database

import (
	"context"
	"fmt"
)

// GetBazarrSettings returns the singleton Bazarr settings.
func (db *DB) GetBazarrSettings(ctx context.Context) (*BazarrSettings, error) {
	var s BazarrSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT id, url, api_key, enabled, timeout, created_at, updated_at
		 FROM bazarr_settings WHERE id = 1`).
		Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.Timeout, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get bazarr settings: %w", err)
	}
	return &s, nil
}

// UpdateBazarrSettings updates the singleton Bazarr settings.
func (db *DB) UpdateBazarrSettings(ctx context.Context, s *BazarrSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE bazarr_settings
		 SET url = $1, api_key = $2, enabled = $3, timeout = $4, updated_at = now()
		 WHERE id = 1`,
		s.URL, s.APIKey, s.Enabled, s.Timeout)
	if err != nil {
		return fmt.Errorf("update bazarr settings: %w", err)
	}
	return nil
}
