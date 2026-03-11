package queuecleaner

//go:generate mockgen -destination=mock_store_test.go -package=queuecleaner github.com/lusoris/lurkarr/internal/queuecleaner Store

	import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	downloadclient "github.com/lusoris/lurkarr/internal/downloadclients"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/deluge"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/qbittorrent"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/transmission"
	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/nzbget"
	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/sabnzbd"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/lurking"
	"github.com/lusoris/lurkarr/internal/metrics"
	"github.com/lusoris/lurkarr/internal/notifications"
)

// Store abstracts the database operations needed by the Cleaner.
type Store interface {
	GetQueueCleanerSettings(ctx context.Context, appType database.AppType) (*database.QueueCleanerSettings, error)
	ListEnabledInstances(ctx context.Context, appType database.AppType) ([]database.AppInstance, error)
	GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error)
	GetScoringProfile(ctx context.Context, appType database.AppType) (*database.ScoringProfile, error)
	LogBlocklist(ctx context.Context, appType database.AppType, instanceID uuid.UUID, downloadID, title, reason string) error
	ResetStrikes(ctx context.Context, appType database.AppType, instanceID uuid.UUID, downloadID string) error
	AddStrike(ctx context.Context, appType database.AppType, instanceID uuid.UUID, downloadID, title, reason string) error
	CountStrikes(ctx context.Context, appType database.AppType, instanceID uuid.UUID, downloadID string, windowHours int) (int, error)
	GetSABnzbdSettings(ctx context.Context) (*database.SABnzbdSettings, error)
	GetDownloadClientSettings(ctx context.Context, appType database.AppType) (*database.DownloadClientSettings, error)
}

// Cleaner monitors download queues and removes stalled/slow/duplicate items.
type Cleaner struct {
	db       Store
	logger   *logging.Logger
	notifier notifications.Notifier
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// New creates a new queue cleaner.
func New(db Store, logger *logging.Logger) *Cleaner {
	return &Cleaner{db: db, logger: logger}
}

// SetNotifier sets an optional notification manager.
func (c *Cleaner) SetNotifier(n notifications.Notifier) {
	c.notifier = n
}

// Start launches cleaner goroutines for each app type.
func (c *Cleaner) Start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)
	for _, appType := range database.AllAppTypes() {
		if lurking.LurkerFor(appType) == nil {
			continue
		}
		c.wg.Add(1)
		go c.cleanLoop(ctx, appType)
	}
	slog.Info("queue cleaner started")
}

// Stop cancels all cleaner goroutines.
func (c *Cleaner) Stop() {
	if c.cancel != nil {
		c.cancel()
	}
	c.wg.Wait()
}

func (c *Cleaner) cleanLoop(ctx context.Context, appType database.AppType) {
	defer c.wg.Done()
	log := c.logger.ForApp(string(appType))

	for {
		settings, err := c.db.GetQueueCleanerSettings(ctx, appType)
		if err != nil {
			log.Error("failed to load queue cleaner settings", "error", err)
			if !sleep(ctx, 60*time.Second) {
				return
			}
			continue
		}

		if !settings.Enabled {
			if !sleep(ctx, 60*time.Second) {
				return
			}
			continue
		}

		instances, err := c.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("failed to list instances", "error", err)
			if !sleep(ctx, time.Duration(settings.CheckIntervalSeconds)*time.Second) {
				return
			}
			continue
		}

		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			start := time.Now()
			c.cleanInstance(ctx, log, appType, settings, inst)
			metrics.QueueCleanerRunDuration.WithLabelValues(string(appType), inst.Name).Observe(time.Since(start).Seconds())
		}

		// 5. Orphan detection — runs once per app type across all instances.
		if settings.OrphanEnabled {
			c.cleanOrphans(ctx, log, appType, settings, instances)
		}

		if !sleep(ctx, time.Duration(settings.CheckIntervalSeconds)*time.Second) {
			return
		}
	}
}

