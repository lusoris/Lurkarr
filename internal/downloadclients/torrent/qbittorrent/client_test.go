package qbittorrent

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
	mux.HandleFunc("/api/v2/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad form", http.StatusBadRequest)
			return
		}
		if r.FormValue("username") == "admin" && r.FormValue("password") == "secret" {
			http.SetCookie(w, &http.Cookie{Name: "SID", Value: "test-session-id"})
			w.Write([]byte("Ok."))
		} else {
			w.Write([]byte("Fails."))
		}
	})
	mux.HandleFunc("/", handler)
	server := httptest.NewServer(mux)
	client := NewClient(server.URL, "admin", "secret", 5*time.Second)
	return server, client
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8080/", "admin", "pass", 30*time.Second)
	if c.BaseURL != "http://localhost:8080" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", c.BaseURL)
	}
	if c.Username != "admin" {
		t.Errorf("Username = %q", c.Username)
	}
	if c.Password != "pass" {
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

func TestLogin_BadCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/auth/login" {
			w.Write([]byte("Fails."))
			return
		}
	}))
	defer server.Close()

	client := NewClient(server.URL, "wrong", "creds", 5*time.Second)
	err := client.ensureAuth(context.Background())
	if err == nil {
		t.Fatal("expected error for bad credentials")
	}
}

func TestGetTorrents(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/info" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		torrents := []Torrent{
			{Hash: "abc123", Name: "test.torrent", Size: 1024, State: "downloading", Progress: 0.5},
			{Hash: "def456", Name: "done.torrent", Size: 2048, State: "uploading", Progress: 1.0},
		}
		json.NewEncoder(w).Encode(torrents)
	})
	defer server.Close()

	torrents, err := client.GetTorrents(context.Background(), "", "")
	if err != nil {
		t.Fatalf("GetTorrents error: %v", err)
	}
	if len(torrents) != 2 {
		t.Fatalf("got %d torrents, want 2", len(torrents))
	}
	if torrents[0].Hash != "abc123" {
		t.Errorf("Hash = %q", torrents[0].Hash)
	}
	if torrents[0].Name != "test.torrent" {
		t.Errorf("Name = %q", torrents[0].Name)
	}
}

func TestGetTorrents_WithFilter(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/info" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.URL.Query().Get("filter") != "downloading" {
			t.Errorf("filter = %q, want downloading", r.URL.Query().Get("filter"))
		}
		if r.URL.Query().Get("category") != "movies" {
			t.Errorf("category = %q, want movies", r.URL.Query().Get("category"))
		}
		json.NewEncoder(w).Encode([]Torrent{})
	})
	defer server.Close()

	_, err := client.GetTorrents(context.Background(), "downloading", "movies")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestPauseTorrents(t *testing.T) {
	var gotHashes string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/pause" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		r.ParseForm()
		gotHashes = r.FormValue("hashes")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.PauseTorrents(context.Background(), []string{"abc", "def"})
	if err != nil {
		t.Fatalf("PauseTorrents error: %v", err)
	}
	if gotHashes != "abc|def" {
		t.Errorf("hashes = %q, want abc|def", gotHashes)
	}
}

func TestResumeTorrents(t *testing.T) {
	var gotHashes string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/resume" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		r.ParseForm()
		gotHashes = r.FormValue("hashes")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.ResumeTorrents(context.Background(), []string{"all"})
	if err != nil {
		t.Fatalf("ResumeTorrents error: %v", err)
	}
	if gotHashes != "all" {
		t.Errorf("hashes = %q, want all", gotHashes)
	}
}

func TestDeleteTorrents(t *testing.T) {
	var gotHashes, gotDelete string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/delete" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		r.ParseForm()
		gotHashes = r.FormValue("hashes")
		gotDelete = r.FormValue("deleteFiles")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.DeleteTorrents(context.Background(), []string{"abc123"}, true)
	if err != nil {
		t.Fatalf("DeleteTorrents error: %v", err)
	}
	if gotHashes != "abc123" {
		t.Errorf("hashes = %q", gotHashes)
	}
	if gotDelete != "true" {
		t.Errorf("deleteFiles = %q, want true", gotDelete)
	}
}

