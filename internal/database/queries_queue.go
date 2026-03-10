package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// --- Queue Cleaner Settings ---

func (db *DB) GetQueueCleanerSettings(ctx context.Context, appType AppType) (*QueueCleanerSettings, error) {
	var s QueueCleanerSettings
	err := db.Pool.QueryRow(ctx,
		`SELECT app_type, enabled, stalled_threshold_minutes, slow_threshold_bytes_per_sec,
		        max_strikes, strike_window_hours, check_interval_seconds,
		        remove_from_client, blocklist_on_remove
		 FROM queue_cleaner_settings WHERE app_type = $1`, appType,
	).Scan(&s.AppType, &s.Enabled, &s.StalledThresholdMinutes, &s.SlowThresholdBytesPerSec,
		&s.MaxStrikes, &s.StrikeWindowHours, &s.CheckIntervalSeconds,
		&s.RemoveFromClient, &s.BlocklistOnRemove)
	if err != nil {
		return nil, fmt.Errorf("get queue cleaner settings: %w", err)
	}
	return &s, nil
}

func (db *DB) UpdateQueueCleanerSettings(ctx context.Context, s *QueueCleanerSettings) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE queue_cleaner_settings SET
		        enabled = $2, stalled_threshold_minutes = $3, slow_threshold_bytes_per_sec = $4,
		        max_strikes = $5, strike_window_hours = $6, check_interval_seconds = $7,
		        remove_from_client = $8, blocklist_on_remove = $9
		 WHERE app_type = $1`,
		s.AppType, s.Enabled, s.StalledThresholdMinutes, s.SlowThresholdBytesPerSec,
		s.MaxStrikes, s.StrikeWindowHours, s.CheckIntervalSeconds,
		s.RemoveFromClient, s.BlocklistOnRemove)
	if err != nil {
		return fmt.Errorf("update queue cleaner settings: %w", err)
	}
	return nil
}

// --- Strikes ---

func (db *DB) AddStrike(ctx context.Context, appType AppType, instanceID uuid.UUID, downloadID, title, reason string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO queue_strikes (app_type, instance_id, download_id, title, reason) VALUES ($1, $2, $3, $4, $5)`,
		appType, instanceID, downloadID, title, reason)
	if err != nil {
		return fmt.Errorf("add strike: %w", err)
	}
	return nil
}

func (db *DB) CountStrikes(ctx context.Context, appType AppType, instanceID uuid.UUID, downloadID string, windowHours int) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM queue_strikes
		 WHERE app_type = $1 AND instance_id = $2 AND download_id = $3
		   AND struck_at > NOW() - make_interval(hours => $4)`,
		appType, instanceID, downloadID, windowHours).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count strikes: %w", err)
	}
	return count, nil
}

func (db *DB) PruneStrikes(ctx context.Context, olderThan time.Duration) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM queue_strikes WHERE struck_at < $1`, time.Now().Add(-olderThan))
	if err != nil {
		return fmt.Errorf("prune strikes: %w", err)
	}
	return nil
}

// --- Auto Import Log ---

func (db *DB) LogAutoImport(ctx context.Context, appType AppType, instanceID uuid.UUID, mediaID int, mediaTitle string, queueItemID int, action, reason string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO auto_import_log (app_type, instance_id, media_id, media_title, queue_item_id, action, reason)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		appType, instanceID, mediaID, mediaTitle, queueItemID, action, reason)
	if err != nil {
		return fmt.Errorf("log auto import: %w", err)
	}
	return nil
}

func (db *DB) GetAutoImportLog(ctx context.Context, appType AppType, limit int) ([]AutoImportLog, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, app_type, instance_id, media_id, media_title, queue_item_id, action, reason, created_at
		 FROM auto_import_log WHERE app_type = $1 ORDER BY created_at DESC LIMIT $2`, appType, limit)
	if err != nil {
		return nil, fmt.Errorf("get auto import log: %w", err)
	}
	defer rows.Close()

	var logs []AutoImportLog
	for rows.Next() {
		var l AutoImportLog
		if err := rows.Scan(&l.ID, &l.AppType, &l.InstanceID, &l.MediaID, &l.MediaTitle, &l.QueueItemID, &l.Action, &l.Reason, &l.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan auto import log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (db *DB) PruneAutoImportLog(ctx context.Context, olderThan time.Duration) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM auto_import_log WHERE created_at < $1`, time.Now().Add(-olderThan))
	if err != nil {
		return fmt.Errorf("prune auto import log: %w", err)
	}
	return nil
}

// --- Scoring Profiles ---

func (db *DB) GetScoringProfile(ctx context.Context, appType AppType) (*ScoringProfile, error) {
	var p ScoringProfile
	err := db.Pool.QueryRow(ctx,
		`SELECT id, app_type, name, prefer_higher_quality, prefer_larger_size, prefer_indexer_flags,
		        custom_format_weight, size_weight, age_weight, seeders_weight, created_at
		 FROM scoring_profiles WHERE app_type = $1`, appType,
	).Scan(&p.ID, &p.AppType, &p.Name, &p.PreferHigherQuality, &p.PreferLargerSize, &p.PreferIndexerFlags,
		&p.CustomFormatWeight, &p.SizeWeight, &p.AgeWeight, &p.SeedersWeight, &p.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get scoring profile: %w", err)
	}
	return &p, nil
}

func (db *DB) UpdateScoringProfile(ctx context.Context, p *ScoringProfile) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE scoring_profiles SET
		        name = $2, prefer_higher_quality = $3, prefer_larger_size = $4, prefer_indexer_flags = $5,
		        custom_format_weight = $6, size_weight = $7, age_weight = $8, seeders_weight = $9
		 WHERE id = $1`,
		p.ID, p.Name, p.PreferHigherQuality, p.PreferLargerSize, p.PreferIndexerFlags,
		p.CustomFormatWeight, p.SizeWeight, p.AgeWeight, p.SeedersWeight)
	if err != nil {
		return fmt.Errorf("update scoring profile: %w", err)
	}
	return nil
}

// --- Blocklist Log ---

func (db *DB) LogBlocklist(ctx context.Context, appType AppType, instanceID uuid.UUID, downloadID, title, reason string) error {
	_, err := db.Pool.Exec(ctx,
		`INSERT INTO blocklist_log (app_type, instance_id, download_id, title, reason)
		 VALUES ($1, $2, $3, $4, $5)`,
		appType, instanceID, downloadID, title, reason)
	if err != nil {
		return fmt.Errorf("log blocklist: %w", err)
	}
	return nil
}

func (db *DB) GetBlocklistLog(ctx context.Context, appType AppType, limit int) ([]BlocklistLog, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, app_type, instance_id, download_id, title, reason, blocklisted_at
		 FROM blocklist_log WHERE app_type = $1 ORDER BY blocklisted_at DESC LIMIT $2`, appType, limit)
	if err != nil {
		return nil, fmt.Errorf("get blocklist log: %w", err)
	}
	defer rows.Close()

	var logs []BlocklistLog
	for rows.Next() {
		var l BlocklistLog
		if err := rows.Scan(&l.ID, &l.AppType, &l.InstanceID, &l.DownloadID, &l.Title, &l.Reason, &l.BlocklistedAt); err != nil {
			return nil, fmt.Errorf("scan blocklist log: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (db *DB) PruneBlocklistLog(ctx context.Context, olderThan time.Duration) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM blocklist_log WHERE blocklisted_at < $1`, time.Now().Add(-olderThan))
	if err != nil {
		return fmt.Errorf("prune blocklist log: %w", err)
	}
	return nil
}
