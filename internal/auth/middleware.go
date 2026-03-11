package auth

//go:generate mockgen -destination=mock_authstore_test.go -package=auth github.com/lusoris/lurkarr/internal/auth AuthStore

import (
	"context"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/lusoris/lurkarr/internal/database"
)

type contextKey string

const userContextKey contextKey = "user"

const sessionCookieName = "lurkarr_session"
const sessionDuration = 7 * 24 * time.Hour

// AuthStore abstracts the database operations needed by auth middleware.
type AuthStore interface {
	GetUserByUsername(ctx context.Context, username string) (*database.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*database.User, error)
	GetSession(ctx context.Context, id uuid.UUID) (*database.Session, error)
	CreateSession(ctx context.Context, userID uuid.UUID, duration time.Duration) (*database.Session, error)
	DeleteSession(ctx context.Context, id uuid.UUID) error
	CreateUser(ctx context.Context, username, passwordHash string) (*database.User, error)
}

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *database.User {
	u, _ := ctx.Value(userContextKey).(*database.User)
	return u
}

// ContextWithUser returns a new context with the given user stored in it.
func ContextWithUser(ctx context.Context, user *database.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}

// Middleware provides auth and CSRF middleware.
type Middleware struct {
	DB              AuthStore
	ProxyAuthBypass bool
	ProxyHeaders    []string
	TrustedProxies  []*net.IPNet
	ProxyAutoCreate bool
	CSRFKey         []byte
	SecureCookie    bool
}

// RequireAuth is middleware that checks for a valid session cookie.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proxy auth bypass (e.g., Authelia, Authentik)
		if m.ProxyAuthBypass && len(m.ProxyHeaders) > 0 {
			// Try each configured header in order; use the first non-empty one.
			var username string
			for _, hdr := range m.ProxyHeaders {
				if v := r.Header.Get(hdr); v != "" {
					username = v
					break
				}
			}
			if username != "" {
				// Validate that the request came from a trusted proxy.
				remoteIP := extractRemoteIP(r)
				if !isTrustedProxy(m.TrustedProxies, remoteIP) {
					slog.Warn("proxy auth header from untrusted IP", "ip", remoteIP, "headers", m.ProxyHeaders) //nolint:gosec // G706
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				user, err := m.DB.GetUserByUsername(r.Context(), username)
				if err != nil && m.ProxyAutoCreate {
					// Auto-create user on first proxy auth login.
					randomPass, genErr := GenerateSecretKey(32)
					if genErr != nil {
						slog.Error("failed to generate password for proxy user", "error", genErr)
						http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
						return
					}
					hash, hashErr := HashPassword(randomPass)
					if hashErr != nil {
						slog.Error("failed to hash password for proxy user", "error", hashErr)
						http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
						return
					}
					user, err = m.DB.CreateUser(r.Context(), username, hash)
					if err != nil {
						slog.Error("failed to auto-create proxy user", "error", err, "username", username) //nolint:gosec // G706
						http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
						return
					}
					slog.Info("auto-created user from proxy auth", "username", username) //nolint:gosec // G706
				} else if err != nil {
					slog.Warn("proxy auth bypass user not found", "username", username) //nolint:gosec // G706
					http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				ctx := context.WithValue(r.Context(), userContextKey, user)
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
		}

		cookie, err := r.Cookie(sessionCookieName)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		sessionID, err := uuid.Parse(cookie.Value)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		session, err := m.DB.GetSession(r.Context(), sessionID)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		user, err := m.DB.GetUserByID(r.Context(), session.UserID)
		if err != nil {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// CSRFProtect returns gorilla/csrf middleware configured for SPA usage.
func (m *Middleware) CSRFProtect() func(http.Handler) http.Handler {
	return csrf.Protect(
		m.CSRFKey,
		csrf.SameSite(csrf.SameSiteLaxMode),
		csrf.Secure(false), // Handled by reverse proxy TLS
		csrf.Path("/"),
		csrf.RequestHeader("X-CSRF-Token"),
		csrf.ErrorHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error":"csrf validation failed"}`, http.StatusForbidden)
		})),
	)
}

// SetSessionCookie creates a session and sets the cookie.
func (m *Middleware) SetSessionCookie(ctx context.Context, w http.ResponseWriter, r *http.Request, userID uuid.UUID) error {
	session, err := m.DB.CreateSession(ctx, userID, sessionDuration)
	if err != nil {
		return err
	}

	secure := m.SecureCookie || r.Header.Get("X-Forwarded-Proto") == "https"

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionDuration.Seconds()),
	})
	return nil
}

// ClearSessionCookie deletes the session cookie and DB record.
func (m *Middleware) ClearSessionCookie(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		if id, err := uuid.Parse(cookie.Value); err == nil {
			_ = m.DB.DeleteSession(ctx, id)
		}
	}
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
}

// extractRemoteIP gets the direct remote IP (no proxy header parsing).
func extractRemoteIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// isTrustedProxy checks if an IP is within one of the trusted proxy CIDRs.
func isTrustedProxy(trustedNets []*net.IPNet, ipStr string) bool {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range trustedNets {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
