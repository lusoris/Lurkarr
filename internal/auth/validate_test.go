package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	"go.uber.org/mock/gomock"
)

func TestValidatePassword_Valid(t *testing.T) {
	if err := ValidatePassword("StrongP1ss"); err != nil {
		t.Errorf("expected valid password, got error: %v", err)
	}
}

func TestValidatePassword_TooShort(t *testing.T) {
	err := ValidatePassword("Ab1")
	if err == nil {
		t.Fatal("expected error for short password")
	}
}

func TestValidatePassword_NoUppercase(t *testing.T) {
	err := ValidatePassword("lowercase1")
	if err == nil {
		t.Fatal("expected error for missing uppercase")
	}
}

func TestValidatePassword_NoLowercase(t *testing.T) {
	err := ValidatePassword("UPPERCASE1")
	if err == nil {
		t.Fatal("expected error for missing lowercase")
	}
}

func TestValidatePassword_NoDigit(t *testing.T) {
	err := ValidatePassword("NoDigitsHere")
	if err == nil {
		t.Fatal("expected error for missing digit")
	}
}

func TestValidatePassword_ExactlyEightChars(t *testing.T) {
	if err := ValidatePassword("Abcdef1!"); err != nil {
		t.Errorf("8-char password should be valid: %v", err)
	}
}

func TestValidatePassword_SevenChars(t *testing.T) {
	err := ValidatePassword("Abcde1!")
	if err == nil {
		t.Fatal("7-char password should be rejected")
	}
}

func TestValidatePassword_Empty(t *testing.T) {
	err := ValidatePassword("")
	if err == nil {
		t.Fatal("empty password should be rejected")
	}
}

func TestExtractRemoteIP_WithPort(t *testing.T) {
	tests := []struct {
		addr string
		want string
	}{
		{"192.168.1.1:8080", "192.168.1.1"},
		{"10.0.0.1:443", "10.0.0.1"},
		{"[::1]:8080", "::1"},
	}
	for _, tt := range tests {
		t.Run(tt.addr, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
			req.RemoteAddr = tt.addr
			got := extractRemoteIP(req)
			if got != tt.want {
				t.Errorf("extractRemoteIP(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

func TestExtractRemoteIP_NoPort(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.RemoteAddr = "192.168.1.1"
	got := extractRemoteIP(req)
	if got != "192.168.1.1" {
		t.Errorf("expected raw addr, got %q", got)
	}
}

func TestIsTrustedProxy_Match(t *testing.T) {
	nets := testTrustedNets() // 192.0.2.0/24
	if !isTrustedProxy(nets, "192.0.2.50") {
		t.Error("expected 192.0.2.50 to be trusted")
	}
}

func TestIsTrustedProxy_NoMatch(t *testing.T) {
	nets := testTrustedNets()
	if isTrustedProxy(nets, "10.0.0.1") {
		t.Error("expected 10.0.0.1 to be untrusted")
	}
}

func TestIsTrustedProxy_InvalidIP(t *testing.T) {
	nets := testTrustedNets()
	if isTrustedProxy(nets, "not-an-ip") {
		t.Error("expected invalid IP to be untrusted")
	}
}

func TestIsTrustedProxy_EmptyNets(t *testing.T) {
	if isTrustedProxy(nil, "192.0.2.1") {
		t.Error("expected no trusted nets to reject all IPs")
	}
}

func TestIsTrustedProxy_EmptyIP(t *testing.T) {
	nets := testTrustedNets()
	if isTrustedProxy(nets, "") {
		t.Error("expected empty IP to be untrusted")
	}
}

func TestSetSessionCookie_UserAgentTruncation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockAuthStore(ctrl)
	userID := uuid.New()
	longUA := strings.Repeat("A", 600)

	store.EXPECT().CreateSessionWithMeta(
		gomock.Any(), userID, gomock.Any(),
		gomock.Any(),
		gomock.Len(512), // UserAgent should be truncated to 512
	).Return(&database.Session{
		ID: uuid.New(), UserID: userID,
	}, nil)

	m := &Middleware{DB: store}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", http.NoBody)
	req.Header.Set("User-Agent", longUA)

	err := m.SetSessionCookie(context.Background(), rec, req, userID)
	if err != nil {
		t.Fatalf("SetSessionCookie error: %v", err)
	}
}

func TestContextWithUser(t *testing.T) {
	user := &database.User{ID: uuid.New(), Username: "testuser"}
	ctx := ContextWithUser(context.Background(), user)
	got := UserFromContext(ctx)
	if got == nil || got.Username != "testuser" {
		t.Errorf("expected user 'testuser', got %v", got)
	}
}
