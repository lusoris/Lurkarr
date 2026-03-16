package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// ListDownloadClientInstances returns all download client instances ordered by name.
func (db *DB) ListDownloadClientInstances(ctx context.Context) ([]DownloadClientInstance, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, name, client_type, url, api_key, username, password, category, enabled, timeout, created_at
		 FROM download_client_instances ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list download client instances: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[DownloadClientInstance])
}

// GetDownloadClientInstance returns a single download client instance by ID.
func (db *DB) GetDownloadClientInstance(ctx context.Context, id uuid.UUID) (*DownloadClientInstance, error) {
	d, err := queryOne[DownloadClientInstance](ctx, db,
		`SELECT id, name, client_type, url, api_key, username, password, category, enabled, timeout, created_at
		 FROM download_client_instances WHERE id = $1`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("get download client instance: %w", err)
	}
	return d, nil
}

// CreateDownloadClientInstance creates a new download client instance.
func (db *DB) CreateDownloadClientInstance(ctx context.Context, d *DownloadClientInstance) (*DownloadClientInstance, error) {
	out, err := queryOne[DownloadClientInstance](ctx, db,
		`INSERT INTO download_client_instances (name, client_type, url, api_key, username, password, category, enabled, timeout)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING id, name, client_type, url, api_key, username, password, category, enabled, timeout, created_at`,
		d.Name, d.ClientType, d.URL, d.APIKey, d.Username, d.Password, d.Category, d.Enabled, d.Timeout,
	)
	if err != nil {
		return nil, fmt.Errorf("create download client instance: %w", err)
	}
	return out, nil
}

// UpdateDownloadClientInstance updates an existing download client instance.
func (db *DB) UpdateDownloadClientInstance(ctx context.Context, d *DownloadClientInstance) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE download_client_instances
		 SET name = $1, client_type = $2, url = $3, api_key = $4, username = $5, password = $6,
		     category = $7, enabled = $8, timeout = $9
		 WHERE id = $10`,
		d.Name, d.ClientType, d.URL, d.APIKey, d.Username, d.Password, d.Category, d.Enabled, d.Timeout, d.ID)
	if err != nil {
		return fmt.Errorf("update download client instance: %w", err)
	}
	return nil
}

// DeleteDownloadClientInstance deletes a download client instance by ID.
func (db *DB) DeleteDownloadClientInstance(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM download_client_instances WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete download client instance: %w", err)
	}
	return nil
}

// ListEnabledDownloadClientInstances returns all enabled download client instances.
func (db *DB) ListEnabledDownloadClientInstances(ctx context.Context) ([]DownloadClientInstance, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, name, client_type, url, api_key, username, password, category, enabled, timeout, created_at
		 FROM download_client_instances WHERE enabled = true ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list enabled download client instances: %w", err)
	}
	return pgx.CollectRows(rows, pgx.RowToStructByPos[DownloadClientInstance])
}
