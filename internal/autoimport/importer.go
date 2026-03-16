package autoimport

//go:generate mockgen -destination=mock_store_test.go -package=autoimport github.com/lusoris/lurkarr/internal/autoimport Store

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/lurking"
	"github.com/lusoris/lurkarr/internal/metrics"
	"github.com/lusoris/lurkarr/internal/notifications"
)

// Store abstracts the database operations needed by the Importer.
type Store interface {
	ListEnabledInstances(ctx context.Context, appType database.AppType) ([]database.AppInstance, error)
	GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error)
	LogAutoImport(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int, mediaTitle string, queueItemID int, action, reason string) error
}

// Importer watches for downloads stuck with "Unable to Import Automatically"
// and attempts to resolve them by triggering a manual import if the content
// matches the expected media by ID and the custom format score is acceptable.
type Importer struct {
	db       Store
	logger   *logging.Logger
	notifier notifications.Notifier
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// New creates a new auto-importer.
func New(db Store, logger *logging.Logger) *Importer {
	return &Importer{db: db, logger: logger}
}

// SetNotifier sets an optional notification manager.
func (imp *Importer) SetNotifier(n notifications.Notifier) {
	imp.notifier = n
}

// Start launches importer goroutines for each app type.
func (imp *Importer) Start(ctx context.Context) {
	ctx, imp.cancel = context.WithCancel(ctx)
	for _, appType := range database.AllAppTypes() {
		if lurking.LurkerFor(appType) == nil {
			continue
		}
		imp.wg.Add(1)
		go imp.importLoop(ctx, appType)
	}
	slog.Info("auto-importer started")
}

// RunOnce performs a single auto-import pass for all app types and returns.
func (imp *Importer) RunOnce(ctx context.Context) {
	for _, appType := range database.AllAppTypes() {
		if lurking.LurkerFor(appType) == nil {
			continue
		}
		log := imp.logger.ForApp(string(appType))
		instances, err := imp.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("run-once: failed to list instances", "error", err)
			continue
		}
		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			imp.checkInstance(ctx, log, appType, inst)
		}
	}
	slog.Info("auto-importer run-once complete")
}

// Stop cancels all importer goroutines.
func (imp *Importer) Stop() {
	if imp.cancel != nil {
		imp.cancel()
	}
	imp.wg.Wait()
}

func (imp *Importer) importLoop(ctx context.Context, appType database.AppType) {
	defer imp.wg.Done()
	log := imp.logger.ForApp(string(appType))

	for {
		interval := 5 * time.Minute
		if settings, err := imp.db.GetGeneralSettings(ctx); err == nil && settings.AutoImportIntervalMinutes > 0 {
			interval = time.Duration(settings.AutoImportIntervalMinutes) * time.Minute
		}
		if !sleep(ctx, interval) {
			return
		}

		instances, err := imp.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("failed to list instances", "error", err)
			continue
		}

		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			imp.checkInstance(ctx, log, appType, inst)
		}
	}
}

