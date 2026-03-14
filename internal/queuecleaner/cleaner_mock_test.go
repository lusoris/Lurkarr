package queuecleaner

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

func defaultQCSettings() *database.QueueCleanerSettings {
	return &database.QueueCleanerSettings{
		Enabled:               true,
		CheckIntervalSeconds:  60,
		MaxStrikes:            3,
		StrikeWindowHours:     24,
		RemoveFromClient:      true,
		BlocklistOnRemove:     true,
		StrikePublic:          true,
		StrikePrivate:         true,
		FailedImportRemove:    true,
		FailedImportBlocklist: true,
	}
}

func defaultGeneralSettings() *database.GeneralSettings {
	return &database.GeneralSettings{APITimeout: 30, SSLVerify: true}
}

func TestNewCleaner(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	c := New(store, nil)
	if c == nil {
		t.Fatal("New() returned nil")
	}
}

func TestCleanerStartStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetQueueCleanerSettings(gomock.Any(), gomock.Any()).Return(defaultQCSettings(), nil).AnyTimes()
	store.EXPECT().ListEnabledInstances(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	ctx := context.Background()
	c.Start(ctx)
	c.Stop()
}

func TestCleanerStopNilCancel(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	c := New(store, nil)
	c.Stop() // Should not panic
}

func TestCleanLoopCancellation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetQueueCleanerSettings(gomock.Any(), database.AppSonarr).Return(defaultQCSettings(), nil).AnyTimes()
	store.EXPECT().ListEnabledInstances(gomock.Any(), database.AppSonarr).Return(nil, nil).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c.wg.Add(1)
	go c.cleanLoop(ctx, database.AppSonarr)
	c.wg.Wait()
}

func TestCleanLoopSettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetQueueCleanerSettings(gomock.Any(), database.AppSonarr).Return(nil, context.DeadlineExceeded).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	c.wg.Add(1)
	c.cleanLoop(ctx, database.AppSonarr)
}

func TestCleanLoopDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	disabled := defaultQCSettings()
	disabled.Enabled = false
	store.EXPECT().GetQueueCleanerSettings(gomock.Any(), database.AppSonarr).Return(disabled, nil).AnyTimes()

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	c.wg.Add(1)
	c.cleanLoop(ctx, database.AppSonarr)
}

func TestCleanInstance_EmptyQueue(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	log := logger.ForApp("sonarr")
	inst := database.AppInstance{ID: uuid.New(), Name: "test", APIURL: srv.URL, APIKey: "k"}

	c.cleanInstance(context.Background(), log, database.AppSonarr, defaultQCSettings(), inst)
}

func TestCleanInstance_StalledItem(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/queue" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(arrclient.QueueResponse{
				Records: []arrclient.QueueRecord{
					{
						ID:                    1,
						Title:                 "Stalled.Download",
						DownloadID:            "dl-1",
						Status:                "warning",
						TrackedDownloadStatus: "warning",
						TrackedDownloadState:  "downloading",
						Protocol:              "torrent",
						Size:                  1000,
						Sizeleft:              900,
					},
				},
			})
			return
		}
		// DELETE queue item
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	instID := uuid.New()

	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().ListEnabledBlocklistRules(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppSonarr).Return(&database.ScoringProfile{}, nil)
	store.EXPECT().ListEnabledDownloadClientInstances(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{Enabled: false}, nil)
	store.EXPECT().AddStrikeAndCount(gomock.Any(), database.AppSonarr, instID, "dl-1", gomock.Any(), gomock.Any(), gomock.Any()).Return(3, nil)
	store.EXPECT().LogBlocklist(gomock.Any(), database.AppSonarr, instID, "dl-1", gomock.Any(), gomock.Any()).Return(nil)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	log := logger.ForApp("sonarr")
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"}

	c.cleanInstance(context.Background(), log, database.AppSonarr, defaultQCSettings(), inst)
}

func TestCleanInstance_ProgressResetStrikes(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{
			Records: []arrclient.QueueRecord{
				{
					ID:                    1,
					Title:                 "Good.Download",
					DownloadID:            "dl-good",
					Size:                  1000,
					Sizeleft:              100, // >50% done
					TrackedDownloadStatus: "ok",
					TrackedDownloadState:  "downloading",
				},
			},
		})
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	instID := uuid.New()

	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().ListEnabledBlocklistRules(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppSonarr).Return(&database.ScoringProfile{}, nil)
	store.EXPECT().ListEnabledDownloadClientInstances(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{Enabled: false}, nil)
	store.EXPECT().ResetStrikes(gomock.Any(), database.AppSonarr, instID, "dl-good").Return(nil)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	log := logger.ForApp("sonarr")
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"}

	c.cleanInstance(context.Background(), log, database.AppSonarr, defaultQCSettings(), inst)
}

