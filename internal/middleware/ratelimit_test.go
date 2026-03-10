package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/time/rate"
)

func TestIPRateLimiterAllow(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(5), 5) // 5/sec, burst 5

	// First 5 should be allowed (burst)
	for i := 0; i < 5; i++ {
		if !rl.Allow("1.2.3.4") {
			t.Fatalf("request %d should be allowed", i)
		}
	}

	// 6th should be denied (burst exhausted, not enough time elapsed)
	if rl.Allow("1.2.3.4") {
		t.Fatal("6th request should be denied")
	}

	// Different IP should still be allowed
	if !rl.Allow("5.6.7.8") {
		t.Fatal("different IP should be allowed")
	}
}

func TestRateLimitMiddleware(t *testing.T) {
	rl := NewIPRateLimiter(rate.Limit(1), 1) // 1/sec, burst 1

	handler := RateLimit(rl)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// First request: OK
	req := httptest.NewRequest("POST", "/api/auth/login", http.NoBody)
	req.RemoteAddr = "10.0.0.1:12345"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	// Second request: rate limited
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
}

func TestExtractIP(t *testing.T) {
	tests := []struct {
		name     string
		remote   string
		xff      string
		expected string
	}{
		{"remote with port", "1.2.3.4:5678", "", "1.2.3.4"},
		{"remote no port", "1.2.3.4", "", "1.2.3.4"},
		{"xff single", "10.0.0.1:1234", "9.8.7.6", "9.8.7.6"},
		{"xff multiple", "10.0.0.1:1234", "9.8.7.6, 10.0.0.2", "9.8.7.6"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest("GET", "/", http.NoBody)
			r.RemoteAddr = tt.remote
			if tt.xff != "" {
				r.Header.Set("X-Forwarded-For", tt.xff)
			}
			got := extractIP(r)
			if got != tt.expected {
				t.Fatalf("extractIP() = %q, want %q", got, tt.expected)
			}
		})
	}
}
