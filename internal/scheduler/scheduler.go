package scheduler

//go:generate mockgen -destination=mock_store_test.go -package=scheduler github.com/lusoris/lurkarr/internal/scheduler Store

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/metrics"
	"github.com/lusoris/lurkarr/internal/notifications"
)

// Store defines the database methods used by the scheduler.
type Store interface {
	CleanupOldHourlyCaps(ctx context.Context) (int64, error)
	ListSchedules(ctx context.Context) ([]database.Schedule, error)
	AddScheduleExecution(ctx context.Context, scheduleID uuid.UUID, result string) error
	ListInstances(ctx context.Context, appType database.AppType) ([]database.AppInstance, error)
	UpdateInstance(ctx context.Context, id uuid.UUID, name, apiURL, apiKey string, enabled bool) error
	GetAppSettings(ctx context.Context, appType database.AppType) (*database.AppSettings, error)
	UpdateAppSettings(ctx context.Context, s *database.AppSettings) error
}

// LurkRunner can perform an on-demand lurk pass.
type LurkRunner interface {
	RunOnceForApp(ctx context.Context, appType database.AppType, mode string) error
}

// CleanRunner can perform an on-demand queue clean pass.
type CleanRunner interface {
	RunOnceForApp(ctx context.Context, appType database.AppType) error
}

// Scheduler manages cron-based scheduling via gocron/v2.
type Scheduler struct {
	db       Store
	logger   *logging.Logger
	notifier notifications.Notifier
	lurker   LurkRunner
	cleaner  CleanRunner
	cron     gocron.Scheduler
	mu       sync.Mutex
}

// New creates a new scheduler.
func New(db Store, logger *logging.Logger) (*Scheduler, error) {
	cron, err := gocron.NewScheduler(
		gocron.WithLocation(time.UTC),
	)
	if err != nil {
		return nil, fmt.Errorf("create scheduler: %w", err)
	}
	return &Scheduler{
		db:     db,
		logger: logger,
		cron:   cron,
	}, nil
}

// SetNotifier sets an optional notification manager.
func (s *Scheduler) SetNotifier(n notifications.Notifier) {
	s.notifier = n
}

// SetLurker sets the lurking engine for on-demand lurk runs.
func (s *Scheduler) SetLurker(l LurkRunner) {
	s.lurker = l
}

// SetCleaner sets the queue cleaner for on-demand clean runs.
func (s *Scheduler) SetCleaner(c CleanRunner) {
	s.cleaner = c
}

// Start loads schedules from DB and starts the cron scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.Reload(ctx); err != nil {
		return err
	}

	// Built-in daily cleanup of old hourly_caps entries
	_, err := s.cron.NewJob(
		gocron.DailyJob(1, gocron.NewAtTimes(gocron.NewAtTime(3, 0, 0))),
		gocron.NewTask(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			deleted, err := s.db.CleanupOldHourlyCaps(ctx)
			if err != nil {
				slog.Error("hourly_caps cleanup failed", "error", err)
				return
			}
			if deleted > 0 {
				slog.Info("cleaned up old hourly_caps", "deleted", deleted)
			}
		}),
	)
	if err != nil {
		slog.Warn("failed to schedule hourly_caps cleanup", "error", err)
	}

	s.cron.Start()
	slog.Info("scheduler started")
	return nil
}

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop() error {
	return s.cron.Shutdown()
}

// Reload clears all jobs and reloads from database.
func (s *Scheduler) Reload(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Remove all existing jobs
	for _, job := range s.cron.Jobs() {
		if err := s.cron.RemoveJob(job.ID()); err != nil {
			slog.Warn("failed to remove job", "id", job.ID(), "error", err)
		}
	}

	schedules, err := s.db.ListSchedules(ctx)
	if err != nil {
		return fmt.Errorf("list schedules: %w", err)
	}

	for _, sched := range schedules {
		if !sched.Enabled {
			continue
		}
		if err := s.addJob(sched); err != nil {
			slog.Error("failed to add schedule job", "id", sched.ID, "error", err)
		}
	}

	slog.Info("scheduler reloaded", "active_jobs", len(s.cron.Jobs()))
	return nil
}

func (s *Scheduler) addJob(sched database.Schedule) error {
	cronExpr := buildCronExpr(sched)
	_, err := s.cron.NewJob(
		gocron.CronJob(cronExpr, false),
		gocron.NewTask(func() {
			s.executeSchedule(sched)
		}),
		gocron.WithSingletonMode(gocron.LimitModeReschedule),
	)
	return err
}

func buildCronExpr(sched database.Schedule) string {
	// Standard cron: minute hour day-of-month month day-of-week
	dayOfWeek := "*"
	if len(sched.Days) > 0 {
		dayOfWeek = strings.Join(sched.Days, ",")
	}
	return fmt.Sprintf("%d %d * * %s", sched.Minute, sched.Hour, dayOfWeek)
}

