package database

import "context"

// GetSeerrSettings returns the singleton Seerr settings row.
func (db *DB) GetSeerrSettings(ctx context.Context) (*SeerrSettings, error) {
	var s SeerrSettings
	err := db.Pool.QueryRow(ctx, `
		SELECT id, url, api_key, enabled, sync_interval_minutes, auto_approve, created_at, updated_at
		FROM seerr_settings
		LIMIT 1
	`).Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.SyncIntervalMinutes, &s.AutoApprove, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSeerrSettings updates the Seerr settings.
func (db *DB) UpdateSeerrSettings(ctx context.Context, s *SeerrSettings) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE seerr_settings
		SET url = $1, api_key = $2, enabled = $3, sync_interval_minutes = $4, auto_approve = $5, updated_at = now()
		WHERE id = $6
	`, s.URL, s.APIKey, s.Enabled, s.SyncIntervalMinutes, s.AutoApprove, s.ID)
	return err
}
