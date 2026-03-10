package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSonarrGetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/wanted/missing" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 2,
			"records": []SonarrEpisode{
				{ID: 1, Title: "Episode 1", SeriesID: 100},
				{ID: 2, Title: "Episode 2", SeriesID: 101},
			},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	episodes, err := c.SonarrGetMissing(context.Background())
	if err != nil {
		t.Fatalf("SonarrGetMissing error: %v", err)
	}
	if len(episodes) != 2 {
		t.Fatalf("got %d episodes, want 2", len(episodes))
	}
	if episodes[0].Title != "Episode 1" {
		t.Errorf("episodes[0].Title = %q", episodes[0].Title)
	}
}

func TestSonarrGetCutoffUnmet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []SonarrEpisode{{ID: 5, Title: "Upgrade Me"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	episodes, err := c.SonarrGetCutoffUnmet(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("got %d, want 1", len(episodes))
	}
}

func TestSonarrSearchEpisode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s, want POST", r.Method)
		}
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "EpisodeSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.SonarrSearchEpisode(context.Background(), []int{1, 2, 3})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "EpisodeSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestSonarrTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Sonarr", Version: "4.0.2"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.SonarrTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Sonarr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}

func TestSonarrGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueueResponse{
			TotalRecords: 1,
			Records:      []QueueRecord{{ID: 1, Title: "Downloading", Status: "downloading"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	queue, err := c.SonarrGetQueue(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}
