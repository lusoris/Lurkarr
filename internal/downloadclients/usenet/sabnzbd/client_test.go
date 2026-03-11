package sabnzbd

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(handler http.HandlerFunc) (*httptest.Server, *Client) {
	server := httptest.NewServer(handler)
	client := NewClient(server.URL, "testkey", 5*time.Second)
	return server, client
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080/", "mykey", 30*time.Second)
	if c.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", c.BaseURL)
	}
	if c.APIKey != "mykey" {
		t.Errorf("APIKey = %q", c.APIKey)
	}
}

func TestGetQueue(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mode") != "queue" {
			t.Errorf("mode = %s, want queue", r.URL.Query().Get("mode"))
		}
		if r.URL.Query().Get("apikey") != "testkey" {
			t.Error("missing apikey param")
		}
		json.NewEncoder(w).Encode(map[string]any{
			"queue": Queue{
				Status:    "Downloading",
				Paused:    false,
				NoOfSlots: 2,
				Slots: []QueueSlot{
					{NzoID: "nzo1", Filename: "file1.nzb", Status: "Downloading"},
					{NzoID: "nzo2", Filename: "file2.nzb", Status: "Queued"},
				},
			},
		})
	})
	defer server.Close()

	queue, err := client.GetQueue(context.Background())
	if err != nil {
		t.Fatalf("GetQueue error: %v", err)
	}
	if queue.Status != "Downloading" {
		t.Errorf("Status = %q", queue.Status)
	}
	if len(queue.Slots) != 2 {
		t.Fatalf("Slots len = %d, want 2", len(queue.Slots))
	}
}

func TestGetHistory(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mode") != "history" {
			t.Errorf("mode = %s", r.URL.Query().Get("mode"))
		}
		json.NewEncoder(w).Encode(map[string]any{
			"history": History{
				NoOfSlots: 1,
				Slots: []HistorySlot{
					{NzoID: "h1", Name: "completed.nzb", Status: "Completed"},
				},
			},
		})
	})
	defer server.Close()

	hist, err := client.GetHistory(context.Background(), 10)
	if err != nil {
		t.Fatalf("GetHistory error: %v", err)
	}
	if hist.NoOfSlots != 1 {
		t.Errorf("NoOfSlots = %d", hist.NoOfSlots)
	}
}

func TestGetServerStats(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ServerStats{Total: 1024, Day: 100, Week: 500, Month: 900})
	})
	defer server.Close()

	stats, err := client.GetServerStats(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if stats.Total != 1024 {
		t.Errorf("Total = %d", stats.Total)
	}
}

func TestPauseResume(t *testing.T) {
	var lastMode string
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		lastMode = r.URL.Query().Get("mode")
		json.NewEncoder(w).Encode(map[string]bool{"status": true})
	})
	defer server.Close()

	if err := client.Pause(context.Background()); err != nil {
		t.Fatalf("Pause error: %v", err)
	}
	if lastMode != "pause" {
		t.Errorf("last mode = %q, want pause", lastMode)
	}

	if err := client.Resume(context.Background()); err != nil {
		t.Fatalf("Resume error: %v", err)
	}
	if lastMode != "resume" {
		t.Errorf("last mode = %q, want resume", lastMode)
	}
}

func TestGetVersion(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mode") != "version" {
			t.Error("wrong mode")
		}
		json.NewEncoder(w).Encode("4.2.1")
	})
	defer server.Close()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "4.2.1" {
		t.Errorf("version = %q", ver)
	}
}

func TestTestConnection(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode("4.2.1")
	})
	defer server.Close()

	ver, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "4.2.1" {
		t.Errorf("version = %q", ver)
	}
}

func TestAPICallServerError(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.GetQueue(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestDeleteQueueItem(t *testing.T) {
	server, client := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("mode") != "queue" {
			t.Errorf("mode = %s, want queue", r.URL.Query().Get("mode"))
		}
		if r.URL.Query().Get("name") != "delete" {
			t.Errorf("name = %s, want delete", r.URL.Query().Get("name"))
		}
		if r.URL.Query().Get("value") != "SABnzbd_nzo_abc123" {
			t.Errorf("value = %s, want SABnzbd_nzo_abc123", r.URL.Query().Get("value"))
		}
		json.NewEncoder(w).Encode(map[string]bool{"status": true})
	})
	defer server.Close()

	if err := client.DeleteQueueItem(context.Background(), "SABnzbd_nzo_abc123"); err != nil {
		t.Fatalf("DeleteQueueItem error: %v", err)
	}
}
