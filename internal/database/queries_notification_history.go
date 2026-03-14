package database

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// InsertNotificationHistory records a single notification delivery.
func (db *DB) InsertNotificationHistory(ctx context.Context, h *NotificationHistory) error {
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO notification_history
			(provider_id, provider_type, provider_name, event_type, title, message, app_type, instance, status, error, duration_ms)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		 RETURNING id, created_at`,
		h.ProviderID, h.ProviderType, h.ProviderName, h.EventType,
		h.Title, h.Message, h.AppType, h.Instance,
		h.Status, h.Error, h.DurationMs).
		Scan(&h.ID, &h.CreatedAt)
	if err != nil {
		return fmt.Errorf("insert notification history: %w", err)
	}
	return nil
}

// ListNotificationHistory returns recent notification history ordered by creation time descending.
func (db *DB) ListNotificationHistory(ctx context.Context, limit int) ([]NotificationHistory, error) {
	if limit <= 0 || limit > 500 {
		limit = 200
	}

	rows, err := db.Pool.Query(ctx,
		`SELECT id, provider_id, provider_type, provider_name, event_type, title, message,
		        app_type, instance, status, error, duration_ms, created_at
		 FROM notification_history
		 ORDER BY created_at DESC
		 LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("list notification history: %w", err)
	}
	defer rows.Close()

	var items []NotificationHistory
	for rows.Next() {
		var h NotificationHistory
		if err := rows.Scan(
			&h.ID, &h.ProviderID, &h.ProviderType, &h.ProviderName, &h.EventType,
			&h.Title, &h.Message, &h.AppType, &h.Instance,
			&h.Status, &h.Error, &h.DurationMs, &h.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan notification history: %w", err)
		}
		items = append(items, h)
	}
	return items, rows.Err()
}

// DeleteOldNotificationHistory removes entries older than the given time.
func (db *DB) DeleteOldNotificationHistory(ctx context.Context, olderThan time.Time) (int64, error) {
	tag, err := db.Pool.Exec(ctx,
		`DELETE FROM notification_history WHERE created_at < $1`, olderThan)
	if err != nil {
		return 0, fmt.Errorf("delete old notification history: %w", err)
	}
	return tag.RowsAffected(), nil
}

// DeleteNotificationHistoryByProvider removes all history for a specific provider.
func (db *DB) DeleteNotificationHistoryByProvider(ctx context.Context, providerID uuid.UUID) (int64, error) {
	tag, err := db.Pool.Exec(ctx,
		`DELETE FROM notification_history WHERE provider_id = $1`, providerID)
	if err != nil {
		return 0, fmt.Errorf("delete notification history by provider: %w", err)
	}
	return tag.RowsAffected(), nil
}
