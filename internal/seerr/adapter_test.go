package seerr

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDBSettingsFunc_Success(t *testing.T) {
	fn := DBSettingsFunc(func(ctx context.Context) (string, string, bool, int, bool, error) {
		return "http://seerr:5055", "apikey123", true, 15, false, nil
	})

	s, err := fn.GetSeerrSettings(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.URL != "http://seerr:5055" {
		t.Errorf("URL = %q", s.URL)
	}
	if s.APIKey != "apikey123" {
		t.Errorf("APIKey = %q", s.APIKey)
	}
	if !s.Enabled {
		t.Error("expected Enabled = true")
	}
	if s.SyncIntervalMinutes != 15 {
		t.Errorf("SyncIntervalMinutes = %d, want 15", s.SyncIntervalMinutes)
	}
	if s.AutoApprove {
		t.Error("expected AutoApprove = false")
	}
}

func TestDBSettingsFunc_Error(t *testing.T) {
	fn := DBSettingsFunc(func(ctx context.Context) (string, string, bool, int, bool, error) {
		return "", "", false, 0, false, errors.New("db error")
	})

	_, err := fn.GetSeerrSettings(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:5055", "key", 30*time.Second)
	if c == nil {
		t.Fatal("NewClient returned nil")
	}
}

func TestSyncEngine_SyncWithServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/request/count" {
			json.NewEncoder(w).Encode(RequestCount{
				Total:   50,
				Pending: 3,
			})
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	provider := &mockSettingsProvider{
		settings: &Settings{
			URL:                 srv.URL,
			APIKey:              "test-key",
			Enabled:             true,
			SyncIntervalMinutes: 1,
		},
	}

	engine := NewSyncEngine(provider, nil)
	// Call sync directly to avoid the 15-second startup delay.
	engine.sync(context.Background(), provider.settings)
	// If we get here without panic/error, sync worked.
}

func TestSyncEngine_SyncAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	provider := &mockSettingsProvider{
		settings: &Settings{
			URL:    srv.URL,
			APIKey: "bad-key",
		},
	}

	engine := NewSyncEngine(provider, nil)
	// Should log error but not panic.
	engine.sync(context.Background(), provider.settings)
}

func TestSyncEngine_StopWithoutStart(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: &Settings{Enabled: false},
	}
	engine := NewSyncEngine(provider, nil)
	// Stop without Start should not panic.
	engine.Stop()
}

func TestSyncEngine_SettingsError(t *testing.T) {
	provider := &mockSettingsProvider{
		err: errors.New("db unavailable"),
	}
	engine := NewSyncEngine(provider, nil)
	ctx, cancel := context.WithCancel(context.Background())

	engine.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	cancel()
	engine.Stop()
}
