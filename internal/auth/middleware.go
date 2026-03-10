package auth

import (
	"context"
	"log/slog"
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

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *database.User {
	u, _ := ctx.Value(userContextKey).(*database.User)
	return u
}

// Middleware provides auth and CSRF middleware.
type Middleware struct {
	DB              *database.DB
	ProxyAuthBypass bool
	ProxyHeader     string
	CSRFKey         []byte
}

// RequireAuth is middleware that checks for a valid session cookie.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Proxy auth bypass (e.g., Authelia, Authentik)
		if m.ProxyAuthBypass && m.ProxyHeader != "" {
			if username := r.Header.Get(m.ProxyHeader); username != "" {
				user, err := m.DB.GetUserByUsername(r.Context(), username)
				if err == nil {
					ctx := context.WithValue(r.Context(), userContextKey, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
				slog.Warn("proxy auth bypass user not found", "username", username)
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
func (m *Middleware) SetSessionCookie(ctx context.Context, w http.ResponseWriter, userID uuid.UUID) error {
	session, err := m.DB.CreateSession(ctx, userID, sessionDuration)
	if err != nil {
		return err
	}

	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    session.ID.String(),
		Path:     "/",
		HttpOnly: true,
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
