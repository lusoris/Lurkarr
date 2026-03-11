package lurking

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/mocks"
)

func newTestLogger(ctrl *gomock.Controller) *logging.Logger {
	logStore := mocks.NewMockLogStore(ctrl)
	logStore.EXPECT().InsertLogs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	hub := logging.NewHub()
	return logging.New(logStore, hub)
}

// defaultGeneralSettings returns a standard GeneralSettings for tests.
func defaultGeneralSettings() *database.GeneralSettings {
	return &database.GeneralSettings{
		APITimeout:         30,
		SSLVerify:          true,
		StatefulResetHours: 24,
	}
}

func TestEngineNewAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	logger := newTestLogger(ctrl)
	defer logger.Close()

	e := New(store, logger)
	if e == nil {
		t.Fatal("New() returned nil")
	}
	e.Stop() // nil cancel should not panic
}

func TestEngineStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	// lurkLoop calls GetAppSettings for each app type; return defaults
	store.EXPECT().GetAppSettings(gomock.Any(), gomock.Any()).Return(
		&database.AppSettings{SleepDuration: 1, HourlyCap: 10, LurkMissingCount: 5, LurkUpgradeCount: 5}, nil,
	).AnyTimes()
	store.EXPECT().ListEnabledInstances(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	logger := newTestLogger(ctrl)
	defer logger.Close()

	e := New(store, logger)
	ctx := context.Background()
	e.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	e.Stop()
}

func TestEngineSleep(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)

	ok := e.sleep(context.Background(), 1*time.Millisecond)
	if !ok {
		t.Error("expected true from sleep")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok = e.sleep(ctx, time.Minute)
	if ok {
		t.Error("expected false from cancelled sleep")
	}
}

// arrServer creates a fake arr API server for testing.
func arrServer(t *testing.T, missingItems, cutoffItems []arrclient.SonarrEpisode) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/v3/wanted/missing":
			json.NewEncoder(w).Encode(struct {
				TotalRecords int                       `json:"totalRecords"`
				Records      []arrclient.SonarrEpisode `json:"records"`
			}{TotalRecords: len(missingItems), Records: missingItems})
		case r.URL.Path == "/api/v3/wanted/cutoff":
			json.NewEncoder(w).Encode(struct {
				TotalRecords int                       `json:"totalRecords"`
				Records      []arrclient.SonarrEpisode `json:"records"`
			}{TotalRecords: len(cutoffItems), Records: cutoffItems})
		case r.URL.Path == "/api/v3/queue":
			json.NewEncoder(w).Encode(arrclient.QueueResponse{})
		default:
			json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1})
		}
	})
	return httptest.NewServer(mux)
}

func TestLurkInstance_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	missing := []arrclient.SonarrEpisode{
		{ID: 1, Title: "Episode 1"},
		{ID: 2, Title: "Episode 2"},
	}
	srv := arrServer(t, missing, nil)
	defer srv.Close()

	instID := uuid.New()
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"}

	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(0, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().GetLastReset(gomock.Any(), database.AppSonarr, instID).Return(nil, nil)
	store.EXPECT().IsProcessed(gomock.Any(), database.AppSonarr, instID, gomock.Any(), "missing").Return(false, nil).Times(2)
	store.EXPECT().MarkProcessed(gomock.Any(), database.AppSonarr, instID, gomock.Any(), "missing").Return(nil).Times(2)
	store.EXPECT().AddLurkHistory(gomock.Any(), database.AppSonarr, instID, "test", gomock.Any(), gomock.Any(), "missing").Return(nil).Times(2)
	store.EXPECT().IncrementStats(gomock.Any(), database.AppSonarr, instID, int64(2), int64(0)).Return(nil)
	store.EXPECT().IncrementHourlyHits(gomock.Any(), database.AppSonarr, instID, 2).Return(nil)

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{
		SleepDuration:    1,
		HourlyCap:        10,
		LurkMissingCount: 5,
		LurkUpgradeCount: 5,
	}

	err := e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	if err != nil {
		t.Fatalf("lurkInstance error: %v", err)
	}
}

