package main

import (
	"context"
	"crypto/rand"
	"log/slog"
	"os"
	"time"

	"go.uber.org/fx"

	"github.com/lusoris/lurkarr/frontend"
	"github.com/lusoris/lurkarr/internal/autoimport"
	"github.com/lusoris/lurkarr/internal/config"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/healthpoller"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/lurking"
	"github.com/lusoris/lurkarr/internal/notifications"
	"github.com/lusoris/lurkarr/internal/openapi"
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
	fx.Provide(provideLogger),
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
		startHealthPoller,
	),
)

var maintenanceModule = fx.Module("maintenance",
	fx.Invoke(startMaintenance),
)

var runOnceModule = fx.Module("run-once",
	fx.Invoke(executeRunOnce),
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

// provideLogger creates the structured logger.
func provideLogger() *logging.Logger {
	return logging.New()
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

// provideServerConfig builds the server configuration from app config and
// persisted CSRF key. If no CSRF_KEY env var is set, the key is loaded from
// the database. If absent there too, a new key is generated and stored.
func provideServerConfig(cfg *config.Config, db *database.DB) server.Config {
	csrfKey := []byte(cfg.CSRFKey)
	if len(csrfKey) < 32 {
		// Try loading persisted key from database.
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		stored, err := db.GetCSRFKey(ctx)
		if err == nil && len(stored) >= 32 {
			csrfKey = []byte(stored)
			slog.Info("loaded persisted CSRF key from database")
		} else {
			// Generate a new key and persist it.
			csrfKey = make([]byte, 32)
			if _, err := rand.Read(csrfKey); err != nil {
				slog.Error("failed to generate CSRF key", "error", err)
				os.Exit(1)
			}
			if err := db.SetCSRFKey(ctx, string(csrfKey)); err != nil {
				slog.Warn("failed to persist CSRF key to database", "error", err)
			} else {
				slog.Info("generated and persisted new CSRF key to database")
			}
		}
	}
	return server.Config{
		Addr:             cfg.ListenAddr,
		CSRFKey:          csrfKey[:32],
		AllowedOrigins:   cfg.AllowedOrigins,
		ProxyAuth:        cfg.ProxyAuth,
		ProxyHeaders:     cfg.ProxyHeaders,
		TrustedProxies:   cfg.TrustedProxies,
		SecureCookie:     cfg.SecureCookie,
		BasePath:         cfg.BasePath,
		OpenAPISpec:      openapi.Spec,
		FrontendFS:       frontend.BuildFS(),
		OIDCEnabled:      cfg.OIDCEnabled,
		OIDCIssuerURL:    cfg.OIDCIssuerURL,
		OIDCClientID:     cfg.OIDCClientID,
		OIDCClientSecret: cfg.OIDCClientSecret,
		OIDCRedirectURL:  cfg.OIDCRedirectURL,
		OIDCScopes:       cfg.OIDCScopes,
		OIDCAutoCreate:   cfg.OIDCAutoCreate,
		OIDCAdminGroup:   cfg.OIDCAdminGroup,

		// WebAuthn
		WebAuthnRPID:          cfg.WebAuthnRPID,
		WebAuthnRPDisplayName: cfg.WebAuthnRPDisplayName,
		WebAuthnRPOrigins:     cfg.WebAuthnRPOrigins,
	}
}

// --- Lifecycle starters (fx.Invoke targets) ---

// startServer creates the HTTP server and manages its lifecycle.
func startServer(lc fx.Lifecycle, shutdowner fx.Shutdowner, cfg server.Config, db *database.DB, sched *scheduler.Scheduler, notifMgr *notifications.Manager) {
	ctx, cancel := context.WithCancel(context.Background())
	srv := server.New(ctx, cfg, db, sched, notifMgr)
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
			cancel()
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
	router := &seerr.RequestRouter{DB: db}
	se := seerr.NewSyncEngine(seerr.DBSettingsFunc(func(ctx context.Context) (string, string, bool, int, bool, error) {
		s, err := db.GetSeerrSettings(ctx)
		if err != nil {
			return "", "", false, 0, false, err
		}
		return s.URL, s.APIKey, s.Enabled, s.SyncIntervalMinutes, s.AutoApprove, nil
	}), router)
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

// startHealthPoller creates the arr health poller and manages its lifecycle.
func startHealthPoller(lc fx.Lifecycle, db *database.DB) {
	hp := healthpoller.New(db)
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			hp.Start(context.Background()) //nolint:gosec // G118: poller manages its own lifecycle via Stop()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			hp.Stop()
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

// executeRunOnce runs all services once synchronously and triggers shutdown.
func executeRunOnce(lc fx.Lifecycle, shutdowner fx.Shutdowner, db *database.DB, logger *logging.Logger, notifMgr *notifications.Manager) {
	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			go func() {
				ctx := context.Background()
				slog.Info("run-once mode: starting single pass")

				e := lurking.New(db, logger)
				e.SetNotifier(notifMgr)
				e.RunOnce(ctx)

				c := queuecleaner.New(db, logger)
				c.SetNotifier(notifMgr)
				c.RunOnce(ctx)

				imp := autoimport.New(db, logger)
				imp.SetNotifier(notifMgr)
				imp.RunOnce(ctx)

				slog.Info("run-once mode: all passes complete, shutting down")
				_ = shutdowner.Shutdown()
			}()
			return nil
		},
	})
}
