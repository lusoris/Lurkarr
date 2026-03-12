package logging

import (
	"log/slog"
)

// Logger provides app-scoped structured logging via slog.
// Log output goes to stdout via the default slog handler; use Loki for
// aggregation and Grafana for exploration.
type Logger struct{}

// New creates a new Logger.
func New() *Logger {
	return &Logger{}
}

// Log writes a structured log entry via slog.
func (l *Logger) Log(appType, level, message string) {
	lvl := slog.LevelInfo
	switch level {
	case "DEBUG":
		lvl = slog.LevelDebug
	case "WARN":
		lvl = slog.LevelWarn
	case "ERROR":
		lvl = slog.LevelError
	}
	slog.Log(nil, lvl, message, "app_type", appType) //nolint:staticcheck // nil context is fine for slog.Log
}

// ForApp returns an slog.Logger scoped to a specific app type.
func (l *Logger) ForApp(appType string) *slog.Logger {
	return slog.Default().With("app_type", appType)
}

// Close is a no-op (kept for interface compatibility).
func (l *Logger) Close() {}
