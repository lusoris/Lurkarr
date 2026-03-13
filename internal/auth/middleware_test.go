package auth

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// testTrustedNets returns CIDRs that include the httptest default RemoteAddr (192.0.2.1).
func testTrustedNets() []*net.IPNet {
	_, cidr, _ := net.ParseCIDR("192.0.2.0/24")
	return []*net.IPNet{cidr}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
}

func TestRequireAuth_NoCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	m := &Middleware{DB: store}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_InvalidCookieValue(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	m := &Middleware{DB: store}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "not-a-uuid"})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_SessionNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	sessID := uuid.New()
	store.EXPECT().GetSession(gomock.Any(), sessID).Return(nil, errors.New("not found"))
	m := &Middleware{DB: store}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessID.String()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	sessID := uuid.New()
	userID := uuid.New()
	store.EXPECT().GetSession(gomock.Any(), sessID).Return(&database.Session{ID: sessID, UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, nil)
	store.EXPECT().GetUserByID(gomock.Any(), userID).Return(nil, errors.New("user not found"))
	m := &Middleware{DB: store}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessID.String()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ValidSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	sessID := uuid.New()
	user := &database.User{ID: userID, Username: "testuser"}
	store.EXPECT().GetSession(gomock.Any(), sessID).Return(&database.Session{ID: sessID, UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, nil)
	store.EXPECT().GetUserByID(gomock.Any(), userID).Return(user, nil)
	m := &Middleware{DB: store}

	var gotUser *database.User
	handler := m.RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotUser = UserFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessID.String()})
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUser == nil || gotUser.Username != "testuser" {
		t.Errorf("expected user 'testuser' in context, got %v", gotUser)
	}
}

func TestRequireAuth_ProxyBypass(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	user := &database.User{ID: userID, Username: "proxyuser"}
	store.EXPECT().GetUserByUsername(gomock.Any(), "proxyuser").Return(user, nil)
	m := &Middleware{
		DB:              store,
		ProxyAuthBypass: true,
		ProxyHeaders:    []string{"X-Forwarded-User"},
		TrustedProxies:  testTrustedNets(),
	}

	var gotUser *database.User
	handler := m.RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotUser = UserFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Forwarded-User", "proxyuser")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUser == nil || gotUser.Username != "proxyuser" {
		t.Errorf("expected user 'proxyuser' in context, got %v", gotUser)
	}
}

func TestRequireAuth_ProxyBypassUserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	store.EXPECT().GetUserByUsername(gomock.Any(), "unknown").Return(nil, errors.New("not found"))
	m := &Middleware{
		DB:              store,
		ProxyAuthBypass: true,
		ProxyHeaders:    []string{"X-Forwarded-User"},
		TrustedProxies:  testTrustedNets(),
	}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Forwarded-User", "unknown")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ProxyBypassDisabled(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	m := &Middleware{
		DB:              store,
		ProxyAuthBypass: false,
		ProxyHeaders:    []string{"X-Forwarded-User"},
	}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Forwarded-User", "proxyuser")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Proxy bypass disabled, so falls through to cookie auth
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuth_ProxyBypassUntrustedIP(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	// No DB calls expected — rejected before lookup.
	m := &Middleware{
		DB:              store,
		ProxyAuthBypass: true,
		ProxyHeaders:    []string{"X-Forwarded-User"},
		TrustedProxies:  testTrustedNets(), // only 192.0.2.0/24
	}

	handler := m.RequireAuth(okHandler())
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.RemoteAddr = "5.5.5.5:1234" // untrusted IP
	req.Header.Set("X-Forwarded-User", "proxyuser")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401 from untrusted proxy, got %d", rec.Code)
	}
}

func TestRequireAuth_ProxyBypassAutoCreate(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	store.EXPECT().GetUserByUsername(gomock.Any(), "newuser").Return(nil, errors.New("not found"))
	store.EXPECT().CreateUser(gomock.Any(), "newuser", gomock.Any()).Return(&database.User{
		ID: userID, Username: "newuser",
	}, nil)
	m := &Middleware{
		DB:              store,
		ProxyAuthBypass: true,
		ProxyHeaders:    []string{"X-Forwarded-User"},
		TrustedProxies:  testTrustedNets(),
		ProxyAutoCreate: true,
	}

	var gotUser *database.User
	handler := m.RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotUser = UserFromContext(r.Context())
	}))

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Forwarded-User", "newuser")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUser == nil || gotUser.Username != "newuser" {
		t.Errorf("expected auto-created user 'newuser' in context, got %v", gotUser)
	}
}

