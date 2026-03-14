package database

import (
	"context"
	"fmt"
)

// GetAllCounters returns all persisted counter rows.
func (db *DB) GetAllCounters(ctx context.Context) ([]PersistentCounter, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT metric_name, label_key, value, updated_at FROM persistent_counters`)
	if err != nil {
		return nil, fmt.Errorf("get all counters: %w", err)
	}
	defer rows.Close()

	var counters []PersistentCounter
	for rows.Next() {
		var c PersistentCounter
		if err := rows.Scan(&c.MetricName, &c.LabelKey, &c.Value, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scan counter: %w", err)
		}
		counters = append(counters, c)
	}
	return counters, rows.Err()
}

// UpsertCounters upserts a batch of counter values in a single statement.
func (db *DB) UpsertCounters(ctx context.Context, counters []PersistentCounter) error {
	if len(counters) == 0 {
		return nil
	}

	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	for _, c := range counters {
		_, err := tx.Exec(ctx,
			`INSERT INTO persistent_counters (metric_name, label_key, value, updated_at)
			 VALUES ($1, $2, $3, now())
			 ON CONFLICT (metric_name, label_key) DO UPDATE
			 SET value = $3, updated_at = now()`,
			c.MetricName, c.LabelKey, c.Value)
		if err != nil {
			return fmt.Errorf("upsert counter %s/%s: %w", c.MetricName, c.LabelKey, err)
		}
	}

	return tx.Commit(ctx)
}
