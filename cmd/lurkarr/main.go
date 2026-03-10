package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/lusoris/lurkarr/internal/config"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/hunting"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/scheduler"
	"github.com/lusoris/lurkarr/internal/server"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
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

	ctx := context.Background()
	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	hub := logging.NewHub()
	logger := logging.New(db, hub)
	defer logger.Close()

	engine := hunting.New(db, logger)
	engine.Start(ctx)
	defer engine.Stop()

	sched, err := scheduler.New(db, logger)
	if err != nil {
		slog.Error("failed to create scheduler", "error", err)
		os.Exit(1)
	}
	if err := sched.Start(ctx); err != nil {
		slog.Error("failed to start scheduler", "error", err)
		os.Exit(1)
	}
	defer sched.Stop()

	csrfKey := []byte(cfg.CSRFKey)
	if len(csrfKey) < 32 {
		padded := make([]byte, 32)
		copy(padded, csrfKey)
		csrfKey = padded
	}

	srv := server.New(server.Config{
		Addr:           cfg.ListenAddr,
		CSRFKey:        csrfKey[:32],
		AllowedOrigins: cfg.AllowedOrigins,
		ProxyAuth:      cfg.ProxyAuth,
		ProxyHeader:    cfg.ProxyHeader,
	}, db, logger, hub, sched)

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
				mCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				cleaned, _ := db.CleanExpiredSessions(mCtx)
				if cleaned > 0 {
					slog.Info("cleaned expired sessions", "count", cleaned)
				}
				reset, _ := db.AutoResetExpiredStates(mCtx, 168)
				if reset > 0 {
					slog.Info("auto-reset expired states", "count", reset)
				}
				pruned, _ := db.PruneLogs(mCtx, 30)
				if pruned > 0 {
					slog.Info("pruned old logs", "count", pruned)
				}
				cancel()
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

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("server shutdown error", "error", err)
	}

	slog.Info("Lurkarr stopped")
}
