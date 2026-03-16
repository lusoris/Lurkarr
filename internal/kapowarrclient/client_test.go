package kapowarrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(t *testing.T, handlers map[string]any) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	for path, resp := range handlers {
		mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("api_key") != "testkey" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		})
	}
	return httptest.NewServer(mux)
}

func TestTestConnection(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/v1/system/about": APIResponse[AboutInfo]{Result: AboutInfo{Version: "1.0.0"}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	info, err := c.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info.Version != "1.0.0" {
		t.Fatalf("expected version 1.0.0, got %s", info.Version)
	}
}

func TestTestConnection_APIError(t *testing.T) {
	errMsg := "unauthorized"
	srv := newTestServer(t, map[string]any{
		"/api/v1/system/about": APIResponse[AboutInfo]{Error: &errMsg},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	_, err := c.TestConnection(context.Background())
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestGetVolumeStats(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/v1/volumes/stats": APIResponse[VolumeStats]{Result: VolumeStats{Total: 42, Monitored: 30}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	stats, err := c.GetVolumeStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 42 {
		t.Fatalf("expected total=42, got %d", stats.Total)
	}
}

func TestGetQueue(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/v1/activity/queue": APIResponse[[]QueueItem]{Result: []QueueItem{{ID: 1, Status: "downloading"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	items, err := c.GetQueue(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].ID != 1 {
		t.Fatalf("unexpected queue items: %+v", items)
	}
}

func TestGetTasks(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/v1/system/tasks": APIResponse[[]TaskInfo]{Result: []TaskInfo{{ID: 1, Action: "scan"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	tasks, err := c.GetTasks(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 1 || tasks[0].Action != "scan" {
		t.Fatalf("unexpected tasks: %+v", tasks)
	}
}

func TestHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	_, err := c.TestConnection(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestUnreachableServer(t *testing.T) {
	c := NewClient("http://127.0.0.1:1", "testkey", 1*time.Second, true)
	_, err := c.TestConnection(context.Background())
	if err == nil {
		t.Fatal("expected error for unreachable server")
	}
}
