package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// =============================================================================
// ProwlarrHandler success-path tests (using httptest backend server)
// =============================================================================

func TestProwlarrGetIndexers_Success(t *testing.T) {
	// Fake Prowlarr API server
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/indexer" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing or wrong API key")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "NZBgeek", "protocol": "usenet", "enable": true, "priority": 1},
			{"id": 2, "name": "TorrentLeech", "protocol": "torrent", "enable": false, "priority": 5},
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(&database.ProwlarrSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SSLVerify: true}, nil)

	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetIndexers(w, httptest.NewRequest("GET", "/api/prowlarr/indexers", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var result []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0]["name"] != "NZBgeek" {
		t.Errorf("name = %v, want NZBgeek", result[0]["name"])
	}
}

func TestProwlarrGetIndexerStats_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/indexerstats" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"indexers": []map[string]any{
				{"indexerId": 1, "numberOfQueries": 100, "numberOfGrabs": 50},
			},
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(&database.ProwlarrSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, nil)

	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetIndexerStats(w, httptest.NewRequest("GET", "/api/prowlarr/indexer-stats", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestProwlarrGetIndexers_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal error"))
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(&database.ProwlarrSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, nil)

	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetIndexers(w, httptest.NewRequest("GET", "/api/prowlarr/indexers", http.NoBody))

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

func TestProwlarrTestConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/system/status" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"appName": "Prowlarr", "version": "1.25.0"})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SSLVerify: true}, nil)

	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": srv.URL, "api_key": "test-key"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/prowlarr/test", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["version"] != "1.25.0" {
		t.Errorf("version = %q, want 1.25.0", resp["version"])
	}
}

func TestProwlarrTestConnection_MaskedKeyFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "stored-real-key" {
			t.Errorf("api key = %q, want stored-real-key", r.Header.Get("X-Api-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"appName": "Prowlarr", "version": "1.25.0"})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(&database.ProwlarrSettings{
		URL:    srv.URL,
		APIKey: "stored-real-key",
	}, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, nil)

	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": srv.URL, "api_key": "****-key"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/prowlarr/test", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

// =============================================================================
// SABnzbdHandler success-path tests (using httptest backend server)
// =============================================================================

func sabFakeServer(t *testing.T, expectedMode string, response any) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode")
		if mode != expectedMode {
			t.Errorf("mode = %q, want %q", mode, expectedMode)
		}
		if r.URL.Query().Get("apikey") == "" {
			t.Error("missing apikey")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	}))
}

func TestSABnzbdGetQueue_Success(t *testing.T) {
	srv := sabFakeServer(t, "queue", map[string]any{
		"queue": map[string]any{
			"status":    "Downloading",
			"speed":     "5.0 M",
			"sizeleft":  "1.2 GB",
			"timeleft":  "00:15:30",
			"noofslots": 3,
			"slots":     []any{},
		},
	})
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetQueue(w, httptest.NewRequest("GET", "/api/sabnzbd/queue", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSABnzbdGetHistory_Success(t *testing.T) {
	srv := sabFakeServer(t, "history", map[string]any{
		"history": map[string]any{
			"noofslots": 1,
			"slots":     []any{},
		},
	})
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetHistory(w, httptest.NewRequest("GET", "/api/sabnzbd/history", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSABnzbdGetStats_Success(t *testing.T) {
	srv := sabFakeServer(t, "server_stats", map[string]any{
		"total": 123456,
		"day":   5000,
		"week":  30000,
		"month": 100000,
	})
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetStats(w, httptest.NewRequest("GET", "/api/sabnzbd/stats", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSABnzbdPause_Success(t *testing.T) {
	srv := sabFakeServer(t, "pause", map[string]any{"status": true})
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandlePause(w, httptest.NewRequest("POST", "/api/sabnzbd/pause", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "paused" {
		t.Errorf("status = %q, want paused", resp["status"])
	}
}

func TestSABnzbdResume_Success(t *testing.T) {
	srv := sabFakeServer(t, "resume", map[string]any{"status": true})
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResume(w, httptest.NewRequest("POST", "/api/sabnzbd/resume", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "resumed" {
		t.Errorf("status = %q, want resumed", resp["status"])
	}
}

func TestSABnzbdGetQueue_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "test-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetQueue(w, httptest.NewRequest("GET", "/api/sabnzbd/queue", http.NoBody))

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

func TestSABnzbdTestConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mode := r.URL.Query().Get("mode")
		switch mode {
		case "version":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"version": "4.4.1"})
		case "fullstatus", "queue":
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]any{"status": true})
		default:
			// Return valid JSON for any mode
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, `{"version": "4.4.1"}`)
		}
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": srv.URL, "api_key": "test-key"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["version"] != "4.4.1" {
		t.Errorf("version = %q, want 4.4.1", resp["version"])
	}
}

func TestSABnzbdTestConnection_MaskedKeyFallback(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("apikey") != "stored-sab-key" {
			t.Errorf("apikey = %q, want stored-sab-key", r.URL.Query().Get("apikey"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"version": "4.4.1"})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{
		URL:    srv.URL,
		APIKey: "stored-sab-key",
	}, nil)

	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": srv.URL, "api_key": "****-key"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSABnzbdTestConnection_BadBody_Detailed(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader([]byte("not json"))))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSABnzbdTestConnection_MissingFieldsFallback(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, fmt.Errorf("not found"))
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": "", "api_key": ""})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader(body)))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}