func TestRequireAuth_ProxyMultiHeader(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	user := &database.User{ID: userID, Username: "authuser"}
	store.EXPECT().GetUserByUsername(gomock.Any(), "authuser").Return(user, nil)
	m := &Middleware{
		DB:              store,
		ProxyAuthBypass: true,
		ProxyHeaders:    []string{"Remote-User", "X-Forwarded-User", "X-authentik-username"},
		TrustedProxies:  testTrustedNets(),
	}

	var gotUser *database.User
	handler := m.RequireAuth(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		gotUser = UserFromContext(r.Context())
	}))

	// Only the second header is set — should use it.
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Forwarded-User", "authuser")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
	if gotUser == nil || gotUser.Username != "authuser" {
		t.Errorf("expected user 'authuser' in context, got %v", gotUser)
	}
}

func TestSetSessionCookie_XForwardedProto(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	store.EXPECT().CreateSessionWithMeta(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any()).Return(&database.Session{
		ID: uuid.New(), UserID: userID, ExpiresAt: time.Now().Add(time.Hour), CreatedAt: time.Now(),
	}, nil)
	m := &Middleware{DB: store, SecureCookie: false} // SecureCookie=false but XFP=https

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("X-Forwarded-Proto", "https")
	err := m.SetSessionCookie(context.Background(), rec, req, userID)
	if err != nil {
		t.Fatalf("SetSessionCookie error: %v", err)
	}
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName {
			if !c.Secure {
				t.Error("expected Secure when X-Forwarded-Proto=https")
			}
			return
		}
	}
	t.Error("session cookie not found")
}

func TestSetSessionCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	store.EXPECT().CreateSessionWithMeta(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any()).Return(&database.Session{
		ID:        uuid.New(),
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
		CreatedAt: time.Now(),
	}, nil)
	m := &Middleware{DB: store, SecureCookie: true}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	err := m.SetSessionCookie(context.Background(), rec, req, userID)
	if err != nil {
		t.Fatalf("SetSessionCookie error: %v", err)
	}

	cookies := rec.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == sessionCookieName {
			found = true
			if c.Value == "" {
				t.Error("cookie value is empty")
			}
			if !c.HttpOnly {
				t.Error("expected HttpOnly")
			}
			if !c.Secure {
				t.Error("expected Secure when SecureCookie=true")
			}
		}
	}
	if !found {
		t.Error("session cookie not found")
	}
}

func TestSetSessionCookie_CreateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	store.EXPECT().CreateSessionWithMeta(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
	m := &Middleware{DB: store}

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	err := m.SetSessionCookie(context.Background(), rec, req, uuid.New())
	if err == nil {
		t.Fatal("expected error from SetSessionCookie")
	}
}

func TestClearSessionCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	sessID := uuid.New()
	store.EXPECT().DeleteSession(gomock.Any(), sessID).Return(nil)
	m := &Middleware{DB: store}

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessID.String()})
	rec := httptest.NewRecorder()

	m.ClearSessionCookie(context.Background(), rec, req)

	// Cookie should be cleared
	cookies := rec.Result().Cookies()
	for _, c := range cookies {
		if c.Name == sessionCookieName && c.MaxAge != -1 {
			t.Error("expected MaxAge=-1 to clear cookie")
		}
	}
}

func TestClearSessionCookie_NoCookie(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	m := &Middleware{DB: store}

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	rec := httptest.NewRecorder()

	// Should not panic
	m.ClearSessionCookie(context.Background(), rec, req)
}

func TestClearSessionCookie_InvalidUUID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	m := &Middleware{DB: store}

	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "not-uuid"})
	rec := httptest.NewRecorder()

	// Should not panic
	m.ClearSessionCookie(context.Background(), rec, req)
}

func TestUserFromContext_Nil(t *testing.T) {
	u := UserFromContext(context.Background())
	if u != nil {
		t.Errorf("expected nil, got %v", u)
	}
}

func TestCSRFProtect(t *testing.T) {
	m := &Middleware{CSRFKey: make([]byte, 32)}
	protect := m.CSRFProtect()
	if protect == nil {
		t.Fatal("CSRFProtect() returned nil")
	}
}
