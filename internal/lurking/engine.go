package lurking

//go:generate mockgen -destination=mock_store_test.go -package=lurking github.com/lusoris/lurkarr/internal/lurking Store

import (
	"context"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"slices"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/metrics"
	"github.com/lusoris/lurkarr/internal/notifications"
)

// Store abstracts the database operations needed by the lurking Engine.
type Store interface {
	GetAppSettings(ctx context.Context, appType database.AppType) (*database.AppSettings, error)
	ListEnabledInstances(ctx context.Context, appType database.AppType) ([]database.AppInstance, error)
	GetCurrentHourHits(ctx context.Context, appType database.AppType, instanceID uuid.UUID) (int, error)
	GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error)
	GetLastReset(ctx context.Context, appType database.AppType, instanceID uuid.UUID) (*time.Time, error)
	ResetState(ctx context.Context, appType database.AppType, instanceID uuid.UUID) error
	IsProcessed(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int, lurkType string) (bool, error)
	MarkProcessed(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int, lurkType string) error
	GetProcessedTimes(ctx context.Context, appType database.AppType, instanceID uuid.UUID, operation string) (map[int]time.Time, error)
	AddLurkHistory(ctx context.Context, appType database.AppType, instanceID uuid.UUID, instanceName string, mediaID int, title, lurkType string) error
	IncrementStats(ctx context.Context, appType database.AppType, instanceID uuid.UUID, missing, upgrades int64) error
	IncrementHourlyHits(ctx context.Context, appType database.AppType, instanceID uuid.UUID, count int) error
	RecordSearchFailure(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int) error
	ClearSearchFailure(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int) error
	GetSearchFailureCounts(ctx context.Context, appType database.AppType, instanceID uuid.UUID) (map[int]int, error)
}

// Engine manages lurking goroutines for all app types.
type Engine struct {
	db       Store
	logger   *logging.Logger
	notifier notifications.Notifier
	cancel   context.CancelFunc
}

// New creates a new lurking engine.
func New(db Store, logger *logging.Logger) *Engine {
	return &Engine{db: db, logger: logger}
}

// SetNotifier sets an optional notification manager.
func (e *Engine) SetNotifier(n notifications.Notifier) {
	e.notifier = n
}

// Start launches a lurking goroutine for each app type that has a registered lurker.
func (e *Engine) Start(ctx context.Context) {
	ctx, e.cancel = context.WithCancel(ctx)
	started := 0
	for _, appType := range database.AllAppTypes() {
		if LurkerFor(appType) == nil {
			continue
		}
		go e.lurkLoop(ctx, appType)
		started++
	}
	slog.Info("lurking engine started", "app_types", started)
}

// RunOnce performs a single lurk pass for all app types and returns.
func (e *Engine) RunOnce(ctx context.Context) {
	for _, appType := range database.AllAppTypes() {
		if LurkerFor(appType) == nil {
			continue
		}
		log := e.logger.ForApp(string(appType))
		settings, err := e.db.GetAppSettings(ctx, appType)
		if err != nil {
			log.Error("run-once: failed to load settings", "error", err)
			continue
		}
		instances, err := e.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("run-once: failed to list instances", "error", err)
			continue
		}
		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			if err := e.lurkInstance(ctx, log, appType, settings, inst); err != nil {
				metrics.LurkErrors.WithLabelValues(string(appType), inst.Name).Inc()
			}
		}
	}
	slog.Info("lurking engine run-once complete")
}

// RunOnceForApp performs a single lurk pass for a specific app type and mode.
// Mode can be "missing", "upgrade", or "all".
func (e *Engine) RunOnceForApp(ctx context.Context, appType database.AppType, mode string) error {
	if LurkerFor(appType) == nil {
		return fmt.Errorf("no lurker for app type: %s", appType)
	}
	log := e.logger.ForApp(string(appType))
	settings, err := e.db.GetAppSettings(ctx, appType)
	if err != nil {
		return fmt.Errorf("load settings: %w", err)
	}

	// Temporarily override counts based on mode
	switch mode {
	case "missing":
		settings.LurkUpgradeCount = 0
	case "upgrade":
		settings.LurkMissingCount = 0
	case "all":
		// keep both
	default:
		return fmt.Errorf("unknown lurk mode: %s", mode)
	}

	instances, err := e.db.ListEnabledInstances(ctx, appType)
	if err != nil {
		return fmt.Errorf("list instances: %w", err)
	}
	for _, inst := range instances {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if err := e.lurkInstance(ctx, log, appType, settings, inst); err != nil {
			log.Error("run-once-for-app: instance error", "instance", inst.Name, "error", err)
		}
	}
	return nil
}

