package config

import (
	"os"
	"testing"
)

func clearEnv() {
	for _, key := range []string{
		"DATABASE_URL", "LISTEN_ADDR", "CSRF_KEY",
		"ALLOWED_ORIGINS", "PROXY_AUTH", "PROXY_HEADER", "LOG_LEVEL",
		"TRUSTED_PROXIES", "BASE_PATH", "SECURE_COOKIE",
		"OIDC_ENABLED", "OIDC_ISSUER_URL", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET",
		"OIDC_REDIRECT_URL", "OIDC_SCOPES", "OIDC_AUTO_CREATE_USER", "OIDC_ADMIN_GROUP",
	} {
		os.Unsetenv(key)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	clearEnv()
	_, err := Load()
	if err == nil {
		t.Fatal("Load() should fail when DATABASE_URL is not set")
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv()
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.ListenAddr != ":8484" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":8484")
	}
	if cfg.ProxyAuth != false {
		t.Error("ProxyAuth should default to false")
	}
	if len(cfg.ProxyHeaders) != 1 || cfg.ProxyHeaders[0] != "Remote-User" {
		t.Errorf("ProxyHeaders = %v, want [Remote-User]", cfg.ProxyHeaders)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}

func TestLoadCustomValues(t *testing.T) {
	clearEnv()
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	os.Setenv("LISTEN_ADDR", ":9000")
	os.Setenv("CSRF_KEY", "my-csrf-key")
	os.Setenv("PROXY_AUTH", "true")
	os.Setenv("PROXY_HEADER", "X-Auth-User")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("ALLOWED_ORIGINS", "http://localhost:3000, http://example.com")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.ListenAddr != ":9000" {
		t.Errorf("ListenAddr = %q, want %q", cfg.ListenAddr, ":9000")
	}
	if cfg.CSRFKey != "my-csrf-key" {
		t.Errorf("CSRFKey = %q, want %q", cfg.CSRFKey, "my-csrf-key")
	}
	if !cfg.ProxyAuth {
		t.Error("ProxyAuth should be true")
	}
	if len(cfg.ProxyHeaders) != 1 || cfg.ProxyHeaders[0] != "X-Auth-User" {
		t.Errorf("ProxyHeaders = %v, want [X-Auth-User]", cfg.ProxyHeaders)
	}
	if len(cfg.AllowedOrigins) != 2 {
		t.Fatalf("AllowedOrigins len = %d, want 2", len(cfg.AllowedOrigins))
	}
	if cfg.AllowedOrigins[0] != "http://localhost:3000" {
		t.Errorf("AllowedOrigins[0] = %q", cfg.AllowedOrigins[0])
	}
	if cfg.AllowedOrigins[1] != "http://example.com" {
		t.Errorf("AllowedOrigins[1] = %q", cfg.AllowedOrigins[1])
	}
}

func TestGetEnv(t *testing.T) {
	os.Unsetenv("TEST_GET_ENV_KEY")
	if v := getEnv("TEST_GET_ENV_KEY", "fallback"); v != "fallback" {
		t.Errorf("getEnv returned %q, want %q", v, "fallback")
	}
	os.Setenv("TEST_GET_ENV_KEY", "actual")
	defer os.Unsetenv("TEST_GET_ENV_KEY")
	if v := getEnv("TEST_GET_ENV_KEY", "fallback"); v != "actual" {
		t.Errorf("getEnv returned %q, want %q", v, "actual")
	}
}

func TestGetEnvBool(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		set      bool
		fallback bool
		want     bool
	}{
		{"unset returns fallback true", "", false, true, true},
		{"unset returns fallback false", "", false, false, false},
		{"true string", "true", true, false, true},
		{"false string", "false", true, true, false},
		{"1 string", "1", true, false, true},
		{"0 string", "0", true, true, false},
		{"invalid returns fallback", "notabool", true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Unsetenv("TEST_BOOL_KEY")
			if tt.set {
				os.Setenv("TEST_BOOL_KEY", tt.value)
			}
			got := getEnvBool("TEST_BOOL_KEY", tt.fallback)
			if got != tt.want {
				t.Errorf("getEnvBool(%q, %v) = %v, want %v", tt.value, tt.fallback, got, tt.want)
			}
		})
	}
	os.Unsetenv("TEST_BOOL_KEY")
}

