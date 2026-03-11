package queuecleaner

import (
	"context"
	"testing"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	downloadclient "github.com/lusoris/lurkarr/internal/downloadclients"
)

func TestDetectProblemStalled(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Status:                "warning",
		TrackedDownloadStatus: "warning",
		Protocol:              "torrent",
	}
	settings := &database.QueueCleanerSettings{
		StrikePublic:  true,
		StrikePrivate: true,
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "stalled" {
		t.Errorf("detectProblem() = %q, want stalled", reason)
	}
}

func TestDetectProblemStalledPublicSkip(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Status:                "warning",
		TrackedDownloadStatus: "warning",
		Protocol:              "torrent",
	}
	settings := &database.QueueCleanerSettings{
		StrikePublic:  false,
		StrikePrivate: true,
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "" {
		t.Errorf("expected empty reason for public torrent with StrikePublic=false, got %q", reason)
	}
}

func TestDetectProblemUsenetSABnzbdQueued(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Protocol:              "usenet",
		DownloadID:            "sab123",
		Status:                "warning",
		TrackedDownloadStatus: "warning",
	}
	settings := &database.QueueCleanerSettings{StrikePublic: true}
	sabStatuses := map[string]string{"sab123": "Queued"}

	reason := c.detectProblem(record, settings, sabStatuses)
	if reason != "" {
		t.Errorf("expected empty reason for SABnzbd Queued item, got %q", reason)
	}
}

func TestDetectProblemUsenetSABnzbdGrabbing(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Protocol:              "usenet",
		DownloadID:            "sab456",
		Status:                "warning",
		TrackedDownloadStatus: "warning",
	}
	settings := &database.QueueCleanerSettings{StrikePublic: true}
	sabStatuses := map[string]string{"sab456": "Grabbing"}

	reason := c.detectProblem(record, settings, sabStatuses)
	if reason != "" {
		t.Errorf("expected empty reason for SABnzbd Grabbing, got %q", reason)
	}
}

func TestDetectProblemUsenetSABnzbdPaused(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Protocol:   "usenet",
		DownloadID: "sab789",
	}
	settings := &database.QueueCleanerSettings{}
	sabStatuses := map[string]string{"sab789": "Paused"}

	reason := c.detectProblem(record, settings, sabStatuses)
	if reason != "paused_in_sabnzbd" {
		t.Errorf("detectProblem() = %q, want paused_in_sabnzbd", reason)
	}
}

func TestDetectProblemMetadataStuck(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Size:     0,
		Sizeleft: 0,
		Status:   "downloading",
	}
	settings := &database.QueueCleanerSettings{
		MetadataStuckMinutes: 15,
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "metadata_stuck" {
		t.Errorf("detectProblem() = %q, want metadata_stuck", reason)
	}
}

func TestDetectProblemMetadataStuckDelay(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Size:     0,
		Sizeleft: 0,
		Status:   "delay",
	}
	settings := &database.QueueCleanerSettings{
		MetadataStuckMinutes: 15,
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "metadata_stuck" {
		t.Errorf("detectProblem() = %q, want metadata_stuck", reason)
	}
}

func TestDetectProblemMetadataStuckDisabled(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Size:     0,
		Sizeleft: 0,
		Status:   "downloading",
	}
	settings := &database.QueueCleanerSettings{
		MetadataStuckMinutes: 0,
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "" {
		t.Errorf("expected empty for disabled metadata stuck, got %q", reason)
	}
}

func TestDetectProblemSlowDownload(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Size:        10 * 1024 * 1024 * 1024, // 10 GB
		Sizeleft:    5 * 1024 * 1024 * 1024,  // 5 GB left
		TimeleftStr: "100:00:00",             // 100 hours = very slow
	}
	settings := &database.QueueCleanerSettings{
		SlowThresholdBytesPerSec: 100 * 1024, // 100 KB/s
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "slow" {
		t.Errorf("detectProblem() = %q, want slow", reason)
	}
}

func TestDetectProblemSlowIgnoreAboveBytes(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Size:        50 * 1024 * 1024 * 1024,
		Sizeleft:    40 * 1024 * 1024 * 1024,
		TimeleftStr: "100:00:00",
	}
	settings := &database.QueueCleanerSettings{
		SlowThresholdBytesPerSec: 100 * 1024,
		SlowIgnoreAboveBytes:     30 * 1024 * 1024 * 1024, // Ignore if >30GB remaining
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "" {
		t.Errorf("expected empty for large remaining download, got %q", reason)
	}
}

func TestDetectProblemNoProblem(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Status:                "downloading",
		TrackedDownloadStatus: "ok",
		Size:                  10 * 1024 * 1024 * 1024,
		Sizeleft:              1 * 1024 * 1024 * 1024,
		TimeleftStr:           "00:10:00",
	}
	settings := &database.QueueCleanerSettings{
		SlowThresholdBytesPerSec: 100 * 1024,
	}

	reason := c.detectProblem(record, settings, nil)
	if reason != "" {
		t.Errorf("expected empty for healthy download, got %q", reason)
	}
}

