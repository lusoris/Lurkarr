package database

import (
	"context"
	"fmt"

	"github.com/google/uuid"
)

// --- Blocklist Sources ---

func (db *DB) ListBlocklistSources(ctx context.Context) ([]BlocklistSource, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, name, url, enabled, sync_interval_hours, last_synced_at, etag, created_at
		 FROM blocklist_sources ORDER BY created_at`)
	if err != nil {
		return nil, fmt.Errorf("list blocklist sources: %w", err)
	}
	defer rows.Close()

	var sources []BlocklistSource
	for rows.Next() {
		var s BlocklistSource
		if err := rows.Scan(&s.ID, &s.Name, &s.URL, &s.Enabled, &s.SyncIntervalHours,
			&s.LastSyncedAt, &s.ETag, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan blocklist source: %w", err)
		}
		sources = append(sources, s)
	}
	return sources, rows.Err()
}

func (db *DB) GetBlocklistSource(ctx context.Context, id uuid.UUID) (*BlocklistSource, error) {
	var s BlocklistSource
	err := db.Pool.QueryRow(ctx,
		`SELECT id, name, url, enabled, sync_interval_hours, last_synced_at, etag, created_at
		 FROM blocklist_sources WHERE id = $1`, id,
	).Scan(&s.ID, &s.Name, &s.URL, &s.Enabled, &s.SyncIntervalHours,
		&s.LastSyncedAt, &s.ETag, &s.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("get blocklist source: %w", err)
	}
	return &s, nil
}

func (db *DB) CreateBlocklistSource(ctx context.Context, s *BlocklistSource) error {
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO blocklist_sources (name, url, enabled, sync_interval_hours)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at`,
		s.Name, s.URL, s.Enabled, s.SyncIntervalHours,
	).Scan(&s.ID, &s.CreatedAt)
	if err != nil {
		return fmt.Errorf("create blocklist source: %w", err)
	}
	return nil
}

func (db *DB) UpdateBlocklistSource(ctx context.Context, s *BlocklistSource) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE blocklist_sources SET name = $2, url = $3, enabled = $4, sync_interval_hours = $5
		 WHERE id = $1`,
		s.ID, s.Name, s.URL, s.Enabled, s.SyncIntervalHours)
	if err != nil {
		return fmt.Errorf("update blocklist source: %w", err)
	}
	return nil
}

func (db *DB) UpdateBlocklistSourceSync(ctx context.Context, id uuid.UUID, etag string) error {
	_, err := db.Pool.Exec(ctx,
		`UPDATE blocklist_sources SET last_synced_at = NOW(), etag = $2 WHERE id = $1`,
		id, etag)
	if err != nil {
		return fmt.Errorf("update blocklist source sync: %w", err)
	}
	return nil
}

func (db *DB) DeleteBlocklistSource(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM blocklist_sources WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete blocklist source: %w", err)
	}
	return nil
}

// --- Blocklist Rules ---

func (db *DB) ListEnabledBlocklistRules(ctx context.Context) ([]BlocklistRule, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, source_id, pattern, pattern_type, reason, enabled, created_at
		 FROM blocklist_rules WHERE enabled = true ORDER BY pattern_type, pattern`)
	if err != nil {
		return nil, fmt.Errorf("list enabled blocklist rules: %w", err)
	}
	defer rows.Close()

	var rules []BlocklistRule
	for rows.Next() {
		var r BlocklistRule
		if err := rows.Scan(&r.ID, &r.SourceID, &r.Pattern, &r.PatternType, &r.Reason,
			&r.Enabled, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan blocklist rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (db *DB) ListBlocklistRules(ctx context.Context) ([]BlocklistRule, error) {
	rows, err := db.Pool.Query(ctx,
		`SELECT id, source_id, pattern, pattern_type, reason, enabled, created_at
		 FROM blocklist_rules ORDER BY pattern_type, pattern`)
	if err != nil {
		return nil, fmt.Errorf("list blocklist rules: %w", err)
	}
	defer rows.Close()

	var rules []BlocklistRule
	for rows.Next() {
		var r BlocklistRule
		if err := rows.Scan(&r.ID, &r.SourceID, &r.Pattern, &r.PatternType, &r.Reason,
			&r.Enabled, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan blocklist rule: %w", err)
		}
		rules = append(rules, r)
	}
	return rules, rows.Err()
}

func (db *DB) CreateBlocklistRule(ctx context.Context, r *BlocklistRule) error {
	err := db.Pool.QueryRow(ctx,
		`INSERT INTO blocklist_rules (source_id, pattern, pattern_type, reason, enabled)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`,
		r.SourceID, r.Pattern, r.PatternType, r.Reason, r.Enabled,
	).Scan(&r.ID, &r.CreatedAt)
	if err != nil {
		return fmt.Errorf("create blocklist rule: %w", err)
	}
	return nil
}

func (db *DB) DeleteBlocklistRule(ctx context.Context, id uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM blocklist_rules WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete blocklist rule: %w", err)
	}
	return nil
}

func (db *DB) DeleteBlocklistRulesBySource(ctx context.Context, sourceID uuid.UUID) error {
	_, err := db.Pool.Exec(ctx, `DELETE FROM blocklist_rules WHERE source_id = $1`, sourceID)
	if err != nil {
		return fmt.Errorf("delete blocklist rules by source: %w", err)
	}
	return nil
}

func (db *DB) CountBlocklistRulesBySource(ctx context.Context, sourceID uuid.UUID) (int, error) {
	var count int
	err := db.Pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM blocklist_rules WHERE source_id = $1`, sourceID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count blocklist rules by source: %w", err)
	}
	return count, nil
}
