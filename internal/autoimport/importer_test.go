package autoimport

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
)

func newTestLogger() *logging.Logger {
	return logging.New()
}

func defaultGeneralSettings() *database.GeneralSettings {
	return &database.GeneralSettings{APITimeout: 30, SSLVerify: true}
}

func TestIsImportStuck(t *testing.T) {
	tests := []struct {
		name   string
		record arrclient.QueueRecord
		want   bool
	}{
		{
			name: "not importPending",
			record: arrclient.QueueRecord{
				TrackedDownloadState: "downloading",
			},
			want: false,
		},
		{
			name: "importPending with unable to import",
			record: arrclient.QueueRecord{
				TrackedDownloadState: "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Messages: []string{"Unable to import - file not found"}},
				},
			},
			want: true,
		},
		{
			name: "importPending with import failed",
			record: arrclient.QueueRecord{
				TrackedDownloadState: "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Messages: []string{"Import failed for some reason"}},
				},
			},
			want: true,
		},
		{
			name: "importPending with no matching",
			record: arrclient.QueueRecord{
				TrackedDownloadState: "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Messages: []string{"No matching series found"}},
				},
			},
			want: true,
		},
		{
			name: "importPending with warning status only",
			record: arrclient.QueueRecord{
				TrackedDownloadState:  "importPending",
				TrackedDownloadStatus: "warning",
			},
			want: true,
		},
		{
			name: "importPending with unrelated messages",
			record: arrclient.QueueRecord{
				TrackedDownloadState: "importPending",
				StatusMessages: []arrclient.StatusMessage{
					{Messages: []string{"some other message"}},
				},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isImportStuck(tt.record)
			if got != tt.want {
				t.Errorf("isImportStuck() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatStatusMessages(t *testing.T) {
	msgs := []arrclient.StatusMessage{
		{Messages: []string{"msg1", "msg2"}},
		{Messages: []string{"msg3"}},
	}
	got := formatStatusMessages(msgs)
	if got != "msg1; msg2; msg3" {
		t.Errorf("formatStatusMessages() = %q", got)
	}

	// Empty
	got = formatStatusMessages(nil)
	if got != "" {
		t.Errorf("formatStatusMessages(nil) = %q", got)
	}
}

func TestAPIVersionFor(t *testing.T) {
	tests := []struct {
		appType database.AppType
		want    string
	}{
		{database.AppLidarr, "v1"},
		{database.AppReadarr, "v1"},
		{database.AppSonarr, "v3"},
		{database.AppRadarr, "v3"},
		{database.AppWhisparr, "v3"},
		{database.AppEros, "v3"},
	}
	for _, tt := range tests {
		t.Run(string(tt.appType), func(t *testing.T) {
			got := apiVersionFor(tt.appType)
			if got != tt.want {
				t.Errorf("apiVersionFor(%s) = %q, want %q", tt.appType, got, tt.want)
			}
		})
	}
}

func TestTriggerRescan(t *testing.T) {
	tests := []struct {
		name    string
		appType database.AppType
	}{
		{"sonarr", database.AppSonarr},
		{"radarr", database.AppRadarr},
		{"whisparr", database.AppWhisparr},
		{"eros", database.AppEros},
		{"lidarr", database.AppLidarr},
		{"readarr", database.AppReadarr},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1})
			}))
			defer server.Close()

			client := arrclient.NewClient(server.URL, "key", 5*time.Second, true)
			err := triggerRescan(context.Background(), client, tt.appType, 42)
			if err != nil {
				t.Errorf("triggerRescan(%s) error: %v", tt.appType, err)
			}
		})
	}
}

func TestTriggerRescanUnsupportedType(t *testing.T) {
	client := arrclient.NewClient("http://localhost", "key", 5*time.Second, true)
	err := triggerRescan(context.Background(), client, database.AppProwlarr, 1)
	if err != nil {
		t.Errorf("expected nil for unsupported type, got: %v", err)
	}
}

func TestSleep(t *testing.T) {
	ok := sleep(context.Background(), 1*time.Millisecond)
	if !ok {
		t.Error("expected true")
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok = sleep(ctx, 1*time.Minute)
	if ok {
		t.Error("expected false with cancelled context")
	}
}

func TestNewImporter(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	imp := New(store, nil)
	if imp == nil {
		t.Fatal("New() returned nil")
	}
}

func TestImporterStopNilCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	imp := New(store, nil)
	// Should not panic
	imp.Stop()
}

func TestImporterStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListEnabledInstances(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx := context.Background()
	imp.Start(ctx)
	imp.Stop()
}

func TestCheckInstance_NoInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx := context.Background()

	instID := uuid.New()
	inst := database.AppInstance{
		ID:     instID,
		Name:   "test-sonarr",
		APIURL: "http://unreachable:9999",
		APIKey: "fake",
	}
	log := logger.ForApp("sonarr")
	// This will fail to connect to the API, which is expected
	imp.checkInstance(ctx, log, database.AppSonarr, inst)
}

