package server

import (
	"context"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/gorilla/csrf"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/time/rate"

	"github.com/lusoris/lurkarr/internal/api"
	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/middleware"
	"github.com/lusoris/lurkarr/internal/notifications"
	"github.com/lusoris/lurkarr/internal/scheduler"
	"github.com/lusoris/lurkarr/internal/seerr"
)

// scalarHTML is the Scalar API reference page.
var scalarHTML = []byte(`<!doctype html>
<html>
<head>
  <title>Lurkarr API Reference</title>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
</head>
<body>
  <script id="api-reference" data-url="spec"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>
`)

// Config holds server configuration.
type Config struct {
	Addr           string
	CSRFKey        []byte
	AllowedOrigins []string
	ProxyAuth      bool
	ProxyHeaders   []string
	TrustedProxies []*net.IPNet
	SecureCookie   bool
	BasePath       string
	OpenAPISpec    []byte
	FrontendFS     fs.FS

	// OIDC
	OIDCEnabled      bool
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCScopes       []string
	OIDCAutoCreate   bool
	OIDCAdminGroup   string

	// WebAuthn / Passkeys
	WebAuthnRPID          string   // e.g. "localhost" or "lurkarr.example.com"
	WebAuthnRPDisplayName string   // e.g. "Lurkarr"
	WebAuthnRPOrigins     []string // e.g. ["http://localhost:9705"]
}

// Server is the main HTTP server.
type Server struct {
	httpServer *http.Server
}

// csrfInjectToken wraps a handler to expose the CSRF token in a response header
// so the SPA can read it and send it back on mutating requests.
func csrfInjectToken(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-CSRF-Token", csrf.Token(r))
		next.ServeHTTP(w, r)
	})
}

// csrfPlaintextHTTP marks requests as plaintext HTTP for gorilla/csrf so that
// it does not enforce strict Referer/Origin checks meant for TLS connections.
func csrfPlaintextHTTP(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.TLS == nil && r.Header.Get("X-Forwarded-Proto") != "https" {
			r = csrf.PlaintextHTTPRequest(r)
		}
		next.ServeHTTP(w, r)
	})
}

