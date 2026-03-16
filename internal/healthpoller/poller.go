// Package healthpoller periodically polls all enabled arr instances
// and publishes Prometheus metrics about their health, disk, and queue status.
package healthpoller

import (
	"context"
	"log/slog"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/metrics"
	"golang.org/x/sync/errgroup"
)

// Store abstracts the DB operations needed by the health poller.
type Store interface {
	ListEnabledInstances(ctx context.Context, appType database.AppType) ([]database.AppInstance, error)
	GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error)
}

// Poller checks arr instances on a fixed interval.
type Poller struct {
	db     Store
	cancel context.CancelFunc
}

// New creates a health poller.
func New(db Store) *Poller {
	return &Poller{db: db}
}

// Start begins periodic polling in a background goroutine.
func (p *Poller) Start(ctx context.Context) {
	ctx, p.cancel = context.WithCancel(ctx)
	go p.run(ctx)
}

// Stop halts the poller.
func (p *Poller) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *Poller) run(ctx context.Context) {
	p.pollAll(ctx)
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.pollAll(ctx)
		}
	}
}

func (p *Poller) pollAll(ctx context.Context) {
	genSettings, err := p.db.GetGeneralSettings(ctx)
	if err != nil {
		slog.Warn("healthpoller: failed to get general settings", "error", err)
		return
	}
	timeout := time.Duration(genSettings.APITimeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	for _, appType := range database.AllAppTypes() {
		instances, err := p.db.ListEnabledInstances(ctx, appType)
		if err != nil {
			slog.Warn("healthpoller: failed to list instances", "app_type", appType, "error", err)
			continue
		}
		apiVersion := arrclient.APIVersionFor(string(appType))
		g, gctx := errgroup.WithContext(ctx)
		g.SetLimit(10)
		for _, inst := range instances {
			g.Go(func() error {
				p.pollInstance(gctx, appType, inst, apiVersion, timeout, genSettings.SSLVerify)
				return nil
			})
		}
		_ = g.Wait()
	}
}

func (p *Poller) pollInstance(ctx context.Context, appType database.AppType, inst database.AppInstance, apiVersion string, timeout time.Duration, sslVerify bool) {
	client := arrclient.NewClient(inst.APIURL, inst.APIKey, timeout, sslVerify)
	name := inst.Name

	status, err := client.TestConnection(ctx, apiVersion)
	if err != nil {
		metrics.ArrInstanceUp.WithLabelValues(string(appType), name, "").Set(0)
		slog.Debug("healthpoller: instance unreachable", "app_type", appType, "instance", name, "error", err)
		return
	}
	metrics.ArrInstanceUp.WithLabelValues(string(appType), name, status.Version).Set(1)

	checks, err := client.GetHealth(ctx, apiVersion)
	if err != nil {
		slog.Debug("healthpoller: health check failed", "app_type", appType, "instance", name, "error", err)
	} else {
		counts := map[string]float64{"warning": 0, "error": 0, "notice": 0}
		for _, c := range checks {
			counts[c.Type]++
		}
		for typ, count := range counts {
			metrics.ArrHealthIssues.WithLabelValues(string(appType), name, typ).Set(count)
		}
	}

	disks, err := client.GetDiskSpace(ctx, apiVersion)
	if err != nil {
		slog.Debug("healthpoller: disk space failed", "app_type", appType, "instance", name, "error", err)
	} else {
		for _, d := range disks {
			metrics.ArrDiskFreeBytes.WithLabelValues(string(appType), name, d.Path).Set(float64(d.FreeSpace))
			metrics.ArrDiskTotalBytes.WithLabelValues(string(appType), name, d.Path).Set(float64(d.TotalSpace))
		}
	}

	queue, err := client.GetQueue(ctx, apiVersion)
	if err != nil {
		slog.Debug("healthpoller: queue fetch failed", "app_type", appType, "instance", name, "error", err)
	} else {
		metrics.ArrQueueTotal.WithLabelValues(string(appType), name).Set(float64(queue.TotalRecords))
	}
}
