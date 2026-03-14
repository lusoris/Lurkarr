package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	"golang.org/x/oauth2"
)

// OIDCStore abstracts database operations needed by the OIDC handler.
type OIDCStore interface {
	GetOrCreateExternalUser(ctx context.Context, provider, externalID, username string) (*database.User, error)
	CreateSession(ctx context.Context, userID uuid.UUID, duration time.Duration) (*database.Session, error)
	UpdateUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error
}

// OIDCConfig holds the OIDC provider configuration.
type OIDCConfig struct {
	Enabled      bool
	IssuerURL    string
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
	AutoCreate   bool
	AdminGroup   string
}

// OIDCHandler manages OIDC authentication flows.
type OIDCHandler struct {
	Config       OIDCConfig
	DB           OIDCStore
	Auth         *Middleware
	provider     *oidc.Provider
	verifier     *oidc.IDTokenVerifier
	oauth2Config *oauth2.Config
	initOnce     sync.Once

	// ConfigLoader optionally loads config from DB (overrides static Config).
	ConfigLoader func() (*OIDCConfig, error)

	// Tracks the issuer URL used for the current provider so we can re-init on change.
	currentIssuer string
	providerMu    sync.RWMutex

	// CSRF state storage (in-memory, short-lived).
	states   map[string]time.Time
	statesMu sync.Mutex
}

// NewOIDCHandler creates a new OIDC handler. Call Init() before use.
func NewOIDCHandler(cfg OIDCConfig, db OIDCStore, authMw *Middleware) *OIDCHandler {
	return &OIDCHandler{
		Config: cfg,
		DB:     db,
		Auth:   authMw,
		states: make(map[string]time.Time),
	}
}

// loadConfig returns the effective OIDC config, preferring DB if a loader is set.
func (h *OIDCHandler) loadConfig() OIDCConfig {
	if h.ConfigLoader != nil {
		cfg, err := h.ConfigLoader()
		if err != nil {
			slog.Warn("failed to load OIDC config from DB, using static config", "error", err)
			return h.Config
		}
		return *cfg
	}
	return h.Config
}

// Init lazily initializes the OIDC provider (requires network call to issuer).
func (h *OIDCHandler) Init(ctx context.Context) error {
	cfg := h.loadConfig()
	if !cfg.Enabled || cfg.IssuerURL == "" {
		return fmt.Errorf("oidc not enabled or issuer URL not configured")
	}

	h.providerMu.RLock()
	sameIssuer := h.currentIssuer == cfg.IssuerURL && h.provider != nil
	h.providerMu.RUnlock()

	if sameIssuer {
		// Update oauth2Config in case other fields changed (client ID, secret, etc.).
		h.providerMu.Lock()
		h.oauth2Config = &oauth2.Config{
			ClientID:     cfg.ClientID,
			ClientSecret: cfg.ClientSecret,
			RedirectURL:  cfg.RedirectURL,
			Endpoint:     h.provider.Endpoint(),
			Scopes:       cfg.Scopes,
		}
		h.verifier = h.provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
		h.Config = cfg
		h.providerMu.Unlock()
		return nil
	}

	// Issuer changed or first init — do provider discovery.
	h.providerMu.Lock()
	defer h.providerMu.Unlock()

	provider, err := oidc.NewProvider(ctx, cfg.IssuerURL)
	if err != nil {
		return fmt.Errorf("oidc provider discovery: %w", err)
	}
	h.provider = provider
	h.currentIssuer = cfg.IssuerURL
	h.verifier = provider.Verifier(&oidc.Config{ClientID: cfg.ClientID})
	h.oauth2Config = &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RedirectURL:  cfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       cfg.Scopes,
	}
	h.Config = cfg

	// Start state cleanup goroutine (only once).
	h.initOnce.Do(func() {
		go h.cleanupStates(ctx)
	})

	return nil
}

