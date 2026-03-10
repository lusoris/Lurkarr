package scheduler

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/go-co-op/gocron/v2"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
)

// Scheduler manages cron-based scheduling via gocron/v2.
type Scheduler struct {
	db     *database.DB
	logger *logging.Logger
	cron   gocron.Scheduler
	mu     sync.Mutex
}

// New creates a new scheduler.
func New(db *database.DB, logger *logging.Logger) (*Scheduler, error) {
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

// Start loads schedules from DB and starts the cron scheduler.
func (s *Scheduler) Start(ctx context.Context) error {
	if err := s.Reload(ctx); err != nil {
		return err
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Info("executing schedule", "action", sched.Action, "app_type", sched.AppType)
	result := "ok"
	if err := s.performAction(ctx, sched); err != nil {
		result = fmt.Sprintf("error: %v", err)
		log.Error("schedule execution failed", "action", sched.Action, "error", err)
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
	default:
		// api-{N} pattern — set hourly cap
		var cap int
		if _, err := fmt.Sscanf(sched.Action, "api-%d", &cap); err == nil {
			return s.setHourlyCap(ctx, sched.AppType, cap)
		}
		return fmt.Errorf("unknown action: %s", sched.Action)
	}
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

func (s *Scheduler) setHourlyCap(ctx context.Context, appType string, cap int) error {
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
		settings.HourlyCap = cap
		if err := s.db.UpdateAppSettings(ctx, settings); err != nil {
			return err
		}
	}
	return nil
}
