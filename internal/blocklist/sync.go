package blocklist

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/metrics"
)

// SyncStore defines the database operations needed by the syncer.
type SyncStore interface {
	ListBlocklistSources(ctx context.Context) ([]database.BlocklistSource, error)
	ReplaceBlocklistRulesForSource(ctx context.Context, sourceID uuid.UUID, rules []database.BlocklistRule) error
	UpdateBlocklistSourceSync(ctx context.Context, id uuid.UUID, etag string) error
}

// Syncer periodically fetches community blocklists and updates rules.
type Syncer struct {
	db     SyncStore
	client *http.Client
}

// NewSyncer creates a new blocklist syncer.
func NewSyncer(db SyncStore) *Syncer {
	return &Syncer{
		db: db,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SyncAll fetches all enabled sources and updates their rules.
func (s *Syncer) SyncAll(ctx context.Context) {
	sources, err := s.db.ListBlocklistSources(ctx)
	if err != nil {
		slog.Error("blocklist sync: failed to list sources", "error", err)
		return
	}

	for _, src := range sources {
		if !src.Enabled {
			continue
		}
		if err := s.SyncSource(ctx, src); err != nil {
			slog.Error("blocklist sync: failed to sync source", "source", src.Name, "error", err)
		}
	}
}

// SyncSource fetches a single blocklist source and updates its rules.
func (s *Syncer) SyncSource(ctx context.Context, src database.BlocklistSource) error {
	start := time.Now()
	err := s.syncSource(ctx, src)
	dur := time.Since(start)
	metrics.BlocklistSyncDuration.WithLabelValues(src.Name).Observe(dur.Seconds())
	if err != nil {
		metrics.BlocklistSyncErrorsTotal.WithLabelValues(src.Name).Inc()
	}
	return err
}

func (s *Syncer) syncSource(ctx context.Context, src database.BlocklistSource) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, src.URL, http.NoBody)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	// Use ETag for conditional fetching.
	if src.ETag != "" {
		req.Header.Set("If-None-Match", src.ETag)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("fetch source %s: %w", src.Name, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotModified {
		slog.Debug("blocklist sync: source not modified", "source", src.Name)
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("source %s returned status %d", src.Name, resp.StatusCode)
	}

	// Limit response body to 5 MB to prevent abuse.
	limited := io.LimitReader(resp.Body, 5<<20)
	rules, err := ParseBlocklist(limited)
	if err != nil {
		return fmt.Errorf("parse source %s: %w", src.Name, err)
	}

	// Replace all rules for this source atomically within a transaction.
	for i := range rules {
		rules[i].SourceID = &src.ID
		rules[i].Enabled = true
	}
	if err := s.db.ReplaceBlocklistRulesForSource(ctx, src.ID, rules); err != nil {
		return fmt.Errorf("replace rules for %s: %w", src.Name, err)
	}

	etag := resp.Header.Get("ETag")
	if err := s.db.UpdateBlocklistSourceSync(ctx, src.ID, etag); err != nil {
		return fmt.Errorf("update sync time for %s: %w", src.Name, err)
	}

	slog.Info("blocklist sync: updated source", "source", src.Name, "rules", len(rules))
	metrics.BlocklistSyncRulesTotal.WithLabelValues(src.Name).Add(float64(len(rules)))
	return nil
}

// ParseBlocklist parses a blocklist file.
// Format: one rule per line, # comments, empty lines skipped.
// Lines can optionally start with a type prefix: "group:", "regex:", "indexer:", "contains:".
// Lines without a prefix default to "release_group".
func ParseBlocklist(r io.Reader) ([]database.BlocklistRule, error) {
	var rules []database.BlocklistRule
	scanner := bufio.NewScanner(r)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		patternType := "release_group"
		pattern := line
		reason := "community blocklist"

		// Check for type prefix.
		if idx := strings.Index(line, ":"); idx > 0 {
			prefix := strings.ToLower(line[:idx])
			switch prefix {
			case "group":
				patternType = "release_group"
				pattern = strings.TrimSpace(line[idx+1:])
			case "regex":
				patternType = "title_regex"
				pattern = strings.TrimSpace(line[idx+1:])
			case "indexer":
				patternType = "indexer"
				pattern = strings.TrimSpace(line[idx+1:])
			case "contains":
				patternType = "title_contains"
				pattern = strings.TrimSpace(line[idx+1:])
			case "file":
				patternType = "file_pattern"
				pattern = strings.TrimSpace(line[idx+1:])
			}
		}

		if pattern == "" {
			continue
		}

		rules = append(rules, database.BlocklistRule{
			Pattern:     pattern,
			PatternType: patternType,
			Reason:      reason,
		})
	}

	return rules, scanner.Err()
}
