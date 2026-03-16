package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// =============================================================================
// SeerrHandler
// =============================================================================

func TestSeerrGetSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    "http://seerr:5055",
		APIKey: "abcdef123456",
	}, nil)
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/seerr/settings", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp database.SeerrSettings
	json.NewDecoder(w.Body).Decode(&resp)
	// API key should be masked
	if resp.APIKey == "abcdef123456" {
		t.Fatal("expected masked API key")
	}
	if resp.APIKey != "****3456" {
		t.Fatalf("expected ****3456, got %s", resp.APIKey)
	}
}

func TestSeerrGetSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/seerr/settings", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSeerrUpdateSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    "http://seerr:5055",
		APIKey: "oldkey12345678",
	}, nil)
	store.EXPECT().UpdateSeerrSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"url": "http://seerr:5055", "api_key": "newkey12345678",
		"enabled": true, "sync_interval_minutes": 15,
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/seerr/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestSeerrUpdateSettings_MaskedKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    "http://seerr:5055",
		APIKey: "original-real-key",
	}, nil)
	store.EXPECT().UpdateSeerrSettings(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, s *database.SeerrSettings) error {
			if s.APIKey != "original-real-key" {
				t.Errorf("expected preserved API key, got %s", s.APIKey)
			}
			return nil
		})
	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"url": "http://seerr:5055", "api_key": "****-key",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/seerr/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestSeerrUpdateSettings_BadBody(t *testing.T) {
	h := &SeerrHandler{}
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/seerr/settings", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSeerrUpdateSettings_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"url": "http://seerr:5055"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/seerr/settings", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSeerrUpdateSettings_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{APIKey: "key"}, nil)
	store.EXPECT().UpdateSeerrSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"url": "http://seerr:5055"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/seerr/settings", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSeerrTestConnection_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/seerr/test", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSeerrTestConnection_MissingSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{}, nil)
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/seerr/test", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSeerrTestConnection_MaskedKeyFallsBack(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{
		URL:    "http://seerr:5055",
		APIKey: "realkey12345",
	}, nil)
	h := &SeerrHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": "http://127.0.0.1:1", "api_key": "****2345"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/seerr/test", bytes.NewReader(body)))
	// Should use the stored key and try to connect (will fail with 502, but not 400)
	if w.Code == 400 {
		t.Fatalf("expected non-400 (masked key resolved), got 400: %s", w.Body.String())
	}
}

func TestSeerrGetRequests_NotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{}, nil)
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetRequests(w, httptest.NewRequest("GET", "/api/seerr/requests", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSeerrGetRequests_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetRequests(w, httptest.NewRequest("GET", "/api/seerr/requests", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSeerrGetRequestCount_NotConfigured(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSeerrSettings(gomock.Any()).Return(&database.SeerrSettings{}, nil)
	h := &SeerrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetRequestCount(w, httptest.NewRequest("GET", "/api/seerr/requests/count", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// =============================================================================
// OIDCSettingsHandler
// =============================================================================

func TestOIDCGetSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetOIDCSettings(gomock.Any()).Return(&database.OIDCSettings{
		ClientSecret: "supersecret1234",
		IssuerURL:    "https://accounts.google.com",
		ClientID:     "test-id",
	}, nil)
	h := &OIDCSettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/oidc/settings", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp database.OIDCSettings
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.ClientSecret == "supersecret1234" {
		t.Fatal("expected masked client secret")
	}
}

func TestOIDCGetSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetOIDCSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &OIDCSettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/oidc/settings", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateOIDCSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"enabled":       false,
		"issuer_url":    "https://accounts.google.com",
		"client_id":     "test-id",
		"client_secret": "newsecret1234",
		"redirect_url":  "http://localhost:9705/auth/oidc/callback",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestOIDCUpdateSettings_MaskedSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetOIDCSettings(gomock.Any()).Return(&database.OIDCSettings{
		ClientSecret: "original-secret-1234",
	}, nil)
	store.EXPECT().UpdateOIDCSettings(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, s *database.OIDCSettings) error {
			if s.ClientSecret != "original-secret-1234" {
				t.Errorf("expected preserved secret, got %s", s.ClientSecret)
			}
			return nil
		})
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"client_secret": "****1234",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings_MaskedSecret_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetOIDCSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"client_secret": "****1234",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings_EnabledMissingIssuer(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"enabled":       true,
		"client_id":     "test",
		"client_secret": "secret123456",
		"redirect_url":  "http://localhost/callback",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400 for missing issuer_url, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings_EnabledMissingClientID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"enabled":       true,
		"issuer_url":    "https://accounts.google.com",
		"client_secret": "secret123456",
		"redirect_url":  "http://localhost/callback",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400 for missing client_id, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings_EnabledMissingRedirect(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"enabled":       true,
		"issuer_url":    "https://accounts.google.com",
		"client_id":     "test",
		"client_secret": "secret123456",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400 for missing redirect_url, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings_BadBody(t *testing.T) {
	h := &OIDCSettingsHandler{}
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOIDCUpdateSettings_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateOIDCSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &OIDCSettingsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{
		"client_secret": "newsecret12345",
	})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/oidc/settings", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
