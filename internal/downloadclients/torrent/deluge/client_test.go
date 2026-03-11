package deluge

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestServer creates a test server that handles auth and returns the client.
func newTestServer(t *testing.T, handler http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		if req.Method == "auth.login" {
			// Check password from params.
			params, _ := json.Marshal(req.Params)
			var args []interface{}
			json.Unmarshal(params, &args)
			if len(args) > 0 && args[0] == "secret" {
				json.NewEncoder(w).Encode(jsonRPCResponse{ID: req.ID, Result: json.RawMessage("true")})
			} else {
				json.NewEncoder(w).Encode(jsonRPCResponse{ID: req.ID, Result: json.RawMessage("false")})
			}
			return
		}
		// For non-auth methods, delegate to the handler after wrapping request in context.
		// Store the parsed RPC request in a header for the handler to inspect.
		r.Header.Set("X-RPC-Method", req.Method)
		paramsJSON, _ := json.Marshal(req.Params)
		r.Header.Set("X-RPC-Params", string(paramsJSON))
		r.Header.Set("X-RPC-ID", json.Number(json.Number(string(rune(req.ID)))).String())
		handler(w, r)
	})
	server := httptest.NewServer(mux)
	client := NewClient(server.URL, "secret", 5*time.Second)
	return server, client
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8112/", "deluge", 30*time.Second)
	if c.BaseURL != "http://localhost:8112" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", c.BaseURL)
	}
	if c.Password != "deluge" {
		t.Errorf("Password = %q", c.Password)
	}
	if c.HTTPClient.Jar == nil {
		t.Error("expected cookie jar")
	}
}

func TestLogin_Success(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {})
	defer server.Close()

	err := client.ensureAuth(context.Background())
	if err != nil {
		t.Fatalf("login error: %v", err)
	}
	if !client.authenticated {
		t.Error("expected authenticated = true")
	}
}

func TestLogin_BadPassword(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/json", func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(jsonRPCResponse{ID: req.ID, Result: json.RawMessage("false")})
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(server.URL, "wrong", 5*time.Second)
	err := client.ensureAuth(context.Background())
	if err == nil {
		t.Fatal("expected error for bad password")
	}
}

func TestGetTorrents(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-RPC-Method") != "core.get_torrents_status" {
			t.Errorf("method = %q", r.Header.Get("X-RPC-Method"))
		}
		result := map[string]interface{}{
			"abc123": map[string]interface{}{
				"name":       "test.torrent",
				"state":      "Downloading",
				"total_size": 1024,
				"progress":   50.0,
			},
			"def456": map[string]interface{}{
				"name":       "done.torrent",
				"state":      "Seeding",
				"total_size": 2048,
				"progress":   100.0,
			},
		}
		data, _ := json.Marshal(result)
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage(data)})
	})
	defer server.Close()

	torrents, err := client.GetTorrents(context.Background())
	if err != nil {
		t.Fatalf("GetTorrents error: %v", err)
	}
	if len(torrents) != 2 {
		t.Fatalf("got %d torrents, want 2", len(torrents))
	}
}

func TestPauseTorrents(t *testing.T) {
	var gotMethod string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Header.Get("X-RPC-Method")
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage("null")})
	})
	defer server.Close()

	err := client.PauseTorrents(context.Background(), []string{"abc123"})
	if err != nil {
		t.Fatalf("PauseTorrents error: %v", err)
	}
	if gotMethod != "core.pause_torrents" {
		t.Errorf("method = %q, want core.pause_torrents", gotMethod)
	}
}

func TestResumeTorrents(t *testing.T) {
	var gotMethod string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Header.Get("X-RPC-Method")
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage("null")})
	})
	defer server.Close()

	err := client.ResumeTorrents(context.Background(), []string{"abc123"})
	if err != nil {
		t.Fatalf("ResumeTorrents error: %v", err)
	}
	if gotMethod != "core.resume_torrents" {
		t.Errorf("method = %q, want core.resume_torrents", gotMethod)
	}
}

func TestDeleteTorrents(t *testing.T) {
	var gotMethod string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Header.Get("X-RPC-Method")
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage("true")})
	})
	defer server.Close()

	err := client.DeleteTorrents(context.Background(), []string{"abc123"}, true)
	if err != nil {
		t.Fatalf("DeleteTorrents error: %v", err)
	}
	if gotMethod != "core.remove_torrent" {
		t.Errorf("method = %q, want core.remove_torrent", gotMethod)
	}
}

func TestGetVersion(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-RPC-Method") != "daemon.info" {
			t.Errorf("method = %q, want daemon.info", r.Header.Get("X-RPC-Method"))
		}
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage(`"2.1.1"`)})
	})
	defer server.Close()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion error: %v", err)
	}
	if ver != "2.1.1" {
		t.Errorf("version = %q", ver)
	}
}

func TestTestConnection(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage(`"2.1.1"`)})
	})
	defer server.Close()

	ver, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "2.1.1" {
		t.Errorf("version = %q", ver)
	}
}

func TestAddTorrentByURL(t *testing.T) {
	var gotMethod string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Header.Get("X-RPC-Method")
		json.NewEncoder(w).Encode(jsonRPCResponse{Result: json.RawMessage(`"abc123"`)})
	})
	defer server.Close()

	err := client.AddTorrentByURL(context.Background(), "magnet:?xt=urn:btih:abc", nil)
	if err != nil {
		t.Fatalf("AddTorrentByURL error: %v", err)
	}
	if gotMethod != "core.add_torrent_url" {
		t.Errorf("method = %q, want core.add_torrent_url", gotMethod)
	}
}

func TestRPCError(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(jsonRPCResponse{
			Error: &jsonRPCError{Message: "something broke", Code: -1},
		})
	})
	defer server.Close()

	_, err := client.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected error for RPC error response")
	}
}

func TestHTTPError(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	defer server.Close()

	// Pre-authenticate to skip login.
	client.mu.Lock()
	client.authenticated = true
	client.mu.Unlock()

	_, err := client.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
