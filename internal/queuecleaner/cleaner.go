package queuecleaner

import (
	"context"
	"log/slog"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/hunting"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/sabnzbd"
)

// Cleaner monitors download queues and removes stalled/slow/duplicate items.
type Cleaner struct {
	db     *database.DB
	logger *logging.Logger
	cancel context.CancelFunc
	wg     sync.WaitGroup
}

// New creates a new queue cleaner.
func New(db *database.DB, logger *logging.Logger) *Cleaner {
	return &Cleaner{db: db, logger: logger}
}

// Start launches cleaner goroutines for each app type.
func (c *Cleaner) Start(ctx context.Context) {
	ctx, c.cancel = context.WithCancel(ctx)
	for _, appType := range database.AllAppTypes() {
		if hunting.HunterFor(appType) == nil {
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
			c.cleanInstance(ctx, log, appType, settings, inst)
		}

		if !sleep(ctx, time.Duration(settings.CheckIntervalSeconds)*time.Second) {
			return
		}
	}
}

func (c *Cleaner) cleanInstance(ctx context.Context, log *slog.Logger, appType database.AppType, settings *database.QueueCleanerSettings, inst database.AppInstance) {
	log = log.With("instance", inst.Name)

	hunter := hunting.HunterFor(appType)
	if hunter == nil {
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

	queue, err := hunter.GetQueue(ctx, client)
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
			_ = c.db.LogBlocklist(ctx, appType, inst.ID, "", d.RemoveTitle, "duplicate_lower_score")
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
				_ = c.db.ResetStrikes(ctx, appType, inst.ID, record.DownloadID)
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
			_ = c.db.LogBlocklist(ctx, appType, inst.ID, record.DownloadID, record.Title, reason+"_max_strikes")
		}
	}
}

// detectProblem checks if a queue item is stalled or slow.
// Returns the reason string, or "" if no problem.
func (c *Cleaner) detectProblem(record arrclient.QueueRecord, settings *database.QueueCleanerSettings, sabStatuses map[string]string) string {
	// For Usenet via SABnzbd: check actual SABnzbd status
	// SABnzbd items show as "Queued" when they're just waiting for a slot,
	// NOT because they're stalled. This fixes the Cleanuparr bug.
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

	// Check for stalled torrents
	if record.Status == "warning" && record.TrackedDownloadStatus == "warning" {
		return "stalled"
	}

	// Check download speed for active downloads
	if record.Size > 0 && record.Sizeleft > 0 && record.Sizeleft < record.Size && settings.SlowThresholdBytesPerSec > 0 {
		downloaded := record.Size - record.Sizeleft
		// Parse timeleft to estimate speed
		if tl := parseTimeleft(record.TimeleftStr); tl > 0 && downloaded > 0 {
			estimatedSpeed := record.Sizeleft / int64(tl.Seconds())
			if estimatedSpeed > 0 && estimatedSpeed < settings.SlowThresholdBytesPerSec {
				return "slow"
			}
		}
	}

	return ""
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
