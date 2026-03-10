package database

import (
	"context"
	"fmt"
)

// GetProwlarrSettings returns the singleton Prowlarr settings.
func (db *DB) GetProwlarrSettings(ctx context.Context) (*ProwlarrSettings, error) {
	var s ProwlarrSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT id, url, api_key, enabled, sync_indexers, timeout, created_at, updated_at
		 FROM prowlarr_settings WHERE id = 1`).
		Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.SyncIndexers, &s.Timeout, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get prowlarr settings: %w", err)
	}
	return &s, nil
}

// UpdateProwlarrSettings updates the singleton Prowlarr settings.
func (db *DB) UpdateProwlarrSettings(ctx context.Context, s *ProwlarrSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE prowlarr_settings
		 SET url = $1, api_key = $2, enabled = $3, sync_indexers = $4, timeout = $5, updated_at = now()
		 WHERE id = 1`,
		s.URL, s.APIKey, s.Enabled, s.SyncIndexers, s.Timeout)
	if err != nil {
		return fmt.Errorf("update prowlarr settings: %w", err)
	}
	return nil
}

// GetSABnzbdSettings returns the singleton SABnzbd settings.
func (db *DB) GetSABnzbdSettings(ctx context.Context) (*SABnzbdSettings, error) {
	var s SABnzbdSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT id, url, api_key, enabled, timeout, category, created_at, updated_at
		 FROM sabnzbd_settings WHERE id = 1`).
		Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.Timeout, &s.Category, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get sabnzbd settings: %w", err)
	}
	return &s, nil
}

// UpdateSABnzbdSettings updates the singleton SABnzbd settings.
func (db *DB) UpdateSABnzbdSettings(ctx context.Context, s *SABnzbdSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE sabnzbd_settings
		 SET url = $1, api_key = $2, enabled = $3, timeout = $4, category = $5, updated_at = now()
		 WHERE id = 1`,
		s.URL, s.APIKey, s.Enabled, s.Timeout, s.Category)
	if err != nil {
		return fmt.Errorf("update sabnzbd settings: %w", err)
	}
	return nil
}
