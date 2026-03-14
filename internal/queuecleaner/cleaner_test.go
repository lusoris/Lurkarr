package queuecleaner

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	downloadclient "github.com/lusoris/lurkarr/internal/downloadclients"
	"github.com/lusoris/lurkarr/internal/logging"
	"go.uber.org/mock/gomock"
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, sabStatuses, false)
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

	reason := c.detectProblem(record, settings, sabStatuses, false)
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

	reason := c.detectProblem(record, settings, sabStatuses, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

	reason := c.detectProblem(record, settings, nil, false)
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

func TestParseExcludedCategories(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  map[string]bool
	}{
		{"empty", "", map[string]bool{}},
		{"single", "cross-seed", map[string]bool{"cross-seed": true}},
		{"multiple", "cross-seed, manual, tv", map[string]bool{"cross-seed": true, "manual": true, "tv": true}},
		{"spaces", " foo , bar , ", map[string]bool{"foo": true, "bar": true}},
		{"case insensitive", "CrossSeed,Manual", map[string]bool{"crossseed": true, "manual": true}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseExcludedCategories(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("parseExcludedCategories(%q) = %v, want %v", tt.input, got, tt.want)
			}
			for k := range tt.want {
				if !got[k] {
					t.Errorf("expected key %q in result %v", k, got)
				}
			}
		})
	}
}

func TestIsOrphan(t *testing.T) {
	// Utility test: given a set of known IDs and an item, verify orphan detection logic.
	// This tests the core filtering that cleanOrphans applies.
	knownIDs := map[string]bool{
		"abc123": true,
		"def456": true,
	}
	excludedCats := map[string]bool{
		"cross-seed": true,
	}
	now := time.Now().Unix()
	graceSeconds := int64(120 * 60) // 120 minutes

	tests := []struct {
		name     string
		item     downloadclient.DownloadItem
		isOrphan bool
	}{
		{
			name:     "tracked by arr — not orphan",
			item:     downloadclient.DownloadItem{ID: "abc123", AddedAt: now - 9999},
			isOrphan: false,
		},
		{
			name:     "not tracked, past grace — orphan",
			item:     downloadclient.DownloadItem{ID: "unknown1", AddedAt: now - graceSeconds - 1},
			isOrphan: true,
		},
		{
			name:     "not tracked, within grace — not orphan",
			item:     downloadclient.DownloadItem{ID: "unknown2", AddedAt: now - 60},
			isOrphan: false,
		},
		{
			name:     "excluded category — not orphan",
			item:     downloadclient.DownloadItem{ID: "unknown3", Category: "cross-seed", AddedAt: now - graceSeconds - 1},
			isOrphan: false,
		},
		{
			name:     "no AddedAt, CompletedAt past grace — orphan",
			item:     downloadclient.DownloadItem{ID: "unknown4", CompletedAt: now - graceSeconds - 1},
			isOrphan: true,
		},
		{
			name:     "no AddedAt, CompletedAt within grace — not orphan",
			item:     downloadclient.DownloadItem{ID: "unknown5", CompletedAt: now - 60},
			isOrphan: false,
		},
		{
			name:     "no timestamps at all — orphan (no way to enforce grace)",
			item:     downloadclient.DownloadItem{ID: "unknown6"},
			isOrphan: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id := strings.ToLower(tt.item.ID)
			if knownIDs[id] {
				if tt.isOrphan {
					t.Error("expected orphan but item is tracked")
				}
				return
			}
			if excludedCats[strings.ToLower(tt.item.Category)] {
				if tt.isOrphan {
					t.Error("expected orphan but category is excluded")
				}
				return
			}
			// Grace period check
			withinGrace := tt.item.AddedAt > 0 && (now-tt.item.AddedAt) < graceSeconds
			if tt.item.AddedAt == 0 && tt.item.CompletedAt > 0 && (now-tt.item.CompletedAt) < graceSeconds {
				withinGrace = true
			}
			isOrphan := !withinGrace
			if isOrphan != tt.isOrphan {
				t.Errorf("orphan detection = %v, want %v", isOrphan, tt.isOrphan)
			}
		})
	}
}

