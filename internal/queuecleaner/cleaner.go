package queuecleaner

//go:generate mockgen -destination=mock_store_test.go -package=queuecleaner github.com/lusoris/lurkarr/internal/queuecleaner Store

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/blocklist"
	"github.com/lusoris/lurkarr/internal/database"
	downloadclient "github.com/lusoris/lurkarr/internal/downloadclients"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/deluge"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/qbittorrent"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/rtorrent"
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
	AddStrikeAndCount(ctx context.Context, appType database.AppType, instanceID uuid.UUID, downloadID, title, reason string, windowHours int) (int, error)
	GetSABnzbdSettings(ctx context.Context) (*database.SABnzbdSettings, error)
	GetDownloadClientSettings(ctx context.Context, appType database.AppType) (*database.DownloadClientSettings, error)
	ListEnabledDownloadClientInstances(ctx context.Context) ([]database.DownloadClientInstance, error)
	ListEnabledBlocklistRules(ctx context.Context) ([]database.BlocklistRule, error)
	IsSearchOnCooldown(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID, cooldownHours int) (bool, error)
	RecordSearch(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int) error
	MarkProcessed(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int, operation string) error
	RecordSearchFailure(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int) error
	ClearSearchFailure(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID int) error
	IsSearchFailureLimitReached(ctx context.Context, appType database.AppType, instanceID uuid.UUID, mediaID, maxFailures int) (bool, error)
	ListSeedingRuleGroups(ctx context.Context) ([]database.SeedingRuleGroup, error)
}

// removal tracks a removed queue item with both its title and media key for
// cross-instance matching. MediaKey uses external IDs (TMDB, TVDB+season) so
// different releases of the same media can be matched across instances that use
// different quality profiles or custom format scores.
type removal struct {
	Title    string // release title (e.g., "Movie.2024.2160p.x264-GROUP")
	MediaKey string // external media key (e.g., "tmdb:12345" or "tvdb:67890:s02")
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

// RunOnce performs a single queue clean pass for all app types and returns.
func (c *Cleaner) RunOnce(ctx context.Context) {
	for _, appType := range database.AllAppTypes() {
		if lurking.LurkerFor(appType) == nil {
			continue
		}
		log := c.logger.ForApp(string(appType))
		settings, err := c.db.GetQueueCleanerSettings(ctx, appType)
		if err != nil {
			log.Error("run-once: failed to load queue cleaner settings", "error", err)
			continue
		}
		if !settings.Enabled {
			continue
		}
		instances, err := c.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("run-once: failed to list instances", "error", err)
			continue
		}
		removals := make(map[uuid.UUID][]removal)
		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			removed := c.cleanInstance(ctx, log, appType, settings, inst)
			if len(removed) > 0 {
				removals[inst.ID] = removed
			}
		}
		if settings.OrphanEnabled {
			c.cleanOrphans(ctx, log, appType, settings, instances)
		}
		if settings.CrossArrSync && len(instances) > 1 && len(removals) > 0 {
			c.syncBlocklistAcross(ctx, log, appType, settings, instances, removals)
		}
	}
	slog.Info("queue cleaner run-once complete")
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

		if settings.DryRun {
			log.Info("[DRY-RUN] queue cleaner running in preview mode — no actions will be executed")
		}

		instances, err := c.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			log.Error("failed to list instances", "error", err)
			if !sleep(ctx, time.Duration(settings.CheckIntervalSeconds)*time.Second) {
				return
			}
			continue
		}

		removals := make(map[uuid.UUID][]removal) // instanceID -> removed items
		for _, inst := range instances {
			if ctx.Err() != nil {
				return
			}
			start := time.Now()
			removed := c.cleanInstance(ctx, log, appType, settings, inst)
			if len(removed) > 0 {
				removals[inst.ID] = removed
			}
			metrics.QueueCleanerRunDuration.WithLabelValues(string(appType), inst.Name).Observe(time.Since(start).Seconds())
		}

		// 5. Orphan detection — runs once per app type across all instances.
		if settings.OrphanEnabled {
			c.cleanOrphans(ctx, log, appType, settings, instances)
		}

		// 6. Cross-Arr blocklist sync — propagate removals to sibling instances.
		if settings.CrossArrSync && len(instances) > 1 && len(removals) > 0 {
			c.syncBlocklistAcross(ctx, log, appType, settings, instances, removals)
		}

		if !sleep(ctx, time.Duration(settings.CheckIntervalSeconds)*time.Second) {
			return
		}
	}
}

