package database

import "context"

// GetSeerrSettings returns the singleton Seerr settings row.
func (db *DB) GetSeerrSettings(ctx context.Context) (*SeerrSettings, error) {
	var s SeerrSettings
	err := db.Pool.QueryRow(ctx, `
		SELECT id, url, api_key, enabled, sync_interval_minutes, auto_approve, cleanup_enabled, cleanup_after_days, created_at, updated_at
		FROM seerr_settings
		LIMIT 1
	`).Scan(&s.ID, &s.URL, &s.APIKey, &s.Enabled, &s.SyncIntervalMinutes, &s.AutoApprove, &s.CleanupEnabled, &s.CleanupAfterDays, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

// UpdateSeerrSettings updates the Seerr settings.
func (db *DB) UpdateSeerrSettings(ctx context.Context, s *SeerrSettings) error {
	_, err := db.Pool.Exec(ctx, `
		UPDATE seerr_settings
		SET url = $1, api_key = $2, enabled = $3, sync_interval_minutes = $4, auto_approve = $5,
		    cleanup_enabled = $6, cleanup_after_days = $7, updated_at = now()
		WHERE id = $8
	`, s.URL, s.APIKey, s.Enabled, s.SyncIntervalMinutes, s.AutoApprove, s.CleanupEnabled, s.CleanupAfterDays, s.ID)
	return err
}
