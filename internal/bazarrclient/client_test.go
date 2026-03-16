package bazarrclient

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
			if got := r.Header.Get("X-API-Key"); got != "testkey" {
				t.Errorf("expected X-API-Key=testkey, got %q", got)
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
		"/api/system/status": SystemStatus{Version: "1.4.5", StartTime: "2024-01-01T00:00:00Z"},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	status, err := c.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status.Version != "1.4.5" {
		t.Fatalf("expected version 1.4.5, got %s", status.Version)
	}
}

func TestGetHealth(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/system/health": struct {
			Data []HealthItem `json:"data"`
		}{Data: []HealthItem{{Object: "test", Issue: "warning"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	items, err := c.GetHealth(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(items) != 1 || items[0].Object != "test" {
		t.Fatalf("unexpected items: %+v", items)
	}
}

func TestGetWantedEpisodes(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/episodes/wanted": WantedResponse{Total: 5, Data: []WantedItem{{Title: "ep1"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	resp, err := c.GetWantedEpisodes(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 5 {
		t.Fatalf("expected total=5, got %d", resp.Total)
	}
}

func TestGetWantedMovies(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/movies/wanted": WantedResponse{Total: 3, Data: []WantedItem{{Title: "movie1"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	resp, err := c.GetWantedMovies(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 3 {
		t.Fatalf("expected total=3, got %d", resp.Total)
	}
}

func TestGetEpisodeHistory(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/history/episodes": HistoryResponse{Total: 2, Data: []HistoryItem{{Title: "ep"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	resp, err := c.GetEpisodeHistory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 2 {
		t.Fatalf("expected total=2, got %d", resp.Total)
	}
}

func TestGetMovieHistory(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/history/movies": HistoryResponse{Total: 1, Data: []HistoryItem{{Title: "mov"}}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	resp, err := c.GetMovieHistory(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total=1, got %d", resp.Total)
	}
}

func TestAPIError(t *testing.T) {
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