func TestCleanInstance_FailedImports(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v3/queue" && r.Method == http.MethodGet {
			json.NewEncoder(w).Encode(arrclient.QueueResponse{
				Records: []arrclient.QueueRecord{
					{
						ID:                    1,
						Title:                 "Failed.Import",
						DownloadID:            "dl-fail",
						TrackedDownloadStatus: "warning",
						TrackedDownloadState:  "importPending",
						StatusMessages: []arrclient.StatusMessage{
							{Messages: []string{"Import failed: no files found"}},
						},
					},
				},
			})
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	instID := uuid.New()

	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(defaultGeneralSettings(), nil)
	store.EXPECT().ListEnabledBlocklistRules(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetScoringProfile(gomock.Any(), database.AppSonarr).Return(&database.ScoringProfile{}, nil)
	store.EXPECT().ListEnabledDownloadClientInstances(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{Enabled: false}, nil)
	store.EXPECT().LogBlocklist(gomock.Any(), database.AppSonarr, instID, "dl-fail", "Failed.Import", gomock.Any()).Return(nil)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	log := logger.ForApp("sonarr")
	inst := database.AppInstance{ID: instID, Name: "test", APIURL: srv.URL, APIKey: "k"}

	c.cleanInstance(context.Background(), log, database.AppSonarr, defaultQCSettings(), inst)
}

func TestCleanInstance_GeneralSettingsError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetGeneralSettings(gomock.Any()).Return(nil, context.DeadlineExceeded)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	log := logger.ForApp("sonarr")
	inst := database.AppInstance{ID: uuid.New(), Name: "test", APIURL: "http://localhost", APIKey: "k"}

	// Should not panic
	c.cleanInstance(context.Background(), log, database.AppSonarr, defaultQCSettings(), inst)
}

func TestCleanInstance_UnsupportedAppType(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	log := logger.ForApp("prowlarr")
	inst := database.AppInstance{ID: uuid.New(), Name: "test"}

	// LurkerFor returns nil for prowlarr
	c.cleanInstance(context.Background(), log, database.AppProwlarr, defaultQCSettings(), inst)
}

func TestHasImportFailureVariants(t *testing.T) {
	tests := []struct {
		name   string
		record arrclient.QueueRecord
		want   bool
	}{
		{
			name: "no warning status",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "ok",
				TrackedDownloadState:  "importPending",
			},
			want: false,
		},
		{
			name: "wrong state",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "downloading",
			},
			want: false,
		},
		{
			name: "import failed message",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importFailed",
				StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"Import failed"}}},
			},
			want: true,
		},
		{
			name: "unable to import",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"Unable to import"}}},
			},
			want: true,
		},
		{
			name: "no files found",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"No files found"}}},
			},
			want: true,
		},
		{
			name: "sample file",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"File is a sample"}}},
			},
			want: true,
		},
		{
			name: "not a valid",
			record: arrclient.QueueRecord{
				TrackedDownloadStatus: "warning",
				TrackedDownloadState:  "importPending",
				StatusMessages:        []arrclient.StatusMessage{{Messages: []string{"Not a valid media file"}}},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasImportFailure(tt.record)
			if got != tt.want {
				t.Errorf("hasImportFailure() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestImportFailureReason_Extended(t *testing.T) {
	record := arrclient.QueueRecord{
		StatusMessages: []arrclient.StatusMessage{
			{Messages: []string{"Import failed: bad quality"}},
		},
	}
	r := importFailureReason(record)
	if r != "Import failed: bad quality" {
		t.Errorf("importFailureReason() = %q", r)
	}

	empty := arrclient.QueueRecord{}
	r = importFailureReason(empty)
	if r != "unknown_import_failure" {
		t.Errorf("importFailureReason(empty) = %q", r)
	}
}

func TestGetSABnzbdStatuses_Disabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListEnabledDownloadClientInstances(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(&database.SABnzbdSettings{Enabled: false}, nil)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	statuses := c.getSABnzbdStatuses(context.Background())
	if len(statuses) != 0 {
		t.Errorf("expected empty statuses when SABnzbd disabled, got %d", len(statuses))
	}
}

func TestGetSABnzbdStatuses_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListEnabledDownloadClientInstances(gomock.Any()).Return(nil, nil)
	store.EXPECT().GetSABnzbdSettings(gomock.Any()).Return(nil, context.DeadlineExceeded)

	logger := newTestLogger()
	defer logger.Close()

	c := New(store, logger)
	statuses := c.getSABnzbdStatuses(context.Background())
	if len(statuses) != 0 {
		t.Errorf("expected empty statuses on error, got %d", len(statuses))
	}
}
