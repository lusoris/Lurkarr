package transmission

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestServer creates a test server that handles CSRF and returns the client.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate CSRF: if no session header, return 409 with the token.
		if r.Header.Get(csrfHeader) == "" {
			w.Header().Set(csrfHeader, "test-csrf-token")
			w.WriteHeader(http.StatusConflict)
			return
		}
		handler(w, r)
	}))
	client := NewClient(server.URL, "admin", "secret", 5*time.Second)
	return server, client
}

// newTestServerNoCSRF creates a test server without CSRF enforcement.
func newTestServerNoCSRF(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(handler)
	client := NewClient(server.URL, "admin", "secret", 5*time.Second)
	// Pre-set CSRF token to skip initial 409 handshake.
	client.csrfToken = "pre-set-token"
	return server, client
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:9091/", "admin", "pass", 30*time.Second)
	if c.BaseURL != "http://localhost:9091" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", c.BaseURL)
	}
	if c.Username != "admin" {
		t.Errorf("Username = %q", c.Username)
	}
}

func TestCSRFHandshake(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := rpcResponse{
			Result:    "success",
			Arguments: mustJSON(t, map[string]string{"version": "4.0.5"}),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "4.0.5" {
		t.Errorf("version = %q", ver)
	}
	if callCount != 1 {
		t.Errorf("expected 1 successful call after CSRF handshake, got %d", callCount)
	}
	if client.csrfToken != "test-csrf-token" {
		t.Errorf("CSRF token = %q", client.csrfToken)
	}
}

func TestGetTorrents(t *testing.T) {
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Method != "torrent-get" {
			t.Errorf("method = %q, want torrent-get", req.Method)
		}
		resp := rpcResponse{
			Result: "success",
			Arguments: mustJSON(t, map[string]interface{}{
				"torrents": []Torrent{
					{ID: 1, Name: "test.torrent", Status: StatusDownloading, PercentDone: 0.5},
					{ID: 2, Name: "done.torrent", Status: StatusSeeding, PercentDone: 1.0},
				},
			}),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	torrents, err := client.GetTorrents(context.Background())
	if err != nil {
		t.Fatalf("GetTorrents error: %v", err)
	}
	if len(torrents) != 2 {
		t.Fatalf("got %d torrents, want 2", len(torrents))
	}
	if torrents[0].Name != "test.torrent" {
		t.Errorf("Name = %q", torrents[0].Name)
	}
	if torrents[1].Status != StatusSeeding {
		t.Errorf("Status = %d, want %d", torrents[1].Status, StatusSeeding)
	}
}

func TestPauseTorrents(t *testing.T) {
	var gotMethod string
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)
		gotMethod = req.Method
		json.NewEncoder(w).Encode(rpcResponse{Result: "success"})
	})
	defer server.Close()

	err := client.PauseTorrents(context.Background(), []int{1, 2})
	if err != nil {
		t.Fatalf("PauseTorrents error: %v", err)
	}
	if gotMethod != "torrent-stop" {
		t.Errorf("method = %q, want torrent-stop", gotMethod)
	}
}

func TestResumeTorrents(t *testing.T) {
	var gotMethod string
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		var req rpcRequest
		json.NewDecoder(r.Body).Decode(&req)
		gotMethod = req.Method
		json.NewEncoder(w).Encode(rpcResponse{Result: "success"})
	})
	defer server.Close()

	err := client.ResumeTorrents(context.Background(), []int{1, 2})
	if err != nil {
		t.Fatalf("ResumeTorrents error: %v", err)
	}
	if gotMethod != "torrent-start" {
		t.Errorf("method = %q, want torrent-start", gotMethod)
	}
}

func TestDeleteTorrents(t *testing.T) {
	var gotMethod string
	var gotArgs map[string]interface{}
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method    string                 `json:"method"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		gotMethod = req.Method
		gotArgs = req.Arguments
		json.NewEncoder(w).Encode(rpcResponse{Result: "success"})
	})
	defer server.Close()

	err := client.DeleteTorrents(context.Background(), []int{1}, true)
	if err != nil {
		t.Fatalf("DeleteTorrents error: %v", err)
	}
	if gotMethod != "torrent-remove" {
		t.Errorf("method = %q, want torrent-remove", gotMethod)
	}
	if del, ok := gotArgs["delete-local-data"].(bool); !ok || !del {
		t.Errorf("delete-local-data = %v, want true", gotArgs["delete-local-data"])
	}
}

func TestGetSessionStats(t *testing.T) {
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			Result: "success",
			Arguments: mustJSON(t, SessionStats{
				DownloadSpeed: 1048576,
				UploadSpeed:   524288,
				TorrentCount:  10,
				ActiveCount:   3,
				PausedCount:   7,
			}),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	stats, err := client.GetSessionStats(context.Background())
	if err != nil {
		t.Fatalf("GetSessionStats error: %v", err)
	}
	if stats.DownloadSpeed != 1048576 {
		t.Errorf("DownloadSpeed = %d", stats.DownloadSpeed)
	}
	if stats.TorrentCount != 10 {
		t.Errorf("TorrentCount = %d", stats.TorrentCount)
	}
}

func TestGetVersion(t *testing.T) {
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			Result:    "success",
			Arguments: mustJSON(t, map[string]string{"version": "4.0.5"}),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion error: %v", err)
	}
	if ver != "4.0.5" {
		t.Errorf("version = %q", ver)
	}
}

func TestTestConnection(t *testing.T) {
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		resp := rpcResponse{
			Result:    "success",
			Arguments: mustJSON(t, map[string]string{"version": "4.0.5"}),
		}
		json.NewEncoder(w).Encode(resp)
	})
	defer server.Close()

	ver, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "4.0.5" {
		t.Errorf("version = %q", ver)
	}
}

func TestAddTorrentByURL(t *testing.T) {
	var gotMethod string
	var gotArgs map[string]interface{}
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Method    string                 `json:"method"`
			Arguments map[string]interface{} `json:"arguments"`
		}
		json.NewDecoder(r.Body).Decode(&req)
		gotMethod = req.Method
		gotArgs = req.Arguments
		json.NewEncoder(w).Encode(rpcResponse{Result: "success"})
	})
	defer server.Close()

	err := client.AddTorrentByURL(context.Background(), "magnet:?xt=urn:btih:abc", "/downloads")
	if err != nil {
		t.Fatalf("AddTorrentByURL error: %v", err)
	}
	if gotMethod != "torrent-add" {
		t.Errorf("method = %q, want torrent-add", gotMethod)
	}
	if gotArgs["filename"] != "magnet:?xt=urn:btih:abc" {
		t.Errorf("filename = %v", gotArgs["filename"])
	}
	if gotArgs["download-dir"] != "/downloads" {
		t.Errorf("download-dir = %v", gotArgs["download-dir"])
	}
}

func TestRPCError(t *testing.T) {
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(rpcResponse{Result: "no such method"})
	})
	defer server.Close()

	_, err := client.GetTorrents(context.Background())
	if err == nil {
		t.Fatal("expected error for failed RPC result")
	}
}

func TestAuthFailure(t *testing.T) {
	server, client := newTestServerNoCSRF(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
	defer server.Close()

	_, err := client.GetTorrents(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func mustJSON(t *testing.T, v interface{}) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("mustJSON: %v", err)
	}
	return json.RawMessage(data)
}