func TestCountCrossSeeds(t *testing.T) {
	items := []downloadclient.DownloadItem{
		{ID: "hash1", SavePath: "/data/movies/Movie.2024", TotalSize: 5000},
		{ID: "hash2", SavePath: "/data/movies/Movie.2024", TotalSize: 5000}, // same path+size = cross-seed
		{ID: "hash3", SavePath: "/data/movies/Movie.2024", TotalSize: 3000}, // same path, different size
		{ID: "hash4", SavePath: "/data/tv/Show.S01", TotalSize: 8000},       // unique
		{ID: "hash5", SavePath: "", TotalSize: 1000},                        // no save path
		{ID: "hash6", SavePath: "/data/music/Album", TotalSize: 0},          // no size
	}

	counts := countCrossSeeds(items)

	key := pathSizeKey{SavePath: "/data/movies/Movie.2024", TotalSize: 5000}
	if counts[key] != 2 {
		t.Errorf("cross-seed count for matching pair = %d, want 2", counts[key])
	}

	key2 := pathSizeKey{SavePath: "/data/movies/Movie.2024", TotalSize: 3000}
	if counts[key2] != 1 {
		t.Errorf("count for different size = %d, want 1", counts[key2])
	}

	key3 := pathSizeKey{SavePath: "/data/tv/Show.S01", TotalSize: 8000}
	if counts[key3] != 1 {
		t.Errorf("unique item count = %d, want 1", counts[key3])
	}

	if len(counts) != 3 {
		t.Errorf("total keys = %d, want 3 (excluded empty path and zero size)", len(counts))
	}
}

