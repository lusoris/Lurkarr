package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRadarrGetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]RadarrMovie{
			{ID: 1, Title: "Missing Movie", Monitored: true, HasFile: false},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	movies, err := c.RadarrGetMissing(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(movies) != 1 {
		t.Fatalf("got %d movies, want 1", len(movies))
	}
	if movies[0].Title != "Missing Movie" {
		t.Errorf("Title = %q", movies[0].Title)
	}
}

func TestRadarrSearchMovie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "MoviesSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.RadarrSearchMovie(context.Background(), []int{1, 2})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "MoviesSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestRadarrTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Radarr", Version: "5.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.RadarrTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Radarr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}