func (c *Cleaner) cleanInstance(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance) {
	log = log.With("instance", inst.Name)

	lurker := lurking.LurkerFor(appType)
	if lurker == nil {
		return
	}

	genSettings, err := c.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("failed to load general settings", "error", err)
		return
	}

	client := arrclient.NewClient(
		inst.APIURL, inst.APIKey,
		time.Duration(genSettings.APITimeout)*time.Second,
		genSettings.SSLVerify,
	)

	queue, err := lurker.GetQueue(ctx, client)
	if err != nil {
		log.Error("failed to get queue", "error", err)
		return
	}

	if len(queue.Records) == 0 {
		return
	}

	// Get SABnzbd queue status for Usenet items
	sabStatuses := c.getSABnzbdStatuses(ctx)

	apiVersion := apiVersionFor(appType)

	// 1. Queue deduplication
	profile, err := c.db.GetScoringProfile(ctx, appType)
	if err != nil {
		log.Error("failed to load scoring profile", "error", err)
	} else {
		dupes := FindDuplicates(queue.Records, profile)
		for _, d := range dupes {
			log.Info("removing duplicate",
				"remove", d.RemoveTitle, "remove_score", d.RemoveScore,
				"keep", d.KeepTitle, "keep_score", d.KeepScore)
			if err := client.DeleteQueueItem(ctx, apiVersion, d.RemoveQueueID, settings.RemoveFromClient, settings.BlocklistOnRemove); err != nil {
				log.Error("failed to remove duplicate", "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, "", d.RemoveTitle, "duplicate_lower_score"); err != nil {
				log.Warn("failed to log blocklist", "title", d.RemoveTitle, "error", err)
			}
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, d.RemoveTitle, "duplicate_lower_score")
		}
	}

	// 2. Stalled/slow detection with strike system
	for _, record := range queue.Records {
		if record.TrackedDownloadState == "importPending" {
			continue // Don't strike items waiting for import
		}

		// Reset strikes for downloads that are making progress
		if record.Size > 0 && record.Sizeleft < record.Size {
			progress := float64(record.Size-record.Sizeleft) / float64(record.Size)
			if progress > 0.5 && record.TrackedDownloadStatus != "warning" {
				// Over 50% and healthy — clear any old strikes
				if err := c.db.ResetStrikes(ctx, appType, inst.ID, record.DownloadID); err != nil {
					log.Warn("failed to reset strikes", "download_id", record.DownloadID, "error", err)
				}
				continue
			}
		}

		reason := c.detectProblem(record, settings, sabStatuses)
		if reason == "" {
			continue
		}

		if err := c.db.AddStrike(ctx, appType, inst.ID, record.DownloadID, record.Title, reason); err != nil {
			log.Error("failed to add strike", "error", err)
			continue
		}
		metrics.QueueCleanerStrikes.WithLabelValues(string(appType), inst.Name).Inc()

		count, err := c.db.CountStrikes(ctx, appType, inst.ID, record.DownloadID, settings.StrikeWindowHours)
		if err != nil {
			log.Error("failed to count strikes", "error", err)
			continue
		}

		log.Info("strike added", "title", record.Title, "reason", reason, "strikes", count, "max", settings.MaxStrikes)

		if count >= settings.MaxStrikes {
			log.Warn("max strikes reached, removing", "title", record.Title, "download_id", record.DownloadID)
			if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, settings.RemoveFromClient, settings.BlocklistOnRemove); err != nil {
				log.Error("failed to remove struck item", "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, reason+"_max_strikes"); err != nil {
				log.Warn("failed to log blocklist", "title", record.Title, "error", err)
			}
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, record.Title, reason+"_max_strikes")
		}
	}

	// 3. Failed import cleanup
	if settings.FailedImportRemove {
		c.cleanFailedImports(ctx, log, appType, settings, inst, client, apiVersion, queue.Records)
	}

	// 4. Seeding enforcement — remove completed torrents exceeding ratio/time limits
	if settings.SeedingEnabled {
		c.cleanSeeding(ctx, log, appType, settings, inst, client, apiVersion, queue.Records)
	}
}

