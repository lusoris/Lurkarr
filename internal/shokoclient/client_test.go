package shokoclient

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
			if got := r.Header.Get("apikey"); got != "testkey" {
				t.Errorf("expected apikey=testkey, got %q", got)
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
		"/api/v3/Init/Version": VersionInfo{Server: struct {
			Version string `json:"Version"`
			Tag     string `json:"Tag,omitempty"`
			Commit  string `json:"Commit,omitempty"`
		}{Version: "4.2.0"}},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	v, err := c.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Server.Version != "4.2.0" {
		t.Fatalf("expected version 4.2.0, got %s", v.Server.Version)
	}
}

func TestGetStats(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/v3/Dashboard/Stats": CollectionStats{FileCount: 100, SeriesCount: 20},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	stats, err := c.GetStats(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.FileCount != 100 {
		t.Fatalf("expected FileCount=100, got %d", stats.FileCount)
	}
	if stats.SeriesCount != 20 {
		t.Fatalf("expected SeriesCount=20, got %d", stats.SeriesCount)
	}
}

func TestGetSeriesSummary(t *testing.T) {
	srv := newTestServer(t, map[string]any{
		"/api/v3/Dashboard/SeriesSummary": SeriesSummary{Series: 50, Movie: 10, OVA: 5},
	})
	defer srv.Close()

	c := NewClient(srv.URL, "testkey", 5*time.Second, true)
	summary, err := c.GetSeriesSummary(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if summary.Series != 50 {
		t.Fatalf("expected Series=50, got %d", summary.Series)
	}
	if summary.Movie != 10 {
		t.Fatalf("expected Movie=10, got %d", summary.Movie)
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
