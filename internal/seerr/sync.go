package seerr

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// SettingsProvider loads Seerr settings from the database.
type SettingsProvider interface {
	GetSeerrSettings(ctx context.Context) (*Settings, error)
}

// Settings mirrors the database SeerrSettings for the sync engine.
type Settings struct {
	URL                 string
	APIKey              string
	Enabled             bool
	SyncIntervalMinutes int
	AutoApprove         bool
}

// SyncEngine periodically polls Seerr for approved requests and triggers
// searches in the matching arr instance.
type SyncEngine struct {
	settings SettingsProvider
	cancel   context.CancelFunc
	wg       sync.WaitGroup
}

// NewSyncEngine creates a new Seerr sync engine.
func NewSyncEngine(settings SettingsProvider) *SyncEngine {
	return &SyncEngine{settings: settings}
}

// Start launches the background sync loop.
func (e *SyncEngine) Start(ctx context.Context) {
	ctx, e.cancel = context.WithCancel(ctx)
	e.wg.Add(1)
	go e.loop(ctx)
	slog.Info("seerr sync engine started")
}

// Stop stops the sync engine and waits for the loop to exit.
func (e *SyncEngine) Stop() {
	if e.cancel != nil {
		e.cancel()
	}
	e.wg.Wait()
}

func (e *SyncEngine) loop(ctx context.Context) {
	defer e.wg.Done()

	// Initial delay to let other services start.
	select {
	case <-time.After(15 * time.Second):
	case <-ctx.Done():
		return
	}

	for {
		settings, err := e.settings.GetSeerrSettings(ctx)
		if err != nil {
			slog.Error("seerr: failed to load settings", "error", err)
		} else if settings.Enabled && settings.URL != "" && settings.APIKey != "" {
			e.sync(ctx, settings)
		}

		interval := 30 * time.Minute
		if settings != nil && settings.SyncIntervalMinutes > 0 {
			interval = time.Duration(settings.SyncIntervalMinutes) * time.Minute
		}

		select {
		case <-time.After(interval):
		case <-ctx.Done():
			return
		}
	}
}

func (e *SyncEngine) sync(ctx context.Context, settings *Settings) {
	client := NewClient(settings.URL, settings.APIKey, 30*time.Second)

	// Fetch pending requests to log them; actual routing to arr instances
	// is handled by Seerr itself — we just monitor the status.
	count, err := client.GetRequestCount(ctx)
	if err != nil {
		slog.Error("seerr: failed to get request count", "error", err)
		return
	}

	slog.Info("seerr: sync complete",
		"total", count.Total,
		"pending", count.Pending,
		"processing", count.Processing,
		"available", count.Available,
	)

	// If there are pending requests, fetch details.
	if count.Pending > 0 {
		resp, err := client.ListRequests(ctx, "pending", 50, 0)
		if err != nil {
			slog.Error("seerr: failed to list pending requests", "error", err)
			return
		}

		for _, req := range resp.Results {
			slog.Info("seerr: pending request",
				"request_id", req.ID,
				"type", req.Type,
				"tmdb_id", req.Media.TmdbID,
				"requested_by", req.RequestedBy.DisplayName,
			)

			// Auto-approve if configured.
			if settings.AutoApprove {
				if err := client.ApproveRequest(ctx, req.ID); err != nil {
					slog.Error("seerr: failed to auto-approve request",
						"request_id", req.ID, "error", err)
				} else {
					slog.Info("seerr: auto-approved request", "request_id", req.ID)
				}
			}
		}
	}
}