// detectProblem checks if a queue item is stalled, slow, or metadata-stuck.
// Returns the reason string, or "" if no problem.
func (c *Cleaner) detectProblem(record arrclient.QueueRecord, settings *database.QueueCleanerSettings, sabStatuses map[string]string) string {
	// For Usenet via SABnzbd: check actual SABnzbd status.
	// SABnzbd items show as "Queued" when they're just waiting for a slot,
	// NOT because they're stalled.
	if record.Protocol == "usenet" && record.DownloadID != "" {
		if sabStatus, ok := sabStatuses[record.DownloadID]; ok {
			switch sabStatus {
			case "Queued", "Grabbing":
				return "" // Actually just waiting in SABnzbd queue, not stalled
			case "Paused":
				return "paused_in_sabnzbd"
			}
		}
	}

	// Check for metadata stuck (no size yet, been in queue too long)
	if settings.MetadataStuckMinutes > 0 && record.Size == 0 && record.Sizeleft == 0 {
		if record.Status == "downloading" || record.Status == "delay" {
			return "metadata_stuck"
		}
	}

	// Check for stalled torrents — with per-privacy type rules
	if record.Status == "warning" && record.TrackedDownloadStatus == "warning" {
		if record.Protocol == "torrent" {
			isPrivate := isPrivateTracker(record)
			if isPrivate && !settings.StrikePrivate {
				return "" // Skip private trackers per config
			}
			if !isPrivate && !settings.StrikePublic {
				return "" // Skip public trackers per config
			}
		}
		return "stalled"
	}

	// Check download speed for active downloads
	if record.Size > 0 && record.Sizeleft > 0 && record.Sizeleft < record.Size && settings.SlowThresholdBytesPerSec > 0 {
		// Skip slow detection for large downloads if configured
		if settings.SlowIgnoreAboveBytes > 0 && record.Sizeleft > settings.SlowIgnoreAboveBytes {
			return ""
		}

		// Parse timeleft to estimate speed
		if tl := parseTimeleft(record.TimeleftStr); tl > 0 {
			estimatedSpeed := record.Sizeleft / int64(tl.Seconds())
			if estimatedSpeed > 0 && estimatedSpeed < settings.SlowThresholdBytesPerSec {
				return "slow"
			}
		}
	}

	return ""
}

// isPrivateTracker determines if a torrent is from a private tracker.
// It uses indexer flags from the arr API (set only by private indexers) as the
// primary signal, with a known-public-tracker fallback for older arr versions.
func isPrivateTracker(record arrclient.QueueRecord) bool {
	// Indexer flags are only populated by private trackers (freeleech, internal, etc.).
	if record.IndexerFlags != 0 {
		return true
	}

	indexer := strings.ToLower(record.Indexer)
	if indexer == "" {
		return false
	}

	// Well-known public indexers — if matched, definitely not private.
	publicIndexers := []string{
		"1337x", "rarbg", "yts", "nyaa", "eztv", "limetorrents",
		"thepiratebay", "kickasstorrents", "torrentz2", "glodls",
		"magnetdl", "ettv", "isohunt", "bt4g", "solidtorrents",
		"bitsearch", "torrentgalaxy", "fitgirl",
	}
	for _, pub := range publicIndexers {
		if indexer == pub {
			return false
		}
	}

	// Has an indexer name but not recognized as public → treat as private.
	return true
}

// cleanFailedImports removes queue items with import errors (statusMessages containing failure reasons).
func (c *Cleaner) cleanFailedImports(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance, client *arrclient.Client, apiVersion string, records []arrclient.QueueRecord) {
	for _, record := range records {
		if !hasImportFailure(record) {
			continue
		}

		reason := importFailureReason(record)
		log.Warn("removing failed import", "title", record.Title, "reason", reason, "download_id", record.DownloadID)

		if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, settings.RemoveFromClient, settings.FailedImportBlocklist); err != nil {
			log.Error("failed to remove failed import", "error", err)
			continue
		}
		if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, "failed_import: "+reason); err != nil {
			log.Warn("failed to log blocklist", "title", record.Title, "error", err)
		}
		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
		metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
		c.notifyRemoval(ctx, appType, inst.Name, record.Title, "failed_import: "+reason)
	}
}

