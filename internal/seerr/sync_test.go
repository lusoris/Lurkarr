package seerr

import (
	"context"
	"testing"
	"time"
)

type mockSettingsProvider struct {
	settings *Settings
	err      error
}

func (m *mockSettingsProvider) GetSeerrSettings(ctx context.Context) (*Settings, error) {
	return m.settings, m.err
}

func TestSyncEngine_StartStop(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: &Settings{
			Enabled:             false,
			SyncIntervalMinutes: 1,
		},
	}

	engine := NewSyncEngine(provider, nil)
	ctx, cancel := context.WithCancel(context.Background())

	engine.Start(ctx)
	// Give it a moment to enter the loop.
	time.Sleep(50 * time.Millisecond)
	cancel()
	engine.Stop()
}

func TestSyncEngine_SyncDisabled(t *testing.T) {
	provider := &mockSettingsProvider{
		settings: &Settings{
			Enabled: false,
		},
	}

	engine := NewSyncEngine(provider, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	engine.Start(ctx)
	time.Sleep(50 * time.Millisecond)
	engine.Stop()
}
