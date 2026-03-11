package queuecleaner

import (
	"context"
	"testing"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
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
	record := arrclient.QueueRecord{Indexer: "1337x"}
	if isPrivateTracker(record) {
		t.Error("expected 1337x to be public")
	}

	empty := arrclient.QueueRecord{}
	if isPrivateTracker(empty) {
		t.Error("expected empty indexer to be public")
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
