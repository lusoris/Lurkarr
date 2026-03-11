package main

import (
	"context"
	"crypto/rand"
	"log/slog"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/lusoris/lurkarr/internal/autoimport"
	"github.com/lusoris/lurkarr/internal/config"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/lurking"
	"github.com/lusoris/lurkarr/internal/notifications"
	"github.com/lusoris/lurkarr/internal/queuecleaner"
	"github.com/lusoris/lurkarr/internal/scheduler"
	"github.com/lusoris/lurkarr/internal/seerr"
	"github.com/lusoris/lurkarr/internal/server"
)

// --- Modules ---

var configModule = fx.Module("config",
	fx.Provide(provideConfig),
)

var databaseModule = fx.Module("database",
	fx.Provide(provideDatabase),
)

var loggingModule = fx.Module("logging",
	fx.Provide(
		logging.NewHub,
		provideLogger,
	),
)

var notificationsModule = fx.Module("notifications",
	fx.Provide(provideNotifications),
)

var schedulerModule = fx.Module("scheduler",
	fx.Provide(provideScheduler),
)

var serverModule = fx.Module("server",
	fx.Provide(provideServerConfig),
	fx.Invoke(startServer),
)

var servicesModule = fx.Module("services",
	fx.Invoke(
		startLurkingEngine,
		startQueueCleaner,
		startAutoImporter,
		startSeerrSync,
	),
)

var maintenanceModule = fx.Module("maintenance",
	fx.Invoke(startMaintenance),
)

// --- Providers ---

// provideConfig loads configuration and sets up structured logging.
func provideConfig() (*config.Config, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	var level slog.Level
	switch cfg.LogLevel {
	case "debug":
		level = slog.LevelDebug
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})))
	slog.Info("starting Lurkarr", "addr", cfg.ListenAddr)

	return cfg, nil
}

// provideDatabase creates the DB connection pool and runs migrations.
func provideDatabase(lc fx.Lifecycle, cfg *config.Config) (*database.DB, error) {
	db, err := database.New(context.Background(), cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			db.Close()
			return nil
		},
	})
	return db, nil
}

// provideLogger creates the async DB log writer.
func provideLogger(lc fx.Lifecycle, db *database.DB, hub *logging.Hub) *logging.Logger {
	l := logging.New(db, hub)
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			l.Close()
			return nil
		},
	})
	return l
}

// provideNotifications creates the notification manager and loads providers from DB.
func provideNotifications(db *database.DB) *notifications.Manager {
	mgr := notifications.NewManager()
	enabled, err := db.ListEnabledNotificationProviders(context.Background())
	if err != nil {
		slog.Warn("failed to load notification providers", "error", err)
		return mgr
	}
	configs := make([]notifications.ProviderConfig, len(enabled))
	for i, np := range enabled {
		configs[i] = notifications.ProviderConfig{
			Type:   np.Type,
			Config: np.Config,
			Events: np.Events,
		}
	}
	if loadErr := mgr.LoadProviders(configs); loadErr != nil {
		slog.Warn("some notification providers failed to load", "error", loadErr)
	}
	return mgr
}

// provideScheduler creates the scheduler (lifecycle hooks start/stop the cron engine).
func provideScheduler(lc fx.Lifecycle, db *database.DB, logger *logging.Logger, notifMgr *notifications.Manager) (*scheduler.Scheduler, error) {
	s, err := scheduler.New(db, logger)
	if err != nil {
		return nil, err
	}
	s.SetNotifier(notifMgr)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return s.Start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			return s.Stop()
		},
	})
	return s, nil
}

// provideServerConfig builds the server configuration from app config.
func provideServerConfig(cfg *config.Config) server.Config {
	csrfKey := []byte(cfg.CSRFKey)
	if len(csrfKey) < 32 {
		csrfKey = make([]byte, 32)
		if _, err := rand.Read(csrfKey); err != nil {
			slog.Error("failed to generate CSRF key", "error", err)
			os.Exit(1)
		}
		slog.Warn("no CSRF_KEY set, generated random key (sessions will not survive restarts)")
	}
	return server.Config{
		Addr:             cfg.ListenAddr,
		CSRFKey:          csrfKey[:32],
		AllowedOrigins:   cfg.AllowedOrigins,
		ProxyAuth:        cfg.ProxyAuth,
		ProxyHeader:      cfg.ProxyHeader,
		TrustedProxies:   cfg.TrustedProxies,
		SecureCookie:     cfg.SecureCookie,
		BasePath:         cfg.BasePath,
		OIDCEnabled:      cfg.OIDCEnabled,
		OIDCIssuerURL:    cfg.OIDCIssuerURL,
		OIDCClientID:     cfg.OIDCClientID,
		OIDCClientSecret: cfg.OIDCClientSecret,
		OIDCRedirectURL:  cfg.OIDCRedirectURL,
		OIDCScopes:       cfg.OIDCScopes,
		OIDCAutoCreate:   cfg.OIDCAutoCreate,
		OIDCAdminGroup:   cfg.OIDCAdminGroup,
	}
}