// HandleLogin redirects the user to the OIDC provider's authorization endpoint.
func (h *OIDCHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if err := h.Init(r.Context()); err != nil {
		slog.Error("OIDC provider initialization failed", "error", err)
		http.Error(w, `{"error":"oidc provider unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	state, err := generateState()
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	h.statesMu.Lock()
	h.states[state] = time.Now().Add(10 * time.Minute)
	h.statesMu.Unlock()

	url := h.oauth2Config.AuthCodeURL(state, oauth2.S256ChallengeOption(state))
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// HandleCallback processes the OIDC callback after provider authentication.
func (h *OIDCHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	if err := h.Init(r.Context()); err != nil {
		slog.Error("OIDC provider initialization failed", "error", err)
		http.Error(w, `{"error":"oidc provider unavailable"}`, http.StatusServiceUnavailable)
		return
	}

	// Validate state parameter.
	state := r.URL.Query().Get("state")
	if !h.consumeState(state) {
		http.Error(w, `{"error":"invalid or expired state"}`, http.StatusBadRequest)
		return
	}

	// Check for error from the provider.
	if errParam := r.URL.Query().Get("error"); errParam != "" {
		desc := r.URL.Query().Get("error_description")
		slog.Warn("OIDC provider returned error", "error", errParam, "description", desc) //nolint:gosec // G706
		http.Error(w, fmt.Sprintf(`{"error":"oidc: %s"}`, errParam), http.StatusUnauthorized)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, `{"error":"missing authorization code"}`, http.StatusBadRequest)
		return
	}

	// Exchange code for tokens.
	token, err := h.oauth2Config.Exchange(r.Context(), code, oauth2.S256ChallengeOption(state))
	if err != nil {
		slog.Error("OIDC token exchange failed", "error", err)
		http.Error(w, `{"error":"token exchange failed"}`, http.StatusUnauthorized)
		return
	}

	// Extract and verify ID token.
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, `{"error":"no id_token in response"}`, http.StatusUnauthorized)
		return
	}

	idToken, err := h.verifier.Verify(r.Context(), rawIDToken)
	if err != nil {
		slog.Error("OIDC ID token verification failed", "error", err)
		http.Error(w, `{"error":"invalid id_token"}`, http.StatusUnauthorized)
		return
	}

	// Extract claims.
	var claims struct {
		Subject   string   `json:"sub"`
		Email     string   `json:"email"`
		Name      string   `json:"name"`
		Preferred string   `json:"preferred_username"`
		Groups    []string `json:"groups"`
	}
	if err := idToken.Claims(&claims); err != nil {
		slog.Error("OIDC claims extraction failed", "error", err)
		http.Error(w, `{"error":"failed to parse claims"}`, http.StatusInternalServerError)
		return
	}

	// Determine username: prefer preferred_username, then email, then sub.
	username := claims.Preferred
	if username == "" {
		username = claims.Email
	}
	if username == "" {
		username = claims.Subject
	}

	// Get or create the user.
	user, err := h.DB.GetOrCreateExternalUser(r.Context(), "oidc", claims.Subject, username)
	if err != nil {
		slog.Error("OIDC user lookup/creation failed", "error", err, "sub", claims.Subject) //nolint:gosec // G706
		http.Error(w, `{"error":"user creation failed"}`, http.StatusInternalServerError)
		return
	}

	// Apply group/role mapping if an admin group is configured.
	if h.Config.AdminGroup != "" {
		isAdmin := containsGroup(claims.Groups, h.Config.AdminGroup)
		if isAdmin != user.IsAdmin {
			if err := h.DB.UpdateUserAdmin(r.Context(), user.ID, isAdmin); err != nil {
				slog.Error("failed to update user admin status", "error", err, "user", username)
			} else {
				user.IsAdmin = isAdmin
				slog.Info("OIDC group mapping updated admin status", "username", username, "is_admin", isAdmin)
			}
		}
	}

	// Create session.
	if err := h.Auth.SetSessionCookie(r.Context(), w, r, user.ID); err != nil {
		slog.Error("OIDC session creation failed", "error", err)
		http.Error(w, `{"error":"session creation failed"}`, http.StatusInternalServerError)
		return
	}

	slog.Info("OIDC login successful", "username", username, "sub", claims.Subject) //nolint:gosec // G706

	// Redirect to the home page (frontend will pick up the session cookie).
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// HandleProviderInfo returns the OIDC configuration status as JSON (for the frontend login button).
func (h *OIDCHandler) HandleProviderInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	resp := map[string]any{
		"enabled": h.Config.Enabled,
	}
	if h.Config.Enabled {
		resp["issuer"] = h.Config.IssuerURL
	}
	data, _ := json.Marshal(resp)
	_, _ = w.Write(data)
}

func (h *OIDCHandler) consumeState(state string) bool {
	h.statesMu.Lock()
	defer h.statesMu.Unlock()
	exp, ok := h.states[state]
	if !ok {
		return false
	}
	delete(h.states, state)
	return time.Now().Before(exp)
}

func (h *OIDCHandler) cleanupStates(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			h.statesMu.Lock()
			now := time.Now()
			for k, exp := range h.states {
				if now.After(exp) {
					delete(h.states, k)
				}
			}
			h.statesMu.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// containsGroup checks if a group list contains the target group (case-insensitive).
func containsGroup(groups []string, target string) bool {
	for _, g := range groups {
		if strings.EqualFold(g, target) {
			return true
		}
	}
	return false
}