func TestCheckInstance_WithStuckImport(t *testing.T) {
	queueResp := arrclient.QueueResponse{
		Records: []arrclient.QueueRecord{
			{
				ID:                    1,
				Title:                 "Test.Movie.2024",
				TrackedDownloadState:  "importPending",
				TrackedDownloadStatus: "warning",
				SeriesID:              42,
				DownloadID:            "dl-123",
				StatusMessages: []arrclient.StatusMessage{
					{Messages: []string{"Unable to import - file not found"}},
				},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(queueResp)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().LogAutoImport(gomock.Any(), database.AppSonarr, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx := context.Background()

	instID := uuid.New()
	inst := database.AppInstance{
		ID:     instID,
		Name:   "test-sonarr",
		APIURL: srv.URL,
		APIKey: "fake",
	}
	log := logger.ForApp("sonarr")
	imp.checkInstance(ctx, log, database.AppSonarr, inst)
}

func TestCheckInstance_GeneralSettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, context.DeadlineExceeded)

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx := context.Background()

	inst := database.AppInstance{
		ID:     uuid.New(),
		Name:   "test",
		APIURL: "http://localhost",
		APIKey: "key",
	}
	log := logger.ForApp("sonarr")
	// Should return early without panic
	imp.checkInstance(ctx, log, database.AppSonarr, inst)
}

func TestCheckInstance_UnsupportedAppType(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx := context.Background()

	inst := database.AppInstance{ID: uuid.New(), Name: "test"}
	log := logger.ForApp("prowlarr")
	// LurkerFor returns nil for prowlarr → should return early
	imp.checkInstance(ctx, log, database.AppProwlarr, inst)
}

func TestImportLoopCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // immediate cancellation

	imp.wg.Add(1)
	go imp.importLoop(ctx, database.AppSonarr)
	imp.wg.Wait() // Should return immediately
}

func TestImportLoop_WithInstances(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	instID := uuid.New()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{})
	}))
	defer srv.Close()

	store.EXPECT().ListEnabledInstances(gomock.Any(), database.AppSonarr).
		Return([]database.AppInstance{
			{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"},
		}, nil).AnyTimes()
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	imp.wg.Add(1)
	go func() {
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		imp.importLoop(ctx, database.AppSonarr)
	}()
	imp.wg.Wait()
}

func TestImportLoop_ListInstancesError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	store.EXPECT().ListEnabledInstances(gomock.Any(), database.AppSonarr).
		Return(nil, context.DeadlineExceeded).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	imp.wg.Add(1)
	go func() {
		go func() {
			time.Sleep(50 * time.Millisecond)
			cancel()
		}()
		imp.importLoop(ctx, database.AppSonarr)
	}()
	imp.wg.Wait()
}

func TestCheckInstance_ManualImportAvailable(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/queue" {
			json.NewEncoder(w).Encode(arrclient.QueueResponse{
				Records: []arrclient.QueueRecord{
					{
						ID:                    1,
						Title:                 "Test.Movie.2024",
						TrackedDownloadState:  "importPending",
						TrackedDownloadStatus: "warning",
						SeriesID:              42,
						DownloadID:            "dl-123",
						StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"Unable to import"}}},
					},
				},
			})
			return
		}
		if r.URL.Path == "/api/v3/manualimport" {
			json.NewEncoder(w).Encode([]arrclient.ManualImportItem{
				{Name: "file1.mkv", CustomFormatScore: 100, Rejections: nil},
			})
			return
		}
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().LogAutoImport(gomock.Any(), database.AppSonarr, gomock.Any(), gomock.Any(), "Test.Movie.2024", 1, "manual_import_available", "file1.mkv").Return(nil)

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)

	inst := database.AppInstance{ID: uuid.New(), Name: "s", APIURL: srv.URL, APIKey: "k"}
	log := logger.ForApp("sonarr")
	imp.checkInstance(context.Background(), log, database.AppSonarr, inst)
}

func TestCheckInstance_FallbackRescan(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/queue" {
			json.NewEncoder(w).Encode(arrclient.QueueResponse{
				Records: []arrclient.QueueRecord{
					{
						ID:                    1,
						Title:                 "Test.Movie",
						TrackedDownloadState:  "importPending",
						TrackedDownloadStatus: "warning",
						SeriesID:              42,
						DownloadID:            "",
						StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"Unable to import"}}},
					},
				},
			})
			return
		}
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().LogAutoImport(gomock.Any(), database.AppSonarr, gomock.Any(), 42, "Test.Movie", 1, "rescan_triggered", gomock.Any()).Return(nil)

	logger := newTestLogger()
	defer logger.Close()

	imp := New(store, logger)

	inst := database.AppInstance{ID: uuid.New(), Name: "s", APIURL: srv.URL, APIKey: "k"}
	log := logger.ForApp("sonarr")
	imp.checkInstance(context.Background(), log, database.AppSonarr, inst)
}
