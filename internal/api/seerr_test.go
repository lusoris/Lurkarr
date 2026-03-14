package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// =============================================================================
// SeerrHandler success-path tests (using httptest backend server)
// =============================================================================

func TestSeerrTestConnection_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/about" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "real-api-key" {
			t.Errorf("api key = %q, want real-api-key", r.Header.Get("X-Api-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"version":         "2.3.0",
			"totalMediaItems": 500,
			"totalRequests":   42,
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "real-api-key",
	}, nil)

	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": srv.URL, "api_key": "real-api-key"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/seerr/test", bytes.NewReader(body)))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["version"] != "2.3.0" {
		t.Errorf("version = %v, want 2.3.0", resp["version"])
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %v, want ok", resp["status"])
	}
}

func TestSeerrTestConnection_FallbackToStored(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "stored-key-123" {
			t.Errorf("api key = %q, want stored-key-123", r.Header.Get("X-Api-Key"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{"version": "2.3.0"})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "stored-key-123",
	}, nil)

	h := &SeerrHandler{DB: store}
	// Send empty body — handler should fall back to stored settings
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/seerr/test", bytes.NewReader([]byte("{}"))))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSeerrTestConnection_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/seerr/test", bytes.NewReader([]byte("{}"))))

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

func TestSeerrGetRequests_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		// Verify filter parameter is passed through
		if r.URL.Query().Get("filter") != "pending" {
			t.Errorf("filter = %q, want pending", r.URL.Query().Get("filter"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"pageInfo": map[string]any{"pages": 1, "pageSize": 50, "results": 1, "page": 1},
			"results": []map[string]any{
				{"id": 1, "type": "movie", "status": 1},
			},
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/seerr/requests?filter=pending", http.NoBody)
	h.HandleGetRequests(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestSeerrGetRequests_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetRequests(w, httptest.NewRequest("GET", "/api/seerr/requests", http.NoBody))

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

func TestSeerrGetRequestCount_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/count" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"total":   10,
			"movie":   6,
			"tv":      4,
			"pending": 3,
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetRequestCount(w, httptest.NewRequest("GET", "/api/seerr/requests/count", http.NoBody))

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["total"] != float64(10) {
		t.Errorf("total = %v, want 10", resp["total"])
	}
}

func TestSeerrGetRequestCount_BackendError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetRequestCount(w, httptest.NewRequest("GET", "/api/seerr/requests/count", http.NoBody))

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}

// =============================================================================
// HandleReassignRequest tests
// =============================================================================

func TestSeerrReassignRequest_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" || r.URL.Path != "/api/v1/request/42" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		var body map[string]any
		json.NewDecoder(r.Body).Decode(&body)
		if body["serverId"] != float64(2) || body["profileId"] != float64(8) {
			t.Errorf("body = %v, want serverId=2 profileId=8", body)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"id": 42, "type": "movie", "status": 1, "serverId": 2, "profileId": 8,
			"media": map[string]any{"tmdbId": 550, "mediaType": "movie"},
		})
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"server_id": 2, "profile_id": 8})
	r := httptest.NewRequest("POST", "/api/seerr/requests/42/reassign", bytes.NewReader(body))
	r.SetPathValue("id", "42")
	w := httptest.NewRecorder()

	h.HandleReassignRequest(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["serverId"] != float64(2) {
		t.Errorf("serverId = %v, want 2", resp["serverId"])
	}
}

func TestSeerrReassignRequest_InvalidID(t *testing.T) {
	h := &SeerrHandler{}
	r := httptest.NewRequest("POST", "/api/seerr/requests/abc/reassign", bytes.NewReader([]byte("{}")))
	r.SetPathValue("id", "abc")
	w := httptest.NewRecorder()

	h.HandleReassignRequest(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSeerrReassignRequest_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"server_id": 0, "profile_id": 8})
	r := httptest.NewRequest("POST", "/api/seerr/requests/42/reassign", bytes.NewReader(body))
	r.SetPathValue("id", "42")
	w := httptest.NewRecorder()

	h.HandleReassignRequest(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusBadRequest, w.Body.String())
	}
}

func TestSeerrReassignRequest_NotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{}, nil)

	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"server_id": 2, "profile_id": 8})
	r := httptest.NewRequest("POST", "/api/seerr/requests/42/reassign", bytes.NewReader(body))
	r.SetPathValue("id", "42")
	w := httptest.NewRecorder()

	h.HandleReassignRequest(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestSeerrReassignRequest_BackendFails(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    srv.URL,
		APIKey: "key",
	}, nil)

	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"server_id": 2, "profile_id": 8})
	r := httptest.NewRequest("POST", "/api/seerr/requests/42/reassign", bytes.NewReader(body))
	r.SetPathValue("id", "42")
	w := httptest.NewRecorder()

	h.HandleReassignRequest(w, r)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadGateway)
	}
}
