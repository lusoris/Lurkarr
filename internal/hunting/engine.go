package hunting

import (
	"context"
	"log/slog"
	"math/rand/v2"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
)

// Engine manages hunting goroutines for all app types.
type Engine struct {
	db     *database.DB
	logger *logging.Logger
	cancel context.CancelFunc
}

// New creates a new hunting engine.
func New(db *database.DB, logger *logging.Logger) *Engine {
	return &Engine{db: db, logger: logger}
}

// Start launches a hunting goroutine for each app type.
func (e *Engine) Start(ctx context.Context) {
	ctx, e.cancel = context.WithCancel(ctx)
	for _, appType := range database.AllAppTypes() {
		go e.huntLoop(ctx, appType)
	}
	slog.Info("hunting engine started", "app_types", len(database.AllAppTypes()))
}

// Stop cancels all hunting goroutines.
func (e *Engine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
}

func (e *Engine) huntLoop(ctx context.Context, appType database.AppType) {
	log := e.logger.ForApp(string(appType))
	consecutiveErrors := 0
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
			if err := e.huntInstance(ctx, log, appType, settings, inst); err != nil {
				hadError = true
			}
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

type huntableItem struct {
	ID    int
	Title string
}

func (e *Engine) huntInstance(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance) error {
	log = log.With("instance", inst.Name)

	// Check hourly cap
	hits, err := e.db.GetCurrentHourHits(ctx, appType)
	if err != nil {
		log.Error("failed to check hourly cap", "error", err)
		return err
	}
	if hits >= settings.HourlyCap {
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

	client := arrclient.NewClient(
		inst.APIURL, inst.APIKey,
		time.Duration(genSettings.APITimeout)*time.Second,
		genSettings.SSLVerify,
	)

	// Check minimum download queue size — skip if queue is already large
	if genSettings.MinDownloadQueueSize > 0 {
		hunter := HunterFor(appType)
		if hunter != nil {
			queue, err := hunter.GetQueue(ctx, client)
			if err != nil {
				log.Warn("failed to check queue size", "error", err)
			} else if queue != nil && queue.TotalRecords >= genSettings.MinDownloadQueueSize {
				log.Info("download queue at capacity, skipping hunt",
					"queue_size", queue.TotalRecords, "min_size", genSettings.MinDownloadQueueSize)
				return nil
			}
		}
	}

	remaining := settings.HourlyCap - hits

	// Hunt missing items
	if settings.HuntMissingCount > 0 {
		count := min(settings.HuntMissingCount, remaining)
		hunted := e.huntMissing(ctx, log, appType, settings, inst, client, count)
		remaining -= hunted
	}

	// Hunt upgrades
	if settings.HuntUpgradeCount > 0 && remaining > 0 {
		count := min(settings.HuntUpgradeCount, remaining)
		e.huntUpgrades(ctx, log, appType, settings, inst, client, count)
	}
	return nil
}

func (e *Engine) huntMissing(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance, client *arrclient.Client, maxCount int) int {
	items, err := e.getMissingItems(ctx, appType, client)
	if err != nil {
		log.Error("failed to get missing items", "error", err)
		return 0
	}

	// Filter already processed
	var unprocessed []huntableItem
	for _, item := range items {
		processed, err := e.db.IsProcessed(ctx, appType, inst.ID, item.ID, "missing")
		if err != nil {
			log.Error("failed to check processed", "error", err)
			continue
		}
		if !processed {
			unprocessed = append(unprocessed, item)
		}
	}

	if len(unprocessed) == 0 {
		log.Debug("no unprocessed missing items")
		return 0
	}

	// Select items
	selected := selectItems(unprocessed, maxCount, settings.RandomSelection)

	// Trigger searches
	hunted := 0
	for _, item := range selected {
		if err := e.triggerSearch(ctx, appType, client, item.ID); err != nil {
			log.Error("search command failed", "media_id", item.ID, "error", err)
			continue
		}
		if err := e.db.MarkProcessed(ctx, appType, inst.ID, item.ID, "missing"); err != nil {
			log.Error("failed to mark processed", "error", err)
		}
		if err := e.db.AddHuntHistory(ctx, appType, inst.ID, inst.Name, item.ID, item.Title, "missing"); err != nil {
			log.Error("failed to add history", "error", err)
		}
		hunted++
		log.Info("hunted missing item", "title", item.Title, "media_id", item.ID)
	}

	if hunted > 0 {
		_ = e.db.IncrementStats(ctx, appType, int64(hunted), 0)
		_ = e.db.IncrementHourlyHits(ctx, appType, hunted)
	}
	return hunted
}

func (e *Engine) huntUpgrades(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance, client *arrclient.Client, maxCount int) int {
	items, err := e.getUpgradeItems(ctx, appType, client)
	if err != nil {
		log.Error("failed to get upgrade items", "error", err)
		return 0
	}

	var unprocessed []huntableItem
	for _, item := range items {
		processed, err := e.db.IsProcessed(ctx, appType, inst.ID, item.ID, "upgrade")
		if err != nil {
			continue
		}
		if !processed {
			unprocessed = append(unprocessed, item)
		}
	}

	if len(unprocessed) == 0 {
		return 0
	}

	selected := selectItems(unprocessed, maxCount, settings.RandomSelection)

	upgraded := 0
	for _, item := range selected {
		if err := e.triggerSearch(ctx, appType, client, item.ID); err != nil {
			log.Error("upgrade search failed", "media_id", item.ID, "error", err)
			continue
		}
		_ = e.db.MarkProcessed(ctx, appType, inst.ID, item.ID, "upgrade")
		_ = e.db.AddHuntHistory(ctx, appType, inst.ID, inst.Name, item.ID, item.Title, "upgrade")
		upgraded++
		log.Info("hunted upgrade", "title", item.Title, "media_id", item.ID)
	}

	if upgraded > 0 {
		_ = e.db.IncrementStats(ctx, appType, 0, int64(upgraded))
		_ = e.db.IncrementHourlyHits(ctx, appType, upgraded)
	}
	return upgraded
}

func (e *Engine) getMissingItems(ctx context.Context, appType database.AppType, client *arrclient.Client) ([]huntableItem, error) {
	hunter := HunterFor(appType)
	if hunter == nil {
		return nil, nil
	}
	return hunter.GetMissing(ctx, client)
}

func (e *Engine) getUpgradeItems(ctx context.Context, appType database.AppType, client *arrclient.Client) ([]huntableItem, error) {
	hunter := HunterFor(appType)
	if hunter == nil {
		return nil, nil
	}
	return hunter.GetUpgrades(ctx, client)
}

func (e *Engine) triggerSearch(ctx context.Context, appType database.AppType, client *arrclient.Client, mediaID int) error {
	hunter := HunterFor(appType)
	if hunter == nil {
		return nil
	}
	return hunter.Search(ctx, client, mediaID)
}

func selectItems(items []huntableItem, count int, random bool) []huntableItem {
	if count >= len(items) {
		return items
	}
	if random {
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
