package middleware

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/lusoris/lurkarr/internal/metrics"
	"golang.org/x/time/rate"
)

// TrustedProxies holds the configured trusted proxy CIDRs for the middleware package.
// Set during server initialization.
var TrustedProxies []*net.IPNet

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
				metrics.HTTPRateLimitHits.WithLabelValues(normalizePath(r.URL.Path)).Inc()
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"too many requests"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// extractIP gets the client IP, stripping the port.
// Only trusts X-Forwarded-For when the direct connection is from a trusted proxy.
func extractIP(r *http.Request) string {
	directIP, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		directIP = r.RemoteAddr
	}

	// Only parse XFF if the direct connection is from a trusted proxy.
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" && isTrustedProxy(directIP) {
		// Take the rightmost non-trusted IP (real client).
		parts := strings.Split(xff, ",")
		for i := len(parts) - 1; i >= 0; i-- {
			ip := strings.TrimSpace(parts[i])
			if ip != "" && !isTrustedProxy(ip) {
				return ip
			}
		}
		// All XFF IPs are trusted; use the leftmost.
		if first := strings.TrimSpace(parts[0]); first != "" {
			return first
		}
	}
	return directIP
}

func isTrustedProxy(ipStr string) bool {
	if len(TrustedProxies) == 0 {
		return false
	}
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return false
	}
	for _, n := range TrustedProxies {
		if n.Contains(ip) {
			return true
		}
	}
	return false
}