func (c *Cleaner) cleanInstance(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance) []removal {
	log = log.With("instance", inst.Name)

	lurker := lurking.LurkerFor(appType)
	if lurker == nil {
		return nil
	}

	genSettings, err := c.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("failed to load general settings", "error", err)
		return nil
	}

	client := arrclient.NewClient(
		inst.APIURL, inst.APIKey,
		time.Duration(genSettings.APITimeout)*time.Second,
		genSettings.SSLVerify,
	)

	// Use enriched queue when cross-arr sync, protected tags, deletion detection,
	// or unmonitored cleanup is enabled so we get external IDs and media status
	// from the enriched response.
	needEnriched := settings.CrossArrSync || settings.ProtectedTags != "" ||
		settings.DeletionDetectionEnabled || settings.UnmonitoredCleanupEnabled
	var queue *arrclient.QueueResponse
	if needEnriched {
		queue, err = getEnrichedQueue(ctx, client, appType)
	} else {
		queue, err = lurker.GetQueue(ctx, client)
	}
	if err != nil {
		log.Error("failed to get queue", "error", err)
		return nil
	}

	if len(queue.Records) == 0 {
		return nil
	}

	// Filter out records whose media has a protected tag.
	if settings.ProtectedTags != "" {
		queue.Records = filterProtectedTags(ctx, log, client, appType, settings.ProtectedTags, queue.Records)
	}

	// Filter out records from ignored indexers.
	if settings.IgnoredIndexers != "" {
		queue.Records = filterIgnoredIndexers(log, settings.IgnoredIndexers, queue.Records)
	}

	// Filter out records from ignored download clients.
	if settings.IgnoredDownloadClients != "" {
		queue.Records = filterIgnoredDownloadClients(log, settings.IgnoredDownloadClients, queue.Records)
	}

	// Get SABnzbd queue status for Usenet items
	sabStatuses := c.getSABnzbdStatuses(ctx)

	apiVersion := apiVersionFor(appType)

	// Resolve obsolete tag ID once per instance when tag-instead-of-delete is enabled.
	var obsoleteTagID int
	if settings.TagInsteadOfDelete && settings.ObsoleteTagLabel != "" && !settings.DryRun {
		obsoleteTagID = resolveTagID(ctx, log, client, apiVersion, settings.ObsoleteTagLabel)
	}

	// Per-run search budget: when MaxSearchesPerRun > 0, limits how many
	// re-searches fire in a single cleanup cycle for this instance.
	var searchBudget *int
	if settings.MaxSearchesPerRun > 0 {
		b := settings.MaxSearchesPerRun
		searchBudget = &b
	}

	var removed []removal

	// 0. Blocklist rule matching — remove items matching user/community blocklist
	blRules, err := c.db.ListEnabledBlocklistRules(ctx)
	if err != nil {
		log.Warn("failed to load blocklist rules", "error", err)
	} else if len(blRules) > 0 {
		matcher := blocklist.NewMatcher(blRules, func(title string) blocklist.ReleaseInfo {
			parsed := ParseRelease(title)
			return blocklist.ReleaseInfo{ReleaseGroup: parsed.ReleaseGroup}
		})
		for _, record := range queue.Records {
			result := matcher.Check(record)
			if !result.Matched {
				continue
			}
			reason := "blocklist_" + result.Rule.PatternType + ":" + result.Rule.Pattern
			if settings.DryRun {
				log.Warn("[DRY-RUN] would remove blocklist match",
					"title", record.Title,
					"rule_type", result.Rule.PatternType,
					"pattern", result.Rule.Pattern)
				removed = append(removed, removal{Title: record.Title, MediaKey: record.MediaKey()})
				continue
			}
			log.Warn("blocklist match, removing",
				"title", record.Title,
				"rule_type", result.Rule.PatternType,
				"pattern", result.Rule.Pattern)
			if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), true); err != nil {
				log.Error("failed to remove blocklisted item", "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, reason); err != nil {
				log.Warn("failed to log blocklist", "title", record.Title, "error", err)
			}
			removed = append(removed, removal{Title: record.Title, MediaKey: record.MediaKey()})
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, record.Title, reason)
			if settings.SearchOnRemove {
				c.triggerReSearch(ctx, log, lurker, client, record, appType, inst.ID, settings.SearchCooldownHours, searchBudget, settings.MaxSearchFailures)
			}
		}
	}

	// Build queue ID → record lookup for dedup media key resolution.
	recordByQueueID := make(map[int]*arrclient.QueueRecord, len(queue.Records))
	for i := range queue.Records {
		recordByQueueID[queue.Records[i].ID] = &queue.Records[i]
	}

	// 1. Queue deduplication
	profile, err := c.db.GetScoringProfile(ctx, appType)
	if err != nil {
		log.Error("failed to load scoring profile", "error", err)
	} else {
		dupes := FindDuplicates(queue.Records, profile)
		for _, d := range dupes {
			var mediaKey string
			if rec, ok := recordByQueueID[d.RemoveQueueID]; ok {
				mediaKey = rec.MediaKey()
			}
			if settings.DryRun {
				log.Info("[DRY-RUN] would remove duplicate",
					"remove", d.RemoveTitle, "remove_score", d.RemoveScore,
					"keep", d.KeepTitle, "keep_score", d.KeepScore)
				removed = append(removed, removal{Title: d.RemoveTitle, MediaKey: mediaKey})
				continue
			}
			log.Info("removing duplicate",
				"remove", d.RemoveTitle, "remove_score", d.RemoveScore,
				"keep", d.KeepTitle, "keep_score", d.KeepScore)
			if err := client.DeleteQueueItem(ctx, apiVersion, d.RemoveQueueID, effectiveRemoveFromClient(settings), shouldBlocklist("duplicate", settings)); err != nil {
				log.Error("failed to remove duplicate", "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, "", d.RemoveTitle, "duplicate_lower_score"); err != nil {
				log.Warn("failed to log blocklist", "title", d.RemoveTitle, "error", err)
			}
			removed = append(removed, removal{Title: d.RemoveTitle, MediaKey: mediaKey})
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, d.RemoveTitle, "duplicate_lower_score")
		}
	}

	// 2. Stalled/slow detection with strike system
	pipeSaturated := isBandwidthSaturated(log, settings, queue.Records)
	for _, record := range queue.Records {
		if record.TrackedDownloadState == "importPending" {
			continue // Don't strike items waiting for import
		}

		// Reset strikes for downloads that are making progress
		if record.Size > 0 && record.Sizeleft < record.Size {
			progress := float64(record.Size-record.Sizeleft) / float64(record.Size)
			if progress > 0.5 && record.TrackedDownloadStatus != "warning" {
				// Over 50% and healthy — clear any old strikes
				if !settings.DryRun {
					if err := c.db.ResetStrikes(ctx, appType, inst.ID, record.DownloadID); err != nil {
						log.Warn("failed to reset strikes", "download_id", record.DownloadID, "error", err)
					}
				}
				continue
			}
		}

		reason := c.detectProblem(record, settings, sabStatuses, pipeSaturated)
		if reason == "" {
			continue
		}

		if settings.DryRun {
			log.Info("[DRY-RUN] would strike", "title", record.Title, "reason", reason, "max", effectiveMaxStrikes(reason, settings))
			continue
		}

		count, err := c.db.AddStrikeAndCount(ctx, appType, inst.ID, record.DownloadID, record.Title, reason, settings.StrikeWindowHours)
		if err != nil {
			log.Error("failed to add strike", "error", err)
			continue
		}
		metrics.QueueCleanerStrikes.WithLabelValues(string(appType), inst.Name).Inc()

		log.Info("strike added", "title", record.Title, "reason", reason, "strikes", count, "max", effectiveMaxStrikes(reason, settings))

		if count >= effectiveMaxStrikes(reason, settings) {
			log.Warn("max strikes reached, removing", "title", record.Title, "download_id", record.DownloadID)
			if obsoleteTagID > 0 {
				if mid := record.TaggableMediaID(); mid > 0 {
					if err := client.TagMedia(ctx, apiVersion, string(appType), mid, obsoleteTagID); err != nil {
						log.Warn("failed to tag media as obsolete", "title", record.Title, "error", err)
					}
				}
			}
			if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), shouldBlocklist(reason, settings)); err != nil {
				log.Error("failed to remove struck item", "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, reason+"_max_strikes"); err != nil {
				log.Warn("failed to log blocklist", "title", record.Title, "error", err)
			}
			removed = append(removed, removal{Title: record.Title, MediaKey: record.MediaKey()})
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, record.Title, reason+"_max_strikes")
			if settings.SearchOnRemove {
				c.triggerReSearch(ctx, log, lurker, client, record, appType, inst.ID, settings.SearchCooldownHours, searchBudget, settings.MaxSearchFailures)
			}
		}
	}

	// 3. Failed import cleanup
	if settings.FailedImportRemove {
		c.cleanFailedImports(ctx, log, appType, settings, inst, client, apiVersion, queue.Records, searchBudget)
	}

	// 4. Seeding enforcement — remove completed torrents exceeding ratio/time limits
	if settings.SeedingEnabled {
		c.cleanSeeding(ctx, log, appType, settings, inst, client, apiVersion, queue.Records)
	}

	// 4.5. Deletion detection — remove queue items whose media file was deleted externally
	if settings.DeletionDetectionEnabled {
		c.cleanDeletedMedia(ctx, log, appType, settings, inst, client, apiVersion, queue.Records, lurker, searchBudget)
	}

	// 4.6. Unmonitored cleanup — remove queue items for unmonitored media
	if settings.UnmonitoredCleanupEnabled {
		c.cleanUnmonitored(ctx, log, appType, settings, inst, client, apiVersion, queue.Records, lurker, searchBudget)
	}

	// 4.8. Metadata mismatch detection — strike/remove items where download
	// doesn't match the expected media (wrong series/movie/episode).
	if settings.MismatchEnabled {
		c.cleanMismatches(ctx, log, appType, settings, inst, client, apiVersion, queue.Records, lurker, searchBudget)
	}

	// 4.7. Recheck paused torrents — verify integrity and auto-resume if complete
	if settings.RecheckPausedEnabled {
		c.recheckPaused(ctx, log, appType, queue.Records)
	}

	return removed
}

