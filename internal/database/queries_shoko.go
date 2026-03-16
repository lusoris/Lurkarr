package database

import (
	"context"
	"fmt"
)

// GetShokoSettings returns the singleton Shoko settings.
func (db *DB) GetShokoSettings(ctx context.Context) (*ShokoSettings, error) {
	var s ShokoSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT id, url, api_key, enabled, timeout, created_at, updated_at
		 FROM shoko_settings WHERE id = 1`).
		Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.Timeout, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get shoko settings: %w", err)
	}
	return &s, nil
}

// UpdateShokoSettings updates the singleton Shoko settings.
func (db *DB) UpdateShokoSettings(ctx context.Context, s *ShokoSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE shoko_settings
		 SET url = $1, api_key = $2, enabled = $3, timeout = $4, updated_at = now()
		 WHERE id = 1`,
		s.URL, s.APIKey, s.Enabled, s.Timeout)
	if err != nil {
		return fmt.Errorf("update shoko settings: %w", err)
	}
	return nil
}
