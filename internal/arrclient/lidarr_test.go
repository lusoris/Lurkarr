package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLidarrGetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/wanted/missing" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []LidarrAlbum{{ID: 10, Title: "Missing Album", Monitored: true}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	albums, err := c.LidarrGetMissing(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(albums) != 1 || albums[0].Title != "Missing Album" {
		t.Errorf("got %v", albums)
	}
}

func TestLidarrGetCutoffUnmet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []LidarrAlbum{{ID: 20, Title: "Upgrade Album"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	albums, err := c.LidarrGetCutoffUnmet(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(albums) != 1 {
		t.Fatalf("got %d, want 1", len(albums))
	}
}

func TestLidarrSearchAlbum(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("method = %s", r.Method)
		}
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "AlbumSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.LidarrSearchAlbum(context.Background(), []int{10, 20})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "AlbumSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestLidarrGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueueResponse{
			TotalRecords: 2,
			Records:      []QueueRecord{{ID: 1}, {ID: 2}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	queue, err := c.LidarrGetQueue(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 2 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestLidarrTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/system/status" {
			t.Errorf("path = %s, want /api/v1/system/status", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Lidarr", Version: "2.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.LidarrTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Lidarr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}