// detectProblem checks if a queue item is stalled, slow, or metadata-stuck.
// Returns the reason string, or "" if no problem.
func (c *Cleaner) detectProblem(record arrclient.QueueRecord, settings *database.QueueCleanerSettings, sabStatuses map[string]string, pipeSaturated bool) string {
	// Skip all strike-based detection for items above the size threshold
	if settings.IgnoreAboveBytes > 0 && record.Size > settings.IgnoreAboveBytes {
		return ""
	}

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
		// Check for unregistered torrents first (subset of stalled/warning)
		if record.Protocol == "torrent" && settings.UnregisteredEnabled && isUnregisteredTorrent(record) {
			return "unregistered"
		}

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

	// Check for items stuck in queued state (not actively downloading)
	if settings.StrikeQueued && record.Status == "queued" {
		return "queued"
	}

	// Check download speed for active downloads
	if !pipeSaturated && record.Size > 0 && record.Sizeleft > 0 && record.Sizeleft < record.Size && settings.SlowThresholdBytesPerSec > 0 {
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

// unregisteredKeywords are substrings in arr status messages that indicate a
// torrent has been removed from its tracker. Matched case-insensitively.
var unregisteredKeywords = []string{
	"unregistered",
	"not registered",
	"torrent not found",
	"torrent is not found",
	"info hash",
	"infohash",
	"not found on tracker",
	"removed from tracker",
	"tracker returned: not found",
	"pack has been removed",
	"pack has been nuked",
	"trump",
	"trumped",
}

// isUnregisteredTorrent checks whether a queue record's status messages indicate
// that the torrent has been removed or unregistered from its tracker.
func isUnregisteredTorrent(record arrclient.QueueRecord) bool {
	for _, sm := range record.StatusMessages {
		for _, msg := range sm.Messages {
			lower := strings.ToLower(msg)
			for _, kw := range unregisteredKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
		}
	}
	return false
}

// mismatchKeywords are substrings in arr status messages that indicate a
// download's metadata doesn't match the expected media. Matched case-insensitively.
var mismatchKeywords = []string{
	"no matching series",
	"no matching movie",
	"no matching artist",
	"no matching album",
	"no matching author",
	"no matching book",
	"series was matched by series id",
	"unable to identify correct episode",
	"unable to identify correct movie",
}

// isMismatchedRelease checks whether a queue record's status messages indicate
// the download doesn't match the expected media (wrong series/movie/episode).
func isMismatchedRelease(record arrclient.QueueRecord) bool {
	for _, sm := range record.StatusMessages {
		for _, msg := range sm.Messages {
			lower := strings.ToLower(msg)
			for _, kw := range mismatchKeywords {
				if strings.Contains(lower, kw) {
					return true
				}
			}
		}
	}
	return false
}

// cleanFailedImports removes queue items with import errors (statusMessages containing failure reasons).
func (c *Cleaner) cleanFailedImports(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance, client *arrclient.Client, apiVersion string, records []arrclient.QueueRecord, searchBudget *int) {
	// Parse user-configured patterns (empty = use built-in defaults).
	var patterns []string
	for _, p := range strings.Split(settings.FailedImportPatterns, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			patterns = append(patterns, strings.ToLower(p))
		}
	}

	for _, record := range records {
		if !hasImportFailure(record, patterns) {
			continue
		}

		reason := importFailureReason(record)

		if settings.DryRun {
			log.Warn("[DRY-RUN] would remove failed import", "title", record.Title, "reason", reason, "download_id", record.DownloadID)
			continue
		}

		log.Warn("removing failed import", "title", record.Title, "reason", reason, "download_id", record.DownloadID)

		if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), settings.FailedImportBlocklist); err != nil {
			log.Error("failed to remove failed import", "error", err)
			continue
		}
		if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, "failed_import: "+reason); err != nil {
			log.Warn("failed to log blocklist", "title", record.Title, "error", err)
		}
		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
		metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
		c.notifyRemoval(ctx, appType, inst.Name, record.Title, "failed_import: "+reason)
		if settings.SearchOnRemove {
			if l := lurking.LurkerFor(appType); l != nil {
				c.triggerReSearch(ctx, log, l, client, record, appType, inst.ID, settings.SearchCooldownHours, searchBudget, settings.MaxSearchFailures)
			}
		}
	}
}