func TestSplitAndTrim(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"a,b,c", []string{"a", "b", "c"}},
		{" a , b , c ", []string{"a", "b", "c"}},
		{"single", []string{"single"}},
		{"", nil},
		{" , , ", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := splitAndTrim(tt.input)
			if len(got) != len(tt.want) {
				t.Fatalf("splitAndTrim(%q) = %v (len %d), want %v (len %d)", tt.input, got, len(got), tt.want, len(tt.want))
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitAndTrim(%q)[%d] = %q, want %q", tt.input, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestTrimSpace(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "hello"},
		{" hello", "hello"},
		{"hello ", "hello"},
		{" hello ", "hello"},
		{"  spaced  ", "spaced"},
		{"", ""},
		{"   ", ""},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := trimSpace(tt.input)
			if got != tt.want {
				t.Errorf("trimSpace(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizeBasePath(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"/", ""},
		{"/lurkarr", "/lurkarr"},
		{"/lurkarr/", "/lurkarr"},
		{"lurkarr", "/lurkarr"},
		{"lurkarr/", "/lurkarr"},
		{"/app/lurkarr/", "/app/lurkarr"},
		{" /lurkarr ", "/lurkarr"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeBasePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizeBasePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseTrustedProxies(t *testing.T) {
	// Default (empty) returns private ranges.
	nets, err := parseTrustedProxies("")
	if err != nil {
		t.Fatalf("parseTrustedProxies(\"\") error: %v", err)
	}
	if len(nets) != len(defaultPrivateNets) {
		t.Errorf("expected %d default networks, got %d", len(defaultPrivateNets), len(nets))
	}

	// Custom CIDR.
	nets, err = parseTrustedProxies("192.168.1.0/24, 10.0.0.0/8")
	if err != nil {
		t.Fatalf("parseTrustedProxies error: %v", err)
	}
	if len(nets) != 2 {
		t.Fatalf("expected 2 networks, got %d", len(nets))
	}

	// Single IP (no mask).
	nets, err = parseTrustedProxies("1.2.3.4")
	if err != nil {
		t.Fatalf("parseTrustedProxies single IP error: %v", err)
	}
	if len(nets) != 1 {
		t.Fatalf("expected 1 network, got %d", len(nets))
	}

	// Invalid CIDR.
	_, err = parseTrustedProxies("not-a-cidr")
	if err == nil {
		t.Fatal("expected error for invalid CIDR")
	}
}

func TestIsTrustedProxy(t *testing.T) {
	nets, _ := parseTrustedProxies("10.0.0.0/8, 172.16.0.0/12")

	tests := []struct {
		ip   string
		want bool
	}{
		{"10.0.0.1", true},
		{"10.255.255.255", true},
		{"172.16.0.1", true},
		{"172.31.255.255", true},
		{"192.168.1.1", false},
		{"8.8.8.8", false},
		{"invalid", false},
	}
	for _, tt := range tests {
		t.Run(tt.ip, func(t *testing.T) {
			got := IsTrustedProxy(nets, tt.ip)
			if got != tt.want {
				t.Errorf("IsTrustedProxy(%q) = %v, want %v", tt.ip, got, tt.want)
			}
		})
	}
}

func TestLoadOIDCValidation(t *testing.T) {
	clearEnv()
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	os.Setenv("OIDC_ENABLED", "true")
	defer clearEnv()

	// Missing issuer URL.
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when OIDC_ENABLED=true but OIDC_ISSUER_URL missing")
	}

	os.Setenv("OIDC_ISSUER_URL", "https://auth.example.com")
	_, err = Load()
	if err == nil {
		t.Fatal("expected error when OIDC_CLIENT_ID missing")
	}

	os.Setenv("OIDC_CLIENT_ID", "lurkarr")
	_, err = Load()
	if err == nil {
		t.Fatal("expected error when OIDC_REDIRECT_URL missing")
	}

	os.Setenv("OIDC_REDIRECT_URL", "http://localhost:8484/api/auth/oidc/callback")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if !cfg.OIDCEnabled {
		t.Error("OIDCEnabled should be true")
	}
	if cfg.OIDCIssuerURL != "https://auth.example.com" {
		t.Errorf("OIDCIssuerURL = %q", cfg.OIDCIssuerURL)
	}
}

func TestLoadBasePath(t *testing.T) {
	clearEnv()
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	os.Setenv("BASE_PATH", "/lurkarr/")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if cfg.BasePath != "/lurkarr" {
		t.Errorf("BasePath = %q, want %q", cfg.BasePath, "/lurkarr")
	}
}

func TestLoadDefaultOIDCScopes(t *testing.T) {
	clearEnv()
	os.Setenv("DATABASE_URL", "postgres://test:test@localhost/test")
	defer clearEnv()

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if len(cfg.OIDCScopes) != 3 {
		t.Fatalf("expected 3 default scopes, got %d: %v", len(cfg.OIDCScopes), cfg.OIDCScopes)
	}
	expected := []string{"openid", "profile", "email"}
	for i, s := range expected {
		if cfg.OIDCScopes[i] != s {
			t.Errorf("OIDCScopes[%d] = %q, want %q", i, cfg.OIDCScopes[i], s)
		}
	}
}
