package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/mocks"
	"github.com/lusoris/lurkarr/internal/scheduler"
)

// Silence unused import
var _ = slog.Default

// --- helpers ---

func reqWithPathValue(method, path string, body []byte, key, value string) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, path, http.NoBody)
	}
	r.SetPathValue(key, value)
	return r
}

func reqWithUserCtx(r *http.Request, user *database.User) *http.Request {
	ctx := auth.ContextWithUser(r.Context(), user)
	return r.WithContext(ctx)
}

func newTestSchedulerHandler(t *testing.T, ctrl *gomock.Controller) (sh *SchedulerHandler, apiStore *MockStore, schedStore *mocks.MockStore) {
	t.Helper()
	apiStore = NewMockStore(ctrl)
	schedStore = mocks.NewMockStore(ctrl)
	logger := logging.New()
	sched, err := scheduler.New(schedStore, logger)
	if err != nil {
		t.Fatal(err)
	}
	return &SchedulerHandler{DB: apiStore, Scheduler: sched}, apiStore, schedStore
}

// =============================================================================
// StatsHandler tests
// =============================================================================

func TestHandleGetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAllStats(gomock.Any()).Return([]database.LurkStats{{AppType: "sonarr"}}, nil)
	h := &StatsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetStats(w, httptest.NewRequest("GET", "/api/stats", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetStats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAllStats(gomock.Any()).Return(nil, errors.New("fail"))
	h := &StatsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetStats(w, httptest.NewRequest("GET", "/api/stats", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleResetStats(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ResetStats(gomock.Any()).Return(nil)
	h := &StatsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResetStats(w, httptest.NewRequest("POST", "/api/stats/reset", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleResetStats_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ResetStats(gomock.Any()).Return(errors.New("fail"))
	h := &StatsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResetStats(w, httptest.NewRequest("POST", "/api/stats/reset", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetHourlyCaps(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAllHourlyCaps(gomock.Any()).Return(nil, nil)
	h := &StatsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetHourlyCaps(w, httptest.NewRequest("GET", "/api/stats/hourly-caps", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetHourlyCaps_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAllHourlyCaps(gomock.Any()).Return(nil, errors.New("fail"))
	h := &StatsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetHourlyCaps(w, httptest.NewRequest("GET", "/api/stats/hourly-caps", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// HistoryHandler tests
// =============================================================================

func TestHandleListHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListLurkHistory(gomock.Any(), gomock.Any()).Return([]database.LurkHistory{{MediaTitle: "Test"}}, 1, nil)
	h := &HistoryHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListHistory(w, httptest.NewRequest("GET", "/api/history", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleListHistory_WithParams(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListLurkHistory(gomock.Any(), gomock.Any()).DoAndReturn(
		func(_ interface{}, q database.HistoryQuery) ([]database.LurkHistory, int, error) {
			if q.Limit != 100 {
				t.Errorf("expected limit 100, got %d", q.Limit)
			}
			if q.Offset != 10 {
				t.Errorf("expected offset 10, got %d", q.Offset)
			}
			if q.AppType != "sonarr" {
				t.Errorf("expected app=sonarr, got %s", q.AppType)
			}
			return nil, 0, nil
		})
	h := &HistoryHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListHistory(w, httptest.NewRequest("GET", "/api/history?app=sonarr&limit=100&offset=10", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleListHistory_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListLurkHistory(gomock.Any(), gomock.Any()).Return(nil, 0, errors.New("fail"))
	h := &HistoryHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListHistory(w, httptest.NewRequest("GET", "/api/history", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleDeleteHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().DeleteHistory(gomock.Any(), database.AppType("sonarr")).Return(nil)
	h := &HistoryHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteHistory(w, reqWithPathValue("DELETE", "/api/history/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDeleteHistory_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &HistoryHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteHistory(w, reqWithPathValue("DELETE", "/api/history/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteHistory_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().DeleteHistory(gomock.Any(), database.AppType("sonarr")).Return(errors.New("fail"))
	h := &HistoryHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteHistory(w, reqWithPathValue("DELETE", "/api/history/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// SettingsHandler tests
// =============================================================================

func TestHandleGetAppSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppType("sonarr")).Return(&database.AppSettings{AppType: "sonarr"}, nil)
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetAppSettings(w, reqWithPathValue("GET", "/api/settings/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetAppSettings_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetAppSettings(w, reqWithPathValue("GET", "/api/settings/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetAppSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("fail"))
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetAppSettings(w, reqWithPathValue("GET", "/api/settings/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateAppSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.AppSettings{LurkMissingCount: 10, LurkUpgradeCount: 5, SleepDuration: 15})
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/sonarr", []byte("bad"), "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings_NegativeHourlyCap(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.AppSettings{HourlyCap: -1})
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings_NegativeLurkCounts(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.AppSettings{LurkMissingCount: -1})
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings_NegativeSleep(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.AppSettings{SleepDuration: -1})
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateAppSettings_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateAppSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.AppSettings{LurkMissingCount: 10, LurkUpgradeCount: 5, SleepDuration: 15})
	w := httptest.NewRecorder()
	h.HandleUpdateAppSettings(w, reqWithPathValue("PUT", "/api/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetGeneralSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "supersecret", APITimeout: 30, StatefulResetHours: 168}, nil)
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetGeneralSettings(w, httptest.NewRequest("GET", "/api/settings/general", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp database.GeneralSettings
	json.NewDecoder(w.Body).Decode(&resp)
	if resp.SecretKey != "****" {
		t.Fatalf("expected secret key masked, got %q", resp.SecretKey)
	}
}

func TestHandleGetGeneralSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetGeneralSettings(w, httptest.NewRequest("GET", "/api/settings/general", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "supersecret", APITimeout: 30}, nil)
	store.EXPECT().UpsertGeneralSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.GeneralSettings{APITimeout: 60, StatefulResetHours: 168})
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "s"}, nil)
	h := &SettingsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_InvalidTimeout(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "s"}, nil)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.GeneralSettings{APITimeout: 0})
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_NegativeReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "s"}, nil)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.GeneralSettings{APITimeout: 30, StatefulResetHours: -1})
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_NegativeCommandWait(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "s"}, nil)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.GeneralSettings{APITimeout: 30, CommandWaitDelay: -1})
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_NegativeMinQueue(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "s"}, nil)
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.GeneralSettings{APITimeout: 30, MinDownloadQueueSize: -1})
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateGeneralSettings_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{SecretKey: "s"}, nil)
	store.EXPECT().UpsertGeneralSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &SettingsHandler{DB: store}
	body, _ := json.Marshal(database.GeneralSettings{APITimeout: 30, StatefulResetHours: 168})
	w := httptest.NewRecorder()
	h.HandleUpdateGeneralSettings(w, httptest.NewRequest("PUT", "/api/settings/general", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// StateHandler tests
// =============================================================================

func TestHandleGetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	now := time.Now()
	for _, at := range database.AllAppTypes() {
		store.EXPECT().ListInstances(gomock.Any(), at).Return([]database.AppInstance{{ID: id, Name: "main"}}, nil)
		store.EXPECT().GetLastReset(gomock.Any(), at, id).Return(&now, nil)
	}
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetState(w, httptest.NewRequest("GET", "/api/state", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetState_WithFilter(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return(nil, nil)
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetState(w, httptest.NewRequest("GET", "/api/state?app=sonarr", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetState_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetState(w, httptest.NewRequest("GET", "/api/state?app=bogus", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleResetState(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().ResetState(gomock.Any(), database.AppType("sonarr"), id).Return(nil)
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResetState(w, httptest.NewRequest("POST", "/api/state/reset?app=sonarr&instance_id="+id.String(), http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleResetState_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResetState(w, httptest.NewRequest("POST", "/api/state/reset?app=bogus", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleResetState_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResetState(w, httptest.NewRequest("POST", "/api/state/reset?app=sonarr&instance_id=bad", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleResetState_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().ResetState(gomock.Any(), database.AppType("sonarr"), id).Return(errors.New("fail"))
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResetState(w, httptest.NewRequest("POST", "/api/state/reset?app=sonarr&instance_id="+id.String(), http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetState_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	for _, at := range database.AllAppTypes() {
		store.EXPECT().ListInstances(gomock.Any(), at).Return(nil, errors.New("fail"))
	}
	h := &StateHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetState(w, httptest.NewRequest("GET", "/api/state", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

// =============================================================================
// QueueHandler tests
// =============================================================================

func TestHandleGetQueueCleanerSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetQueueCleanerSettings(gomock.Any(), database.AppType("sonarr")).Return(&database.QueueCleanerSettings{AppType: "sonarr"}, nil)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetQueueCleanerSettings(w, reqWithPathValue("GET", "/api/queue/settings/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetQueueCleanerSettings_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetQueueCleanerSettings(w, reqWithPathValue("GET", "/api/queue/settings/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetQueueCleanerSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetQueueCleanerSettings(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("fail"))
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetQueueCleanerSettings(w, reqWithPathValue("GET", "/api/queue/settings/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateQueueCleanerSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateQueueCleanerSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &QueueHandler{DB: store}
	body, _ := json.Marshal(database.QueueCleanerSettings{Enabled: true})
	w := httptest.NewRecorder()
	h.HandleUpdateQueueCleanerSettings(w, reqWithPathValue("PUT", "/api/queue/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateQueueCleanerSettings_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateQueueCleanerSettings(w, reqWithPathValue("PUT", "/api/queue/settings/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateQueueCleanerSettings_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateQueueCleanerSettings(w, reqWithPathValue("PUT", "/api/queue/settings/sonarr", []byte("bad"), "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateQueueCleanerSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateQueueCleanerSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &QueueHandler{DB: store}
	body, _ := json.Marshal(database.QueueCleanerSettings{Enabled: true})
	w := httptest.NewRecorder()
	h.HandleUpdateQueueCleanerSettings(w, reqWithPathValue("PUT", "/api/queue/settings/sonarr", body, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetScoringProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppType("sonarr")).Return(&database.ScoringProfile{ID: uuid.New()}, nil)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetScoringProfile(w, reqWithPathValue("GET", "/api/queue/scoring/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetScoringProfile_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetScoringProfile(w, reqWithPathValue("GET", "/api/queue/scoring/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetScoringProfile_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("fail"))
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetScoringProfile(w, reqWithPathValue("GET", "/api/queue/scoring/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateScoringProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	profileID := uuid.New()
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppType("sonarr")).Return(&database.ScoringProfile{ID: profileID}, nil)
	store.EXPECT().UpdateScoringProfile(gomock.Any(), gomock.Any()).Return(nil)
	h := &QueueHandler{DB: store}
	body, _ := json.Marshal(database.ScoringProfile{})
	w := httptest.NewRecorder()
	h.HandleUpdateScoringProfile(w, reqWithPathValue("PUT", "/api/queue/scoring/sonarr", body, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateScoringProfile_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateScoringProfile(w, reqWithPathValue("PUT", "/api/queue/scoring/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateScoringProfile_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateScoringProfile(w, reqWithPathValue("PUT", "/api/queue/scoring/sonarr", []byte("bad"), "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateScoringProfile_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("fail"))
	h := &QueueHandler{DB: store}
	body, _ := json.Marshal(database.ScoringProfile{})
	w := httptest.NewRecorder()
	h.HandleUpdateScoringProfile(w, reqWithPathValue("PUT", "/api/queue/scoring/sonarr", body, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateScoringProfile_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppType("sonarr")).Return(&database.ScoringProfile{ID: uuid.New()}, nil)
	store.EXPECT().UpdateScoringProfile(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &QueueHandler{DB: store}
	body, _ := json.Marshal(database.ScoringProfile{})
	w := httptest.NewRecorder()
	h.HandleUpdateScoringProfile(w, reqWithPathValue("PUT", "/api/queue/scoring/sonarr", body, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetBlocklistLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetBlocklistLog(gomock.Any(), database.AppType("sonarr"), 100).Return(nil, nil)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetBlocklistLog(w, reqWithPathValue("GET", "/api/queue/blocklist/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetBlocklistLog_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetBlocklistLog(w, reqWithPathValue("GET", "/api/queue/blocklist/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetBlocklistLog_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetBlocklistLog(gomock.Any(), database.AppType("sonarr"), 100).Return(nil, errors.New("fail"))
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetBlocklistLog(w, reqWithPathValue("GET", "/api/queue/blocklist/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetAutoImportLog(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAutoImportLog(gomock.Any(), database.AppType("sonarr"), 100).Return(nil, nil)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetAutoImportLog(w, reqWithPathValue("GET", "/api/queue/imports/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetAutoImportLog_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetAutoImportLog(w, reqWithPathValue("GET", "/api/queue/imports/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetAutoImportLog_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAutoImportLog(gomock.Any(), database.AppType("sonarr"), 100).Return(nil, errors.New("fail"))
	h := &QueueHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetAutoImportLog(w, reqWithPathValue("GET", "/api/queue/imports/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// ProwlarrHandler tests
// =============================================================================

func TestHandleGetProwlarrSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(&database.ProwlarrSettings{URL: "http://localhost:9696", APIKey: "abcdef123456"}, nil)
	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/prowlarr/settings", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetProwlarrSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/prowlarr/settings", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateProwlarrSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateProwlarrSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(database.ProwlarrSettings{URL: "http://localhost:9696", APIKey: "newkey123456"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/prowlarr/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateProwlarrSettings_MaskedKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(&database.ProwlarrSettings{URL: "http://localhost:9696", APIKey: "realkey123"}, nil)
	store.EXPECT().UpdateProwlarrSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(database.ProwlarrSettings{URL: "http://localhost:9696", APIKey: "****y123"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/prowlarr/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateProwlarrSettings_ShortKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateProwlarrSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(database.ProwlarrSettings{URL: "http://localhost:9696", APIKey: "ab"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/prowlarr/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: short API key should not panic", w.Code)
	}
}

func TestHandleUpdateProwlarrSettings_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/prowlarr/settings", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateProwlarrSettings_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateProwlarrSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(database.ProwlarrSettings{URL: "http://localhost:9696", APIKey: "newkey123456"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/prowlarr/settings", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestProwlarrGetIndexers_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetIndexers(w, httptest.NewRequest("GET", "/api/prowlarr/indexers", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProwlarrGetIndexerStats_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetProwlarrSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetIndexerStats(w, httptest.NewRequest("GET", "/api/prowlarr/indexer-stats", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProwlarrTestConnection_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &ProwlarrHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/prowlarr/test", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProwlarrTestConnection_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": "http://example.com"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/prowlarr/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestProwlarrTestConnection_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &ProwlarrHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": "://bad", "api_key": "key123"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/prowlarr/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// =============================================================================
// SABnzbdHandler tests
// =============================================================================

func TestHandleGetSABnzbdSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{URL: "http://localhost:8080", APIKey: "abcdef123456"}, nil)
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/sabnzbd/settings", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetSABnzbdSettings_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetSettings(w, httptest.NewRequest("GET", "/api/sabnzbd/settings", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateSABnzbdSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateSABnzbdSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(database.SABnzbdSettings{URL: "http://localhost:8080", APIKey: "newkey123456"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/sabnzbd/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateSABnzbdSettings_MaskedKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{URL: "http://localhost:8080", APIKey: "realkey123"}, nil)
	store.EXPECT().UpdateSABnzbdSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(database.SABnzbdSettings{URL: "http://localhost:8080", APIKey: "****y123"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/sabnzbd/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateSABnzbdSettings_ShortKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateSABnzbdSettings(gomock.Any(), gomock.Any()).Return(nil)
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(database.SABnzbdSettings{URL: "http://localhost:8080", APIKey: "ab"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/sabnzbd/settings", bytes.NewReader(body)))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: short API key should not panic", w.Code)
	}
}

func TestHandleUpdateSABnzbdSettings_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/sabnzbd/settings", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateSABnzbdSettings_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UpdateSABnzbdSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(database.SABnzbdSettings{URL: "http://localhost:8080", APIKey: "newkey123456"})
	w := httptest.NewRecorder()
	h.HandleUpdateSettings(w, httptest.NewRequest("PUT", "/api/sabnzbd/settings", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestSABnzbdGetQueue_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetQueue(w, httptest.NewRequest("GET", "/api/sabnzbd/queue", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdGetHistory_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetHistory(w, httptest.NewRequest("GET", "/api/sabnzbd/history", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdGetStats_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetStats(w, httptest.NewRequest("GET", "/api/sabnzbd/stats", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdPause_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandlePause(w, httptest.NewRequest("POST", "/api/sabnzbd/pause", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdResume_SettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, errors.New("no settings"))
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleResume(w, httptest.NewRequest("POST", "/api/sabnzbd/resume", http.NoBody))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdTestConnection_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SABnzbdHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdTestConnection_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": "http://example.com"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestSABnzbdTestConnection_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SABnzbdHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"url": "://bad", "api_key": "key123"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/sabnzbd/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// =============================================================================
// AppsHandler tests
// =============================================================================

func TestHandleListInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return([]database.AppInstance{
		{ID: uuid.New(), Name: "main", APIKey: "secret123456"},
	}, nil)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListInstances(w, reqWithPathValue("GET", "/api/instances/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleListInstances_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListInstances(w, reqWithPathValue("GET", "/api/instances/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleListInstances_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("fail"))
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListInstances(w, reqWithPathValue("GET", "/api/instances/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleCreateInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateInstance(gomock.Any(), database.AppType("sonarr"), "main", "http://localhost:8989", "key123").
		Return(&database.AppInstance{ID: uuid.New(), AppType: "sonarr", Name: "main", APIURL: "http://localhost:8989", APIKey: "key123"}, nil)
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "main", "api_url": "http://localhost:8989", "api_key": "key123"})
	w := httptest.NewRecorder()
	h.HandleCreateInstance(w, reqWithPathValue("POST", "/api/instances/sonarr", body, "app", "sonarr"))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestHandleCreateInstance_InvalidApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleCreateInstance(w, reqWithPathValue("POST", "/api/instances/bogus", nil, "app", "bogus"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateInstance_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "main"})
	w := httptest.NewRecorder()
	h.HandleCreateInstance(w, reqWithPathValue("POST", "/api/instances/sonarr", body, "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateInstance_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateInstance(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("fail"))
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "main", "api_url": "http://localhost:8989", "api_key": "key123"})
	w := httptest.NewRecorder()
	h.HandleCreateInstance(w, reqWithPathValue("POST", "/api/instances/sonarr", body, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateInstance(gomock.Any(), id, "main", "http://localhost:8989", "newkey", true).Return(nil)
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"name": "main", "api_url": "http://localhost:8989", "api_key": "newkey", "enabled": true})
	w := httptest.NewRecorder()
	h.HandleUpdateInstance(w, reqWithPathValue("PUT", "/api/instances/"+id.String(), body, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateInstance_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateInstance(w, reqWithPathValue("PUT", "/api/instances/bad", nil, "id", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateInstance_MaskedKey(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetInstance(gomock.Any(), id).Return(&database.AppInstance{ID: id, APIKey: "secret123"}, nil)
	store.EXPECT().UpdateInstance(gomock.Any(), id, "main", "http://localhost:8989", "secret123", true).Return(nil)
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"name": "main", "api_url": "http://localhost:8989", "api_key": "****t123", "enabled": true})
	w := httptest.NewRecorder()
	h.HandleUpdateInstance(w, reqWithPathValue("PUT", "/api/instances/"+id.String(), body, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateInstance_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateInstance(gomock.Any(), id, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"name": "main", "api_url": "http://localhost:8989", "api_key": "newkey", "enabled": true})
	w := httptest.NewRecorder()
	h.HandleUpdateInstance(w, reqWithPathValue("PUT", "/api/instances/"+id.String(), body, "id", id.String()))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleDeleteInstance(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteInstance(gomock.Any(), id).Return(nil)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteInstance(w, reqWithPathValue("DELETE", "/api/instances/"+id.String(), nil, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDeleteInstance_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteInstance(w, reqWithPathValue("DELETE", "/api/instances/bad", nil, "id", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteInstance_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteInstance(gomock.Any(), id).Return(errors.New("fail"))
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteInstance(w, reqWithPathValue("DELETE", "/api/instances/"+id.String(), nil, "id", id.String()))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestAppsTestConnection_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/instances/test", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAppsTestConnection_InvalidAppType(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"api_url": "http://example.com", "api_key": "key123", "app_type": "bogus"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/instances/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAppsTestConnection_InvalidURL(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AppsHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"api_url": "://bad", "api_key": "key123", "app_type": "sonarr"})
	w := httptest.NewRecorder()
	h.HandleTestConnection(w, httptest.NewRequest("POST", "/api/instances/test", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// =============================================================================
// UserHandler tests
// =============================================================================

func TestHandleGetUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	h := &UserHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/user", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.HandleGetUser(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetUser_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &UserHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetUser(w, httptest.NewRequest("GET", "/api/user", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleUpdateUsername(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().UpdateUsername(gomock.Any(), user.ID, "newname").Return(nil)
	h := &UserHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "newname"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/username", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.HandleUpdateUsername(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateUsername_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &UserHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdateUsername(w, httptest.NewRequest("POST", "/api/user/username", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleUpdateUsername_EmptyName(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	h := &UserHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": ""})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/username", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.HandleUpdateUsername(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateUsername_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().UpdateUsername(gomock.Any(), user.ID, "newname").Return(errors.New("fail"))
	h := &UserHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "newname"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/username", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.HandleUpdateUsername(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdatePassword_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &UserHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleUpdatePassword(w, httptest.NewRequest("POST", "/api/user/password", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleUpdatePassword_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	h := &UserHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/password", bytes.NewReader([]byte("bad")))
	r = reqWithUserCtx(r, user)
	h.HandleUpdatePassword(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdatePassword_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	hash, _ := auth.HashPassword("correct")
	user := &database.User{ID: uuid.New(), Username: "admin", Password: hash}
	h := &UserHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"current_password": "wrong", "new_password": "newpass"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/password", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.HandleUpdatePassword(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleUpdatePassword_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	hash, _ := auth.HashPassword("correct")
	user := &database.User{ID: uuid.New(), Username: "admin", Password: hash}
	store.EXPECT().UpdatePassword(gomock.Any(), user.ID, gomock.Any()).Return(nil)
	h := &UserHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"current_password": "correct", "new_password": "Newpass1"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/password", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.HandleUpdatePassword(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdatePassword_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	hash, _ := auth.HashPassword("correct")
	user := &database.User{ID: uuid.New(), Username: "admin", Password: hash}
	store.EXPECT().UpdatePassword(gomock.Any(), user.ID, gomock.Any()).Return(errors.New("fail"))
	h := &UserHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"current_password": "correct", "new_password": "Newpass1"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/user/password", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.HandleUpdatePassword(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// SchedulerHandler tests
// =============================================================================

func TestHandleListSchedules(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return(nil, nil)
	h := &SchedulerHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListSchedules(w, httptest.NewRequest("GET", "/api/schedules", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleListSchedules_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return(nil, errors.New("fail"))
	h := &SchedulerHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListSchedules(w, httptest.NewRequest("GET", "/api/schedules", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleScheduleHistory(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListScheduleExecutions(gomock.Any(), 50).Return(nil, nil)
	h := &SchedulerHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleScheduleHistory(w, httptest.NewRequest("GET", "/api/schedules/history", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleScheduleHistory_WithLimit(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListScheduleExecutions(gomock.Any(), 100).Return(nil, nil)
	h := &SchedulerHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleScheduleHistory(w, httptest.NewRequest("GET", "/api/schedules/history?limit=100", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleScheduleHistory_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListScheduleExecutions(gomock.Any(), 50).Return(nil, errors.New("fail"))
	h := &SchedulerHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleScheduleHistory(w, httptest.NewRequest("GET", "/api/schedules/history", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleCreateSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, apiStore, schedStore := newTestSchedulerHandler(t, ctrl)
	apiStore.EXPECT().CreateSchedule(gomock.Any(), gomock.Any()).Return(nil)
	schedStore.EXPECT().ListSchedules(gomock.Any()).Return(nil, nil)
	body, _ := json.Marshal(database.Schedule{AppType: "sonarr", Action: "disable"})
	w := httptest.NewRecorder()
	h.HandleCreateSchedule(w, httptest.NewRequest("POST", "/api/schedules", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestHandleCreateSchedule_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, _, _ := newTestSchedulerHandler(t, ctrl)
	w := httptest.NewRecorder()
	h.HandleCreateSchedule(w, httptest.NewRequest("POST", "/api/schedules", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateSchedule_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, _, _ := newTestSchedulerHandler(t, ctrl)
	body, _ := json.Marshal(database.Schedule{AppType: "sonarr"})
	w := httptest.NewRecorder()
	h.HandleCreateSchedule(w, httptest.NewRequest("POST", "/api/schedules", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateSchedule_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, apiStore, _ := newTestSchedulerHandler(t, ctrl)
	apiStore.EXPECT().CreateSchedule(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	body, _ := json.Marshal(database.Schedule{AppType: "sonarr", Action: "disable"})
	w := httptest.NewRecorder()
	h.HandleCreateSchedule(w, httptest.NewRequest("POST", "/api/schedules", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleUpdateSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, apiStore, schedStore := newTestSchedulerHandler(t, ctrl)
	id := uuid.New()
	apiStore.EXPECT().UpdateSchedule(gomock.Any(), gomock.Any()).Return(nil)
	schedStore.EXPECT().ListSchedules(gomock.Any()).Return(nil, nil)
	body, _ := json.Marshal(database.Schedule{AppType: "sonarr", Action: "enable"})
	w := httptest.NewRecorder()
	h.HandleUpdateSchedule(w, reqWithPathValue("PUT", "/api/schedules/"+id.String(), body, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateSchedule_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, _, _ := newTestSchedulerHandler(t, ctrl)
	w := httptest.NewRecorder()
	h.HandleUpdateSchedule(w, reqWithPathValue("PUT", "/api/schedules/bad", nil, "id", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateSchedule_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, _, _ := newTestSchedulerHandler(t, ctrl)
	id := uuid.New()
	w := httptest.NewRecorder()
	h.HandleUpdateSchedule(w, reqWithPathValue("PUT", "/api/schedules/"+id.String(), []byte("bad"), "id", id.String()))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleUpdateSchedule_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, apiStore, _ := newTestSchedulerHandler(t, ctrl)
	id := uuid.New()
	apiStore.EXPECT().UpdateSchedule(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	body, _ := json.Marshal(database.Schedule{AppType: "sonarr", Action: "enable"})
	w := httptest.NewRecorder()
	h.HandleUpdateSchedule(w, reqWithPathValue("PUT", "/api/schedules/"+id.String(), body, "id", id.String()))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleDeleteSchedule(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, apiStore, schedStore := newTestSchedulerHandler(t, ctrl)
	id := uuid.New()
	apiStore.EXPECT().DeleteSchedule(gomock.Any(), id).Return(nil)
	schedStore.EXPECT().ListSchedules(gomock.Any()).Return(nil, nil)
	w := httptest.NewRecorder()
	h.HandleDeleteSchedule(w, reqWithPathValue("DELETE", "/api/schedules/"+id.String(), nil, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDeleteSchedule_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, _, _ := newTestSchedulerHandler(t, ctrl)
	w := httptest.NewRecorder()
	h.HandleDeleteSchedule(w, reqWithPathValue("DELETE", "/api/schedules/bad", nil, "id", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteSchedule_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	h, apiStore, _ := newTestSchedulerHandler(t, ctrl)
	id := uuid.New()
	apiStore.EXPECT().DeleteSchedule(gomock.Any(), id).Return(errors.New("fail"))
	w := httptest.NewRecorder()
	h.HandleDeleteSchedule(w, reqWithPathValue("DELETE", "/api/schedules/"+id.String(), nil, "id", id.String()))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// AuthHandler tests
// =============================================================================

func TestHandleSetup_AlreadyDone(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UserCount(gomock.Any()).Return(1, nil)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleSetup(w, httptest.NewRequest("POST", "/api/auth/setup", http.NoBody))
	if w.Code != 409 {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestHandleSetup_CountError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UserCount(gomock.Any()).Return(0, errors.New("fail"))
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleSetup(w, httptest.NewRequest("POST", "/api/auth/setup", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleSetup_MissingFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UserCount(gomock.Any()).Return(0, nil)
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "", "password": ""})
	w := httptest.NewRecorder()
	h.HandleSetup(w, httptest.NewRequest("POST", "/api/auth/setup", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSetup_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UserCount(gomock.Any()).Return(0, nil)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleSetup(w, httptest.NewRequest("POST", "/api/auth/setup", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSetup_CreateUserError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UserCount(gomock.Any()).Return(0, nil)
	store.EXPECT().CreateUser(gomock.Any(), "admin", gomock.Any()).Return(nil, errors.New("fail"))
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "Testpass1"})
	w := httptest.NewRecorder()
	h.HandleSetup(w, httptest.NewRequest("POST", "/api/auth/setup", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleSetup_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().UserCount(gomock.Any()).Return(0, nil)
	var createdUser string
	store.EXPECT().CreateUser(gomock.Any(), "admin", gomock.Any()).DoAndReturn(
		func(_ interface{}, username, _ string) (*database.User, error) {
			createdUser = username
			return &database.User{ID: uuid.New(), Username: username}, nil
		})
	store.EXPECT().UpsertGeneralSettings(gomock.Any(), gomock.Any()).Return(nil)
	// Setup calls Auth.SetSessionCookie which will panic since Auth is nil.
	h := &AuthHandler{DB: store, Auth: nil}
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "Testpass1"})
	w := httptest.NewRecorder()
	func() {
		defer func() { recover() }()
		h.HandleSetup(w, httptest.NewRequest("POST", "/api/auth/setup", bytes.NewReader(body)))
	}()
	if createdUser != "admin" {
		t.Fatalf("expected user 'admin' was created, got %q", createdUser)
	}
}

func TestHandleLogin_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleLogin(w, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleLogin_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetUserByUsername(gomock.Any(), "nope").Return(nil, errors.New("not found"))
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "nope", "password": "pass"})
	w := httptest.NewRecorder()
	h.HandleLogin(w, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleLogin_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	hash, _ := auth.HashPassword("correct")
	store.EXPECT().GetUserByUsername(gomock.Any(), "admin").Return(&database.User{ID: uuid.New(), Username: "admin", Password: hash}, nil)
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "wrong"})
	w := httptest.NewRecorder()
	h.HandleLogin(w, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleLogin_TOTPRequired(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	secret := "JBSWY3DPEHPK3PXP"
	hash, _ := auth.HashPassword("correct")
	store.EXPECT().GetUserByUsername(gomock.Any(), "admin").Return(&database.User{ID: uuid.New(), Username: "admin", Password: hash, TOTPSecret: &secret}, nil)
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "correct"})
	w := httptest.NewRecorder()
	h.HandleLogin(w, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["totp_required"] != true {
		t.Fatal("expected totp_required in response")
	}
}

func TestHandleLogin_InvalidTOTP(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	secret := "JBSWY3DPEHPK3PXP"
	hash, _ := auth.HashPassword("correct")
	store.EXPECT().GetUserByUsername(gomock.Any(), "admin").Return(&database.User{ID: uuid.New(), Username: "admin", Password: hash, TOTPSecret: &secret}, nil)
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"username": "admin", "password": "correct", "totp_code": "000000"})
	w := httptest.NewRecorder()
	h.HandleLogin(w, httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(body)))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleLogout(t *testing.T) {
	ctrl := gomock.NewController(t)
	authStore := mocks.NewMockAuthStore(ctrl)
	mw := &auth.Middleware{DB: authStore}
	store := NewMockStore(ctrl)
	h := &AuthHandler{DB: store, Auth: mw}
	w := httptest.NewRecorder()
	h.HandleLogout(w, httptest.NewRequest("POST", "/api/auth/logout", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandle2FAEnable_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.Handle2FAEnable(w, httptest.NewRequest("POST", "/api/auth/2fa/enable", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandle2FAEnable_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().SetTOTPSecret(gomock.Any(), user.ID, gomock.Any()).Return(nil)
	store.EXPECT().SetRecoveryCodes(gomock.Any(), user.ID, gomock.Any()).Return(nil)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/enable", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.Handle2FAEnable(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["secret"] == nil {
		t.Fatal("expected secret in response")
	}
}

func TestHandle2FAEnable_SaveError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().SetTOTPSecret(gomock.Any(), user.ID, gomock.Any()).Return(errors.New("fail"))
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/enable", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.Handle2FAEnable(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandle2FADisable_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.Handle2FADisable(w, httptest.NewRequest("POST", "/api/auth/2fa/disable", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandle2FADisable_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().SetTOTPSecret(gomock.Any(), user.ID, gomock.Nil()).Return(nil)
	store.EXPECT().SetRecoveryCodes(gomock.Any(), user.ID, gomock.Nil()).Return(nil)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/disable", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.Handle2FADisable(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandle2FADisable_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().SetTOTPSecret(gomock.Any(), user.ID, gomock.Nil()).Return(errors.New("fail"))
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/disable", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.Handle2FADisable(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandle2FAVerify_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	h.Handle2FAVerify(w, httptest.NewRequest("POST", "/api/auth/2fa/verify", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandle2FAVerify_BadBody(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	h := &AuthHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/verify", bytes.NewReader([]byte("bad")))
	r = reqWithUserCtx(r, user)
	h.Handle2FAVerify(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandle2FAVerify_NoSecret(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"code": "123456"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/verify", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.Handle2FAVerify(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandle2FAVerify_InvalidCode(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	secret := "JBSWY3DPEHPK3PXP"
	user := &database.User{ID: uuid.New(), Username: "admin", TOTPSecret: &secret}
	h := &AuthHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"code": "000000"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/auth/2fa/verify", bytes.NewReader(body))
	r = reqWithUserCtx(r, user)
	h.Handle2FAVerify(w, r)
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

// =============================================================================
// SessionHandler tests
// =============================================================================

func TestHandleListSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	sessionID := uuid.New()
	sessions := []database.Session{
		{ID: sessionID, UserID: user.ID, CreatedAt: time.Now(), ExpiresAt: time.Now().Add(7 * 24 * time.Hour), IPAddress: "127.0.0.1", UserAgent: "Test"},
	}
	store.EXPECT().ListUserSessions(gomock.Any(), user.ID).Return(sessions, nil)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/sessions", http.NoBody)
	r = reqWithUserCtx(r, user)
	r.AddCookie(&http.Cookie{Name: "lurkarr_session", Value: sessionID.String()})
	h.HandleListSessions(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var result []map[string]any
	json.NewDecoder(w.Body).Decode(&result)
	if len(result) != 1 {
		t.Fatalf("expected 1 session, got %d", len(result))
	}
	if result[0]["current"] != true {
		t.Fatal("expected current session to be marked")
	}
}

func TestHandleListSessions_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListSessions(w, httptest.NewRequest("GET", "/api/sessions", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleListSessions_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().ListUserSessions(gomock.Any(), user.ID).Return(nil, errors.New("db error"))
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/sessions", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.HandleListSessions(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleRevokeSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	sessionID := uuid.New()
	sessions := []database.Session{{ID: sessionID, UserID: user.ID}}
	store.EXPECT().ListUserSessions(gomock.Any(), user.ID).Return(sessions, nil)
	store.EXPECT().DeleteSession(gomock.Any(), sessionID).Return(nil)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/sessions/"+sessionID.String(), nil, "id", sessionID.String())
	r = reqWithUserCtx(r, user)
	h.HandleRevokeSession(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleRevokeSession_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleRevokeSession(w, httptest.NewRequest("DELETE", "/api/sessions/"+uuid.New().String(), http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleRevokeSession_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/sessions/bad", nil, "id", "bad")
	r = reqWithUserCtx(r, user)
	h.HandleRevokeSession(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleRevokeSession_NotOwned(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	otherSessionID := uuid.New()
	sessions := []database.Session{{ID: uuid.New(), UserID: user.ID}} // different ID
	store.EXPECT().ListUserSessions(gomock.Any(), user.ID).Return(sessions, nil)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/sessions/"+otherSessionID.String(), nil, "id", otherSessionID.String())
	r = reqWithUserCtx(r, user)
	h.HandleRevokeSession(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleRevokeAllSessions(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	currentSID := uuid.New()
	store.EXPECT().DeleteUserSessionsExcept(gomock.Any(), user.ID, currentSID).Return(nil)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/sessions", http.NoBody)
	r = reqWithUserCtx(r, user)
	r.AddCookie(&http.Cookie{Name: "lurkarr_session", Value: currentSID.String()})
	h.HandleRevokeAllSessions(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleRevokeAllSessions_NoUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleRevokeAllSessions(w, httptest.NewRequest("DELETE", "/api/sessions", http.NoBody))
	if w.Code != 401 {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestHandleRevokeAllSessions_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "admin"}
	store.EXPECT().DeleteUserSessionsExcept(gomock.Any(), user.ID, gomock.Any()).Return(errors.New("db error"))
	h := &SessionHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("DELETE", "/api/sessions", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.HandleRevokeAllSessions(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// AdminHandler tests
// =============================================================================

func TestHandleListUsers(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	users := []database.User{{ID: uuid.New(), Username: "user1", AuthProvider: "local", IsAdmin: false, CreatedAt: time.Now()}}
	store.EXPECT().ListUsers(gomock.Any()).Return(users, nil)
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/admin/users", http.NoBody)
	r = reqWithUserCtx(r, admin)
	h.HandleListUsers(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleListUsers_NonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "user", IsAdmin: false}
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/admin/users", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.HandleListUsers(w, r)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleListUsers_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	store.EXPECT().ListUsers(gomock.Any()).Return(nil, errors.New("db error"))
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/api/admin/users", http.NoBody)
	r = reqWithUserCtx(r, admin)
	h.HandleListUsers(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	newUser := &database.User{ID: uuid.New(), Username: "newuser"}
	store.EXPECT().CreateUser(gomock.Any(), "newuser", gomock.Any()).Return(newUser, nil)
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"username": "newuser", "password": "StrongPass1!"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/users", bytes.NewReader(body))
	r = reqWithUserCtx(r, admin)
	h.HandleCreateUser(w, r)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestHandleCreateUser_NonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "user", IsAdmin: false}
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/users", http.NoBody)
	r = reqWithUserCtx(r, user)
	h.HandleCreateUser(w, r)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleCreateUser_EmptyFields(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"username": "", "password": ""})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/users", bytes.NewReader(body))
	r = reqWithUserCtx(r, admin)
	h.HandleCreateUser(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateUser_WeakPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"username": "newuser", "password": "short"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/users", bytes.NewReader(body))
	r = reqWithUserCtx(r, admin)
	h.HandleCreateUser(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateUser_WithAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	newUser := &database.User{ID: uuid.New(), Username: "admin2"}
	store.EXPECT().CreateUser(gomock.Any(), "admin2", gomock.Any()).Return(newUser, nil)
	store.EXPECT().UpdateUserAdmin(gomock.Any(), newUser.ID, true).Return(nil)
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"username": "admin2", "password": "StrongPass1!", "is_admin": true})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/users", bytes.NewReader(body))
	r = reqWithUserCtx(r, admin)
	h.HandleCreateUser(w, r)
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestHandleCreateUser_Duplicate(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	store.EXPECT().CreateUser(gomock.Any(), "existing", gomock.Any()).Return(nil, errors.New("duplicate"))
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]any{"username": "existing", "password": "StrongPass1!"})
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/api/admin/users", bytes.NewReader(body))
	r = reqWithUserCtx(r, admin)
	h.HandleCreateUser(w, r)
	if w.Code != 409 {
		t.Fatalf("expected 409, got %d", w.Code)
	}
}

func TestHandleDeleteUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	targetID := uuid.New()
	store.EXPECT().DeleteUser(gomock.Any(), targetID).Return(nil)
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/admin/users/"+targetID.String(), nil, "id", targetID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleDeleteUser(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDeleteUser_SelfDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/admin/users/"+admin.ID.String(), nil, "id", admin.ID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleDeleteUser(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteUser_NonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "user", IsAdmin: false}
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/admin/users/"+uuid.New().String(), nil, "id", uuid.New().String())
	r = reqWithUserCtx(r, user)
	h.HandleDeleteUser(w, r)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleResetUserPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	targetID := uuid.New()
	store.EXPECT().UpdatePassword(gomock.Any(), targetID, gomock.Any()).Return(nil)
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"password": "NewStrongPass1!"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+targetID.String()+"/reset-password", body, "id", targetID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleResetUserPassword(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleResetUserPassword_NonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "user", IsAdmin: false}
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+uuid.New().String()+"/reset-password", nil, "id", uuid.New().String())
	r = reqWithUserCtx(r, user)
	h.HandleResetUserPassword(w, r)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleResetUserPassword_WeakPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	targetID := uuid.New()
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"password": "weak"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+targetID.String()+"/reset-password", body, "id", targetID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleResetUserPassword(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleToggleAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	targetID := uuid.New()
	store.EXPECT().UpdateUserAdmin(gomock.Any(), targetID, true).Return(nil)
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]bool{"is_admin": true})
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+targetID.String()+"/toggle-admin", body, "id", targetID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleToggleAdmin(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleToggleAdmin_SelfDemotion(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]bool{"is_admin": false})
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+admin.ID.String()+"/toggle-admin", body, "id", admin.ID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleToggleAdmin(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleToggleAdmin_NonAdmin(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	user := &database.User{ID: uuid.New(), Username: "user", IsAdmin: false}
	h := &AdminHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+uuid.New().String()+"/toggle-admin", nil, "id", uuid.New().String())
	r = reqWithUserCtx(r, user)
	h.HandleToggleAdmin(w, r)
	if w.Code != 403 {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}

func TestHandleToggleAdmin_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	admin := &database.User{ID: uuid.New(), Username: "admin", IsAdmin: true}
	targetID := uuid.New()
	store.EXPECT().UpdateUserAdmin(gomock.Any(), targetID, true).Return(errors.New("db error"))
	h := &AdminHandler{DB: store}
	body, _ := json.Marshal(map[string]bool{"is_admin": true})
	w := httptest.NewRecorder()
	r := reqWithPathValue("POST", "/api/admin/users/"+targetID.String()+"/toggle-admin", body, "id", targetID.String())
	r = reqWithUserCtx(r, admin)
	h.HandleToggleAdmin(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