// hasImportFailure checks if a queue record has import failure messages.
// When patterns is non-empty, only messages containing one of the patterns match.
// When patterns is empty, built-in defaults are used.
func hasImportFailure(record arrclient.QueueRecord, patterns []string) bool {
	if record.TrackedDownloadStatus != "warning" {
		return false
	}
	if record.TrackedDownloadState != "importPending" && record.TrackedDownloadState != "importFailed" {
		return false
	}
	for _, sm := range record.StatusMessages {
		for _, msg := range sm.Messages {
			lower := strings.ToLower(msg)
			if len(patterns) > 0 {
				for _, p := range patterns {
					if strings.Contains(lower, p) {
						return true
					}
				}
			} else if strings.Contains(lower, "import failed") ||
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
	case downloadclient.TypeRTorrent:
		native := rtorrent.NewClient(dcs.URL, dcs.Username, dcs.Password, timeout)
		return downloadclient.NewRTorrentAdapter(native)
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

	// Build cross-seed map if enabled: count items sharing the same save path & size.
	crossSeedCounts := countCrossSeeds(items)

	// Load seeding rule groups (priority-sorted, first match wins).
	groups, err := c.db.ListSeedingRuleGroups(ctx)
	if err != nil {
		log.Error("failed to load seeding rule groups", "error", err)
		groups = nil // fall back to global settings only
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

		// Find matching seeding rule group for this item.
		group := matchSeedingGroup(item, groups)

		// If a group matches and has skip_removal, skip this item entirely.
		if group != nil && group.SkipRemoval {
			log.Debug("seeding group skip_removal", "title", record.Title, "group", group.Name)
			continue
		}

		// Determine limits: use group overrides or global settings.
		maxRatio := settings.SeedingMaxRatio
		maxHours := settings.SeedingMaxHours
		seedingMode := settings.SeedingMode
		deleteFiles := settings.SeedingDeleteFiles
		if group != nil {
			maxRatio = group.MaxRatio
			maxHours = group.MaxHours
			seedingMode = group.SeedingMode
			deleteFiles = group.DeleteFiles
		}

		if !seedingLimitReachedEx(maxRatio, maxHours, seedingMode, item) {
			continue
		}

		// Skip cross-seeded items: multiple torrents sharing the same content.
		if settings.SkipCrossSeeds && isCrossSeeded(item, crossSeedCounts) {
			log.Info("cross-seed detected, skipping removal", "title", record.Title, "path", item.SavePath)
			continue
		}

		if deleteFiles && settings.KeepArchives {
			deleteFiles = false
		}
		if deleteFiles && settings.HardlinkProtection && item.SavePath != "" && hasHardlinks(item.SavePath) {
			log.Info("hardlinks detected, skipping file deletion", "title", record.Title, "path", item.SavePath)
			deleteFiles = false
		}

		// RecycleBin: move files to recycle folder instead of permanent deletion
		if deleteFiles && settings.RecycleBinEnabled && settings.RecycleBinPath != "" && item.SavePath != "" {
			if err := moveToRecycleBin(item.SavePath, settings.RecycleBinPath); err != nil {
				log.Error("recycle bin move failed, falling back to deletion", "title", record.Title, "error", err)
			} else {
				log.Info("moved to recycle bin", "title", record.Title, "source", item.SavePath, "dest", settings.RecycleBinPath)
				deleteFiles = false
			}
		}

		groupName := "global"
		if group != nil {
			groupName = group.Name
		}
		log.Info("seeding limit reached, removing",
			"title", record.Title,
			"ratio", item.Ratio,
			"seeding_hours", float64(item.SeedingTime)/3600,
			"max_ratio", maxRatio,
			"max_hours", maxHours,
			"mode", seedingMode,
			"group", groupName,
			"delete_files", deleteFiles,
		)

		if settings.DryRun {
			log.Info("[DRY-RUN] would remove seeded torrent", "title", record.Title)
			continue
		}

		// Remove from the torrent client directly.
		if err := torrentClient.RemoveItem(ctx, item.ID, deleteFiles); err != nil {
			log.Error("failed to remove seeded torrent", "title", record.Title, "error", err)
			continue
		}

		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
		c.notifyRemoval(ctx, appType, inst.Name, record.Title, "seeding_limit_reached")
	}
}

// matchSeedingGroup finds the first matching seeding rule group for a download item.
// Groups are pre-sorted by priority DESC so first match wins. Returns nil if no match.
func matchSeedingGroup(item downloadclient.DownloadItem, groups []database.SeedingRuleGroup) *database.SeedingRuleGroup {
	for i := range groups {
		g := &groups[i]
		switch g.MatchType {
		case "tracker":
			if g.MatchPattern != "" && item.TrackerURL != "" &&
				strings.Contains(strings.ToLower(item.TrackerURL), strings.ToLower(g.MatchPattern)) {
				return g
			}
		case "category":
			if strings.EqualFold(item.Category, g.MatchPattern) {
				return g
			}
		case "tag":
			for _, tag := range item.Tags {
				if strings.EqualFold(tag, g.MatchPattern) {
					return g
				}
			}
		}
	}
	return nil
}

// seedingLimitReachedEx evaluates whether a torrent has exceeded the given
// seeding limits. Extracted to support both global and per-group limits.
func seedingLimitReachedEx(maxRatio float64, maxHours int, mode string, item downloadclient.DownloadItem) bool {
	ratioMet := maxRatio > 0 && item.Ratio >= maxRatio
	timeMet := maxHours > 0 && item.SeedingTime >= int64(maxHours)*3600

	if maxRatio <= 0 && maxHours <= 0 {
		return false
	}
	if maxRatio <= 0 {
		return timeMet
	}
	if maxHours <= 0 {
		return ratioMet
	}
	if mode == "and" {
		return ratioMet && timeMet
	}
	return ratioMet || timeMet
}

// seedingLimitReached evaluates whether a torrent has exceeded the configured
// seeding limits. In "or" mode either condition triggers; in "and" mode both must be met.
// recheckPaused triggers a data integrity recheck on paused torrents and
// auto-resumes them if they're already complete. This recovers torrents that
// were paused due to transient errors but have valid data.
func (c *Cleaner) recheckPaused(ctx context.Context, log *slog.Logger, appType database.AppType, records []arrclient.QueueRecord) {
	torrentClient := c.getDownloadClient(ctx, appType)
	if torrentClient == nil {
		return
	}

	items, err := torrentClient.GetItems(ctx)
	if err != nil {
		log.Error("failed to get torrent items for recheck", "error", err)
		return
	}

	itemsByID := make(map[string]downloadclient.DownloadItem, len(items))
	for _, item := range items {
		itemsByID[strings.ToLower(item.ID)] = item
	}

	for _, record := range records {
		if record.Protocol != "torrent" || record.DownloadID == "" {
			continue
		}

		item, ok := itemsByID[strings.ToLower(record.DownloadID)]
		if !ok {
			continue
		}

		if !isPausedStatus(item.Status) {
			continue
		}

		if item.Progress >= 1.0 {
			// Already complete — just resume, no recheck needed.
			log.Info("resuming complete paused torrent", "title", record.Title, "hash", item.ID)
			if err := torrentClient.ResumeItem(ctx, item.ID); err != nil {
				log.Error("failed to resume torrent", "title", record.Title, "error", err)
			}
			continue
		}

		// Incomplete and paused — trigger recheck so the client verifies data integrity.
		log.Info("rechecking paused torrent", "title", record.Title, "hash", item.ID, "progress", item.Progress)
		if err := torrentClient.RecheckItem(ctx, item.ID); err != nil {
			log.Error("failed to trigger recheck", "title", record.Title, "error", err)
		}
	}
}

// isPausedStatus returns true if the download client status indicates a paused torrent.
func isPausedStatus(status string) bool {
	switch status {
	case "pausedDL", "pausedUP": // qBittorrent
		return true
	case "stopped": // Transmission
		return true
	case "Paused": // Deluge
		return true
	case "paused": // rTorrent
		return true
	}
	return false
}

func (c *Cleaner) seedingLimitReached(settings *database.QueueCleanerSettings, item downloadclient.DownloadItem) bool {
	return seedingLimitReachedEx(settings.SeedingMaxRatio, settings.SeedingMaxHours, settings.SeedingMode, item)
}

// cleanDeletedMedia removes queue items whose media file has been externally deleted.
// Uses enriched queue data — the embedded media object's HasFile field tells us
// whether the *arr instance still sees a file on disk for this media.
func (c *Cleaner) cleanDeletedMedia(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance, client *arrclient.Client, apiVersion string, records []arrclient.QueueRecord, lurker lurking.ArrLurker, searchBudget *int) {
	for _, record := range records {
		// Only check items that have already been imported — these are the ones
		// where the media file was present and may have been externally deleted.
		if record.TrackedDownloadState != "imported" {
			continue
		}

		hasFile, ok := record.MediaHasFile()
		if !ok {
			continue // enriched data unavailable, skip
		}
		if hasFile {
			continue // file still present, nothing to do
		}

		if settings.DryRun {
			log.Warn("[DRY-RUN] would remove (media file deleted)", "title", record.Title, "media_id", record.MediaID())
			continue
		}

		log.Warn("media file deleted externally, removing queue item",
			"title", record.Title, "media_id", record.MediaID())
		if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), false); err != nil {
			log.Error("failed to remove deleted-media item", "title", record.Title, "error", err)
			continue
		}
		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
		c.notifyRemoval(ctx, appType, inst.Name, record.Title, "media_file_deleted")
		if settings.SearchOnRemove {
			c.triggerReSearch(ctx, log, lurker, client, record, appType, inst.ID, settings.SearchCooldownHours, searchBudget, settings.MaxSearchFailures)
		}
	}
}

