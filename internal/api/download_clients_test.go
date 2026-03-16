package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// =============================================================================
// DownloadClientHandler
// =============================================================================

func TestDLClientList(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListDownloadClientInstances(gomock.Any()).Return([]database.DownloadClientInstance{
		{ID: uuid.New(), Name: "qb", ClientType: "qbittorrent", URL: "http://qb:8080", APIKey: "abcdef1234", Password: "secret"},
	}, nil)
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleList(w, httptest.NewRequest("GET", "/api/download-clients", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var items []database.DownloadClientInstance
	json.NewDecoder(w.Body).Decode(&items)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	// API key should be masked
	if items[0].APIKey == "abcdef1234" {
		t.Fatal("expected masked api key")
	}
	// Password should be masked
	if items[0].Password == "secret" {
		t.Fatal("expected masked password")
	}
}

func TestDLClientList_NilSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListDownloadClientInstances(gomock.Any()).Return(nil, nil)
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleList(w, httptest.NewRequest("GET", "/api/download-clients", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() == "null\n" {
		t.Fatal("expected empty array, got null")
	}
}

func TestDLClientList_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListDownloadClientInstances(gomock.Any()).Return(nil, errors.New("fail"))
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleList(w, httptest.NewRequest("GET", "/api/download-clients", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestDLClientCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateDownloadClientInstance(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, d *database.DownloadClientInstance) (*database.DownloadClientInstance, error) {
			d.ID = uuid.New()
			return d, nil
		})
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"name": "test-qb", "client_type": "qbittorrent",
		"url": "http://qb:8080", "username": "admin", "password": "pass123",
	})
	w := httptest.NewRecorder()
	h.HandleCreate(w, httptest.NewRequest("POST", "/api/download-clients", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestDLClientCreate_DefaultTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateDownloadClientInstance(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, d *database.DownloadClientInstance) (*database.DownloadClientInstance, error) {
			if d.Timeout != 30 {
				t.Errorf("expected default timeout 30, got %d", d.Timeout)
			}
			d.ID = uuid.New()
			return d, nil
		})
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"name": "test", "client_type": "sabnzbd", "url": "http://sab:8080",
	})
	w := httptest.NewRecorder()
	h.HandleCreate(w, httptest.NewRequest("POST", "/api/download-clients", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestDLClientCreate_MissingFields(t *testing.T) {
	h := &DownloadClientHandler{}
	body, _ := json.Marshal(map[string]string{"name": "test"})
	w := httptest.NewRecorder()
	h.HandleCreate(w, httptest.NewRequest("POST", "/api/download-clients", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDLClientCreate_InvalidType(t *testing.T) {
	h := &DownloadClientHandler{}
	body, _ := json.Marshal(map[string]string{"name": "test", "url": "http://x", "client_type": "invalid"})
	w := httptest.NewRecorder()
	h.HandleCreate(w, httptest.NewRequest("POST", "/api/download-clients", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDLClientCreate_BadBody(t *testing.T) {
	h := &DownloadClientHandler{}
	w := httptest.NewRecorder()
	h.HandleCreate(w, httptest.NewRequest("POST", "/api/download-clients", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDLClientCreate_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateDownloadClientInstance(gomock.Any(), gomock.Any()).Return(nil, errors.New("fail"))
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"name": "test", "url": "http://x", "client_type": "qbittorrent"})
	w := httptest.NewRecorder()
	h.HandleCreate(w, httptest.NewRequest("POST", "/api/download-clients", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestDLClientUpdate(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(
		&database.DownloadClientInstance{ID: id, APIKey: "realkey", Password: "realpass"}, nil,
	)
	store.EXPECT().UpdateDownloadClientInstance(gomock.Any(), gomock.Any()).Return(nil)
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"name": "updated", "client_type": "qbittorrent", "url": "http://qb:8080",
		"api_key": "newkey", "password": "newpass", "enabled": true,
	})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/download-clients/"+id.String(), body, "id", id.String())
	h.HandleUpdate(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDLClientUpdate_MaskedSecrets(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(
		&database.DownloadClientInstance{ID: id, APIKey: "original-apikey", Password: "original-pass"}, nil,
	)
	store.EXPECT().UpdateDownloadClientInstance(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, d *database.DownloadClientInstance) error {
			if d.APIKey != "original-apikey" {
				t.Errorf("expected preserved api key, got %s", d.APIKey)
			}
			if d.Password != "original-pass" {
				t.Errorf("expected preserved password, got %s", d.Password)
			}
			return nil
		})
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"name": "updated", "client_type": "qbittorrent", "url": "http://qb:8080",
		"api_key": "****key1", "password": "****", "enabled": true,
	})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/download-clients/"+id.String(), body, "id", id.String())
	h.HandleUpdate(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDLClientUpdate_InvalidID(t *testing.T) {
	h := &DownloadClientHandler{}
	body, _ := json.Marshal(map[string]any{"client_type": "qbittorrent"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/download-clients/bad", body, "id", "bad")
	h.HandleUpdate(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDLClientUpdate_InvalidType(t *testing.T) {
	id := uuid.New()
	h := &DownloadClientHandler{}
	body, _ := json.Marshal(map[string]any{"client_type": "invalid"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/download-clients/"+id.String(), body, "id", id.String())
	h.HandleUpdate(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDLClientUpdate_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"client_type": "qbittorrent", "url": "http://x"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/download-clients/"+id.String(), body, "id", id.String())
	h.HandleUpdate(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestDLClientUpdate_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(
		&database.DownloadClientInstance{ID: id}, nil)
	store.EXPECT().UpdateDownloadClientInstance(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"client_type": "qbittorrent", "url": "http://x"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/download-clients/"+id.String(), body, "id", id.String())
	h.HandleUpdate(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestDLClientDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteDownloadClientInstance(gomock.Any(), id).Return(nil)
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/download-clients/"+id.String(), nil, "id", id.String())
	h.HandleDelete(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestDLClientDelete_InvalidID(t *testing.T) {
	h := &DownloadClientHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/download-clients/bad", nil, "id", "bad")
	h.HandleDelete(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestDLClientDelete_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteDownloadClientInstance(gomock.Any(), id).Return(errors.New("fail"))
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/download-clients/"+id.String(), nil, "id", id.String())
	h.HandleDelete(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestDLClientTest_BadBody tests /api/download-clients/test with bad JSON.
func TestDLClientTest_BadBody(t *testing.T) {
	h := &DownloadClientHandler{}
	w := httptest.NewRecorder()
	h.HandleTest(w, httptest.NewRequest("POST", "/api/download-clients/test", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestDLClientTest_MissingFields tests /api/download-clients/test with no url.
func TestDLClientTest_MissingFields(t *testing.T) {
	h := &DownloadClientHandler{}
	body, _ := json.Marshal(map[string]string{"client_type": "qbittorrent"})
	w := httptest.NewRecorder()
	h.HandleTest(w, httptest.NewRequest("POST", "/api/download-clients/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestDLClientTest_MaskedCredentials verifies that masked credentials fall back to stored values.
func TestDLClientTest_MaskedCredentials(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(
		&database.DownloadClientInstance{
			ID: id, ClientType: "qbittorrent", URL: "http://qb:8080",
			APIKey: "realkey", Password: "realpass", Username: "admin",
		}, nil,
	)
	h := &DownloadClientHandler{DB: store}
	body, _ := json.Marshal(map[string]string{
		"id": id.String(), "client_type": "qbittorrent",
		"url": "http://qb:8080", "api_key": "****key1", "password": "****",
	})
	w := httptest.NewRecorder()
	h.HandleTest(w, httptest.NewRequest("POST", "/api/download-clients/test", bytes.NewReader(body)))
	// Will fail with 502 because the qbittorrent server isn't running, but that's fine —
	// the important thing is it didn't return 400 (credentials resolved).
	if w.Code == 400 {
		t.Fatalf("expected non-400 (credentials should have resolved), got 400")
	}
}

// TestDLClientHealthCheck_InvalidID tests health check with invalid UUID.
func TestDLClientHealthCheck_InvalidID(t *testing.T) {
	h := &DownloadClientHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/download-clients/bad/health", nil, "id", "bad")
	h.HandleHealthCheck(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestDLClientHealthCheck_NotFound tests health check for non-existent client.
func TestDLClientHealthCheck_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/download-clients/"+id.String()+"/health", nil, "id", id.String())
	h.HandleHealthCheck(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestDLClientStatus_InvalidID tests status endpoint with bad ID.
func TestDLClientStatus_InvalidID(t *testing.T) {
	h := &DownloadClientHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/download-clients/bad/status", nil, "id", "bad")
	h.HandleStatus(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestDLClientStatus_NotFound tests status endpoint for non-existent client.
func TestDLClientStatus_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/download-clients/"+id.String()+"/status", nil, "id", id.String())
	h.HandleStatus(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestDLClientItems_InvalidID tests items endpoint with bad ID.
func TestDLClientItems_InvalidID(t *testing.T) {
	h := &DownloadClientHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/download-clients/bad/items", nil, "id", "bad")
	h.HandleItems(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestDLClientItems_NotFound tests items endpoint for non-existent client.
func TestDLClientItems_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetDownloadClientInstance(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &DownloadClientHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/download-clients/"+id.String()+"/items", nil, "id", id.String())
	h.HandleItems(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestBuildClient_AllTypes verifies buildClient returns non-nil for all valid types.
func TestBuildClient_AllTypes(t *testing.T) {
	for _, ct := range []string{"qbittorrent", "transmission", "deluge", "sabnzbd", "nzbget"} {
		c := buildClient(ct, "http://localhost:8080", "key", "user", "pass", 30)
		if c == nil {
			t.Errorf("buildClient(%q) returned nil", ct)
		}
	}
}

// TestBuildClient_InvalidType verifies buildClient returns nil for unknown types.
func TestBuildClient_InvalidType(t *testing.T) {
	c := buildClient("unknown", "http://localhost:8080", "key", "user", "pass", 30)
	if c != nil {
		t.Fatal("expected nil for unknown client type")
	}
}
