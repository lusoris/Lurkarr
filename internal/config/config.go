package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"
)

// Config holds all application configuration from environment variables.
type Config struct {
	DatabaseURL    string
	DBMaxConns     int32
	ListenAddr     string
	CSRFKey        string
	AllowedOrigins []string
	ProxyAuth      bool
	ProxyHeaders   []string
	TrustedProxies []*net.IPNet
	SecureCookie   bool
	LogLevel       string
	BasePath       string

	// OIDC configuration
	OIDCEnabled      bool
	OIDCIssuerURL    string
	OIDCClientID     string
	OIDCClientSecret string
	OIDCRedirectURL  string
	OIDCScopes       []string
	OIDCAutoCreate   bool
	OIDCAdminGroup   string

	// WebAuthn / Passkeys
	WebAuthnRPID          string
	WebAuthnRPDisplayName string
	WebAuthnRPOrigins     []string

	// Run-once mode
	RunOnce bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL:  getEnv("DATABASE_URL", ""),
		DBMaxConns:   defaultDBMaxConns(),
		ListenAddr:   getEnv("LISTEN_ADDR", ":8484"),
		CSRFKey:      getEnv("CSRF_KEY", ""),
		ProxyAuth:    getEnvBool("PROXY_AUTH", false),
		ProxyHeaders: splitAndTrim(getEnv("PROXY_HEADER", "Remote-User")),
		SecureCookie: getEnvBool("SECURE_COOKIE", false),
		LogLevel:     getEnv("LOG_LEVEL", "info"),
		BasePath:     normalizeBasePath(getEnv("BASE_PATH", "")),

		// OIDC
		OIDCEnabled:      getEnvBool("OIDC_ENABLED", false),
		OIDCIssuerURL:    getEnv("OIDC_ISSUER_URL", ""),
		OIDCClientID:     getEnv("OIDC_CLIENT_ID", ""),
		OIDCClientSecret: getEnv("OIDC_CLIENT_SECRET", ""),
		OIDCRedirectURL:  getEnv("OIDC_REDIRECT_URL", ""),
		OIDCAutoCreate:   getEnvBool("OIDC_AUTO_CREATE_USER", true),
		OIDCAdminGroup:   getEnv("OIDC_ADMIN_GROUP", ""),

		// WebAuthn / Passkeys
		WebAuthnRPID:          getEnv("WEBAUTHN_RP_ID", ""),
		WebAuthnRPDisplayName: getEnv("WEBAUTHN_RP_NAME", "Lurkarr"),

		// Run-once
		RunOnce: getEnvBool("RUN_ONCE", false),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	if v := os.Getenv("DB_MAX_CONNS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 10000 {
			cfg.DBMaxConns = int32(n) // #nosec G109 G115 -- n is bounds-checked above
		}
	}

	origins := getEnv("ALLOWED_ORIGINS", "")
	if origins != "" {
		for _, o := range splitAndTrim(origins) {
			if o != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
			}
		}
	}

	// Parse trusted proxy CIDRs (default: RFC1918 private ranges).
	trustedStr := getEnv("TRUSTED_PROXIES", "")
	nets, err := parseTrustedProxies(trustedStr)
	if err != nil {
		return nil, fmt.Errorf("parse TRUSTED_PROXIES: %w", err)
	}
	cfg.TrustedProxies = nets

	if cfg.ProxyAuth && len(cfg.TrustedProxies) == 0 {
		slog.Warn("PROXY_AUTH enabled without TRUSTED_PROXIES; accepting proxy headers from private IPs only")
	}

	// Parse OIDC scopes.
	scopeStr := getEnv("OIDC_SCOPES", "openid,profile,email")
	for _, s := range splitAndTrim(scopeStr) {
		if s != "" {
			cfg.OIDCScopes = append(cfg.OIDCScopes, s)
		}
	}

	// Validate OIDC configuration.
	if cfg.OIDCEnabled {
		if cfg.OIDCIssuerURL == "" {
			return nil, fmt.Errorf("OIDC_ISSUER_URL is required when OIDC_ENABLED=true")
		}
		if cfg.OIDCClientID == "" {
			return nil, fmt.Errorf("OIDC_CLIENT_ID is required when OIDC_ENABLED=true")
		}
		if cfg.OIDCRedirectURL == "" {
			return nil, fmt.Errorf("OIDC_REDIRECT_URL is required when OIDC_ENABLED=true")
		}
	}

	// Parse WebAuthn origins.
	waOrigins := getEnv("WEBAUTHN_RP_ORIGINS", "")
	if waOrigins != "" {
		for _, o := range splitAndTrim(waOrigins) {
			if o != "" {
				cfg.WebAuthnRPOrigins = append(cfg.WebAuthnRPOrigins, o)
			}
		}
	}

	return cfg, nil
}

func defaultDBMaxConns() int32 {
	cpus := runtime.NumCPU()
	if cpus < 4 {
		cpus = 4
	}
	if cpus > 10000 {
		cpus = 10000
	}
	return int32(cpus) // #nosec G115 -- cpus is clamped to [4,10000]
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func splitAndTrim(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			part := trimSpace(s[start:i])
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	part := trimSpace(s[start:])
	if part != "" {
		result = append(result, part)
	}
	return result
}

func trimSpace(s string) string {
	i, j := 0, len(s)
	for i < j && s[i] == ' ' {
		i++
	}
	for j > i && s[j-1] == ' ' {
		j--
	}
	return s[i:j]
}

// defaultPrivateNets returns RFC1918 + loopback + link-local CIDRs.
var defaultPrivateNets = []string{
	"10.0.0.0/8",
	"172.16.0.0/12",
	"192.168.0.0/16",
	"127.0.0.0/8",
	"::1/128",
	"fc00::/7",
	"fe80::/10",
}

// parseTrustedProxies parses CIDR strings. Empty input returns default private ranges.
func parseTrustedProxies(s string) ([]*net.IPNet, error) {
	cidrs := splitAndTrim(s)
	if len(cidrs) == 0 || (len(cidrs) == 1 && cidrs[0] == "") {
		cidrs = defaultPrivateNets
	}
	var nets []*net.IPNet
	for _, cidr := range cidrs {
		if cidr == "" {
			continue
		}
		// If no slash, treat as single IP.
		if !strings.Contains(cidr, "/") {
			if strings.Contains(cidr, ":") {
				cidr += "/128"
			} else {
				cidr += "/32"
			}
		}
		_, ipNet, err := net.ParseCIDR(cidr)
		if err != nil {
			return nil, fmt.Errorf("invalid CIDR %q: %w", cidr, err)
		}
		nets = append(nets, ipNet)
	}
	return nets, nil
}

// IsTrustedProxy checks if an IP is within one of the trusted proxy CIDRs.
func IsTrustedProxy(trustedNets []*net.IPNet, ipStr string) bool {
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

// normalizeBasePath ensures the base path starts with "/" and does not end with "/".
func normalizeBasePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" || p == "/" {
		return ""
	}
	if !strings.HasPrefix(p, "/") {
		p = "/" + p
	}
	return strings.TrimRight(p, "/")
}
