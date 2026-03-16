package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestProwlarrGetIndexers(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/indexer" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]ProwlarrIndexer{
			{ID: 1, Name: "NZBgeek", Protocol: "usenet", Enable: true, Priority: 25},
			{ID: 2, Name: "1337x", Protocol: "torrent", Enable: true, Priority: 50},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	indexers, err := c.ProwlarrGetIndexers(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(indexers) != 2 {
		t.Fatalf("got %d indexers, want 2", len(indexers))
	}
	if indexers[0].Name != "NZBgeek" {
		t.Errorf("indexers[0].Name = %q", indexers[0].Name)
	}
}

func TestProwlarrGetIndexerStats(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/indexerstats" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"indexers": []ProwlarrIndexerStats{
				{IndexerID: 1, NumberOfQueries: 100, NumberOfGrabs: 50, NumberOfFailures: 2},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	stats, err := c.ProwlarrGetIndexerStats(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(stats) != 1 {
		t.Fatalf("got %d stats, want 1", len(stats))
	}
	if stats[0].NumberOfQueries != 100 {
		t.Errorf("NumberOfQueries = %d", stats[0].NumberOfQueries)
	}
}

func TestProwlarrTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/system/status" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Prowlarr", Version: "1.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.ProwlarrTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Prowlarr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}

func TestDeleteQueueItem(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		if r.URL.Path != "/api/v3/queue/42" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("removeFromClient") != "true" {
			t.Error("expected removeFromClient=true")
		}
		if r.URL.Query().Get("blocklist") != "false" {
			t.Error("expected blocklist=false")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	err := c.DeleteQueueItem(context.Background(), "v3", 42, true, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestGetManualImport(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/manualimport" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("downloadId") != "abc123" {
			t.Errorf("downloadId = %s", r.URL.Query().Get("downloadId"))
		}
		json.NewEncoder(w).Encode([]ManualImportItem{
			{ID: 1, Path: "/downloads/movie.mkv", Name: "movie.mkv", Size: 5000000000},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	items, err := c.GetManualImport(context.Background(), "v3", "abc123")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
	if items[0].Name != "movie.mkv" {
		t.Errorf("Name = %q", items[0].Name)
	}
}

func TestQueueRecordHasImportError(t *testing.T) {
	tests := []struct {
		name   string
		record QueueRecord
		want   bool
	}{
		{
			name:   "warning status",
			record: QueueRecord{TrackedDownloadStatus: "warning"},
			want:   true,
		},
		{
			name:   "importPending state",
			record: QueueRecord{TrackedDownloadState: "importPending"},
			want:   true,
		},
		{
			name:   "ok status, downloading state",
			record: QueueRecord{TrackedDownloadStatus: "ok", TrackedDownloadState: "downloading"},
			want:   false,
		},
		{
			name:   "warning and importPending",
			record: QueueRecord{TrackedDownloadStatus: "warning", TrackedDownloadState: "importPending"},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.HasImportError(); got != tt.want {
				t.Errorf("HasImportError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueueRecordMediaID(t *testing.T) {
	tests := []struct {
		name   string
		record QueueRecord
		want   int
	}{
		{"movie", QueueRecord{MovieID: 10}, 10},
		{"episode", QueueRecord{EpisodeID: 20}, 20},
		{"album", QueueRecord{AlbumID: 30}, 30},
		{"book", QueueRecord{BookID: 40}, 40},
		{"series fallback", QueueRecord{SeriesID: 50}, 50},
		{"movie priority over series", QueueRecord{MovieID: 10, SeriesID: 50}, 10},
		{"none", QueueRecord{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.record.MediaID(); got != tt.want {
				t.Errorf("MediaID() = %d, want %d", got, tt.want)
			}
		})
	}
}

func TestClientDel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("method = %s, want DELETE", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	err := c.del(context.Background(), "/test")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestClientPostNilResult(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	err := c.post(context.Background(), "/test", nil, nil)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestNewClientSSLVerifyFalse(t *testing.T) {
	c := NewClient("https://localhost:8989", "key", 30*time.Second, false)
	if c.BaseURL != "https://localhost:8989" {
		t.Errorf("BaseURL = %q", c.BaseURL)
	}
}

func TestRadarrGetCutoffUnmet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/cutoff" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []RadarrMovie{{ID: 1, Title: "Upgrade Movie"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	movies, err := c.RadarrGetCutoffUnmet(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(movies) != 1 {
		t.Fatalf("got %d, want 1", len(movies))
	}
}

func TestRadarrGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueueResponse{TotalRecords: 2, Records: []QueueRecord{{ID: 1}, {ID: 2}}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	queue, err := c.RadarrGetQueue(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 2 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestSonarrSearchSeason(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "SeasonSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.SonarrSearchSeason(context.Background(), 100, 1)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "SeasonSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestSonarrSearchSeries(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "SeriesSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.SonarrSearchSeries(context.Background(), 100)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "SeriesSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestClientGetServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	var result map[string]string
	err := c.get(context.Background(), "/error", &result)
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestQueueRecordMediaHasFile(t *testing.T) {
	tests := []struct {
		name   string
		record QueueRecord
		want   bool
		wantOK bool
	}{
		{"movie has file", QueueRecord{Movie: &QueueMovie{HasFile: true}}, true, true},
		{"movie no file", QueueRecord{Movie: &QueueMovie{HasFile: false}}, false, true},
		{"episode has file", QueueRecord{Episode: &QueueEpisode{HasFile: true}}, true, true},
		{"episode no file", QueueRecord{Episode: &QueueEpisode{HasFile: false}}, false, true},
		{"album has tracks", QueueRecord{Album: &QueueAlbum{Statistics: struct {
			TrackFileCount int `json:"trackFileCount"`
		}{TrackFileCount: 3}}}, true, true},
		{"album no tracks", QueueRecord{Album: &QueueAlbum{}}, false, true},
		{"book has files", QueueRecord{Book: &QueueBook{Statistics: struct {
			BookFileCount int `json:"bookFileCount"`
		}{BookFileCount: 1}}}, true, true},
		{"book no files", QueueRecord{Book: &QueueBook{}}, false, true},
		{"no enriched data", QueueRecord{}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.record.MediaHasFile()
			if ok != tt.wantOK {
				t.Errorf("MediaHasFile() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("MediaHasFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueueRecordMediaMonitored(t *testing.T) {
	tests := []struct {
		name   string
		record QueueRecord
		want   bool
		wantOK bool
	}{
		{"movie monitored", QueueRecord{Movie: &QueueMovie{Monitored: true}}, true, true},
		{"movie unmonitored", QueueRecord{Movie: &QueueMovie{Monitored: false}}, false, true},
		{"episode monitored, series monitored", QueueRecord{
			Episode: &QueueEpisode{Monitored: true},
			Series:  &QueueSeries{Monitored: true},
		}, true, true},
		{"episode monitored, series unmonitored", QueueRecord{
			Episode: &QueueEpisode{Monitored: true},
			Series:  &QueueSeries{Monitored: false},
		}, false, true},
		{"episode unmonitored", QueueRecord{Episode: &QueueEpisode{Monitored: false}}, false, true},
		{"album monitored", QueueRecord{Album: &QueueAlbum{Monitored: true}}, true, true},
		{"album unmonitored", QueueRecord{Album: &QueueAlbum{Monitored: false}}, false, true},
		{"book monitored", QueueRecord{Book: &QueueBook{Monitored: true}}, true, true},
		{"book unmonitored", QueueRecord{Book: &QueueBook{Monitored: false}}, false, true},
		{"no enriched data", QueueRecord{}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := tt.record.MediaMonitored()
			if ok != tt.wantOK {
				t.Errorf("MediaMonitored() ok = %v, want %v", ok, tt.wantOK)
			}
			if got != tt.want {
				t.Errorf("MediaMonitored() = %v, want %v", got, tt.want)
			}
		})
	}
}
