package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all application configuration from environment variables.
type Config struct {
	DatabaseURL    string
	ListenAddr     string
	CSRFKey        string
	AllowedOrigins []string
	ProxyAuth      bool
	ProxyHeader    string
	LogLevel       string
}

// Load reads configuration from environment variables with sensible defaults.
func Load() (*Config, error) {
	cfg := &Config{
		DatabaseURL: getEnv("DATABASE_URL", ""),
		ListenAddr:  getEnv("LISTEN_ADDR", ":8484"),
		CSRFKey:     getEnv("CSRF_KEY", ""),
		ProxyAuth:   getEnvBool("PROXY_AUTH", false),
		ProxyHeader: getEnv("PROXY_HEADER", "Remote-User"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	origins := getEnv("ALLOWED_ORIGINS", "")
	if origins != "" {
		for _, o := range splitAndTrim(origins) {
			if o != "" {
				cfg.AllowedOrigins = append(cfg.AllowedOrigins, o)
			}
		}
	}

	return cfg, nil
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
