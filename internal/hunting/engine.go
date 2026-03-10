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
	for {
		settings, err := e.db.GetAppSettings(ctx, appType)
		if err != nil {
			log.Error("failed to load settings", "error", err)
			if !e.sleep(ctx, 60*time.Second) {
				return
			}
			continue
		}

		instances, err := e.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("failed to list instances", "error", err)
			if !e.sleep(ctx, time.Duration(settings.SleepDuration)*time.Second) {
				return
			}
			continue
		}

		if len(instances) == 0 {
			if !e.sleep(ctx, time.Duration(settings.SleepDuration)*time.Second) {
				return
			}
			continue
		}

		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			e.huntInstance(ctx, log, appType, settings, inst)
		}

		if !e.sleep(ctx, time.Duration(settings.SleepDuration)*time.Second) {
			return
		}
	}
}

type huntableItem struct {
	ID    int
	Title string
}

func (e *Engine) huntInstance(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.AppSettings, inst database.AppInstance) {
	log = log.With("instance", inst.Name)

	// Check hourly cap
	hits, err := e.db.GetCurrentHourHits(ctx, appType)
	if err != nil {
		log.Error("failed to check hourly cap", "error", err)
		return
	}
	if hits >= settings.HourlyCap {
		log.Info("hourly cap reached, skipping", "hits", hits, "cap", settings.HourlyCap)
		return
	}

	// Check state reset
	genSettings, err := e.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("failed to load general settings", "error", err)
		return
	}
	lastReset, err := e.db.GetLastReset(ctx, appType, inst.ID)
	if err != nil {
		log.Error("failed to get last reset", "error", err)
		return
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
	switch appType {
	case database.AppSonarr:
		eps, err := client.SonarrGetMissing(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(eps))
		for i, ep := range eps {
			items[i] = huntableItem{ID: ep.ID, Title: ep.Title}
		}
		return items, nil
	case database.AppRadarr:
		movies, err := client.RadarrGetMissing(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(movies))
		for i, m := range movies {
			items[i] = huntableItem{ID: m.ID, Title: m.Title}
		}
		return items, nil
	case database.AppLidarr:
		albums, err := client.LidarrGetMissing(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(albums))
		for i, a := range albums {
			items[i] = huntableItem{ID: a.ID, Title: a.Title}
		}
		return items, nil
	case database.AppReadarr:
		books, err := client.ReadarrGetMissing(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(books))
		for i, b := range books {
			items[i] = huntableItem{ID: b.ID, Title: b.Title}
		}
		return items, nil
	case database.AppWhisparr:
		movies, err := client.WhisparrGetMissing(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(movies))
		for i, m := range movies {
			items[i] = huntableItem{ID: m.ID, Title: m.Title}
		}
		return items, nil
	case database.AppEros:
		movies, err := client.ErosGetMissing(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(movies))
		for i, m := range movies {
			items[i] = huntableItem{ID: m.ID, Title: m.Title}
		}
		return items, nil
	default:
		return nil, nil
	}
}

func (e *Engine) getUpgradeItems(ctx context.Context, appType database.AppType, client *arrclient.Client) ([]huntableItem, error) {
	switch appType {
	case database.AppSonarr:
		eps, err := client.SonarrGetCutoffUnmet(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(eps))
		for i, ep := range eps {
			items[i] = huntableItem{ID: ep.ID, Title: ep.Title}
		}
		return items, nil
	case database.AppRadarr:
		movies, err := client.RadarrGetCutoffUnmet(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(movies))
		for i, m := range movies {
			items[i] = huntableItem{ID: m.ID, Title: m.Title}
		}
		return items, nil
	case database.AppLidarr:
		albums, err := client.LidarrGetCutoffUnmet(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(albums))
		for i, a := range albums {
			items[i] = huntableItem{ID: a.ID, Title: a.Title}
		}
		return items, nil
	case database.AppReadarr:
		books, err := client.ReadarrGetCutoffUnmet(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(books))
		for i, b := range books {
			items[i] = huntableItem{ID: b.ID, Title: b.Title}
		}
		return items, nil
	case database.AppWhisparr:
		movies, err := client.WhisparrGetCutoffUnmet(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(movies))
		for i, m := range movies {
			items[i] = huntableItem{ID: m.ID, Title: m.Title}
		}
		return items, nil
	case database.AppEros:
		movies, err := client.ErosGetCutoffUnmet(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]huntableItem, len(movies))
		for i, m := range movies {
			items[i] = huntableItem{ID: m.ID, Title: m.Title}
		}
		return items, nil
	default:
		return nil, nil
	}
}

func (e *Engine) triggerSearch(ctx context.Context, appType database.AppType, client *arrclient.Client, mediaID int) error {
	switch appType {
	case database.AppSonarr:
		_, err := client.SonarrSearchEpisode(ctx, []int{mediaID})
		return err
	case database.AppRadarr:
		_, err := client.RadarrSearchMovie(ctx, []int{mediaID})
		return err
	case database.AppLidarr:
		_, err := client.LidarrSearchAlbum(ctx, []int{mediaID})
		return err
	case database.AppReadarr:
		_, err := client.ReadarrSearchBook(ctx, []int{mediaID})
		return err
	case database.AppWhisparr:
		_, err := client.WhisparrSearchMovie(ctx, []int{mediaID})
		return err
	case database.AppEros:
		_, err := client.ErosSearchMovie(ctx, []int{mediaID})
		return err
	default:
		return nil
	}
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
