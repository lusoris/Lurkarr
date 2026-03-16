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
	"github.com/lusoris/lurkarr/internal/notifications"
)

// =============================================================================
// NotificationHandler
// =============================================================================

func TestNotificationListProviders(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListNotificationProviders(gomock.Any()).Return([]database.NotificationProvider{
		{ID: uuid.New(), Type: "discord", Name: "my-discord"},
	}, nil)
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListProviders(w, httptest.NewRequest("GET", "/api/notifications/providers", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNotificationListProviders_NilSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListNotificationProviders(gomock.Any()).Return(nil, nil)
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListProviders(w, httptest.NewRequest("GET", "/api/notifications/providers", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() == "null\n" {
		t.Fatal("expected empty array, got null")
	}
}

func TestNotificationListProviders_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListNotificationProviders(gomock.Any()).Return(nil, errors.New("fail"))
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListProviders(w, httptest.NewRequest("GET", "/api/notifications/providers", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNotificationGetProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetNotificationProvider(gomock.Any(), id).Return(
		&database.NotificationProvider{ID: id, Type: "discord"}, nil)
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/notifications/providers/"+id.String(), nil, "id", id.String())
	h.HandleGetProvider(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestNotificationGetProvider_InvalidID(t *testing.T) {
	h := &NotificationHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/notifications/providers/bad", nil, "id", "bad")
	h.HandleGetProvider(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationGetProvider_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetNotificationProvider(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/notifications/providers/"+id.String(), nil, "id", id.String())
	h.HandleGetProvider(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestNotificationCreateProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateNotificationProvider(gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().ListEnabledNotificationProviders(gomock.Any()).Return(nil, nil)
	mgr := notifications.NewManager()
	h := &NotificationHandler{DB: store, Manager: mgr}
	body, _ := json.Marshal(map[string]any{
		"type": "discord", "name": "test", "enabled": true,
		"config": map[string]string{"webhook_url": "https://discord.com/api/webhooks/test"},
	})
	w := httptest.NewRecorder()
	h.HandleCreateProvider(w, httptest.NewRequest("POST", "/api/notifications/providers", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNotificationCreateProvider_InvalidType(t *testing.T) {
	h := &NotificationHandler{}
	body, _ := json.Marshal(map[string]any{"type": "invalid", "name": "test"})
	w := httptest.NewRecorder()
	h.HandleCreateProvider(w, httptest.NewRequest("POST", "/api/notifications/providers", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationCreateProvider_BadBody(t *testing.T) {
	h := &NotificationHandler{}
	w := httptest.NewRecorder()
	h.HandleCreateProvider(w, httptest.NewRequest("POST", "/api/notifications/providers", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationCreateProvider_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateNotificationProvider(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &NotificationHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"type": "discord", "name": "test"})
	w := httptest.NewRecorder()
	h.HandleCreateProvider(w, httptest.NewRequest("POST", "/api/notifications/providers", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNotificationUpdateProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateNotificationProvider(gomock.Any(), gomock.Any()).Return(nil)
	store.EXPECT().ListEnabledNotificationProviders(gomock.Any()).Return(nil, nil)
	mgr := notifications.NewManager()
	h := &NotificationHandler{DB: store, Manager: mgr}
	body, _ := json.Marshal(map[string]any{
		"type": "telegram", "name": "updated", "enabled": true,
	})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/notifications/providers/"+id.String(), body, "id", id.String())
	h.HandleUpdateProvider(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestNotificationUpdateProvider_InvalidID(t *testing.T) {
	h := &NotificationHandler{}
	body, _ := json.Marshal(map[string]any{"type": "discord"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/notifications/providers/bad", body, "id", "bad")
	h.HandleUpdateProvider(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationUpdateProvider_InvalidType(t *testing.T) {
	id := uuid.New()
	h := &NotificationHandler{}
	body, _ := json.Marshal(map[string]any{"type": "invalid"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/notifications/providers/"+id.String(), body, "id", id.String())
	h.HandleUpdateProvider(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationUpdateProvider_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateNotificationProvider(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &NotificationHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"type": "discord", "name": "test"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/notifications/providers/"+id.String(), body, "id", id.String())
	h.HandleUpdateProvider(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestNotificationDeleteProvider(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteNotificationProvider(gomock.Any(), id).Return(nil)
	store.EXPECT().ListEnabledNotificationProviders(gomock.Any()).Return(nil, nil)
	mgr := notifications.NewManager()
	h := &NotificationHandler{DB: store, Manager: mgr}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/notifications/providers/"+id.String(), nil, "id", id.String())
	h.HandleDeleteProvider(w, r)
	if w.Code != 204 {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestNotificationDeleteProvider_InvalidID(t *testing.T) {
	h := &NotificationHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/notifications/providers/bad", nil, "id", "bad")
	h.HandleDeleteProvider(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestNotificationDeleteProvider_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteNotificationProvider(gomock.Any(), id).Return(errors.New("fail"))
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/notifications/providers/"+id.String(), nil, "id", id.String())
	h.HandleDeleteProvider(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// TestNotificationTestProvider_InvalidID tests test endpoint with bad UUID.
func TestNotificationTestProvider_InvalidID(t *testing.T) {
	h := &NotificationHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/notifications/providers/bad/test", nil, "id", "bad")
	h.HandleTestProvider(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// TestNotificationTestProvider_NotFound tests test endpoint for non-existent provider.
func TestNotificationTestProvider_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetNotificationProvider(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &NotificationHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/notifications/providers/"+id.String()+"/test", nil, "id", id.String())
	h.HandleTestProvider(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// TestValidProviderType tests all valid and invalid provider types.
func TestValidProviderType(t *testing.T) {
	validTypes := []string{"discord", "telegram", "pushover", "gotify", "ntfy", "apprise", "email", "webhook"}
	for _, pt := range validTypes {
		if !validProviderType(pt) {
			t.Errorf("expected %q to be valid", pt)
		}
	}
	invalidTypes := []string{"", "slack", "invalid", "sms"}
	for _, pt := range invalidTypes {
		if validProviderType(pt) {
			t.Errorf("expected %q to be invalid", pt)
		}
	}
}
