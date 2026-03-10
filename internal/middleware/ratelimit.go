package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// IPRateLimiter tracks per-IP rate limiters.
type IPRateLimiter struct {
	mu       sync.Mutex
	limiters map[string]*entry
	rate     rate.Limit
	burst    int
}

type entry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewIPRateLimiter creates a limiter allowing r requests/sec with the given burst.
func NewIPRateLimiter(r rate.Limit, burst int) *IPRateLimiter {
	rl := &IPRateLimiter{
		limiters: make(map[string]*entry),
		rate:     r,
		burst:    burst,
	}
	go rl.cleanup()
	return rl
}

// Allow reports whether a request from the given IP is allowed.
func (rl *IPRateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	e, ok := rl.limiters[ip]
	if !ok {
		e = &entry{
			limiter: rate.NewLimiter(rl.rate, rl.burst),
		}
		rl.limiters[ip] = e
	}
	e.lastSeen = time.Now()
	rl.mu.Unlock()
	return e.limiter.Allow()
}

// cleanup removes stale entries every 3 minutes.
func (rl *IPRateLimiter) cleanup() {
	for {
		time.Sleep(3 * time.Minute)
		rl.mu.Lock()
		for ip, e := range rl.limiters {
			if time.Since(e.lastSeen) > 5*time.Minute {
				delete(rl.limiters, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit wraps a handler with per-IP rate limiting.
func RateLimit(limiter *IPRateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := extractIP(r)
			if !limiter.Allow(ip) {
					slog.Warn("rate limit exceeded", "ip", ip, "path", r.URL.Path) //nolint:gosec // G706: slog structured logging
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractIP gets the client IP, stripping the port.
func extractIP(r *http.Request) string {
	// Check X-Forwarded-For first (common behind reverse proxies).
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		// Take the first IP (client IP).
		if idx := len(xff); idx > 0 {
			for i, c := range xff {
				if c == ',' {
					return xff[:i]
				}
			}
			return xff
		}
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