// cleanUnmonitored removes queue items for media that has been unmonitored.
// Uses enriched queue data — the embedded media object's Monitored field tells us
// whether the user still wants this media downloaded.
func (c *Cleaner) cleanUnmonitored(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance, client *arrclient.Client, apiVersion string, records []arrclient.QueueRecord, lurker lurking.ArrLurker, searchBudget *int) {
	for _, record := range records {
		// Skip items that have already been imported — they completed successfully.
		if record.TrackedDownloadState == "imported" {
			continue
		}

		monitored, ok := record.MediaMonitored()
		if !ok {
			continue // enriched data unavailable, skip
		}
		if monitored {
			continue // still monitored, nothing to do
		}

		if settings.DryRun {
			log.Warn("[DRY-RUN] would remove (media unmonitored)", "title", record.Title, "media_id", record.MediaID())
			continue
		}

		log.Warn("media unmonitored, removing queue item",
			"title", record.Title, "media_id", record.MediaID())
		if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), false); err != nil {
			log.Error("failed to remove unmonitored item", "title", record.Title, "error", err)
			continue
		}
		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
		c.notifyRemoval(ctx, appType, inst.Name, record.Title, "media_unmonitored")
	}
}

// cleanMismatches detects and strikes queue items where the download doesn't
// match the expected media (wrong series/movie/episode) based on *arr status messages.
// Uses the strike system since mismatches can be transient during initial processing.
func (c *Cleaner) cleanMismatches(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance, client *arrclient.Client, apiVersion string, records []arrclient.QueueRecord, lurker lurking.ArrLurker, searchBudget *int) {
	for _, record := range records {
		// Only check items in warning state with import pending/failed — that's where
		// mismatch messages appear. Skip items that imported successfully.
		if record.TrackedDownloadStatus != "warning" {
			continue
		}
		if record.TrackedDownloadState != "importPending" && record.TrackedDownloadState != "importFailed" {
			continue
		}
		if !isMismatchedRelease(record) {
			continue
		}

		if settings.DryRun {
			log.Info("[DRY-RUN] would strike (metadata mismatch)", "title", record.Title, "download_id", record.DownloadID)
			continue
		}

		count, err := c.db.AddStrikeAndCount(ctx, appType, inst.ID, record.DownloadID, record.Title, "mismatch", settings.StrikeWindowHours)
		if err != nil {
			log.Error("failed to add mismatch strike", "error", err)
			continue
		}
		metrics.QueueCleanerStrikes.WithLabelValues(string(appType), inst.Name).Inc()

		maxStrikes := effectiveMaxStrikes("mismatch", settings)
		log.Info("mismatch strike added", "title", record.Title, "strikes", count, "max", maxStrikes)

		if count >= maxStrikes {
			log.Warn("metadata mismatch — max strikes reached, removing", "title", record.Title, "download_id", record.DownloadID)
			if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), shouldBlocklist("mismatch", settings)); err != nil {
				log.Error("failed to remove mismatched item", "title", record.Title, "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, "mismatch"); err != nil {
				log.Warn("failed to log blocklist", "title", record.Title, "error", err)
			}
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, record.Title, "mismatch")
			if settings.SearchOnRemove {
				c.triggerReSearch(ctx, log, lurker, client, record, appType, inst.ID, settings.SearchCooldownHours, searchBudget, settings.MaxSearchFailures)
			}
		}
	}
}

