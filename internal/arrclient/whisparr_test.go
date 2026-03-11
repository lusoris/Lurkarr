package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestWhisparrGetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/movie" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode([]WhisparrMovie{
			{ID: 1, Title: "Has File", Monitored: true, HasFile: true},
			{ID: 2, Title: "Missing", Monitored: true, HasFile: false},
			{ID: 3, Title: "Unmonitored Missing", Monitored: false, HasFile: false},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	movies, err := c.WhisparrGetMissing(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	// Only monitored without file should be returned
	if len(movies) != 1 || movies[0].Title != "Missing" {
		t.Errorf("got %v, want 1 missing movie", movies)
	}
}

func TestWhisparrGetCutoffUnmet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []WhisparrMovie{{ID: 5, Title: "Upgrade"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	movies, err := c.WhisparrGetCutoffUnmet(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(movies) != 1 {
		t.Fatalf("got %d, want 1", len(movies))
	}
}

func TestWhisparrSearchMovie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "MoviesSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.WhisparrSearchMovie(context.Background(), []int{2})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "MoviesSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestWhisparrGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueueResponse{TotalRecords: 0, Records: []QueueRecord{}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	queue, err := c.WhisparrGetQueue(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 0 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestWhisparrTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Whisparr", Version: "2.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.WhisparrTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Whisparr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}