// --- Lifecycle starters (fx.Invoke targets) ---

// startServer creates the HTTP server and manages its lifecycle.
func startServer(lc fx.Lifecycle, shutdowner fx.Shutdowner, cfg server.Config, db *database.DB, logger *logging.Logger, hub *logging.Hub, sched *scheduler.Scheduler, notifMgr *notifications.Manager) {
	srv := server.New(cfg, db, logger, hub, sched, notifMgr)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			go func() {
				if err := srv.Start(); err != nil {
					slog.Error("server error", "error", err)
					_ = shutdowner.Shutdown()
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return srv.Shutdown(ctx)
		},
	})
}

// startLurkingEngine creates the lurking engine and manages its lifecycle.
func startLurkingEngine(lc fx.Lifecycle, db *database.DB, logger *logging.Logger, notifMgr *notifications.Manager) {
	e := lurking.New(db, logger)
	e.SetNotifier(notifMgr)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			e.Start(context.Background()) //nolint:gosec // G118: engine manages its own lifecycle via Stop()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			e.Stop()
			return nil
		},
	})
}

// startQueueCleaner creates the queue cleaner and manages its lifecycle.
func startQueueCleaner(lc fx.Lifecycle, db *database.DB, logger *logging.Logger, notifMgr *notifications.Manager) {
	c := queuecleaner.New(db, logger)
	c.SetNotifier(notifMgr)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			c.Start(context.Background()) //nolint:gosec // G118: cleaner manages its own lifecycle via Stop()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			c.Stop()
			return nil
		},
	})
}

// startAutoImporter creates the auto-importer and manages its lifecycle.
func startAutoImporter(lc fx.Lifecycle, db *database.DB, logger *logging.Logger, notifMgr *notifications.Manager) {
	imp := autoimport.New(db, logger)
	imp.SetNotifier(notifMgr)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			imp.Start(context.Background()) //nolint:gosec // G118: importer manages its own lifecycle via Stop()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			imp.Stop()
			return nil
		},
	})
}

// startSeerrSync creates the Seerr sync engine and manages its lifecycle.
func startSeerrSync(lc fx.Lifecycle, db *database.DB) {
	se := seerr.NewSyncEngine(seerr.DBSettingsFunc(func(ctx context.Context) (string, string, bool, int, bool, error) {
		s, err := db.GetSeerrSettings(ctx)
		if err != nil {
			return "", "", false, 0, false, err
		}
		return s.URL, s.APIKey, s.Enabled, s.SyncIntervalMinutes, s.AutoApprove, nil
	}))
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			se.Start(context.Background()) //nolint:gosec // G118: sync engine manages its own lifecycle via Stop()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			se.Stop()
			return nil
		},
	})
}

// startMaintenance runs the hourly database cleanup goroutine.
func startMaintenance(lc fx.Lifecycle, db *database.DB) {
	var cancel context.CancelFunc
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background()) //nolint:gosec // G118: cancel is called in OnStop
			go runMaintenance(ctx, db)
			return nil
		},
		OnStop: func(_ context.Context) error {
			if cancel != nil {
				cancel()
			}
			return nil
		},
	})
}

func runMaintenance(ctx context.Context, db *database.DB) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			mCtx, mCancel := context.WithTimeout(ctx, 30*time.Second)
			if cleaned, err := db.CleanExpiredSessions(mCtx); err != nil {
				slog.Warn("failed to clean expired sessions", "error", err)
			} else if cleaned > 0 {
				slog.Info("cleaned expired sessions", "count", cleaned)
			}
			if reset, err := db.AutoResetExpiredStates(mCtx, 168); err != nil {
				slog.Warn("failed to auto-reset expired states", "error", err)
			} else if reset > 0 {
				slog.Info("auto-reset expired states", "count", reset)
			}
			if pruned, err := db.PruneLogs(mCtx, 30); err != nil {
				slog.Warn("failed to prune logs", "error", err)
			} else if pruned > 0 {
				slog.Info("pruned old logs", "count", pruned)
			}
			if caps, err := db.CleanupOldHourlyCaps(mCtx); err != nil {
				slog.Warn("failed to cleanup old hourly caps", "error", err)
			} else if caps > 0 {
				slog.Info("cleaned old hourly caps", "count", caps)
			}
			if err := db.PruneStrikes(mCtx, 7*24*time.Hour); err != nil {
				slog.Warn("failed to prune strikes", "error", err)
			}
			if err := db.PruneAutoImportLog(mCtx, 30*24*time.Hour); err != nil {
				slog.Warn("failed to prune auto import log", "error", err)
			}
			if err := db.PruneBlocklistLog(mCtx, 30*24*time.Hour); err != nil {
				slog.Warn("failed to prune blocklist log", "error", err)
			}
			mCancel()
		case <-ctx.Done():
			return
		}
	}
}