func (imp *Importer) checkInstance(ctx context.Context, log *slog.Logger, appType database.AppType, inst database.AppInstance) {
	log = log.With("instance", inst.Name)
	metrics.AutoimportRunsTotal.WithLabelValues(string(appType), inst.Name).Inc()

	lurker := lurking.LurkerFor(appType)
	if lurker == nil {
		return
	}

	genSettings, err := imp.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("failed to load general settings", "error", err)
		return
	}

	client := arrclient.NewClientForInstance(inst.APIURL, inst.APIKey, genSettings.APITimeout, genSettings.SSLVerify)

	queue, err := lurker.GetQueue(ctx, client)
	if err != nil {
		log.Error("failed to get queue", "error", err)
		metrics.AutoimportErrors.WithLabelValues(string(appType), inst.Name).Inc()
		return
	}

	apiVersion := apiVersionFor(appType)

	for _, record := range queue.Records {
		if !isImportStuck(record) {
			continue
		}

		mediaID := record.MediaID()
		if mediaID == 0 {
			continue
		}

		log.Info("found stuck import, checking manual import options",
			"title", record.Title,
			"media_id", mediaID,
			"status", record.TrackedDownloadStatus,
			"messages", formatStatusMessages(record.StatusMessages))

		if imp.notifier != nil {
			imp.notifier.Notify(ctx, notifications.Event{
				Type:     notifications.EventDownloadStuck,
				Title:    "Download Stuck",
				Message:  record.Title,
				AppType:  string(appType),
				Instance: inst.Name,
				Fields: map[string]string{
					"Status": record.TrackedDownloadStatus,
				},
			})
		}

		// Try manual import: check if files are available and have acceptable quality
		if record.DownloadID != "" && apiVersion != "" {
			items, err := client.GetManualImport(ctx, apiVersion, record.DownloadID)
			if err != nil {
				log.Warn("failed to get manual import options", "error", err)
			} else if len(items) > 0 {
				// Check if any available file has a better or equal custom format score
				best := items[0]
				for _, item := range items[1:] {
					if item.CustomFormatScore > best.CustomFormatScore {
						best = item
					}
				}
				if len(best.Rejections) == 0 {
					log.Info("triggering manual import",
						"file", best.Name,
						"score", best.CustomFormatScore,
						"queue_score", record.CustomFormatScore)

					best.ImportMode = "move"
					if err := client.PostManualImport(ctx, apiVersion, []arrclient.ManualImportItem{best}); err != nil {
						log.Error("failed to trigger manual import", "title", record.Title, "error", err)
						metrics.AutoimportErrors.WithLabelValues(string(appType), inst.Name).Inc()
					} else {
						if err := imp.db.LogAutoImport(ctx, appType, inst.ID, mediaID, record.Title, record.ID, "manual_import_triggered", best.Name); err != nil {
							log.Warn("failed to log auto import", "title", record.Title, "error", err)
						}
						metrics.AutoimportActionsTotal.WithLabelValues(string(appType), inst.Name, "manual_import").Inc()
					}
					continue
				}
			}
		}

		// Fallback: trigger rescan which often resolves import issues
		if err := triggerRescan(ctx, client, appType, mediaID); err != nil {
			log.Error("failed to trigger rescan", "title", record.Title, "error", err)
			continue
		}

		if err := imp.db.LogAutoImport(ctx, appType, inst.ID, mediaID, record.Title, record.ID, "rescan_triggered", formatStatusMessages(record.StatusMessages)); err != nil {
			log.Warn("failed to log auto import", "title", record.Title, "error", err)
		}
	}
}

// isImportStuck checks if a queue record has import issues.
func isImportStuck(r arrclient.QueueRecord) bool {
	if r.TrackedDownloadState != "importPending" {
		return false
	}

	// Look for specific import failure messages
	for _, sm := range r.StatusMessages {
		for _, msg := range sm.Messages {
			lower := strings.ToLower(msg)
			if strings.Contains(lower, "unable to import") ||
				strings.Contains(lower, "import failed") ||
				strings.Contains(lower, "no matching") {
				return true
			}
		}
	}

	return r.TrackedDownloadStatus == "warning"
}

// triggerRescan sends a RefreshCommand to the arr for the specified media.
func triggerRescan(ctx context.Context, client *arrclient.Client, appType database.AppType, mediaID int) error {
	switch appType {
	case database.AppSonarr:
		_, err := client.SonarrSearchSeries(ctx, mediaID)
		return err
	case database.AppRadarr, database.AppWhisparr, database.AppEros:
		_, err := client.RadarrSearchMovie(ctx, []int{mediaID})
		return err
	case database.AppLidarr:
		_, err := client.LidarrSearchAlbum(ctx, []int{mediaID})
		return err
	case database.AppReadarr:
		_, err := client.ReadarrSearchBook(ctx, []int{mediaID})
		return err
	default:
		return nil
	}
}

func formatStatusMessages(msgs []arrclient.StatusMessage) string {
	var parts []string
	for _, m := range msgs {
		parts = append(parts, m.Messages...)
	}
	return strings.Join(parts, "; ")
}

func apiVersionFor(appType database.AppType) string {
	switch appType {
	case database.AppLidarr, database.AppReadarr:
		return "v1"
	default:
		return "v3"
	}
}

func sleep(ctx context.Context, d time.Duration) bool {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