func TestAPIVersionFor(t *testing.T) {
	tests := []struct {
		appType database.AppType
		want    string
	}{
		{database.AppLidarr, "v1"},
		{database.AppReadarr, "v1"},
		{database.AppSonarr, "v3"},
		{database.AppRadarr, "v3"},
		{database.AppWhisparr, "v3"},
		{database.AppEros, "v3"},
	}
	for _, tt := range tests {
		t.Run(string(tt.appType), func(t *testing.T) {
			got := apiVersionFor(tt.appType)
			if got != tt.want {
				t.Errorf("apiVersionFor(%s) = %q, want %q", tt.appType, got, tt.want)
			}
		})
	}
}

func TestIsPrivateTracker(t *testing.T) {
	tests := []struct {
		name    string
		record  arrclient.QueueRecord
		private bool
	}{
		{
			name:    "empty indexer is public",
			record:  arrclient.QueueRecord{},
			private: false,
		},
		{
			name:    "known public indexer 1337x",
			record:  arrclient.QueueRecord{Indexer: "1337x"},
			private: false,
		},
		{
			name:    "known public indexer YTS case-insensitive",
			record:  arrclient.QueueRecord{Indexer: "YTS"},
			private: false,
		},
		{
			name:    "known public indexer nyaa",
			record:  arrclient.QueueRecord{Indexer: "nyaa"},
			private: false,
		},
		{
			name:    "indexer flags set means private",
			record:  arrclient.QueueRecord{Indexer: "1337x", IndexerFlags: 1},
			private: true,
		},
		{
			name:    "unknown indexer treated as private",
			record:  arrclient.QueueRecord{Indexer: "MyPrivateTracker"},
			private: true,
		},
		{
			name:    "indexer flags only no name",
			record:  arrclient.QueueRecord{IndexerFlags: 32},
			private: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isPrivateTracker(tt.record)
			if got != tt.private {
				t.Errorf("isPrivateTracker(%+v) = %v, want %v", tt.record, got, tt.private)
			}
		})
	}
}

func TestSleepCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok := sleep(ctx, 1*time.Minute)
	if ok {
		t.Error("expected sleep to return false with cancelled context")
	}
}

func TestSeedingLimitReached(t *testing.T) {
	c := &Cleaner{}

	tests := []struct {
		name     string
		settings *database.QueueCleanerSettings
		item     downloadclient.DownloadItem
		want     bool
	}{
		{
			name:     "no limits configured",
			settings: &database.QueueCleanerSettings{},
			item:     downloadclient.DownloadItem{Ratio: 5.0, SeedingTime: 7200},
			want:     false,
		},
		{
			name:     "ratio exceeded or mode",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{Ratio: 2.5},
			want:     true,
		},
		{
			name:     "ratio not exceeded",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{Ratio: 1.5},
			want:     false,
		},
		{
			name:     "time exceeded or mode",
			settings: &database.QueueCleanerSettings{SeedingMaxHours: 24, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{SeedingTime: 25 * 3600},
			want:     true,
		},
		{
			name:     "time not exceeded",
			settings: &database.QueueCleanerSettings{SeedingMaxHours: 24, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{SeedingTime: 12 * 3600},
			want:     false,
		},
		{
			name:     "or mode either ratio",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMaxHours: 24, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{Ratio: 3.0, SeedingTime: 1 * 3600},
			want:     true,
		},
		{
			name:     "or mode either time",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMaxHours: 24, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{Ratio: 0.5, SeedingTime: 25 * 3600},
			want:     true,
		},
		{
			name:     "or mode neither met",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMaxHours: 24, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{Ratio: 0.5, SeedingTime: 12 * 3600},
			want:     false,
		},
		{
			name:     "and mode both met",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMaxHours: 24, SeedingMode: "and"},
			item:     downloadclient.DownloadItem{Ratio: 3.0, SeedingTime: 25 * 3600},
			want:     true,
		},
		{
			name:     "and mode only ratio met",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMaxHours: 24, SeedingMode: "and"},
			item:     downloadclient.DownloadItem{Ratio: 3.0, SeedingTime: 12 * 3600},
			want:     false,
		},
		{
			name:     "and mode only time met",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMaxHours: 24, SeedingMode: "and"},
			item:     downloadclient.DownloadItem{Ratio: 0.5, SeedingTime: 25 * 3600},
			want:     false,
		},
		{
			name:     "exact ratio boundary",
			settings: &database.QueueCleanerSettings{SeedingMaxRatio: 2.0, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{Ratio: 2.0},
			want:     true,
		},
		{
			name:     "exact time boundary",
			settings: &database.QueueCleanerSettings{SeedingMaxHours: 24, SeedingMode: "or"},
			item:     downloadclient.DownloadItem{SeedingTime: 24 * 3600},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.seedingLimitReached(tt.settings, tt.item)
			if got != tt.want {
				t.Errorf("seedingLimitReached() = %v, want %v (ratio=%.1f, seeding=%ds, maxRatio=%.1f, maxHours=%d, mode=%s)",
					got, tt.want, tt.item.Ratio, tt.item.SeedingTime,
					tt.settings.SeedingMaxRatio, tt.settings.SeedingMaxHours, tt.settings.SeedingMode)
			}
		})
	}
}