func TestIsCrossSeeded(t *testing.T) {
	items := []downloadclient.DownloadItem{
		{ID: "hash1", SavePath: "/data/movies/Movie.2024", TotalSize: 5000},
		{ID: "hash2", SavePath: "/data/movies/Movie.2024", TotalSize: 5000},
		{ID: "hash3", SavePath: "/data/tv/Show.S01", TotalSize: 8000},
		{ID: "hash4", SavePath: "", TotalSize: 1000},
	}
	counts := countCrossSeeds(items)

	tests := []struct {
		name string
		item downloadclient.DownloadItem
		want bool
	}{
		{
			name: "cross-seeded item",
			item: downloadclient.DownloadItem{SavePath: "/data/movies/Movie.2024", TotalSize: 5000},
			want: true,
		},
		{
			name: "unique item",
			item: downloadclient.DownloadItem{SavePath: "/data/tv/Show.S01", TotalSize: 8000},
			want: false,
		},
		{
			name: "empty save path",
			item: downloadclient.DownloadItem{SavePath: "", TotalSize: 5000},
			want: false,
		},
		{
			name: "zero size",
			item: downloadclient.DownloadItem{SavePath: "/data/movies/Movie.2024", TotalSize: 0},
			want: false,
		},
		{
			name: "unknown path",
			item: downloadclient.DownloadItem{SavePath: "/data/other", TotalSize: 9999},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isCrossSeeded(tt.item, counts)
			if got != tt.want {
				t.Errorf("isCrossSeeded() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSyncBlocklistAcross(t *testing.T) {
	// Create a mock arr server for instance B that has the same movie (by TMDB ID)
	// but with a different release title (different quality profile).
	var deleteCount atomic.Int32
	serverB := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v3/queue") && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(arrclient.QueueResponse{
				TotalRecords: 2,
				Records: []arrclient.QueueRecord{
					{
						ID: 10, DownloadID: "dl-10",
						Title: "Bad.Movie.2024.1080p.BluRay-OTHER", Status: "downloading",
						MovieID: 5,
						Movie:   &arrclient.QueueMovie{TmdbID: 99001, Title: "Bad Movie"},
					},
					{
						ID: 11, DownloadID: "dl-11",
						Title: "Good.Movie.2024.1080p.WEB-DL-GOOD", Status: "downloading",
						MovieID: 6,
						Movie:   &arrclient.QueueMovie{TmdbID: 99002, Title: "Good Movie"},
					},
				},
			})
		case strings.HasPrefix(r.URL.Path, "/api/v3/queue/") && r.Method == http.MethodDelete:
			deleteCount.Add(1)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer serverB.Close()

	instA := uuid.New()
	instB := uuid.New()

	instances := []database.AppInstance{
		{ID: instA, AppType: database.AppRadarr, Name: "Radarr-1", APIURL: "http://unused", APIKey: "key1", Enabled: true},
		{ID: instB, AppType: database.AppRadarr, Name: "Radarr-2", APIURL: serverB.URL, APIKey: "key2", Enabled: true},
	}

	// Instance A removed TMDB 99001 (different release title) — should propagate to B by media key.
	removals := map[uuid.UUID][]removal{
		instA: {{Title: "Bad.Movie.2024.2160p.WEB-DL-GROUP", MediaKey: "tmdb:99001"}},
	}

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{
		APITimeout: 10,
		SSLVerify:  true,
	}, nil)
	store.EXPECT().LogBlocklist(gomock.Any(), database.AppRadarr, instB, "dl-10", "Bad.Movie.2024.1080p.BluRay-OTHER", "cross_arr_sync").Return(nil)

	logger := logging.New()
	defer logger.Close()
	c := &Cleaner{db: store, logger: logger}

	settings := &database.QueueCleanerSettings{
		RemoveFromClient: true,
		CrossArrSync:     true,
	}

	c.syncBlocklistAcross(context.Background(), slog.Default(), database.AppRadarr, settings, instances, removals)

	// Should have deleted exactly 1 item (same TMDB ID from instance B, despite different title).
	if got := deleteCount.Load(); got != 1 {
		t.Errorf("delete count = %d, want 1", got)
	}
}

func TestSyncBlocklistAcrossSkipsOwnRemovals(t *testing.T) {
	// Instance A removed a movie, and that same movie is also in A's queue
	// (shouldn't happen in practice, but verifies skip logic).
	var deleteCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/api/v3/queue") && r.Method == http.MethodGet:
			json.NewEncoder(w).Encode(arrclient.QueueResponse{
				TotalRecords: 1,
				Records: []arrclient.QueueRecord{
					{
						ID: 20, DownloadID: "dl-20",
						Title: "Removed.Movie.2024.1080p", Status: "downloading",
						MovieID: 1,
						Movie:   &arrclient.QueueMovie{TmdbID: 55555, Title: "Removed Movie"},
					},
				},
			})
		case strings.HasPrefix(r.URL.Path, "/api/v3/queue/") && r.Method == http.MethodDelete:
			deleteCount.Add(1)
			w.WriteHeader(http.StatusOK)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	instA := uuid.New()

	instances := []database.AppInstance{
		{ID: instA, AppType: database.AppRadarr, Name: "Radarr-1", APIURL: server.URL, APIKey: "key1", Enabled: true},
	}

	// Instance A removed TMDB 55555 — should NOT re-delete from A's own queue.
	removals := map[uuid.UUID][]removal{
		instA: {{Title: "Removed.Movie.2024.2160p", MediaKey: "tmdb:55555"}},
	}

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{
		APITimeout: 10,
		SSLVerify:  true,
	}, nil)

	logger := logging.New()
	defer logger.Close()
	c := &Cleaner{db: store, logger: logger}

	settings := &database.QueueCleanerSettings{
		RemoveFromClient: true,
		CrossArrSync:     true,
	}

	c.syncBlocklistAcross(context.Background(), slog.Default(), database.AppRadarr, settings, instances, removals)

	// Should NOT have deleted anything (the title was removed by this instance itself).
	if got := deleteCount.Load(); got != 0 {
		t.Errorf("delete count = %d, want 0 (own removal should be skipped)", got)
	}
}