// syncBlocklistAcross propagates removals to sibling instances of the same app type.
// Matching uses external media IDs (TMDB for Radarr, TVDB+season for Sonarr, etc.)
// so that different releases of the same media are correctly matched across instances
// with different quality profiles or custom format scores. Falls back to title
// matching only when enriched media data is unavailable.
func (c *Cleaner) syncBlocklistAcross(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, instances []database.AppInstance, removals map[uuid.UUID][]removal) {
	// Build sets of removed media keys and titles across all instances.
	removedMediaKeys := make(map[string]bool)
	removedTitles := make(map[string]bool)
	for _, items := range removals {
		for _, r := range items {
			if r.MediaKey != "" {
				removedMediaKeys[r.MediaKey] = true
			}
			removedTitles[strings.ToLower(strings.TrimSpace(r.Title))] = true
		}
	}
	if len(removedMediaKeys) == 0 && len(removedTitles) == 0 {
		return
	}

	genSettings, err := c.db.GetGeneralSettings(ctx)
	if err != nil {
		log.Error("cross-arr sync: failed to load general settings", "error", err)
		return
	}

	apiVersion := apiVersionFor(appType)

	for _, inst := range instances {
		if ctx.Err() != nil {
			return
		}

		// Build sets of this instance's own removals so we skip them.
		ownMediaKeys := make(map[string]bool)
		ownTitles := make(map[string]bool)
		for _, r := range removals[inst.ID] {
			if r.MediaKey != "" {
				ownMediaKeys[r.MediaKey] = true
			}
			ownTitles[strings.ToLower(strings.TrimSpace(r.Title))] = true
		}

		client := arrclient.NewClient(
			inst.APIURL, inst.APIKey,
			time.Duration(genSettings.APITimeout)*time.Second,
			genSettings.SSLVerify,
		)

		// Fetch enriched queue to get external media IDs for matching.
		queue, err := getEnrichedQueue(ctx, client, appType)
		if err != nil {
			log.Warn("cross-arr sync: failed to get enriched queue", "instance", inst.Name, "error", err)
			continue
		}

		for _, record := range queue.Records {
			mediaKey := record.MediaKey()
			normalTitle := strings.ToLower(strings.TrimSpace(record.Title))

			// Prefer media key matching; fall back to title if no key available.
			matched := false
			if mediaKey != "" && removedMediaKeys[mediaKey] && !ownMediaKeys[mediaKey] {
				matched = true
			} else if mediaKey == "" && removedTitles[normalTitle] && !ownTitles[normalTitle] {
				// Title fallback only when enriched data is unavailable.
				matched = true
			}
			if !matched {
				continue
			}

			log.Info("cross-arr sync: removing matching item",
				"instance", inst.Name, "title", record.Title, "media_key", mediaKey)

			if settings.DryRun {
				log.Info("[DRY-RUN] would remove cross-arr match",
					"instance", inst.Name, "title", record.Title, "media_key", mediaKey)
				continue
			}

			if err := client.DeleteQueueItem(ctx, apiVersion, record.ID, effectiveRemoveFromClient(settings), true); err != nil {
				log.Error("cross-arr sync: failed to remove item",
					"instance", inst.Name, "title", record.Title, "error", err)
				continue
			}
			if err := c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, "cross_arr_sync"); err != nil {
				log.Warn("cross-arr sync: failed to log blocklist", "title", record.Title, "error", err)
			}
			metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), inst.Name).Inc()
			metrics.QueueCleanerBlocklistAdditions.WithLabelValues(string(appType), inst.Name).Inc()
			c.notifyRemoval(ctx, appType, inst.Name, record.Title, "cross_arr_sync")
		}
	}
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

	// Build cross-seed map: count items sharing the same save path & size.
	crossSeedCounts := countCrossSeeds(items)

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

		// Skip cross-seeded items to avoid breaking other seeders.
		if settings.SkipCrossSeeds && isCrossSeeded(item, crossSeedCounts) {
			continue
		}

		deleteFiles := settings.OrphanDeleteFiles
		if deleteFiles && settings.KeepArchives {
			deleteFiles = false
		}
		if deleteFiles && settings.HardlinkProtection && item.SavePath != "" && hasHardlinks(item.SavePath) {
			log.Info("hardlinks detected, skipping file deletion for orphan", "name", item.Name, "path", item.SavePath)
			deleteFiles = false
		}

		// RecycleBin: move files to recycle folder instead of permanent deletion
		if deleteFiles && settings.RecycleBinEnabled && settings.RecycleBinPath != "" && item.SavePath != "" {
			if err := moveToRecycleBin(item.SavePath, settings.RecycleBinPath); err != nil {
				log.Error("recycle bin move failed, falling back to deletion", "name", item.Name, "error", err)
			} else {
				log.Info("moved to recycle bin", "name", item.Name, "source", item.SavePath, "dest", settings.RecycleBinPath)
				deleteFiles = false
			}
		}

		log.Info("removing orphan download",
			"name", item.Name,
			"id", item.ID,
			"category", item.Category,
			"added_at", item.AddedAt,
			"delete_files", deleteFiles,
		)

		if settings.DryRun {
			log.Info("[DRY-RUN] would remove orphan", "name", item.Name, "id", item.ID)
			continue
		}

		if err := dlClient.RemoveItem(ctx, item.ID, deleteFiles); err != nil {
			log.Error("orphan: failed to remove item", "name", item.Name, "error", err)
			continue
		}

		metrics.QueueCleanerItemsRemoved.WithLabelValues(string(appType), "orphan").Inc()
		c.notifyRemoval(ctx, appType, "orphan", item.Name, "orphan_not_tracked")
	}
}

// pathSizeKey uniquely identifies download content by its location and total size.
type pathSizeKey struct {
	SavePath  string
	TotalSize int64
}

// countCrossSeeds builds a map from pathSizeKey → number of items sharing that content.
func countCrossSeeds(items []downloadclient.DownloadItem) map[pathSizeKey]int {
	counts := make(map[pathSizeKey]int, len(items))
	for _, item := range items {
		if item.SavePath == "" || item.TotalSize == 0 {
			continue
		}
		counts[pathSizeKey{SavePath: item.SavePath, TotalSize: item.TotalSize}]++
	}
	return counts
}