func TestLurkInstance_HourlyCapReached(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	instID := uuid.New()
	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(100, nil)

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10}
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: "http://cant-reach", APIKey: "k"}

	err := e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	if err != nil {
		t.Fatalf("expected nil error when cap reached, got: %v", err)
	}
}

func TestLurkInstance_HourHitsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	instID := uuid.New()
	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(0, context.DeadlineExceeded)

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10}
	inst := database.AppInstance{ID: instID, Name: "test"}

	err := e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLurkInstance_GeneralSettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	instID := uuid.New()
	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(0, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, context.DeadlineExceeded)

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10}
	inst := database.AppInstance{ID: instID, Name: "test"}

	err := e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLurkInstance_LastResetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	instID := uuid.New()
	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(0, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().GetLastReset(gomock.Any(), database.AppSonarr, instID).Return(nil, context.DeadlineExceeded)

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10}
	inst := database.AppInstance{ID: instID, Name: "test"}

	err := e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	if err == nil {
		t.Error("expected error")
	}
}

func TestLurkInstance_StateReset(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	old := time.Now().Add(-48 * time.Hour)
	instID := uuid.New()

	missing := []arrclient.SonarrEpisode{{ID: 1, Title: "Ep1"}}
	srv := arrServer(t, missing, nil)
	defer srv.Close()

	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(0, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{
		APITimeout: 30, SSLVerify: true, StatefulResetHours: 24,
	}, nil)
	store.EXPECT().GetLastReset(gomock.Any(), database.AppSonarr, instID).Return(&old, nil)
	store.EXPECT().ResetState(gomock.Any(), database.AppSonarr, instID).Return(nil)
	store.EXPECT().IsProcessed(gomock.Any(), database.AppSonarr, instID, gomock.Any(), "missing").Return(false, nil).AnyTimes()
	store.EXPECT().MarkProcessed(gomock.Any(), database.AppSonarr, instID, gomock.Any(), "missing").Return(nil).AnyTimes()
	store.EXPECT().AddLurkHistory(gomock.Any(), database.AppSonarr, instID, gomock.Any(), gomock.Any(), gomock.Any(), "missing").Return(nil).AnyTimes()
	store.EXPECT().IncrementStats(gomock.Any(), database.AppSonarr, instID, gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	store.EXPECT().IncrementHourlyHits(gomock.Any(), database.AppSonarr, instID, gomock.Any()).Return(nil).AnyTimes()

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10, LurkMissingCount: 5}
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"}

	_ = e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	// ResetState expectation with Times(1) (implicit) verifies it was called
}

func TestLurkInstance_MinDownloadQueueSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	// Server returns queue with 10 items (above threshold)
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/queue" {
			json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 10})
			return
		}
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	instID := uuid.New()
	store.EXPECT().GetCurrentHourHits(gomock.Any(), database.AppSonarr, instID).Return(0, nil)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(&database.GeneralSettings{
		APITimeout: 30, SSLVerify: true, StatefulResetHours: 24, MinDownloadQueueSize: 5,
	}, nil)
	store.EXPECT().GetLastReset(gomock.Any(), database.AppSonarr, instID).Return(nil, nil)
	// No IncrementStats or AddLurkHistory expected — queue at capacity

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10, LurkMissingCount: 5}
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"}

	err := e.lurkInstance(context.Background(), log, database.AppSonarr, settings, inst)
	if err != nil {
		t.Fatalf("lurkInstance error: %v", err)
	}
}

func TestLurkMissing_FetchError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	// Server that returns errors
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10, LurkMissingCount: 5}
	inst := database.AppInstance{ID: uuid.New(), Name: "test", APIURL: srv.URL, APIKey: "k"}
	client := arrclient.NewClient(srv.URL, "k", 5*time.Second, true)

	count := e.lurkMissing(context.Background(), log, database.AppSonarr, settings, inst, client, 5)
	if count != 0 {
		t.Errorf("expected 0 lurked on error, got %d", count)
	}
}

