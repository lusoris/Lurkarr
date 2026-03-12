package logging

import (
	"testing"
)

func TestNew(t *testing.T) {
	l := New()
	if l == nil {
		t.Fatal("New() returned nil")
	}
}

func TestLogLevels(t *testing.T) {
	l := New()
	l.Log("sonarr", "DEBUG", "debug msg")
	l.Log("sonarr", "INFO", "info msg")
	l.Log("sonarr", "WARN", "warn msg")
	l.Log("sonarr", "ERROR", "error msg")
	l.Log("sonarr", "UNKNOWN", "defaults to info")
}

func TestForApp(t *testing.T) {
	l := New()
	slogger := l.ForApp("radarr")
	if slogger == nil {
		t.Fatal("ForApp() returned nil")
	}
	slogger.Info("hello from radarr")
}

func TestClose(t *testing.T) {
	l := New()
	l.Close()
}