// hasImportFailure checks if a queue record has import failure messages.
func hasImportFailure(record arrclient.QueueRecord) bool {
	if record.TrackedDownloadStatus != "warning" {
		return false
	}
	if record.TrackedDownloadState != "importPending" && record.TrackedDownloadState != "importFailed" {
		return false
	}
	for _, sm := range record.StatusMessages {
		for _, msg := range sm.Messages {
			lower := strings.ToLower(msg)
			if strings.Contains(lower, "import failed") ||
				strings.Contains(lower, "unable to import") ||
				strings.Contains(lower, "no files found") ||
				strings.Contains(lower, "sample") ||
				strings.Contains(lower, "not a valid") {
				return true
			}
		}
	}
	return false
}

// importFailureReason extracts the first relevant failure message from a queue record.
func importFailureReason(record arrclient.QueueRecord) string {
	for _, sm := range record.StatusMessages {
		for _, msg := range sm.Messages {
			if msg != "" {
				return msg
			}
		}
	}
	return "unknown_import_failure"
}

// parseTimeleft parses arr's timeleft format "HH:MM:SS" or "D.HH:MM:SS" into a Duration.
func parseTimeleft(s string) time.Duration {
	if s == "" || s == "00:00:00" {
		return 0
	}
	parts := strings.Split(s, ":")
	if len(parts) != 3 {
		return 0
	}

	var hours int
	// Handle "D.HH" format for days
	dayParts := strings.SplitN(parts[0], ".", 2)
	if len(dayParts) == 2 {
		days, err := strconv.Atoi(dayParts[0])
		if err != nil {
			return 0
		}
		h, err := strconv.Atoi(dayParts[1])
		if err != nil {
			return 0
		}
		hours = days*24 + h
	} else {
		h, err := strconv.Atoi(parts[0])
		if err != nil {
			return 0
		}
		hours = h
	}

	mins, err := strconv.Atoi(parts[1])
	if err != nil {
		return 0
	}
	secs, err := strconv.Atoi(parts[2])
	if err != nil {
		return 0
	}

	return time.Duration(hours)*time.Hour + time.Duration(mins)*time.Minute + time.Duration(secs)*time.Second
}

func (c *Cleaner) notifyRemoval(ctx context.Context, appType database.AppType, instName, title, reason string) {
	if c.notifier == nil {
		return
	}
	c.notifier.Notify(ctx, notifications.Event{
		Type:     notifications.EventQueueItemRemoved,
		Title:    "Queue Item Removed",
		Message:  title,
		AppType:  string(appType),
		Instance: instName,
		Fields:   map[string]string{"Reason": reason},
	})
}