func TestDeleteTorrents_NoDeleteFiles(t *testing.T) {
	var gotDelete string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/delete" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		r.ParseForm()
		gotDelete = r.FormValue("deleteFiles")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.DeleteTorrents(context.Background(), []string{"abc123"}, false)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if gotDelete != "false" {
		t.Errorf("deleteFiles = %q, want false", gotDelete)
	}
}

func TestGetTransferInfo(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/transfer/info" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		json.NewEncoder(w).Encode(TransferInfo{
			DownloadSpeed:    1048576,
			UploadSpeed:      524288,
			DownloadTotal:    10737418240,
			UploadTotal:      5368709120,
			ConnectionStatus: "connected",
		})
	})
	defer server.Close()

	info, err := client.GetTransferInfo(context.Background())
	if err != nil {
		t.Fatalf("GetTransferInfo error: %v", err)
	}
	if info.DownloadSpeed != 1048576 {
		t.Errorf("DownloadSpeed = %d", info.DownloadSpeed)
	}
	if info.ConnectionStatus != "connected" {
		t.Errorf("ConnectionStatus = %q", info.ConnectionStatus)
	}
}

func TestGetVersion(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/app/version" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		w.Write([]byte("v4.6.3"))
	})
	defer server.Close()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion error: %v", err)
	}
	if ver != "v4.6.3" {
		t.Errorf("version = %q", ver)
	}
}

func TestTestConnection(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/app/version" {
			w.Write([]byte("v4.6.3"))
			return
		}
	})
	defer server.Close()

	ver, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "v4.6.3" {
		t.Errorf("version = %q", ver)
	}
}

func TestAddTorrentByURL(t *testing.T) {
	var gotURL, gotCategory, gotSavePath string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/add" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		r.ParseForm()
		gotURL = r.FormValue("urls")
		gotCategory = r.FormValue("category")
		gotSavePath = r.FormValue("savepath")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.AddTorrentByURL(context.Background(), "magnet:?xt=urn:btih:abc", "movies", "/downloads/movies")
	if err != nil {
		t.Fatalf("AddTorrentByURL error: %v", err)
	}
	if gotURL != "magnet:?xt=urn:btih:abc" {
		t.Errorf("urls = %q", gotURL)
	}
	if gotCategory != "movies" {
		t.Errorf("category = %q", gotCategory)
	}
	if gotSavePath != "/downloads/movies" {
		t.Errorf("savepath = %q", gotSavePath)
	}
}

func TestAddTorrentByURL_NoOptional(t *testing.T) {
	var gotCategory, gotSavePath string
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/add" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		r.ParseForm()
		gotCategory = r.FormValue("category")
		gotSavePath = r.FormValue("savepath")
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	err := client.AddTorrentByURL(context.Background(), "magnet:?xt=urn:btih:abc", "", "")
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if gotCategory != "" {
		t.Errorf("expected no category, got %q", gotCategory)
	}
	if gotSavePath != "" {
		t.Errorf("expected no savepath, got %q", gotSavePath)
	}
}

func TestSessionExpiry_Reauth(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/app/version" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		callCount++
		if callCount == 1 {
			// First call: simulate expired session.
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.Write([]byte("v4.6.3"))
	})
	defer server.Close()

	// Pre-authenticate to simulate an existing (expired) session.
	client.mu.Lock()
	client.authenticated = true
	client.mu.Unlock()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion error after reauth: %v", err)
	}
	if ver != "v4.6.3" {
		t.Errorf("version = %q", ver)
	}
}

func TestPostSessionExpiry_Reauth(t *testing.T) {
	callCount := 0
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v2/torrents/pause" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		callCount++
		if callCount == 1 {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client.mu.Lock()
	client.authenticated = true
	client.mu.Unlock()

	err := client.PauseTorrents(context.Background(), []string{"abc"})
	if err != nil {
		t.Fatalf("PauseTorrents error after reauth: %v", err)
	}
}

func TestPostServerError(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v2/torrents/resume" {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
	})
	defer server.Close()

	err := client.ResumeTorrents(context.Background(), []string{"abc"})
	if err == nil {
		t.Fatal("expected error for 500 response on POST")
	}
}

func TestAPIServerError(t *testing.T) {
	server, client := newTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	})
	defer server.Close()

	_, err := client.GetTorrents(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
