package scheduler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
)

func newTestLogger() *logging.Logger {
	return logging.New()
}

func newTestScheduler(ctrl *gomock.Controller, store Store) *Scheduler {
	cron, _ := gocron.NewScheduler(gocron.WithLocation(time.UTC))
	return &Scheduler{
		db:     store,
		logger: newTestLogger(),
		cron:   cron,
	}
}

// --- Tests ---

func TestBuildCronExpr(t *testing.T) {
	tests := []struct {
		name  string
		sched database.Schedule
		want  string
	}{
		{
			name:  "every day at 3:30",
			sched: database.Schedule{Hour: 3, Minute: 30},
			want:  "30 3 * * *",
		},
		{
			name:  "specific days",
			sched: database.Schedule{Hour: 14, Minute: 0, Days: []string{"MON", "WED", "FRI"}},
			want:  "0 14 * * MON,WED,FRI",
		},
		{
			name:  "midnight empty days",
			sched: database.Schedule{Hour: 0, Minute: 0, Days: []string{}},
			want:  "0 0 * * *",
		},
		{
			name:  "weekdays at midnight",
			sched: database.Schedule{Hour: 0, Minute: 0, Days: []string{"1", "2", "3", "4", "5"}},
			want:  "0 0 * * 1,2,3,4,5",
		},
		{
			name:  "single day at 23:59",
			sched: database.Schedule{Hour: 23, Minute: 59, Days: []string{"0"}},
			want:  "59 23 * * 0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := buildCronExpr(tt.sched)
			if got != tt.want {
				t.Errorf("buildCronExpr() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestPerformAction_Disable(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	inst := database.AppInstance{ID: uuid.New(), Name: "test", Enabled: true}
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return([]database.AppInstance{inst}, nil)
	store.EXPECT().UpdateInstance(gomock.Any(), inst.ID, inst.Name, inst.APIURL, inst.APIKey, false).Return(nil)

	s := newTestScheduler(ctrl, store)
	err := s.performAction(context.Background(), database.Schedule{Action: "disable", AppType: "sonarr"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPerformAction_Enable(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	inst := database.AppInstance{ID: uuid.New(), Name: "test", Enabled: false}
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("radarr")).Return([]database.AppInstance{inst}, nil)
	store.EXPECT().UpdateInstance(gomock.Any(), inst.ID, inst.Name, inst.APIURL, inst.APIKey, true).Return(nil)

	s := newTestScheduler(ctrl, store)
	err := s.performAction(context.Background(), database.Schedule{Action: "enable", AppType: "radarr"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPerformAction_APICap(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	store.EXPECT().GetAppSettings(gomock.Any(), database.AppType("sonarr")).Return(&database.AppSettings{AppType: "sonarr", HourlyCap: 10}, nil)
	store.EXPECT().UpdateAppSettings(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s *database.AppSettings) error {
		if s.HourlyCap != 25 {
			t.Fatalf("expected HourlyCap=25, got %d", s.HourlyCap)
		}
		return nil
	})

	s := newTestScheduler(ctrl, store)
	err := s.performAction(context.Background(), database.Schedule{Action: "api-25", AppType: "sonarr"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPerformAction_Unknown(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	s := newTestScheduler(ctrl, store)
	err := s.performAction(context.Background(), database.Schedule{Action: "nope", AppType: "sonarr"})
	if err == nil {
		t.Fatal("expected error for unknown action")
	}
}

func TestSetAppEnabled_Global(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	for _, at := range database.AllAppTypes() {
		store.EXPECT().ListInstances(gomock.Any(), at).Return(nil, nil)
	}

	s := newTestScheduler(ctrl, store)
	err := s.setAppEnabled(context.Background(), "global", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetAppEnabled_InvalidAppType(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	s := newTestScheduler(ctrl, store)
	err := s.setAppEnabled(context.Background(), "bogus", true)
	if err == nil {
		t.Fatal("expected error for invalid app type")
	}
}

func TestSetAppEnabled_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("db down"))

	s := newTestScheduler(ctrl, store)
	err := s.setAppEnabled(context.Background(), "sonarr", true)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetAppEnabled_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	inst := database.AppInstance{ID: uuid.New()}
	store.EXPECT().ListInstances(gomock.Any(), database.AppType("radarr")).Return([]database.AppInstance{inst}, nil)
	store.EXPECT().UpdateInstance(gomock.Any(), inst.ID, inst.Name, inst.APIURL, inst.APIKey, false).Return(errors.New("update fail"))

	s := newTestScheduler(ctrl, store)
	err := s.setAppEnabled(context.Background(), "radarr", false)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetHourlyCap_SingleApp(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	store.EXPECT().GetAppSettings(gomock.Any(), database.AppType("lidarr")).Return(&database.AppSettings{AppType: "lidarr", HourlyCap: 5}, nil)
	store.EXPECT().UpdateAppSettings(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, s *database.AppSettings) error {
		if s.HourlyCap != 42 {
			t.Fatalf("expected 42, got %d", s.HourlyCap)
		}
		return nil
	})

	s := newTestScheduler(ctrl, store)
	err := s.setHourlyCap(context.Background(), "lidarr", 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetHourlyCap_Global(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	for _, at := range database.AllAppTypes() {
		store.EXPECT().GetAppSettings(gomock.Any(), at).Return(&database.AppSettings{AppType: at}, nil)
		store.EXPECT().UpdateAppSettings(gomock.Any(), gomock.Any()).Return(nil)
	}

	s := newTestScheduler(ctrl, store)
	err := s.setHourlyCap(context.Background(), "global", 99)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSetHourlyCap_InvalidAppType(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	s := newTestScheduler(ctrl, store)
	err := s.setHourlyCap(context.Background(), "bogus", 10)
	if err == nil {
		t.Fatal("expected error for invalid app type")
	}
}

func TestSetHourlyCap_GetError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppType("sonarr")).Return(nil, errors.New("fail"))

	s := newTestScheduler(ctrl, store)
	err := s.setHourlyCap(context.Background(), "sonarr", 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestSetHourlyCap_UpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().GetAppSettings(gomock.Any(), database.AppType("sonarr")).Return(&database.AppSettings{AppType: "sonarr"}, nil)
	store.EXPECT().UpdateAppSettings(gomock.Any(), gomock.Any()).Return(errors.New("fail"))

	s := newTestScheduler(ctrl, store)
	err := s.setHourlyCap(context.Background(), "sonarr", 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestExecuteSchedule_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return(nil, nil)
	store.EXPECT().AddScheduleExecution(gomock.Any(), gomock.Any(), "ok").Return(nil)

	s := newTestScheduler(ctrl, store)
	s.executeSchedule(database.Schedule{
		ID:      uuid.New(),
		Action:  "enable",
		AppType: "sonarr",
	})
}

func TestExecuteSchedule_ActionError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	// "badaction" will fail in performAction, then record the error
	store.EXPECT().AddScheduleExecution(gomock.Any(), gomock.Any(), gomock.Not("ok")).Return(nil)

	s := newTestScheduler(ctrl, store)
	s.executeSchedule(database.Schedule{
		ID:      uuid.New(),
		Action:  "badaction",
		AppType: "sonarr",
	})
}

func TestExecuteSchedule_RecordError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	store.EXPECT().ListInstances(gomock.Any(), database.AppType("sonarr")).Return(nil, nil)
	store.EXPECT().AddScheduleExecution(gomock.Any(), gomock.Any(), "ok").Return(errors.New("db fail"))

	s := newTestScheduler(ctrl, store)
	// Just verify it doesn't panic
	s.executeSchedule(database.Schedule{
		ID:      uuid.New(),
		Action:  "enable",
		AppType: "sonarr",
	})
}

func TestReload_NoSchedules(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return(nil, nil)

	s := newTestScheduler(ctrl, store)
	err := s.Reload(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.cron.Jobs()) != 0 {
		t.Fatalf("expected 0 jobs, got %d", len(s.cron.Jobs()))
	}
}

func TestReload_WithSchedules(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return([]database.Schedule{
		{ID: uuid.New(), AppType: "sonarr", Action: "enable", Hour: 8, Minute: 0, Enabled: true},
		{ID: uuid.New(), AppType: "radarr", Action: "disable", Hour: 22, Minute: 0, Enabled: true},
		{ID: uuid.New(), AppType: "lidarr", Action: "enable", Hour: 12, Minute: 0, Enabled: false},
	}, nil)

	s := newTestScheduler(ctrl, store)
	err := s.Reload(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.cron.Jobs()) != 2 {
		t.Fatalf("expected 2 active jobs (disabled skipped), got %d", len(s.cron.Jobs()))
	}
}

func TestReload_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return(nil, errors.New("db error"))

	s := newTestScheduler(ctrl, store)
	err := s.Reload(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReload_RemovesOldJobs(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)

	first := store.EXPECT().ListSchedules(gomock.Any()).Return([]database.Schedule{
		{ID: uuid.New(), AppType: "sonarr", Action: "enable", Hour: 8, Minute: 0, Enabled: true},
		{ID: uuid.New(), AppType: "radarr", Action: "enable", Hour: 9, Minute: 0, Enabled: true},
	}, nil)
	store.EXPECT().ListSchedules(gomock.Any()).Return([]database.Schedule{
		{ID: uuid.New(), AppType: "sonarr", Action: "enable", Hour: 8, Minute: 0, Enabled: true},
	}, nil).After(first)

	s := newTestScheduler(ctrl, store)

	if err := s.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(s.cron.Jobs()) != 2 {
		t.Fatalf("expected 2 jobs after first reload, got %d", len(s.cron.Jobs()))
	}

	if err := s.Reload(context.Background()); err != nil {
		t.Fatal(err)
	}
	if len(s.cron.Jobs()) != 1 {
		t.Fatalf("expected 1 job after second reload, got %d", len(s.cron.Jobs()))
	}
}

func TestNew(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	s, err := New(store, newTestLogger())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s == nil {
		t.Fatal("expected non-nil scheduler")
	}
	_ = s.Stop()
}

func TestStartAndStop(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return(nil, nil)

	s, err := New(store, newTestLogger())
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	if err := s.Stop(); err != nil {
		t.Fatalf("Stop: %v", err)
	}
}

func TestStart_ReloadError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListSchedules(gomock.Any()).Return(nil, errors.New("fail"))

	s, err := New(store, newTestLogger())
	if err != nil {
		t.Fatal(err)
	}

	if err := s.Start(context.Background()); err == nil {
		t.Fatal("expected error from Start when Reload fails")
	}
	_ = s.Stop()
}

func TestAddJob(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	s := newTestScheduler(ctrl, store)

	err := s.addJob(database.Schedule{
		ID:      uuid.New(),
		AppType: "sonarr",
		Action:  "enable",
		Hour:    10,
		Minute:  30,
		Days:    []string{"MON", "TUE"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(s.cron.Jobs()) != 1 {
		t.Fatalf("expected 1 job, got %d", len(s.cron.Jobs()))
	}
}