// getDownloadClient builds a unified download client for the given app type's
// configured download client, or returns nil if not configured/enabled.
func (c *Cleaner) getDownloadClient(ctx context.Context, appType database.AppType) downloadclient.Client {
	dcs, err := c.db.GetDownloadClientSettings(ctx, appType)
	if err != nil || !dcs.Enabled || dcs.URL == "" {
		return nil
	}

	timeout := time.Duration(dcs.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	switch downloadclient.ClientType(dcs.ClientType) {
	case downloadclient.TypeQBittorrent:
		native := qbittorrent.NewClient(dcs.URL, dcs.Username, dcs.Password, timeout)
		return downloadclient.NewQBittorrentAdapter(native)
	case downloadclient.TypeTransmission:
		native := transmission.NewClient(dcs.URL, dcs.Username, dcs.Password, timeout)
		return downloadclient.NewTransmissionAdapter(native)
	case downloadclient.TypeDeluge:
		native := deluge.NewClient(dcs.URL, dcs.Password, timeout)
		return downloadclient.NewDelugeAdapter(native)
	case downloadclient.TypeSABnzbd:
		// For SABnzbd, the "password" field stores the API key.
		native := sabnzbd.NewClient(dcs.URL, dcs.Password, timeout)
		return downloadclient.NewSABnzbdAdapter(native)
	case downloadclient.TypeNZBGet:
		native := nzbget.NewClient(dcs.URL, dcs.Username, dcs.Password, timeout)
		return downloadclient.NewNZBGetAdapter(native)
	default:
		return nil
	}
}

// cleanSeeding checks completed torrent downloads against seeding rules and
// removes items that have exceeded the configured ratio or seeding time.
func (c *Cleaner) cleanSeeding(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance, client *arrclient.Client, apiVersion string, records []arrclient.QueueRecord) {
	torrentClient := c.getDownloadClient(ctx, appType)
	if torrentClient == nil {
		return
	}

	items, err := torrentClient.GetItems(ctx)
	if err != nil {
		log.Error("failed to get torrent items for seeding check", "error", err)
		return
	}

	// Build a map of download hash → torrent item for fast lookup.
	itemsByID := make(map[string]downloadclient.DownloadItem, len(items))
	for _, item := range items {
		itemsByID[strings.ToLower(item.ID)] = item
	}

	for _, record := range records {
		if record.Protocol != "torrent" {
			continue
		}
		if record.DownloadID == "" {
			continue
		}
		// Only look at completed downloads (imported or waiting for import).
		if record.TrackedDownloadState != "imported" && record.TrackedDownloadState != "importPending" {
			continue
		}

		item, ok := itemsByID[strings.ToLower(record.DownloadID)]
		if !ok {
			continue
		}

		// Skip private trackers if configured.
		if settings.SeedingSkipPrivate && isPrivateTracker(record) {
			continue
		}

		if !c.seedingLimitReached(settings, item) {
			continue
		}

		log.Info("seeding limit reached, removing",
			"title", record.Title,
			"ratio", item.Ratio,
			"seeding_hours", float64(item.SeedingTime)/3600,
			"max_ratio", settings.SeedingMaxRatio,
			"max_hours", settings.SeedingMaxHours,
			"mode", settings.SeedingMode,
		)

		// Remove from the torrent client directly.
		if err := torrentClient.RemoveItem(ctx, item.ID, settings.SeedingDeleteFiles); err != nil {
			log.Error("failed to remove seeded torrent", "title", record.Title, "error", err)
			continue
		}

		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
		c.notifyRemoval(ctx, appType, inst.Name, record.Title, "seeding_limit_reached")
	}
}

// seedingLimitReached evaluates whether a torrent has exceeded the configured
// seeding limits. In "or" mode either condition triggers; in "and" mode both must be met.
func (c *Cleaner) seedingLimitReached(settings *database.QueueCleanerSettings, item downloadclient.DownloadItem) bool {
	ratioMet := settings.SeedingMaxRatio > 0 && item.Ratio >= settings.SeedingMaxRatio
	timeMet := settings.SeedingMaxHours > 0 && item.SeedingTime >= int64(settings.SeedingMaxHours)*3600

	// If neither limit is configured, nothing to enforce.
	if settings.SeedingMaxRatio <= 0 && settings.SeedingMaxHours <= 0 {
		return false
	}

	// If only one limit is configured, use that one.
	if settings.SeedingMaxRatio <= 0 {
		return timeMet
	}
	if settings.SeedingMaxHours <= 0 {
		return ratioMet
	}

	if settings.SeedingMode == "and" {
		return ratioMet && timeMet
	}
	return ratioMet || timeMet // "or" mode (default)
}

// cleanOrphans detects downloads in the configured download client that are not
// tracked by any *arr instance of this app type. Works for all client types
// (torrent and usenet). Items must exceed the grace period and not match any
// excluded category to be removed.
func (c *Cleaner) cleanOrphans(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, instances []database.AppInstance) {
	dlClient := c.getDownloadClient(ctx, appType)
	if dlClient == nil {
		return
	}

	// Collect all known download IDs from all *arr instances of this app type.
	knownIDs := make(map[string]bool)

	lurker := lurking.LurkerFor(appType)
	if lurker == nil {
		return
	}

	genSettings, err := c.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("orphan: failed to load general settings", "error", err)
		return
	}

	for _, inst := range instances {
		client := arrclient.NewClient(
			inst.APIURL, inst.APIKey,
			time.Duration(genSettings.APITimeout)*time.Second,
			genSettings.SSLVerify,
		)
		queue, err := lurker.GetQueue(ctx, client)
		if err != nil {
			log.Warn("orphan: failed to get queue", "instance", inst.Name, "error", err)
			continue
		}
		for _, r := range queue.Records {
			if r.DownloadID != "" {
				knownIDs[strings.ToLower(r.DownloadID)] = true
			}
		}
	}

	// Get all items from the download client (active + completed).
	items, err := dlClient.GetItems(ctx)
	if err != nil {
		log.Error("orphan: failed to get download client items", "error", err)
		return
	}

	// Also include history/completed items (important for usenet clients).
	history, err := dlClient.GetHistory(ctx)
	if err != nil {
		log.Warn("orphan: failed to get download client history", "error", err)
		// Non-fatal — continue with active items only.
	} else {
		// Merge, deduplicating by ID.
		seen := make(map[string]bool, len(items))
		for _, item := range items {
			seen[strings.ToLower(item.ID)] = true
		}
		for _, item := range history {
			if !seen[strings.ToLower(item.ID)] {
				items = append(items, item)
			}
		}
	}

	// Parse excluded categories.
	excludedCats := parseExcludedCategories(settings.OrphanExcludedCategories)

	now := time.Now().Unix()
	graceSeconds := int64(settings.OrphanGraceMinutes) * 60

	for _, item := range items {
		id := strings.ToLower(item.ID)

		// Already tracked by an *arr instance — not an orphan.
		if knownIDs[id] {
			continue
		}

		// Check excluded categories.
		if excludedCats[strings.ToLower(item.Category)] {
			continue
		}

		// Grace period: skip if the item was added recently.
		if item.AddedAt > 0 && (now-item.AddedAt) < graceSeconds {
			continue
		}
		// For items without AddedAt, use CompletedAt as fallback.
		if item.AddedAt == 0 && item.CompletedAt > 0 && (now-item.CompletedAt) < graceSeconds {
			continue
		}

		log.Info("removing orphan download",
			"name", item.Name,
			"id", item.ID,
			"category", item.Category,
			"added_at", item.AddedAt,
		)

		if err := dlClient.RemoveItem(ctx, item.ID, settings.OrphanDeleteFiles); err != nil {
			log.Error("orphan: failed to remove item", "name", item.Name, "error", err)
			continue
		}

		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), "orphan").Inc()
		c.notifyRemoval(ctx, appType, "orphan", item.Name, "orphan_not_tracked")
	}
}

// parseExcludedCategories splits a comma-separated category string into a lookup set.
func parseExcludedCategories(s string) map[string]bool {
	cats := make(map[string]bool)
	for _, c := range strings.Split(s, ",") {
		c = strings.TrimSpace(strings.ToLower(c))
		if c != "" {
			cats[c] = true
		}
	}
	return cats
}

// getSABnzbdStatuses fetches the SABnzbd queue and returns a map of downloadID -> status.
func (c *Cleaner) getSABnzbdStatuses(ctx context.Context) map[string]string {
	statuses := make(map[string]string)

	sabSettings, err := c.db.GetSABnzbdSettings(ctx)
	if err != nil || !sabSettings.Enabled {
		return statuses
	}

	sabClient := sabnzbd.NewClient(sabSettings.URL, sabSettings.APIKey, time.Duration(sabSettings.Timeout)*time.Second)
	queue, err := sabClient.GetQueue(ctx)
	if err != nil {
		return statuses
	}

	for _, slot := range queue.Slots {
		statuses[slot.NzoID] = slot.Status
	}
	return statuses
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