// Stop cancels all lurking goroutines.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
}

func (e *Engine) lurkLoop(ctx context.Context, appType database.AppType) {
	log := e.logger.ForApp(string(appType))
	consecutiveErrors := 0
	// Per-instance backoff: track consecutive errors per instance so one failing
	// instance doesn't slow down the entire app type.
	instanceErrors := make(map[uuid.UUID]int)
	instanceLastRun := make(map[uuid.UUID]time.Time)
	for {
		settings, err := e.db.GetAppSettings(ctx, appType)
		if err != nil {
			log.Error("failed to load settings", "error", err)
			consecutiveErrors++
			if !e.sleep(ctx, backoff(consecutiveErrors)) {
				return
			}
			continue
		}

		instances, err := e.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("failed to list instances", "error", err)
			consecutiveErrors++
			if !e.sleep(ctx, backoff(consecutiveErrors)) {
				return
			}
			continue
		}

		if len(instances) == 0 {
			consecutiveErrors = 0
			if !e.sleep(ctx, time.Duration(settings.SleepDuration)*time.Second) {
				return
			}
			continue
		}

		hadError := false
		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			// Per-instance backoff: skip if this instance is still cooling down.
			if errCount := instanceErrors[inst.ID]; errCount > 0 {
				cooldown := backoff(errCount)
				if last, ok := instanceLastRun[inst.ID]; ok && time.Since(last) < cooldown {
					log.Debug("instance on backoff, skipping",
						"instance", inst.Name, "errors", errCount, "cooldown", cooldown)
					continue
				}
			}
			start := time.Now()
			instanceLastRun[inst.ID] = start
			if err := e.lurkInstance(ctx, log, appType, settings, inst); err != nil {
				hadError = true
				instanceErrors[inst.ID]++
				metrics.LurkErrors.WithLabelValues(string(appType), inst.Name).Inc()
			} else {
				instanceErrors[inst.ID] = 0
			}
			metrics.LurkDuration.WithLabelValues(string(appType), inst.Name).Observe(time.Since(start).Seconds())
		}

		if hadError {
			consecutiveErrors++
		} else {
			consecutiveErrors = 0
		}

		sleepDur := time.Duration(settings.SleepDuration) * time.Second
		if consecutiveErrors > 0 {
			sleepDur = backoff(consecutiveErrors)
		}
		if !e.sleep(ctx, sleepDur) {
			return
		}
	}
}

// backoff returns an exponential backoff duration capped at 5 minutes.
func backoff(errors int) time.Duration {
	d := time.Duration(1<<min(errors, 8)) * time.Second // 2s, 4s, 8s, ...
	if d > 5*time.Minute {
		d = 5 * time.Minute
	}
	return d
}

type lurkableItem struct {
	ID       int
	Title    string
	SortDate time.Time // Date used for newest/oldest sort; zero if unavailable
}

// parseArrDate parses ISO 8601 date strings returned by *arr APIs.
func parseArrDate(s string) time.Time {
	for _, layout := range []string{time.RFC3339, "2006-01-02T15:04:05Z", "2006-01-02"} {
		if t, err := time.Parse(layout, s); err == nil {
			return t
		}
	}
	return time.Time{}
}