// isCrossSeeded returns true if the item shares its save path and total size
// with at least one other item (indicative of cross-seeding).
func isCrossSeeded(item downloadclient.DownloadItem, counts map[pathSizeKey]int) bool {
	if item.SavePath == "" || item.TotalSize == 0 {
		return false
	}
	return counts[pathSizeKey{SavePath: item.SavePath, TotalSize: item.TotalSize}] > 1
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

// getSABnzbdStatuses fetches the SABnzbd queue from all enabled SABnzbd download
// client instances and returns a combined map of downloadID -> status.
func (c *Cleaner) getSABnzbdStatuses(ctx context.Context) map[string]string {
	statuses := make(map[string]string)

	// Try new multi-instance download clients first.
	instances, err := c.db.ListEnabledDownloadClientInstances(ctx)
	if err == nil {
		for _, inst := range instances {
			if inst.ClientType != "sabnzbd" || inst.URL == "" {
				continue
			}
			timeout := time.Duration(inst.Timeout) * time.Second
			if timeout == 0 {
				timeout = 30 * time.Second
			}
			sabClient := sabnzbd.NewClient(inst.URL, inst.APIKey, timeout)
			queue, err := sabClient.GetQueue(ctx)
			if err != nil {
				continue
			}
			for _, slot := range queue.Slots {
				statuses[slot.NzoID] = slot.Status
			}
		}
		if len(statuses) > 0 {
			return statuses
		}
	}

	// Fallback to legacy singleton SABnzbd settings.
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

// getEnrichedQueue fetches the queue with embedded media data (movie, series+episode,
// album, book) so QueueRecord.MediaKey() returns external IDs for cross-arr matching.
func getEnrichedQueue(ctx context.Context, client *arrclient.Client, appType database.AppType) (*arrclient.QueueResponse, error) {
	switch appType {
	case database.AppSonarr:
		return client.SonarrGetQueueEnriched(ctx)
	case database.AppRadarr:
		return client.RadarrGetQueueEnriched(ctx)
	case database.AppLidarr:
		return client.LidarrGetQueueEnriched(ctx)
	case database.AppReadarr:
		return client.ReadarrGetQueueEnriched(ctx)
	case database.AppWhisparr:
		return client.WhisparrGetQueueEnriched(ctx)
	case database.AppEros:
		return client.ErosGetQueueEnriched(ctx)
	default:
		return nil, fmt.Errorf("unsupported app type for enriched queue: %s", appType)
	}
}

// filterProtectedTags resolves the comma-separated tag labels to IDs via the
// instance's tag API and removes any queue records whose media has a matching tag.
func filterProtectedTags(ctx context.Context, log *slog.Logger, client *arrclient.Client, appType database.AppType, protectedCSV string, records []arrclient.QueueRecord) []arrclient.QueueRecord {
	labels := make(map[string]struct{})
	for _, raw := range strings.Split(protectedCSV, ",") {
		label := strings.TrimSpace(strings.ToLower(raw))
		if label != "" {
			labels[label] = struct{}{}
		}
	}
	if len(labels) == 0 {
		return records
	}

	tags, err := client.GetTags(ctx, apiVersionFor(appType))
	if err != nil {
		log.Warn("failed to fetch tags for protected-tag filtering, skipping filter", "error", err)
		return records
	}

	protectedIDs := make(map[int]struct{})
	for _, t := range tags {
		if _, ok := labels[strings.ToLower(t.Label)]; ok {
			protectedIDs[t.ID] = struct{}{}
		}
	}
	if len(protectedIDs) == 0 {
		return records
	}

	filtered := make([]arrclient.QueueRecord, 0, len(records))
	for _, rec := range records {
		skip := false
		for _, tagID := range rec.MediaTags() {
			if _, ok := protectedIDs[tagID]; ok {
				log.Info("skipping protected item", "title", rec.Title, "tag_id", tagID)
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, rec)
		}
	}
	return filtered
}

// isBandwidthSaturated estimates total download bandwidth from queue records and
// returns true when usage exceeds 80% of the configured limit. When true, slow
// detection is suppressed to avoid false positives from a full pipe.
// effectiveMaxStrikes returns the per-reason override if set (> 0), otherwise
// falls back to the global MaxStrikes setting.
func effectiveMaxStrikes(reason string, settings *database.QueueCleanerSettings) int {
	switch reason {
	case "stalled":
		if settings.MaxStrikesStalled > 0 {
			return settings.MaxStrikesStalled
		}
	case "slow":
		if settings.MaxStrikesSlow > 0 {
			return settings.MaxStrikesSlow
		}
	case "metadata_stuck":
		if settings.MaxStrikesMetadata > 0 {
			return settings.MaxStrikesMetadata
		}
	case "paused_in_sabnzbd":
		if settings.MaxStrikesPaused > 0 {
			return settings.MaxStrikesPaused
		}
	case "queued":
		if settings.MaxStrikesQueued > 0 {
			return settings.MaxStrikesQueued
		}
	case "unregistered":
		if settings.MaxStrikesUnregistered > 0 {
			return settings.MaxStrikesUnregistered
		}
	case "mismatch":
		if settings.MaxStrikesMismatch > 0 {
			return settings.MaxStrikesMismatch
		}
	}
	return settings.MaxStrikes
}

// shouldBlocklist returns the per-reason blocklist flag if available, falling
// back to the global BlocklistOnRemove setting.
func shouldBlocklist(reason string, settings *database.QueueCleanerSettings) bool {
	switch reason {
	case "stalled":
		return settings.BlocklistStalled
	case "slow":
		return settings.BlocklistSlow
	case "metadata_stuck":
		return settings.BlocklistMetadata
	case "duplicate":
		return settings.BlocklistDuplicate
	case "unregistered":
		return settings.BlocklistUnregistered
	case "mismatch":
		return settings.BlocklistMismatch
	}
	return settings.BlocklistOnRemove
}

// effectiveRemoveFromClient returns false when KeepArchives is enabled,
// preserving downloaded files for unpackerr. Otherwise returns the user's
// configured RemoveFromClient preference.
func effectiveRemoveFromClient(settings *database.QueueCleanerSettings) bool {
	if settings.KeepArchives {
		return false
	}
	return settings.RemoveFromClient
}

// resolveTagID finds or creates a tag with the given label, returning its ID.
// Returns 0 if the tag cannot be resolved.
func resolveTagID(ctx context.Context, log *slog.Logger, client *arrclient.Client, apiVersion, label string) int {
	tags, err := client.GetTags(ctx, apiVersion)
	if err != nil {
		log.Warn("failed to get tags for obsolete tagging", "error", err)
		return 0
	}
	lower := strings.ToLower(label)
	for _, t := range tags {
		if strings.ToLower(t.Label) == lower {
			return t.ID
		}
	}
	tag, err := client.CreateTag(ctx, apiVersion, label)
	if err != nil {
		log.Warn("failed to create obsolete tag", "label", label, "error", err)
		return 0
	}
	return tag.ID
}

func isBandwidthSaturated(log *slog.Logger, settings *database.QueueCleanerSettings, records []arrclient.QueueRecord) bool {
	if settings.BandwidthLimitBytesPerSec <= 0 || settings.SlowThresholdBytesPerSec <= 0 {
		return false
	}

	var totalSpeed int64
	for _, rec := range records {
		if rec.Size > 0 && rec.Sizeleft > 0 && rec.Sizeleft < rec.Size {
			if tl := parseTimeleft(rec.TimeleftStr); tl > 0 {
				totalSpeed += rec.Sizeleft / int64(tl.Seconds())
			}
		}
	}

	threshold := settings.BandwidthLimitBytesPerSec * 80 / 100
	if totalSpeed >= threshold {
		log.Info("bandwidth saturated, skipping slow detection",
			"estimated_speed", totalSpeed, "threshold", threshold)
		return true
	}
	return false
}

// filterIgnoredIndexers removes queue records from indexers listed in the
// comma-separated ignoredCSV setting.
func filterIgnoredIndexers(log *slog.Logger, ignoredCSV string, records []arrclient.QueueRecord) []arrclient.QueueRecord {
	ignored := make(map[string]struct{})
	for _, raw := range strings.Split(ignoredCSV, ",") {
		name := strings.TrimSpace(strings.ToLower(raw))
		if name != "" {
			ignored[name] = struct{}{}
		}
	}
	if len(ignored) == 0 {
		return records
	}

	filtered := make([]arrclient.QueueRecord, 0, len(records))
	for _, rec := range records {
		if _, ok := ignored[strings.ToLower(rec.Indexer)]; ok {
			log.Info("skipping ignored indexer item", "title", rec.Title, "indexer", rec.Indexer)
			continue
		}
		filtered = append(filtered, rec)
	}
	return filtered
}

// filterIgnoredDownloadClients removes queue records from download clients
// listed in the comma-separated ignoredCSV setting.
func filterIgnoredDownloadClients(log *slog.Logger, ignoredCSV string, records []arrclient.QueueRecord) []arrclient.QueueRecord {
	ignored := make(map[string]struct{})
	for _, raw := range strings.Split(ignoredCSV, ",") {
		name := strings.TrimSpace(strings.ToLower(raw))
		if name != "" {
			ignored[name] = struct{}{}
		}
	}
	if len(ignored) == 0 {
		return records
	}

	filtered := make([]arrclient.QueueRecord, 0, len(records))
	for _, rec := range records {
		if _, ok := ignored[strings.ToLower(rec.DownloadClient)]; ok {
			log.Info("skipping ignored download client item", "title", rec.Title, "client", rec.DownloadClient)
			continue
		}
		filtered = append(filtered, rec)
	}
	return filtered
}

// triggerReSearch asks the arr instance to search for a replacement after a
// queue item is removed. It is a best-effort operation — failures are logged
// but do not interrupt the cleanup loop. When cooldownHours > 0, recently
// searched media is skipped. When searchBudget is non-nil and points to a
// positive value, it is decremented on each search; when it reaches zero,
// further searches are skipped. Pass nil to disable budget tracking.
func (c *Cleaner) triggerReSearch(ctx context.Context, log *slog.Logger, lurker lurking.ArrLurker, client *arrclient.Client, record arrclient.QueueRecord, appType database.AppType, instID uuid.UUID, cooldownHours int, searchBudget *int, maxSearchFailures int) {
	mediaID := record.MediaID()
	if mediaID <= 0 {
		return
	}
	if searchBudget != nil && *searchBudget <= 0 {
		log.Debug("skipping re-search (per-run budget exhausted)", "title", record.Title, "media_id", mediaID)
		return
	}
	if cooldownHours > 0 {
		onCooldown, err := c.db.IsSearchOnCooldown(ctx, appType, instID, mediaID, cooldownHours)
		if err != nil {
			log.Warn("failed to check search cooldown", "media_id", mediaID, "error", err)
		} else if onCooldown {
			log.Debug("skipping re-search (cooldown active)", "title", record.Title, "media_id", mediaID)
			return
		}
	}
	if maxSearchFailures > 0 {
		limitReached, err := c.db.IsSearchFailureLimitReached(ctx, appType, instID, mediaID, maxSearchFailures)
		if err != nil {
			log.Warn("failed to check search failure limit", "media_id", mediaID, "error", err)
		} else if limitReached {
			log.Debug("skipping re-search (failure limit reached)", "title", record.Title, "media_id", mediaID)
			return
		}
	}
	if err := lurker.Search(ctx, client, mediaID); err != nil {
		log.Warn("re-search failed", "title", record.Title, "media_id", mediaID, "error", err)
		if maxSearchFailures > 0 {
			if ferr := c.db.RecordSearchFailure(ctx, appType, instID, mediaID); ferr != nil {
				log.Warn("failed to record search failure", "media_id", mediaID, "error", ferr)
			}
		}
	} else {
		log.Info("re-search triggered", "title", record.Title, "media_id", mediaID)
		if searchBudget != nil {
			*searchBudget--
		}
		if cooldownHours > 0 {
			if err := c.db.RecordSearch(ctx, appType, instID, mediaID); err != nil {
				log.Warn("failed to record search cooldown", "media_id", mediaID, "error", err)
			}
		}
		// Mark as processed in the lurking engine so it doesn't redundantly
		// search for the same media on its next cycle.
		if err := c.db.MarkProcessed(ctx, appType, instID, mediaID, "missing"); err != nil {
			log.Debug("failed to mark lurk-processed after re-search", "media_id", mediaID, "error", err)
		}
		if maxSearchFailures > 0 {
			if ferr := c.db.ClearSearchFailure(ctx, appType, instID, mediaID); ferr != nil {
				log.Debug("failed to clear search failure", "media_id", mediaID, "error", ferr)
			}
		}
	}
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