func TestLurkUpgrades_NoUpgrades(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	srv := arrServer(t, nil, nil)
	defer srv.Close()

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10, LurkUpgradeCount: 5}
	inst := database.AppInstance{ID: uuid.New(), Name: "test"}
	client := arrclient.NewClient(srv.URL, "k", 5*time.Second, true)

	count := e.lurkUpgrades(context.Background(), log, database.AppSonarr, settings, inst, client, 5)
	if count != 0 {
		t.Errorf("expected 0, got %d", count)
	}
}

func TestLurkUpgrades_WithItems(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	cutoff := []arrclient.SonarrEpisode{
		{ID: 10, Title: "Upgrade1"},
		{ID: 11, Title: "Upgrade2"},
	}
	srv := arrServer(t, nil, cutoff)
	defer srv.Close()

	instID := uuid.New()
	store.EXPECT().IsProcessed(gomock.Any(), database.AppSonarr, instID, gomock.Any(), "upgrade").Return(false, nil).Times(2)
	store.EXPECT().MarkProcessed(gomock.Any(), database.AppSonarr, instID, gomock.Any(), "upgrade").Return(nil).Times(2)
	store.EXPECT().AddLurkHistory(gomock.Any(), database.AppSonarr, instID, "test", gomock.Any(), gomock.Any(), "upgrade").Return(nil).Times(2)
	store.EXPECT().IncrementStats(gomock.Any(), database.AppSonarr, instID, int64(0), int64(2)).Return(nil)
	store.EXPECT().IncrementHourlyHits(gomock.Any(), database.AppSonarr, instID, 2).Return(nil)

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	log := logger.ForApp("sonarr")
	settings := &database.AppSettings{HourlyCap: 10, LurkUpgradeCount: 5}
	inst := database.AppInstance{ID: instID, Name: "test"}
	client := arrclient.NewClient(srv.URL, "k", 5*time.Second, true)

	count := e.lurkUpgrades(context.Background(), log, database.AppSonarr, settings, inst, client, 5)
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

func TestGetMissingItems_NilLurker(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)

	items, err := e.getMissingItems(context.Background(), database.AppProwlarr, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if items != nil {
		t.Errorf("expected nil items for prowlarr")
	}
}

func TestGetUpgradeItems_NilLurker(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)

	items, err := e.getUpgradeItems(context.Background(), database.AppProwlarr, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if items != nil {
		t.Errorf("expected nil items for prowlarr")
	}
}

func TestTriggerSearch_NilLurker(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)

	err := e.triggerSearch(context.Background(), database.AppProwlarr, nil, 1)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestLurkLoopSettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppSonarr).Return(nil, context.DeadlineExceeded).AnyTimes()

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	e.lurkLoop(ctx, database.AppSonarr) // should exit via cancel
}

func TestLurkLoopInstancesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppSonarr).Return(
		&database.AppSettings{SleepDuration: 1, HourlyCap: 10}, nil,
	).AnyTimes()
	store.EXPECT().ListEnabledInstances(gomock.Any(), database.AppSonarr).Return(nil, context.DeadlineExceeded).AnyTimes()

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	e.lurkLoop(ctx, database.AppSonarr)
}

func TestLurkLoopNoInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppSonarr).Return(
		&database.AppSettings{SleepDuration: 1, HourlyCap: 10}, nil,
	).AnyTimes()
	store.EXPECT().ListEnabledInstances(gomock.Any(), database.AppSonarr).Return(nil, nil).AnyTimes()

	logger := newTestLogger(ctrl)
	defer logger.Close()
	e := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	e.lurkLoop(ctx, database.AppSonarr) // zero instances → sleep → cancel
}