func TestDetectProblemUnregistered(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Status:                "warning",
		TrackedDownloadStatus: "warning",
		Protocol:              "torrent",
		StatusMessages: []arrclient.StatusMessage{
			{Title: "test", Messages: []string{"Torrent is not registered with this tracker"}},
		},
	}
	settings := &database.QueueCleanerSettings{
		StrikePublic:        true,
		StrikePrivate:       true,
		UnregisteredEnabled: true,
	}

	reason := c.detectProblem(record, settings, nil, false)
	if reason != "unregistered" {
		t.Errorf("detectProblem() = %q, want unregistered", reason)
	}
}

func TestDetectProblemUnregisteredDisabled(t *testing.T) {
	c := &Cleaner{}
	record := arrclient.QueueRecord{
		Status:                "warning",
		TrackedDownloadStatus: "warning",
		Protocol:              "torrent",
		StatusMessages: []arrclient.StatusMessage{
			{Title: "test", Messages: []string{"Torrent is not registered with this tracker"}},
		},
	}
	settings := &database.QueueCleanerSettings{
		StrikePublic:        true,
		StrikePrivate:       true,
		UnregisteredEnabled: false,
	}

	// When disabled, should fall through to "stalled"
	reason := c.detectProblem(record, settings, nil, false)
	if reason != "stalled" {
		t.Errorf("detectProblem() = %q, want stalled (detection disabled)", reason)
	}
}

func TestDetectProblemUnregisteredUsenet(t *testing.T) {
	c := &Cleaner{}
	// Usenet items should not trigger unregistered detection
	record := arrclient.QueueRecord{
		Status:                "warning",
		TrackedDownloadStatus: "warning",
		Protocol:              "usenet",
		StatusMessages: []arrclient.StatusMessage{
			{Title: "test", Messages: []string{"not registered"}},
		},
	}
	settings := &database.QueueCleanerSettings{
		StrikePublic:        true,
		StrikePrivate:       true,
		UnregisteredEnabled: true,
	}

	reason := c.detectProblem(record, settings, nil, false)
	if reason != "stalled" {
		t.Errorf("detectProblem() = %q, want stalled (usenet can't be unregistered)", reason)
	}
}

func TestIsUnregisteredTorrent(t *testing.T) {
	tests := []struct {
		name     string
		messages []string
		want     bool
	}{
		{"unregistered keyword", []string{"The download is not registered"}, true},
		{"not found", []string{"Torrent not found on tracker"}, true},
		{"info hash", []string{"Could not find info hash"}, true},
		{"trumped", []string{"Release has been trumped"}, true},
		{"normal stall", []string{"The download is stalled with no connections"}, false},
		{"empty messages", []string{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			record := arrclient.QueueRecord{
				StatusMessages: []arrclient.StatusMessage{
					{Title: "test", Messages: tt.messages},
				},
			}
			if got := isUnregisteredTorrent(record); got != tt.want {
				t.Errorf("isUnregisteredTorrent() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEffectiveMaxStrikesUnregistered(t *testing.T) {
	settings := &database.QueueCleanerSettings{
		MaxStrikes:             5,
		MaxStrikesUnregistered: 2,
	}
	if got := effectiveMaxStrikes("unregistered", settings); got != 2 {
		t.Errorf("effectiveMaxStrikes(unregistered) = %d, want 2", got)
	}

	// Fallback to global when override is 0
	settings.MaxStrikesUnregistered = 0
	if got := effectiveMaxStrikes("unregistered", settings); got != 5 {
		t.Errorf("effectiveMaxStrikes(unregistered) = %d, want 5 (global fallback)", got)
	}
}
