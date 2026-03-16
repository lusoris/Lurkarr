package database

import (
	"context"
	"fmt"
)

// GetKapowarrSettings returns the singleton Kapowarr settings.
func (db *DB) GetKapowarrSettings(ctx context.Context) (*KapowarrSettings, error) {
	var s KapowarrSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT id, url, api_key, enabled, timeout, created_at, updated_at
		 FROM kapowarr_settings WHERE id = 1`).
		Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.Timeout, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get kapowarr settings: %w", err)
	}
	return &s, nil
}

// UpdateKapowarrSettings updates the singleton Kapowarr settings.
func (db *DB) UpdateKapowarrSettings(ctx context.Context, s *KapowarrSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE kapowarr_settings
		 SET url = $1, api_key = $2, enabled = $3, timeout = $4, updated_at = now()
		 WHERE id = 1`,
		s.URL, s.APIKey, s.Enabled, s.Timeout)
	if err != nil {
		return fmt.Errorf("update kapowarr settings: %w", err)
	}
	return nil
}