func (e *Engine) lurkInstance(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance) error {
	log = log.With("instance", inst.Name)

	if e.notifier != nil {
		e.notifier.Notify(ctx, notifications.Event{
			Type:     notifications.EventLurkStarted,
			Title:    "Lurk Started",
			Message:  fmt.Sprintf("Starting lurk for %s/%s", appType, inst.Name),
			AppType:  string(appType),
			Instance: inst.Name,
		})
	}

	// Check hourly cap
	hits, err := e.db.GetCurrentHourHits(ctx, appType, inst.ID)
	if err != nil {
		log.Error("failed to check hourly cap", "error", err)
		return err
	}
	if settings.HourlyCap > 0 && hits >= settings.HourlyCap {
		log.Info("hourly cap reached, skipping", "hits", hits, "cap", settings.HourlyCap)
		return nil
	}

	// Check state reset
	genSettings, err := e.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("failed to load general settings", "error", err)
		return err
	}
	lastReset, err := e.db.GetLastReset(ctx, appType, inst.ID)
	if err != nil {
		log.Error("failed to get last reset", "error", err)
		return err
	}
	if lastReset != nil && time.Since(*lastReset) > time.Duration(genSettings.StatefulResetHours)*time.Hour {
		if err := e.db.ResetState(ctx, appType, inst.ID); err != nil {
			log.Error("failed to reset state", "error", err)
		} else {
			log.Info("state reset performed", "hours_since_last", time.Since(*lastReset).Hours())
		}
	}

	client := arrclient.NewClientForInstance(inst.APIURL, inst.APIKey, genSettings.APITimeout, genSettings.SSLVerify)

	// Check download queue size — skip if queue is already large
	if genSettings.MaxDownloadQueueSize > 0 {
		lurker := LurkerFor(appType)
		if lurker != nil {
			queue, err := lurker.GetQueue(ctx, client)
			if err != nil {
				log.Warn("failed to check queue size", "error", err)
			} else if queue != nil && queue.TotalRecords >= genSettings.MaxDownloadQueueSize {
				log.Info("download queue at capacity, skipping lurk",
					"queue_size", queue.TotalRecords, "max_size", genSettings.MaxDownloadQueueSize)
				return nil
			}
		}
	}

	remaining := settings.HourlyCap - hits
	if settings.HourlyCap == 0 {
		remaining = 1<<31 - 1 // unlimited
	}

	// Lurk missing items
	var missingCount, upgradeCount int
	if settings.LurkMissingCount > 0 {
		count := min(settings.LurkMissingCount, remaining)
		missingCount = e.lurkMissing(ctx, log, appType, settings, inst, client, count)
		remaining -= missingCount
	}

	// Lurk upgrades
	if settings.LurkUpgradeCount > 0 && remaining > 0 {
		count := min(settings.LurkUpgradeCount, remaining)
		upgradeCount = e.lurkUpgrades(ctx, log, appType, settings, inst, client, count)
	}

	if e.notifier != nil && (missingCount > 0 || upgradeCount > 0) {
		e.notifier.Notify(ctx, notifications.Event{
			Type:     notifications.EventLurkCompleted,
			Title:    "Lurk Completed",
			Message:  fmt.Sprintf("Found %d missing, %d upgrades", missingCount, upgradeCount),
			AppType:  string(appType),
			Instance: inst.Name,
			Fields: map[string]string{
				"Missing":  fmt.Sprintf("%d", missingCount),
				"Upgrades": fmt.Sprintf("%d", upgradeCount),
			},
		})
	}
	return nil
}

