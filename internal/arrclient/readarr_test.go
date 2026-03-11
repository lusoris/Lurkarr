package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestReadarrGetMissing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/wanted/missing" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []ReadarrBook{{ID: 5, Title: "Missing Book", Monitored: true}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	books, err := c.ReadarrGetMissing(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(books) != 1 || books[0].Title != "Missing Book" {
		t.Errorf("got %v", books)
	}
}

func TestReadarrGetCutoffUnmet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []ReadarrBook{{ID: 6, Title: "Upgrade Book"}},
		})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	books, err := c.ReadarrGetCutoffUnmet(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(books) != 1 {
		t.Fatalf("got %d, want 1", len(books))
	}
}

func TestReadarrSearchBook(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(CommandResponse{ID: 1, Name: "BookSearch", Status: "queued"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	resp, err := c.ReadarrSearchBook(context.Background(), []int{5})
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if resp.Name != "BookSearch" {
		t.Errorf("Name = %q", resp.Name)
	}
}

func TestReadarrGetQueue(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(QueueResponse{TotalRecords: 1, Records: []QueueRecord{{ID: 1}}})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	queue, err := c.ReadarrGetQueue(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestReadarrTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/system/status" {
			t.Errorf("path = %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Readarr", Version: "1.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.ReadarrTestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if status.AppName != "Readarr" {
		t.Errorf("AppName = %q", status.AppName)
	}
}
