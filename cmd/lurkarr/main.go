package main

import (
	"context"
	"crypto/rand"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lusoris/lurkarr/internal/autoimport"
	"github.com/lusoris/lurkarr/internal/config"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/lurking"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/notifications"
	"github.com/lusoris/lurkarr/internal/queuecleaner"
	"github.com/lusoris/lurkarr/internal/scheduler"
	"github.com/lusoris/lurkarr/internal/seerr"
	"github.com/lusoris/lurkarr/internal/server"
)

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, err := database.New(ctx, cfg.DatabaseURL, cfg.DBMaxConns)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer db.Close()

	hub := logging.NewHub()
	logger := logging.New(db, hub)
	defer logger.Close()

	notifMgr := notifications.NewManager()

	// Load notification providers from database.
	enabledProviders, err := db.ListEnabledNotificationProviders(ctx)
	if err != nil {
		slog.Warn("failed to load notification providers", "error", err)
	} else {
		configs := make([]notifications.ProviderConfig, len(enabledProviders))
		for i, np := range enabledProviders {
			configs[i] = notifications.ProviderConfig{
				Type:   np.Type,
				Config: np.Config,
				Events: np.Events,
			}
		}
		if loadErr := notifMgr.LoadProviders(configs); loadErr != nil {
			slog.Warn("some notification providers failed to load", "error", loadErr)
		}
	}

	engine := lurking.New(db, logger)
	engine.SetNotifier(notifMgr)
	engine.Start(ctx)
	defer engine.Stop()

	sched, err := scheduler.New(db, logger)
	if err != nil {
		return fmt.Errorf("create scheduler: %w", err)
	}
	sched.SetNotifier(notifMgr)
	if err := sched.Start(ctx); err != nil {
		return fmt.Errorf("start scheduler: %w", err)
	}
	defer func() {
		if err := sched.Stop(); err != nil {
			slog.Warn("failed to stop scheduler", "error", err)
		}
	}()

	cleaner := queuecleaner.New(db, logger)
	cleaner.SetNotifier(notifMgr)
	cleaner.Start(ctx)
	defer cleaner.Stop()

	importer := autoimport.New(db, logger)
	importer.SetNotifier(notifMgr)
	importer.Start(ctx)
	defer importer.Stop()

	seerrSync := seerr.NewSyncEngine(seerr.DBSettingsFunc(func(ctx context.Context) (string, string, bool, int, bool, error) {
		s, err := db.GetSeerrSettings(ctx)
		if err != nil {
			return "", "", false, 0, false, err
		}
		return s.URL, s.APIKey, s.Enabled, s.SyncIntervalMinutes, s.AutoApprove, nil
	}))
	seerrSync.Start(ctx)
	defer seerrSync.Stop()

	csrfKey := []byte(cfg.CSRFKey)
	if len(csrfKey) < 32 {
		csrfKey = make([]byte, 32)
		if _, err := rand.Read(csrfKey); err != nil {
			return fmt.Errorf("generate CSRF key: %w", err)
		}
		slog.Warn("no CSRF_KEY set, generated random key (sessions will not survive restarts)")
	}

	srv := server.New(server.Config{
		Addr:           cfg.ListenAddr,
		CSRFKey:        csrfKey[:32],
		AllowedOrigins: cfg.AllowedOrigins,
		ProxyAuth:      cfg.ProxyAuth,
		ProxyHeader:    cfg.ProxyHeader,
		SecureCookie:   cfg.SecureCookie,
	}, db, logger, hub, sched, notifMgr)

	errCh := make(chan error, 1)
	go func() {
		errCh <- srv.Start()
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				mCtx, mCancel := context.WithTimeout(ctx, 30*time.Second)
				cleaned, err := db.CleanExpiredSessions(mCtx)
				if err != nil {
					slog.Warn("failed to clean expired sessions", "error", err)
				} else if cleaned > 0 {
					slog.Info("cleaned expired sessions", "count", cleaned)
				}
				reset, err := db.AutoResetExpiredStates(mCtx, 168)
				if err != nil {
					slog.Warn("failed to auto-reset expired states", "error", err)
				} else if reset > 0 {
					slog.Info("auto-reset expired states", "count", reset)
				}
				pruned, err := db.PruneLogs(mCtx, 30)
				if err != nil {
					slog.Warn("failed to prune logs", "error", err)
				} else if pruned > 0 {
					slog.Info("pruned old logs", "count", pruned)
				}
				caps, err := db.CleanupOldHourlyCaps(mCtx)
				if err != nil {
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
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-quit:
		slog.Info("received shutdown signal", "signal", sig)
	case err := <-errCh:
		slog.Error("server error", "error", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("Lurkarr stopped")
	return nil
}