func (e *Engine) lurkMissing(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance, client *arrclient.Client, maxCount int) int {
	items, err := e.getMissingItems(ctx, appType, client)
	if err != nil {
		log.Error("failed to get missing items", "error", err)
		return 0
	}

	var candidates []lurkableItem
	var processedTimes map[int]time.Time

	if settings.SelectionMode == SelectLeastRecent {
		// In least_recent mode, include ALL items and sort by last-processed time.
		processedTimes, err = e.db.GetProcessedTimes(ctx, appType, inst.ID, "missing")
		if err != nil {
			log.Error("failed to get processed times", "error", err)
			return 0
		}
		candidates = items
	} else {
		// Standard modes: filter out already-processed items.
		for _, item := range items {
			processed, err := e.db.IsProcessed(ctx, appType, inst.ID, item.ID, "missing")
			if err != nil {
				log.Error("failed to check processed", "error", err)
				continue
			}
			if !processed {
				candidates = append(candidates, item)
			}
		}
	}

	if len(candidates) == 0 {
		log.Debug("no eligible missing items")
		return 0
	}

	// Select items
	selected := selectItems(candidates, maxCount, settings.SelectionMode, processedTimes)

	// Filter out items that have exceeded the search failure limit.
	var failureCounts map[int]int
	if settings.MaxSearchFailures > 0 {
		failureCounts, err = e.db.GetSearchFailureCounts(ctx, appType, inst.ID)
		if err != nil {
			log.Warn("failed to get search failure counts", "error", err)
		} else {
			filtered := selected[:0]
			for _, item := range selected {
				if count, ok := failureCounts[item.ID]; ok && count >= settings.MaxSearchFailures {
					log.Debug("skipping item (search failure limit reached)", "title", item.Title, "media_id", item.ID, "failures", count)
					continue
				}
				filtered = append(filtered, item)
			}
			selected = filtered
		}
	}

	// Trigger searches
	lurked := 0
	for _, item := range selected {
		if err := e.triggerSearch(ctx, appType, client, item.ID); err != nil {
			log.Error("search command failed", "media_id", item.ID, "error", err)
			if settings.MaxSearchFailures > 0 {
				if ferr := e.db.RecordSearchFailure(ctx, appType, inst.ID, item.ID); ferr != nil {
					log.Warn("failed to record search failure", "media_id", item.ID, "error", ferr)
				}
			}
			continue
		}
		if settings.MaxSearchFailures > 0 {
			if ferr := e.db.ClearSearchFailure(ctx, appType, inst.ID, item.ID); ferr != nil {
				log.Warn("failed to clear search failure", "media_id", item.ID, "error", ferr)
			}
		}
		metrics.LurkSearchesTotal.WithLabelValues(string(appType), inst.Name).Inc()
		if err := e.db.MarkProcessed(ctx, appType, inst.ID, item.ID, "missing"); err != nil {
			log.Error("failed to mark processed", "error", err)
		}
		if err := e.db.AddLurkHistory(ctx, appType, inst.ID, inst.Name, item.ID, item.Title, "missing"); err != nil {
			log.Error("failed to add history", "error", err)
		}
		lurked++
		log.Info("lurked missing item", "title", item.Title, "media_id", item.ID)
	}

	if lurked > 0 {
		if err := e.db.IncrementStats(ctx, appType, inst.ID, int64(lurked), 0); err != nil {
			log.Warn("failed to increment stats", "error", err)
		}
		if err := e.db.IncrementHourlyHits(ctx, appType, inst.ID, lurked); err != nil {
			log.Warn("failed to increment hourly hits", "error", err)
		}
		metrics.LurkMissingFound.WithLabelValues(string(appType), inst.Name).Add(float64(lurked))
	}
	return lurked
}