func (s *Scheduler) executeSchedule(sched database.Schedule) {
	log := s.logger.ForApp("system")

	// Lurk and clean actions need longer timeouts than simple config changes.
	timeout := 30 * time.Second
	switch sched.Action {
	case "lurk_missing", "lurk_upgrade", "lurk_all", "clean_queue":
		timeout = 10 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	start := time.Now()
	taskType := sched.Action
	log.Info("executing schedule", "action", sched.Action, "app_type", sched.AppType)

	result := "ok"
	if err := s.performAction(ctx, sched); err != nil {
		result = fmt.Sprintf("error: %v", err)
		log.Error("schedule execution failed", "action", sched.Action, "error", err)
		metrics.SchedulerErrors.WithLabelValues(taskType).Inc()
		if s.notifier != nil {
			s.notifier.Notify(ctx, notifications.Event{
				Type:     notifications.EventError,
				Title:    "Schedule Failed",
				Message:  fmt.Sprintf("Action %q failed: %v", sched.Action, err),
				AppType:  sched.AppType,
				Instance: sched.Action,
			})
		}
	}

	metrics.SchedulerExecutionsTotal.WithLabelValues(taskType).Inc()
	metrics.SchedulerDuration.WithLabelValues(taskType).Observe(time.Since(start).Seconds())

	if s.notifier != nil && result == "ok" {
		s.notifier.Notify(ctx, notifications.Event{
			Type:     notifications.EventSchedulerAction,
			Title:    "Schedule Executed",
			Message:  fmt.Sprintf("Action %q completed", sched.Action),
			AppType:  sched.AppType,
			Instance: sched.Action,
		})
	}

	if err := s.db.AddScheduleExecution(ctx, sched.ID, result); err != nil {
		log.Error("failed to record schedule execution", "error", err)
	}
}

func (s *Scheduler) performAction(ctx context.Context, sched database.Schedule) error {
	switch sched.Action {
	case "disable":
		return s.setAppEnabled(ctx, sched.AppType, false)
	case "enable":
		return s.setAppEnabled(ctx, sched.AppType, true)
	case "lurk_missing":
		return s.runLurk(ctx, sched.AppType, "missing")
	case "lurk_upgrade":
		return s.runLurk(ctx, sched.AppType, "upgrade")
	case "lurk_all":
		return s.runLurk(ctx, sched.AppType, "all")
	case "clean_queue":
		return s.runClean(ctx, sched.AppType)
	default:
		// api-{N} pattern — set hourly cap
		var capVal int
		if _, err := fmt.Sscanf(sched.Action, "api-%d", &capVal); err == nil {
			return s.setHourlyCap(ctx, sched.AppType, capVal)
		}
		return fmt.Errorf("unknown action: %s", sched.Action)
	}
}

func (s *Scheduler) runLurk(ctx context.Context, appType string, mode string) error {
	if s.lurker == nil {
		return fmt.Errorf("lurking engine not available")
	}
	if appType == "global" {
		for _, at := range database.AllAppTypes() {
			if err := s.lurker.RunOnceForApp(ctx, at, mode); err != nil {
				slog.Warn("scheduled lurk failed for app type", "app_type", at, "error", err)
			}
		}
		return nil
	}
	return s.lurker.RunOnceForApp(ctx, database.AppType(appType), mode)
}

func (s *Scheduler) runClean(ctx context.Context, appType string) error {
	if s.cleaner == nil {
		return fmt.Errorf("queue cleaner not available")
	}
	if appType == "global" {
		for _, at := range database.AllAppTypes() {
			if err := s.cleaner.RunOnceForApp(ctx, at); err != nil {
				slog.Warn("scheduled clean failed for app type", "app_type", at, "error", err)
			}
		}
		return nil
	}
	return s.cleaner.RunOnceForApp(ctx, database.AppType(appType))
}

func (s *Scheduler) setAppEnabled(ctx context.Context, appType string, enabled bool) error {
	if !database.ValidAppType(appType) && appType != "global" {
		return fmt.Errorf("invalid app type: %s", appType)
	}
	appTypes := []database.AppType{database.AppType(appType)}
	if appType == "global" {
		appTypes = database.AllAppTypes()
	}
	for _, at := range appTypes {
		instances, err := s.db.ListInstances(ctx, at)
		if err != nil {
			return err
		}
		for _, inst := range instances {
			if err := s.db.UpdateInstance(ctx, inst.ID, inst.Name, inst.APIURL, inst.APIKey, enabled); err != nil {
				return err
			}
		}
	}
	return nil
}

func (s *Scheduler) setHourlyCap(ctx context.Context, appType string, capVal int) error {
	if !database.ValidAppType(appType) && appType != "global" {
		return fmt.Errorf("invalid app type: %s", appType)
	}
	appTypes := []database.AppType{database.AppType(appType)}
	if appType == "global" {
		appTypes = database.AllAppTypes()
	}
	for _, at := range appTypes {
		settings, err := s.db.GetAppSettings(ctx, at)
		if err != nil {
			return err
		}
		settings.HourlyCap = capVal
		if err := s.db.UpdateAppSettings(ctx, settings); err != nil {
			return err
		}
	}
	return nil
}
