package metrics

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"

	"github.com/lusoris/lurkarr/internal/database"
)

// counterDef maps a Prometheus counter to its persistence name.
type counterDef struct {
	name    string
	counter *prometheus.CounterVec
	labels  []string
}

// persistedCounters defines which counters survive restarts.
var persistedCounters = []counterDef{
	{"lurkarr_queue_cleaner_items_removed_total", QueueCleanerItemsRemoved, []string{"app_type", "instance"}},
	{"lurkarr_queue_cleaner_strikes_total", QueueCleanerStrikes, []string{"app_type", "instance"}},
	{"lurkarr_queue_cleaner_blocklist_additions_total", QueueCleanerBlocklistAdditions, []string{"app_type", "instance"}},
	{"lurkarr_lurk_searches_total", LurkSearchesTotal, []string{"app_type", "instance"}},
	{"lurkarr_lurk_missing_found_total", LurkMissingFound, []string{"app_type", "instance"}},
	{"lurkarr_lurk_upgrades_found_total", LurkUpgradesFound, []string{"app_type", "instance"}},
}

// Persister periodically flushes Prometheus counters to the database
// and restores them on startup so cumulative values survive restarts.
type Persister struct {
	db *database.DB
	mu sync.Mutex
	// baseline tracks the counter value at last flush, so we compute deltas correctly.
	baseline map[string]float64
}

// NewPersister creates a new metrics Persister.
func NewPersister(db *database.DB) *Persister {
	return &Persister{
		db:       db,
		baseline: make(map[string]float64),
	}
}

// Restore loads saved counter values from the database and adds them to
// Prometheus counters. This should be called once at startup.
func (p *Persister) Restore(ctx context.Context) error {
	rows, err := p.db.GetAllCounters(ctx)
	if err != nil {
		return fmt.Errorf("load persisted counters: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	restored := 0
	for _, row := range rows {
		def := findDef(row.MetricName)
		if def == nil {
			continue
		}
		labels := parseLabelKey(row.LabelKey, def.labels)
		if labels == nil {
			continue
		}
		counter, err := def.counter.GetMetricWith(labels)
		if err != nil {
			slog.Warn("failed to restore counter", "metric", row.MetricName, "error", err)
			continue
		}
		counter.Add(float64(row.Value))
		// Set baseline so the first flush writes only the delta above the restored value.
		key := counterKey(row.MetricName, row.LabelKey)
		p.baseline[key] = float64(row.Value)
		restored++
	}

	if restored > 0 {
		slog.Info("restored persisted counters", "count", restored)
	}
	return nil
}

// FlushLoop runs a periodic flush every interval until ctx is cancelled.
func (p *Persister) FlushLoop(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := p.FlushNow(ctx); err != nil {
				slog.Warn("metrics flush failed", "error", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// FlushNow writes the current counter values to the database.
func (p *Persister) FlushNow(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	var batch []database.PersistentCounter

	for _, def := range persistedCounters {
		metrics := collectCounterMetrics(def.counter)
		for labelKey, value := range metrics {
			key := counterKey(def.name, labelKey)
			if value == p.baseline[key] {
				continue // no change, skip
			}
			batch = append(batch, database.PersistentCounter{
				MetricName: def.name,
				LabelKey:   labelKey,
				Value:      int64(value),
			})
			p.baseline[key] = value
		}
	}

	if len(batch) == 0 {
		return nil
	}

	if err := p.db.UpsertCounters(ctx, batch); err != nil {
		return fmt.Errorf("upsert counters: %w", err)
	}

	slog.Debug("flushed metrics", "count", len(batch))
	return nil
}

// collectCounterMetrics gathers all label combinations and their values
// from a CounterVec using the Collect/Describe interface.
func collectCounterMetrics(cv *prometheus.CounterVec) map[string]float64 {
	ch := make(chan prometheus.Metric, 64)
	go func() {
		cv.Collect(ch)
		close(ch)
	}()

	result := make(map[string]float64)
	for m := range ch {
		d := &dto.Metric{}
		if err := m.Write(d); err != nil {
			continue
		}
		if d.Counter == nil {
			continue
		}
		labelKey := buildLabelKey(d.Label)
		result[labelKey] = d.Counter.GetValue()
	}
	return result
}

// buildLabelKey creates "value1/value2" from label pairs, sorted by label name.
func buildLabelKey(labels []*dto.LabelPair) string {
	// Labels from Prometheus are already sorted by name.
	parts := make([]string, 0, len(labels))
	for _, lp := range labels {
		parts = append(parts, lp.GetValue())
	}
	return strings.Join(parts, "/")
}

// parseLabelKey reverses buildLabelKey back into a prometheus.Labels map.
func parseLabelKey(key string, labelNames []string) prometheus.Labels {
	if key == "" && len(labelNames) == 0 {
		return prometheus.Labels{}
	}
	parts := strings.Split(key, "/")
	if len(parts) != len(labelNames) {
		return nil
	}
	labels := make(prometheus.Labels, len(labelNames))
	for i, name := range labelNames {
		labels[name] = parts[i]
	}
	return labels
}

func findDef(name string) *counterDef {
	for i := range persistedCounters {
		if persistedCounters[i].name == name {
			return &persistedCounters[i]
		}
	}
	return nil
}

func counterKey(metricName, labelKey string) string {
	return metricName + "|" + labelKey
}