func (e *Engine) lurkUpgrades(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance, client *arrclient.Client, maxCount int) int {
	items, err := e.getUpgradeItems(ctx, appType, client)
	if err != nil {
		log.Error("failed to get upgrade items", "error", err)
		return 0
	}

	var candidates []lurkableItem
	var processedTimes map[int]time.Time

	if settings.SelectionMode == SelectLeastRecent {
		processedTimes, err = e.db.GetProcessedTimes(ctx, appType, inst.ID, "upgrade")
		if err != nil {
			log.Error("failed to get processed times", "error", err)
			return 0
		}
		candidates = items
	} else {
		for _, item := range items {
			processed, err := e.db.IsProcessed(ctx, appType, inst.ID, item.ID, "upgrade")
			if err != nil {
				continue
			}
			if !processed {
				candidates = append(candidates, item)
			}
		}
	}

	if len(candidates) == 0 {
		return 0
	}

	selected := selectItems(candidates, maxCount, settings.SelectionMode, processedTimes)

	// Filter out items that have exceeded the search failure limit.
	var failureCounts map[int]int
	if settings.MaxSearchFailures > 0 {
		failureCounts, err = e.db.GetSearchFailureCounts(ctx, appType, inst.ID)
		if err != nil {
			log.Warn("failed to get search failure counts", "error", err)
		} else {
			filtered := selected[:0]
			for _, item := range selected {
				if count, ok := failureCounts[item.ID]; ok && count >= settings.MaxSearchFailures {
					log.Debug("skipping upgrade (search failure limit reached)", "title", item.Title, "media_id", item.ID, "failures", count)
					continue
				}
				filtered = append(filtered, item)
			}
			selected = filtered
		}
	}

	upgraded := 0
	for _, item := range selected {
		if err := e.triggerSearch(ctx, appType, client, item.ID); err != nil {
			log.Error("upgrade search failed", "media_id", item.ID, "error", err)
			if settings.MaxSearchFailures > 0 {
				if ferr := e.db.RecordSearchFailure(ctx, appType, inst.ID, item.ID); ferr != nil {
					log.Warn("failed to record search failure", "media_id", item.ID, "error", ferr)
				}
			}
			continue
		}
		if settings.MaxSearchFailures > 0 {
			if ferr := e.db.ClearSearchFailure(ctx, appType, inst.ID, item.ID); ferr != nil {
				log.Warn("failed to clear search failure", "media_id", item.ID, "error", ferr)
			}
		}
		metrics.LurkSearchesTotal.WithLabelValues(string(appType), inst.Name).Inc()
		if err := e.db.MarkProcessed(ctx, appType, inst.ID, item.ID, "upgrade"); err != nil {
			log.Warn("failed to mark processed", "media_id", item.ID, "error", err)
		}
		if err := e.db.AddLurkHistory(ctx, appType, inst.ID, inst.Name, item.ID, item.Title, "upgrade"); err != nil {
			log.Warn("failed to add lurk history", "media_id", item.ID, "error", err)
		}
		upgraded++
		log.Info("lurked upgrade", "title", item.Title, "media_id", item.ID)
	}

	if upgraded > 0 {
		if err := e.db.IncrementStats(ctx, appType, inst.ID, 0, int64(upgraded)); err != nil {
			log.Warn("failed to increment stats", "error", err)
		}
		if err := e.db.IncrementHourlyHits(ctx, appType, inst.ID, upgraded); err != nil {
			log.Warn("failed to increment hourly hits", "error", err)
		}
		metrics.LurkUpgradesFound.WithLabelValues(string(appType), inst.Name).Add(float64(upgraded))
	}
	return upgraded
}

func (e *Engine) getMissingItems(ctx context.Context, appType database.AppType, client *arrclient.Client) ([]lurkableItem, error) {
	lurker := LurkerFor(appType)
	if lurker == nil {
		return nil, nil
	}
	return lurker.GetMissing(ctx, client)
}

func (e *Engine) getUpgradeItems(ctx context.Context, appType database.AppType, client *arrclient.Client) ([]lurkableItem, error) {
	lurker := LurkerFor(appType)
	if lurker == nil {
		return nil, nil
	}
	return lurker.GetUpgrades(ctx, client)
}

func (e *Engine) triggerSearch(ctx context.Context, appType database.AppType, client *arrclient.Client, mediaID int) error {
	lurker := LurkerFor(appType)
	if lurker == nil {
		return nil
	}
	return lurker.Search(ctx, client, mediaID)
}

// Selection mode constants.
const (
	SelectRandom      = "random"
	SelectNewest      = "newest"
	SelectOldest      = "oldest"
	SelectLeastRecent = "least_recent"
)

func selectItems(items []lurkableItem, count int, mode string, processedTimes map[int]time.Time) []lurkableItem {
	if count >= len(items) {
		count = len(items)
	}
	switch mode {
	case SelectNewest:
		slices.SortFunc(items, func(a, b lurkableItem) int {
			return b.SortDate.Compare(a.SortDate) // newest first
		})
	case SelectOldest:
		slices.SortFunc(items, func(a, b lurkableItem) int {
			return a.SortDate.Compare(b.SortDate) // oldest first
		})
	case SelectLeastRecent:
		slices.SortFunc(items, func(a, b lurkableItem) int {
			ta, okA := processedTimes[a.ID]
			tb, okB := processedTimes[b.ID]
			// Never-processed items come first
			if !okA && !okB {
				return 0
			}
			if !okA {
				return -1
			}
			if !okB {
				return 1
			}
			return ta.Compare(tb) // oldest processed first
		})
	default: // "random"
		rand.Shuffle(len(items), func(i, j int) {
			items[i], items[j] = items[j], items[i]
		})
	}
	return items[:count]
}

func (e *Engine) sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
