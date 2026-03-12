// Package metrics provides Prometheus metrics for Lurkarr.
package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ── Lurking metrics ──────────────────────────────────────────────────────────

var (
	LurkSearchesTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "lurk",
		Name:      "searches_total",
		Help:      "Total number of lurk searches triggered.",
	}, []string{"app_type", "instance"})

	LurkMissingFound = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "lurk",
		Name:      "missing_found_total",
		Help:      "Total missing items found during lurks.",
	}, []string{"app_type", "instance"})

	LurkUpgradesFound = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "lurk",
		Name:      "upgrades_found_total",
		Help:      "Total upgrade candidates found during lurks.",
	}, []string{"app_type", "instance"})

	LurkDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lurkarr",
		Subsystem: "lurk",
		Name:      "duration_seconds",
		Help:      "Duration of lurk operations in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"app_type", "instance"})

	LurkErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "lurk",
		Name:      "errors_total",
		Help:      "Total lurk errors.",
	}, []string{"app_type", "instance"})
)

// ── Queue Cleaner metrics ────────────────────────────────────────────────────

var (
	QueueCleanerItemsRemoved = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "queue_cleaner",
		Name:      "items_removed_total",
		Help:      "Total items removed from queue by cleaner.",
	}, []string{"app_type", "instance"})

	QueueCleanerStrikes = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "queue_cleaner",
		Name:      "strikes_total",
		Help:      "Total strikes issued by queue cleaner.",
	}, []string{"app_type", "instance"})

	QueueCleanerBlocklistAdditions = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "queue_cleaner",
		Name:      "blocklist_additions_total",
		Help:      "Total items added to blocklist by queue cleaner.",
	}, []string{"app_type", "instance"})

	QueueCleanerRunDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lurkarr",
		Subsystem: "queue_cleaner",
		Name:      "run_duration_seconds",
		Help:      "Duration of queue cleaner runs in seconds.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"app_type", "instance"})
)

// ── Download Client metrics ──────────────────────────────────────────────────

var (
	DownloadClientQueueSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "lurkarr",
		Subsystem: "download_client",
		Name:      "queue_size",
		Help:      "Current download queue size.",
	}, []string{"client_type", "instance"})

	DownloadClientSpeed = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "lurkarr",
		Subsystem: "download_client",
		Name:      "speed_bytes_per_second",
		Help:      "Current download speed in bytes/sec.",
	}, []string{"client_type", "instance", "direction"})

	DownloadClientPaused = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "lurkarr",
		Subsystem: "download_client",
		Name:      "paused",
		Help:      "Whether the download client is paused (1=paused, 0=active).",
	}, []string{"client_type", "instance"})
)

// ── Scheduler metrics ────────────────────────────────────────────────────────

var (
	SchedulerExecutionsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "scheduler",
		Name:      "executions_total",
		Help:      "Total scheduler task executions.",
	}, []string{"task_type"})

	SchedulerDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lurkarr",
		Subsystem: "scheduler",
		Name:      "duration_seconds",
		Help:      "Duration of scheduled task executions.",
		Buckets:   prometheus.DefBuckets,
	}, []string{"task_type"})

	SchedulerErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "scheduler",
		Name:      "errors_total",
		Help:      "Total scheduler task errors.",
	}, []string{"task_type"})
)

// ── HTTP metrics ─────────────────────────────────────────────────────────────

var (
	HTTPRequestsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "http",
		Name:      "requests_total",
		Help:      "Total HTTP requests.",
	}, []string{"method", "path", "status"})

	HTTPRequestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lurkarr",
		Subsystem: "http",
		Name:      "request_duration_seconds",
		Help:      "HTTP request duration in seconds.",
		Buckets:   []float64{0.001, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"method", "path"})

	HTTPResponseSize = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lurkarr",
		Subsystem: "http",
		Name:      "response_size_bytes",
		Help:      "HTTP response size in bytes.",
		Buckets:   prometheus.ExponentialBuckets(100, 10, 7), // 100B to 100MB
	}, []string{"method", "path"})

	HTTPRateLimitHits = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "http",
		Name:      "rate_limit_hits_total",
		Help:      "Total rate limit hits.",
	}, []string{"path"})
)

// ── Autoimport metrics ───────────────────────────────────────────────────────

var (
	AutoimportRunsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "autoimport",
		Name:      "runs_total",
		Help:      "Total autoimport runs.",
	}, []string{"app_type", "instance"})

	AutoimportErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "autoimport",
		Name:      "errors_total",
		Help:      "Total autoimport errors.",
	}, []string{"app_type", "instance"})
)

// ── Notification metrics ─────────────────────────────────────────────────────

var (
	NotificationSentTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "notifications",
		Name:      "sent_total",
		Help:      "Total notifications sent.",
	}, []string{"provider", "event_type"})

	NotificationErrorsTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "lurkarr",
		Subsystem: "notifications",
		Name:      "errors_total",
		Help:      "Total notification delivery errors.",
	}, []string{"provider", "event_type"})

	NotificationDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: "lurkarr",
		Subsystem: "notifications",
		Name:      "duration_seconds",
		Help:      "Duration of notification delivery in seconds.",
		Buckets:   []float64{0.01, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10},
	}, []string{"provider"})
)
