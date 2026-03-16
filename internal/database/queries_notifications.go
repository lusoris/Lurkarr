package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListNotificationProviders returns all notification providers.
func (db *DB) ListNotificationProviders(ctx context.Context) ([]NotificationProvider, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, type, name, enabled, config, events, created_at, updated_at
		 FROM notification_providers ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list notification providers: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[NotificationProvider])
}

// ListEnabledNotificationProviders returns only enabled providers.
func (db *DB) ListEnabledNotificationProviders(ctx context.Context) ([]NotificationProvider, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, type, name, enabled, config, events, created_at, updated_at
		 FROM notification_providers WHERE enabled = true ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list enabled notification providers: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[NotificationProvider])
}

// GetNotificationProvider returns a single notification provider by ID.
func (db *DB) GetNotificationProvider(ctx context.Context, id uuid.UUID) (*NotificationProvider, error) {
	var p NotificationProvider
	err := db.Pool.QueryRow(ctx,
		`SELECT id, type, name, enabled, config, events, created_at, updated_at
		 FROM notification_providers WHERE id = $1`, id).
		Scan(&p.ID, &p.Type, &p.Name, &p.Enabled, &p.Config, &p.Events, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("get notification provider: %w", err)
	}
	return &p, nil
}

// CreateNotificationProvider inserts a new notification provider.
func (db *DB) CreateNotificationProvider(ctx context.Context, p *NotificationProvider) error {
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO notification_providers (type, name, enabled, config, events)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		p.Type, p.Name, p.Enabled, p.Config, p.Events).
		Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return fmt.Errorf("create notification provider: %w", err)
	}
	return nil
}

// UpdateNotificationProvider updates an existing notification provider.
func (db *DB) UpdateNotificationProvider(ctx context.Context, p *NotificationProvider) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE notification_providers
		 SET type = $1, name = $2, enabled = $3, config = $4, events = $5, updated_at = now()
		 WHERE id = $6`,
		p.Type, p.Name, p.Enabled, p.Config, p.Events, p.ID)
	if err != nil {
		return fmt.Errorf("update notification provider: %w", err)
	}
	return nil
}

// DeleteNotificationProvider deletes a notification provider by ID.
func (db *DB) DeleteNotificationProvider(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx,
		`DELETE FROM notification_providers WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete notification provider: %w", err)
	}
	return nil
}
