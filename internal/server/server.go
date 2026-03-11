package server

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"github.com/lusoris/lurkarr/internal/api"
	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
	"github.com/lusoris/lurkarr/internal/middleware"
	"github.com/lusoris/lurkarr/internal/notifications"
	"github.com/lusoris/lurkarr/internal/scheduler"
)

// Config holds server configuration.
type Config struct {
	Addr           string
	CSRFKey        []byte
	AllowedOrigins []string
	ProxyAuth      bool
	ProxyHeader    string
	SecureCookie   bool
}

// Server is the main HTTP server.
type Server struct {
	httpServer *http.Server
}

// New creates a new Server with all routes registered.
func New(cfg Config, db *database.DB, logger *logging.Logger, hub *logging.Hub, sched *scheduler.Scheduler, notifMgr *notifications.Manager) *Server {
	authMw := &auth.Middleware{
		DB:              db,
		ProxyAuthBypass: cfg.ProxyAuth,
		ProxyHeader:     cfg.ProxyHeader,
		CSRFKey:         cfg.CSRFKey,
		SecureCookie:    cfg.SecureCookie,
	}

	authH := &api.AuthHandler{DB: db, Auth: authMw}
	settingsH := &api.SettingsHandler{DB: db}
	appsH := &api.AppsHandler{DB: db}
	historyH := &api.HistoryHandler{DB: db}
	logsH := &api.LogsHandler{DB: db, Hub: hub}
	statsH := &api.StatsHandler{DB: db}
	stateH := &api.StateHandler{DB: db}
	schedulerH := &api.SchedulerHandler{DB: db, Scheduler: sched}
	prowlarrH := &api.ProwlarrHandler{DB: db}
	sabnzbdH := &api.SABnzbdHandler{DB: db}
	userH := &api.UserHandler{DB: db}
	queueH := &api.QueueHandler{DB: db}
	notificationH := &api.NotificationHandler{DB: db, Manager: notifMgr}

	mux := http.NewServeMux()

	// Rate limiter for login: 5 attempts per minute per IP (burst 5).
	loginRL := middleware.NewIPRateLimiter(rate.Limit(5.0/60.0), 5)

	// --- Public routes (no auth) ---
	mux.Handle("POST /api/auth/login", middleware.RateLimit(loginRL)(http.HandlerFunc(authH.HandleLogin)))
	mux.HandleFunc("POST /api/auth/setup", authH.HandleSetup)

	// --- Health ---
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		if err := db.HealthCheck(r.Context()); err != nil {
			http.Error(w, `{"status":"unhealthy"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"healthy"}`))
	})

	// --- Metrics (Prometheus) ---
	mux.Handle("GET /metrics", promhttp.Handler())

	// --- Protected routes ---
	protected := http.NewServeMux()

	// Auth
	protected.HandleFunc("POST /api/auth/logout", authH.HandleLogout)
	protected.HandleFunc("POST /api/auth/2fa/enable", authH.Handle2FAEnable)
	protected.HandleFunc("POST /api/auth/2fa/disable", authH.Handle2FADisable)
	protected.HandleFunc("POST /api/auth/2fa/verify", authH.Handle2FAVerify)

	// User
	protected.HandleFunc("GET /api/user", userH.HandleGetUser)
	protected.HandleFunc("POST /api/user/username", userH.HandleUpdateUsername)
	protected.HandleFunc("POST /api/user/password", userH.HandleUpdatePassword)

	// Settings
	protected.HandleFunc("GET /api/settings/general", settingsH.HandleGetGeneralSettings)
	protected.HandleFunc("PUT /api/settings/general", settingsH.HandleUpdateGeneralSettings)
	protected.HandleFunc("GET /api/settings/{app}", settingsH.HandleGetAppSettings)
	protected.HandleFunc("PUT /api/settings/{app}", settingsH.HandleUpdateAppSettings)

	// App Instances
	protected.HandleFunc("GET /api/instances/{app}", appsH.HandleListInstances)
	protected.HandleFunc("POST /api/instances/{app}", appsH.HandleCreateInstance)
	protected.HandleFunc("PUT /api/instances/{id}", appsH.HandleUpdateInstance)
	protected.HandleFunc("DELETE /api/instances/{id}", appsH.HandleDeleteInstance)
	protected.HandleFunc("POST /api/instances/test", appsH.HandleTestConnection)

	// History
	protected.HandleFunc("GET /api/history", historyH.HandleListHistory)
	protected.HandleFunc("DELETE /api/history/{app}", historyH.HandleDeleteHistory)

	// Logs
	protected.HandleFunc("GET /api/logs", logsH.HandleGetLogs)

	// Stats
	protected.HandleFunc("GET /api/stats", statsH.HandleGetStats)
	protected.HandleFunc("POST /api/stats/reset", statsH.HandleResetStats)
	protected.HandleFunc("GET /api/stats/hourly-caps", statsH.HandleGetHourlyCaps)

	// State
	protected.HandleFunc("GET /api/state", stateH.HandleGetState)
	protected.HandleFunc("POST /api/state/reset", stateH.HandleResetState)

	// Schedules
	protected.HandleFunc("GET /api/schedules", schedulerH.HandleListSchedules)
	protected.HandleFunc("POST /api/schedules", schedulerH.HandleCreateSchedule)
	protected.HandleFunc("PUT /api/schedules/{id}", schedulerH.HandleUpdateSchedule)
	protected.HandleFunc("DELETE /api/schedules/{id}", schedulerH.HandleDeleteSchedule)
	protected.HandleFunc("GET /api/schedules/history", schedulerH.HandleScheduleHistory)

	// Prowlarr
	protected.HandleFunc("GET /api/prowlarr/settings", prowlarrH.HandleGetSettings)
	protected.HandleFunc("PUT /api/prowlarr/settings", prowlarrH.HandleUpdateSettings)
	protected.HandleFunc("GET /api/prowlarr/indexers", prowlarrH.HandleGetIndexers)
	protected.HandleFunc("GET /api/prowlarr/indexers/stats", prowlarrH.HandleGetIndexerStats)
	protected.HandleFunc("POST /api/prowlarr/test", prowlarrH.HandleTestConnection)

	// SABnzbd
	protected.HandleFunc("GET /api/sabnzbd/settings", sabnzbdH.HandleGetSettings)
	protected.HandleFunc("PUT /api/sabnzbd/settings", sabnzbdH.HandleUpdateSettings)
	protected.HandleFunc("GET /api/sabnzbd/queue", sabnzbdH.HandleGetQueue)
	protected.HandleFunc("GET /api/sabnzbd/history", sabnzbdH.HandleGetHistory)
	protected.HandleFunc("GET /api/sabnzbd/stats", sabnzbdH.HandleGetStats)
	protected.HandleFunc("POST /api/sabnzbd/pause", sabnzbdH.HandlePause)
	protected.HandleFunc("POST /api/sabnzbd/resume", sabnzbdH.HandleResume)
	protected.HandleFunc("POST /api/sabnzbd/test", sabnzbdH.HandleTestConnection)

	// Queue Management
	protected.HandleFunc("GET /api/queue/settings/{app}", queueH.HandleGetQueueCleanerSettings)
	protected.HandleFunc("PUT /api/queue/settings/{app}", queueH.HandleUpdateQueueCleanerSettings)
	protected.HandleFunc("GET /api/queue/scoring/{app}", queueH.HandleGetScoringProfile)
	protected.HandleFunc("PUT /api/queue/scoring/{app}", queueH.HandleUpdateScoringProfile)
	protected.HandleFunc("GET /api/queue/blocklist/{app}", queueH.HandleGetBlocklistLog)
	protected.HandleFunc("GET /api/queue/imports/{app}", queueH.HandleGetAutoImportLog)

	// Notifications
	protected.HandleFunc("GET /api/notifications/providers", notificationH.HandleListProviders)
	protected.HandleFunc("POST /api/notifications/providers", notificationH.HandleCreateProvider)
	protected.HandleFunc("GET /api/notifications/providers/{id}", notificationH.HandleGetProvider)
	protected.HandleFunc("PUT /api/notifications/providers/{id}", notificationH.HandleUpdateProvider)
	protected.HandleFunc("DELETE /api/notifications/providers/{id}", notificationH.HandleDeleteProvider)
	protected.HandleFunc("POST /api/notifications/providers/{id}/test", notificationH.HandleTestProvider)

	// WebSocket (no CSRF, but auth required)
	mux.Handle("GET /ws/logs", authMw.RequireAuth(http.HandlerFunc(logsH.HandleWebSocketLogs)))

	// Mount protected routes with auth + CSRF
	mux.Handle("/api/", authMw.CSRFProtect()(authMw.RequireAuth(protected)))

	// Global middleware chain
	corsMiddleware := middleware.CORS(middleware.CORSConfig{AllowedOrigins: cfg.AllowedOrigins})
	handler := middleware.Chain(mux,
		middleware.Recovery,
		middleware.RequestID,
		middleware.Logging,
		corsMiddleware,
	)

	return &Server{
		httpServer: &http.Server{
			Addr:              cfg.Addr,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
	}
}

// Start begins listening and serving.
func (s *Server) Start() error {
	slog.Info("server starting", "addr", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Shutdown gracefully shuts down the server.
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}
