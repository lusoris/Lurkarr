package autoimport

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/hunting"
	"github.com/lusoris/lurkarr/internal/logging"
)

// Importer watches for downloads stuck with "Unable to Import Automatically"
// and attempts to resolve them by triggering a manual import if the content
// matches the expected media by ID and the custom format score is acceptable.
type Importer struct {
	db     *database.DB
	logger *logging.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new auto-importer.
func New(db *database.DB, logger *logging.Logger) *Importer {
	return &Importer{db: db, logger: logger}
}

// Start launches importer goroutines for each app type.
func (imp *Importer) Start(ctx context.Context) {
	ctx, imp.cancel = context.WithCancel(ctx)
	for _, appType := range database.AllAppTypes() {
		if hunting.HunterFor(appType) == nil {
			continue
		}
		imp.wg.Add(1)
		go imp.importLoop(ctx, appType)
	}
	slog.Info("auto-importer started")
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
		// Check every 5 minutes
		if !sleep(ctx, 5*time.Minute) {
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

	hunter := hunting.HunterFor(appType)
	if hunter == nil {
		return
	}

	genSettings, err := imp.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("failed to load general settings", "error", err)
		return
	}

	client := arrclient.NewClient(
		inst.APIURL, inst.APIKey,
		time.Duration(genSettings.APITimeout)*time.Second,
		genSettings.SSLVerify,
	)

	queue, err := hunter.GetQueue(ctx, client)
	if err != nil {
		log.Error("failed to get queue", "error", err)
		return
	}

	for _, record := range queue.Records {
		if !isImportStuck(record) {
			continue
		}

		mediaID := record.MediaID()
		if mediaID == 0 {
			continue
		}

		// The item is stuck on import. Trigger a rescan/refresh which often
		// resolves "Unable to Import Automatically" issues.
		log.Info("found stuck import, triggering rescan",
			"title", record.Title,
			"media_id", mediaID,
			"status", record.TrackedDownloadStatus,
			"messages", formatStatusMessages(record.StatusMessages))

		if err := triggerRescan(ctx, client, appType, mediaID); err != nil {
			log.Error("failed to trigger rescan", "title", record.Title, "error", err)
			continue
		}

		_ = imp.db.LogAutoImport(ctx, appType, inst.ID, mediaID, record.Title, record.ID, "rescan_triggered", formatStatusMessages(record.StatusMessages))
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
