package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestErosGetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/movie" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]ErosMovie{
			{ID: 1, Title: "Has File", Monitored: true, HasFile: true},
			{ID: 2, Title: "Missing Eros", Monitored: true, HasFile: false},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	movies, err := c.ErosGetMissing(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(movies) != 1 || movies[0].Title != "Missing Eros" {
		t.Errorf("got %v", movies)
	}
}

func TestErosGetCutoffUnmet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []ErosMovie{{ID: 3, Title: "Upgrade Eros"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	movies, err := c.ErosGetCutoffUnmet(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(movies) != 1 {
		t.Fatalf("got %d, want 1", len(movies))
	}
}

func TestErosSearchMovie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "MoviesSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.ErosSearchMovie(context.Background(), []int{2})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "MoviesSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestErosGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueueResponse{TotalRecords: 3, Records: []QueueRecord{{ID: 1}, {ID: 2}, {ID: 3}}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	queue, err := c.ErosGetQueue(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestErosTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Whisparr", Version: "3.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.ErosTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Whisparr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}
