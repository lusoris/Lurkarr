package config

import (
	"os"
	"testing"
)

func clearEnv() {
	for _, key := range []string{
		"DATABASE_URL", "LISTEN_ADDR", "CSRF_KEY",
		"ALLOWED_ORIGINS", "PROXY_AUTH", "PROXY_HEADER", "LOG_LEVEL",
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
	if cfg.ProxyHeader != "Remote-User" {
		t.Errorf("ProxyHeader = %q, want %q", cfg.ProxyHeader, "Remote-User")
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
	if cfg.ProxyHeader != "X-Auth-User" {
		t.Errorf("ProxyHeader = %q, want %q", cfg.ProxyHeader, "X-Auth-User")
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
