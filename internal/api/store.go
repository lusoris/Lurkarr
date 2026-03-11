package api

//go:generate mockgen -destination=mock_store_test.go -package=api github.com/lusoris/lurkarr/internal/api Store

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// Store defines the database methods used by all API handlers.
type Store interface {
	// Users
	GetUserByUsername(ctx context.Context, username string) (*database.User, error)
	CreateUser(ctx context.Context, username, passwordHash string) (*database.User, error)
	UserCount(ctx context.Context) (int, error)
	UpdateUsername(ctx context.Context, id uuid.UUID, username string) error
	UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error
	SetTOTPSecret(ctx context.Context, id uuid.UUID, secret *string) error

	// Sessions
	DeleteSession(ctx context.Context, id uuid.UUID) error

	// Instances
	ListInstances(ctx context.Context, appType database.AppType) ([]database.AppInstance, error)
	GetInstance(ctx context.Context, id uuid.UUID) (*database.AppInstance, error)
	CreateInstance(ctx context.Context, appType database.AppType, name, apiURL, apiKey string) (*database.AppInstance, error)
	UpdateInstance(ctx context.Context, id uuid.UUID, name, apiURL, apiKey string, enabled bool) error
	DeleteInstance(ctx context.Context, id uuid.UUID) error

	// Settings
	GetAppSettings(ctx context.Context, appType database.AppType) (*database.AppSettings, error)
	UpdateAppSettings(ctx context.Context, s *database.AppSettings) error
	GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error)
	UpsertGeneralSettings(ctx context.Context, s *database.GeneralSettings) error

	// History
	ListHuntHistory(ctx context.Context, q database.HistoryQuery) ([]database.HuntHistory, int, error)
	DeleteHistory(ctx context.Context, appType database.AppType) error

	// Logs
	QueryLogs(ctx context.Context, q database.LogQuery) ([]database.LogEntry, error)

	// Stats
	GetAllStats(ctx context.Context) ([]database.HuntStats, error)
	ResetStats(ctx context.Context) error
	GetAllHourlyCaps(ctx context.Context) ([]database.HourlyCap, error)

	// State
	GetLastReset(ctx context.Context, appType database.AppType, instanceID uuid.UUID) (*time.Time, error)
	ResetState(ctx context.Context, appType database.AppType, instanceID uuid.UUID) error

	// Queue
	GetQueueCleanerSettings(ctx context.Context, appType database.AppType) (*database.QueueCleanerSettings, error)
	UpdateQueueCleanerSettings(ctx context.Context, s *database.QueueCleanerSettings) error
	GetScoringProfile(ctx context.Context, appType database.AppType) (*database.ScoringProfile, error)
	UpdateScoringProfile(ctx context.Context, p *database.ScoringProfile) error
	GetBlocklistLog(ctx context.Context, appType database.AppType, limit int) ([]database.BlocklistLog, error)
	GetAutoImportLog(ctx context.Context, appType database.AppType, limit int) ([]database.AutoImportLog, error)

	// Prowlarr
	GetProwlarrSettings(ctx context.Context) (*database.ProwlarrSettings, error)
	UpdateProwlarrSettings(ctx context.Context, s *database.ProwlarrSettings) error

	// SABnzbd
	GetSABnzbdSettings(ctx context.Context) (*database.SABnzbdSettings, error)
	UpdateSABnzbdSettings(ctx context.Context, s *database.SABnzbdSettings) error

	// Schedules
	ListSchedules(ctx context.Context) ([]database.Schedule, error)
	CreateSchedule(ctx context.Context, s *database.Schedule) error
	UpdateSchedule(ctx context.Context, s *database.Schedule) error
	DeleteSchedule(ctx context.Context, id uuid.UUID) error
	ListScheduleExecutions(ctx context.Context, limit int) ([]database.ScheduleExecution, error)

	// Notifications
	ListNotificationProviders(ctx context.Context) ([]database.NotificationProvider, error)
	GetNotificationProvider(ctx context.Context, id uuid.UUID) (*database.NotificationProvider, error)
	CreateNotificationProvider(ctx context.Context, p *database.NotificationProvider) error
	UpdateNotificationProvider(ctx context.Context, p *database.NotificationProvider) error
	DeleteNotificationProvider(ctx context.Context, id uuid.UUID) error

	// Seerr
	GetSeerrSettings(ctx context.Context) (*database.SeerrSettings, error)
	UpdateSeerrSettings(ctx context.Context, s *database.SeerrSettings) error
}