// spaHandler serves an SPA from an fs.FS. It tries the exact path first,
// then falls back to index.html for client-side routing.
func spaHandler(fsys fs.FS) http.Handler {
	fileServer := http.FileServerFS(fsys)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the exact file.
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(fsys, path); err == nil {
			// Cache immutable assets aggressively.
			if strings.HasPrefix(path, "_app/immutable/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}
		// SPA fallback: serve index.html for all other paths.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}

// New creates a new Server with all routes registered.
func New(ctx context.Context, cfg Config, db *database.DB, sched *scheduler.Scheduler, notifMgr *notifications.Manager) *Server {
	// Set trusted proxies for rate limiter IP extraction.
	middleware.TrustedProxies = cfg.TrustedProxies

	authMw := &auth.Middleware{
		DB:              db,
		ProxyAuthBypass: cfg.ProxyAuth,
		ProxyHeaders:    cfg.ProxyHeaders,
		TrustedProxies:  cfg.TrustedProxies,
		ProxyAutoCreate: cfg.ProxyAuth,
		CSRFKey:         cfg.CSRFKey,
		SecureCookie:    cfg.SecureCookie,
	}

	authH := &api.AuthHandler{DB: db, Auth: authMw}
	settingsH := &api.SettingsHandler{DB: db}
	appsH := &api.AppsHandler{DB: db}
	historyH := &api.HistoryHandler{DB: db}
	statsH := &api.StatsHandler{DB: db}
	stateH := &api.StateHandler{DB: db}
	schedulerH := &api.SchedulerHandler{DB: db, Scheduler: sched}
	prowlarrH := &api.ProwlarrHandler{DB: db}
	sabnzbdH := &api.SABnzbdHandler{DB: db}
	userH := &api.UserHandler{DB: db}
	queueH := &api.QueueHandler{DB: db}
	blocklistH := &api.BlocklistHandler{DB: db}
	notificationH := &api.NotificationHandler{DB: db, Manager: notifMgr}
	seerrH := &api.SeerrHandler{DB: db, Router: &seerr.RequestRouter{DB: db}}
	dlClientH := &api.DownloadClientHandler{DB: db}
	sessionH := &api.SessionHandler{DB: db}
	adminH := &api.AdminHandler{DB: db}
	oidcSettingsH := &api.OIDCSettingsHandler{DB: db}
	groupsH := &api.InstanceGroupsHandler{DB: db}

	// WebAuthn / Passkeys
	var passkeyH *api.PasskeyHandler
	if cfg.WebAuthnRPID != "" {
		waConfig := &webauthn.Config{
			RPID:          cfg.WebAuthnRPID,
			RPDisplayName: cfg.WebAuthnRPDisplayName,
			RPOrigins:     cfg.WebAuthnRPOrigins,
			AuthenticatorSelection: protocol.AuthenticatorSelection{
				ResidentKey:      protocol.ResidentKeyRequirementPreferred,
				UserVerification: protocol.VerificationPreferred,
			},
		}
		wa, waErr := webauthn.New(waConfig)
		if waErr != nil {
			slog.Error("failed to create WebAuthn", "error", waErr)
		} else {
			passkeyH = api.NewPasskeyHandler(ctx, db, authMw, wa)
		}
	}

	mux := http.NewServeMux()

	// Rate limiter for login: 5 attempts per minute per IP (burst 5).
	loginRL := middleware.NewIPRateLimiter(rate.Limit(5.0/60.0), 5)

	// --- Public routes (no auth) ---
	mux.Handle("POST /api/auth/login", middleware.RateLimit(loginRL)(http.HandlerFunc(authH.HandleLogin)))
	mux.HandleFunc("GET /api/auth/setup", authH.HandleSetupCheck)
	mux.HandleFunc("POST /api/auth/setup", authH.HandleSetup)

	// Passkey login (public, no auth required)
	if passkeyH != nil {
		mux.Handle("POST /api/auth/passkey/login/begin", middleware.RateLimit(loginRL)(http.HandlerFunc(passkeyH.HandleBeginLogin)))
		mux.Handle("POST /api/auth/passkey/login/finish", middleware.RateLimit(loginRL)(http.HandlerFunc(passkeyH.HandleFinishLogin)))
	}

	// --- Seed OIDC settings from env vars if DB row is empty ---
	if cfg.OIDCEnabled {
		oidcDB, _ := db.GetOIDCSettings(context.Background())
		if oidcDB != nil && oidcDB.IssuerURL == "" {
			oidcDB.Enabled = cfg.OIDCEnabled
			oidcDB.IssuerURL = cfg.OIDCIssuerURL
			oidcDB.ClientID = cfg.OIDCClientID
			oidcDB.ClientSecret = cfg.OIDCClientSecret
			oidcDB.RedirectURL = cfg.OIDCRedirectURL
			oidcDB.AutoCreate = cfg.OIDCAutoCreate
			oidcDB.AdminGroup = cfg.OIDCAdminGroup
			if len(cfg.OIDCScopes) > 0 {
				oidcDB.Scopes = strings.Join(cfg.OIDCScopes, ",")
			}
			_ = db.UpdateOIDCSettings(context.Background(), oidcDB)
			slog.Info("seeded OIDC settings from environment variables")
		}
	}

	// --- OIDC routes (always registered, handler checks if enabled) ---
	oidcH := auth.NewOIDCHandler(auth.OIDCConfig{}, db, authMw)
	oidcH.ConfigLoader = func() (*auth.OIDCConfig, error) {
		s, err := db.GetOIDCSettings(context.Background())
		if err != nil {
			return nil, err
		}
		var scopes []string
		for _, sc := range strings.Split(s.Scopes, ",") {
			sc = strings.TrimSpace(sc)
			if sc != "" {
				scopes = append(scopes, sc)
			}
		}
		return &auth.OIDCConfig{
			Enabled:      s.Enabled,
			IssuerURL:    s.IssuerURL,
			ClientID:     s.ClientID,
			ClientSecret: s.ClientSecret,
			RedirectURL:  s.RedirectURL,
			Scopes:       scopes,
			AutoCreate:   s.AutoCreate,
			AdminGroup:   s.AdminGroup,
		}, nil
	}
	mux.HandleFunc("GET /api/auth/oidc/login", oidcH.HandleLogin)
	mux.HandleFunc("GET /api/auth/oidc/callback", oidcH.HandleCallback)

	mux.HandleFunc("GET /api/auth/oidc/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		s, err := db.GetOIDCSettings(r.Context())
		if err != nil || !s.Enabled {
			_, _ = w.Write([]byte(`{"enabled":false}`))
		} else {
			_, _ = w.Write([]byte(`{"enabled":true}`))
		}
	})

	mux.HandleFunc("GET /api/auth/passkey/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if passkeyH != nil {
			_, _ = w.Write([]byte(`{"enabled":true}`))
		} else {
			_, _ = w.Write([]byte(`{"enabled":false}`))
		}
	})

	// --- Kubernetes health probes ---
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		if err := db.HealthCheck(r.Context()); err != nil {
			http.Error(w, `{"status":"not ready"}`, http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// --- Metrics (Prometheus) ---
	mux.Handle("GET /metrics", promhttp.Handler())

	// --- API Documentation ---
	if len(cfg.OpenAPISpec) > 0 {
		mux.HandleFunc("GET /api/spec", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/yaml")
			w.Header().Set("Cache-Control", "public, max-age=3600")
			_, _ = w.Write(cfg.OpenAPISpec)
		})
		mux.HandleFunc("GET /api/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write(scalarHTML)
		})
	}

	// --- Protected routes ---
	protected := http.NewServeMux()

	// Auth
	protected.HandleFunc("POST /api/auth/logout", authH.HandleLogout)
	protected.HandleFunc("POST /api/auth/2fa/enable", authH.Handle2FAEnable)
	protected.HandleFunc("POST /api/auth/2fa/disable", authH.Handle2FADisable)
	protected.HandleFunc("POST /api/auth/2fa/verify", authH.Handle2FAVerify)
	protected.HandleFunc("POST /api/auth/2fa/recovery-codes", authH.HandleRegenerateRecoveryCodes)

	// User
	protected.HandleFunc("GET /api/user", userH.HandleGetUser)
	protected.HandleFunc("POST /api/user/username", userH.HandleUpdateUsername)
	protected.HandleFunc("POST /api/user/password", userH.HandleUpdatePassword)

	// Sessions
	protected.HandleFunc("GET /api/sessions", sessionH.HandleListSessions)
	protected.HandleFunc("DELETE /api/sessions/{id}", sessionH.HandleRevokeSession)
	protected.HandleFunc("DELETE /api/sessions", sessionH.HandleRevokeAllSessions)

	// Passkeys (protected)
	if passkeyH != nil {
		protected.HandleFunc("GET /api/passkeys", passkeyH.HandleList)
		protected.HandleFunc("POST /api/passkeys/register/begin", passkeyH.HandleBeginRegistration)
		protected.HandleFunc("POST /api/passkeys/register/finish", passkeyH.HandleFinishRegistration)
		protected.HandleFunc("DELETE /api/passkeys/{id}", passkeyH.HandleDelete)
		protected.HandleFunc("POST /api/passkeys/{id}/rename", passkeyH.HandleRename)
	}

	// Admin User Management
	protected.HandleFunc("GET /api/admin/users", adminH.HandleListUsers)
	protected.HandleFunc("POST /api/admin/users", adminH.HandleCreateUser)
	protected.HandleFunc("DELETE /api/admin/users/{id}", adminH.HandleDeleteUser)
	protected.HandleFunc("POST /api/admin/users/{id}/reset-password", adminH.HandleResetUserPassword)
	protected.HandleFunc("POST /api/admin/users/{id}/toggle-admin", adminH.HandleToggleAdmin)

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
	protected.HandleFunc("GET /api/instances/{id}/health", appsH.HandleHealthCheckInstance)
	protected.HandleFunc("POST /api/instances/test", appsH.HandleTestConnection)

	// History
	protected.HandleFunc("GET /api/history", historyH.HandleListHistory)
	protected.HandleFunc("DELETE /api/history/{app}", historyH.HandleDeleteHistory)

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
	protected.HandleFunc("GET /api/queue/download-client/{app}", queueH.HandleGetDownloadClientSettings)
	protected.HandleFunc("PUT /api/queue/download-client/{app}", queueH.HandleUpdateDownloadClientSettings)

	// Blocklist Sources & Rules
	protected.HandleFunc("GET /api/blocklist/sources", blocklistH.HandleListSources)
	protected.HandleFunc("POST /api/blocklist/sources", blocklistH.HandleCreateSource)
	protected.HandleFunc("GET /api/blocklist/sources/{id}", blocklistH.HandleGetSource)
	protected.HandleFunc("PUT /api/blocklist/sources/{id}", blocklistH.HandleUpdateSource)
	protected.HandleFunc("DELETE /api/blocklist/sources/{id}", blocklistH.HandleDeleteSource)
	protected.HandleFunc("GET /api/blocklist/rules", blocklistH.HandleListRules)
	protected.HandleFunc("POST /api/blocklist/rules", blocklistH.HandleCreateRule)
	protected.HandleFunc("DELETE /api/blocklist/rules/{id}", blocklistH.HandleDeleteRule)

	// Notifications
	protected.HandleFunc("GET /api/notifications/providers", notificationH.HandleListProviders)
	protected.HandleFunc("POST /api/notifications/providers", notificationH.HandleCreateProvider)
	protected.HandleFunc("GET /api/notifications/providers/{id}", notificationH.HandleGetProvider)
	protected.HandleFunc("PUT /api/notifications/providers/{id}", notificationH.HandleUpdateProvider)
	protected.HandleFunc("DELETE /api/notifications/providers/{id}", notificationH.HandleDeleteProvider)
	protected.HandleFunc("POST /api/notifications/providers/{id}/test", notificationH.HandleTestProvider)

	// Download Client Instances
	protected.HandleFunc("GET /api/download-clients", dlClientH.HandleList)
	protected.HandleFunc("POST /api/download-clients", dlClientH.HandleCreate)
	protected.HandleFunc("PUT /api/download-clients/{id}", dlClientH.HandleUpdate)
	protected.HandleFunc("DELETE /api/download-clients/{id}", dlClientH.HandleDelete)
	protected.HandleFunc("GET /api/download-clients/{id}/health", dlClientH.HandleHealthCheck)
	protected.HandleFunc("GET /api/download-clients/{id}/status", dlClientH.HandleStatus)
	protected.HandleFunc("GET /api/download-clients/{id}/items", dlClientH.HandleItems)
	protected.HandleFunc("POST /api/download-clients/test", dlClientH.HandleTest)

	// Seerr
	protected.HandleFunc("GET /api/seerr/settings", seerrH.HandleGetSettings)
	protected.HandleFunc("PUT /api/seerr/settings", seerrH.HandleUpdateSettings)
	protected.HandleFunc("POST /api/seerr/test", seerrH.HandleTestConnection)
	protected.HandleFunc("GET /api/seerr/requests", seerrH.HandleGetRequests)
	protected.HandleFunc("GET /api/seerr/requests/count", seerrH.HandleGetRequestCount)
	protected.HandleFunc("POST /api/seerr/scan-duplicates", seerrH.HandleScanDuplicates)

	// OIDC Settings
	protected.HandleFunc("GET /api/oidc/settings", oidcSettingsH.HandleGetSettings)
	protected.HandleFunc("PUT /api/oidc/settings", oidcSettingsH.HandleUpdateSettings)

	// Instance Groups
	protected.HandleFunc("GET /api/instance-groups/{app}", groupsH.HandleListGroups)
	protected.HandleFunc("POST /api/instance-groups/{app}", groupsH.HandleCreateGroup)
	protected.HandleFunc("GET /api/instance-groups/by-id/{id}", groupsH.HandleGetGroup)
	protected.HandleFunc("PUT /api/instance-groups/by-id/{id}", groupsH.HandleUpdateGroup)
	protected.HandleFunc("DELETE /api/instance-groups/by-id/{id}", groupsH.HandleDeleteGroup)
	protected.HandleFunc("PUT /api/instance-groups/by-id/{id}/members", groupsH.HandleSetMembers)
	protected.HandleFunc("GET /api/instance-groups/by-id/{id}/overlaps", groupsH.HandleListOverlaps)
	protected.HandleFunc("GET /api/instance-groups/actions", groupsH.HandleListActions)

	// Mount protected routes with auth + CSRF + token injection.
	// csrfPlaintextHTTP must wrap the CSRF middleware so gorilla/csrf knows when
	// the request arrives over plain HTTP (no TLS, no X-Forwarded-Proto: https).
	mux.Handle("/api/", csrfPlaintextHTTP(authMw.CSRFProtect()(csrfInjectToken(authMw.RequireAuth(protected)))))

	// Serve embedded frontend SPA (if built).
	if cfg.FrontendFS != nil {
		mux.Handle("/", spaHandler(cfg.FrontendFS))
	}

	// Global middleware chain
	corsMiddleware := middleware.CORS(middleware.CORSConfig{AllowedOrigins: cfg.AllowedOrigins})
	var handler http.Handler
	handler = middleware.Chain(mux,
		middleware.Recovery,
		middleware.RequestID,
		middleware.Logging,
		corsMiddleware,
	)

	// Strip base path prefix if configured (for sub-path reverse proxy hosting).
	if cfg.BasePath != "" {
		handler = http.StripPrefix(cfg.BasePath, handler)
		slog.Info("base path configured", "base_path", cfg.BasePath)
	}

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
