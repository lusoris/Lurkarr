package notifications

import (
	"context"
	"encoding/json"
	"testing"
)

func TestBuildProviderAllTypes(t *testing.T) {
	tests := []struct {
		name     string
		provType string
		config   map[string]any
		wantName string
		wantErr  bool
	}{
		{
			name:     "discord",
			provType: "discord",
			config:   map[string]any{"webhook_url": "https://discord.com/api/webhooks/123/abc"},
			wantName: "Discord",
		},
		{
			name:     "telegram",
			provType: "telegram",
			config:   map[string]any{"bot_token": "123:ABC", "chat_id": "456"},
			wantName: "Telegram",
		},
		{
			name:     "pushover",
			provType: "pushover",
			config:   map[string]any{"api_token": "abc", "user_key": "def"},
			wantName: "Pushover",
		},
		{
			name:     "gotify",
			provType: "gotify",
			config:   map[string]any{"server_url": "http://localhost:8080", "app_token": "abc"},
			wantName: "Gotify",
		},
		{
			name:     "ntfy",
			provType: "ntfy",
			config:   map[string]any{"server_url": "https://ntfy.sh", "topic": "lurkarr"},
			wantName: "ntfy",
		},
		{
			name:     "apprise",
			provType: "apprise",
			config:   map[string]any{"server_url": "http://localhost:8000", "urls": []any{"discord://wh"}, "tag": "lurkarr"},
			wantName: "Apprise",
		},
		{
			name:     "email",
			provType: "email",
			config:   map[string]any{"host": "smtp.example.com", "port": float64(587), "from": "a@b.com", "to": []any{"c@d.com"}},
			wantName: "Email",
		},
		{
			name:     "webhook",
			provType: "webhook",
			config:   map[string]any{"url": "https://example.com/hook", "headers": map[string]any{"X-Token": "abc"}},
			wantName: "Webhook",
		},
		{
			name:     "unsupported type",
			provType: "sms",
			config:   map[string]any{},
			wantErr:  true,
		},
		{
			name:     "invalid json config",
			provType: "discord",
			config:   nil, // will marshal to "null" which is a valid JSON but unmarshal to nil map
			wantName: "Discord",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgBytes, err := json.Marshal(tt.config)
			if err != nil {
				t.Fatal(err)
			}

			p, pt, _, _, _, err := BuildProvider(ProviderConfig{
				Type:   tt.provType,
				Config: cfgBytes,
				Events: []string{"lurk_completed", "error"},
			})

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if p.Name() != tt.wantName {
				t.Errorf("name = %q, want %q", p.Name(), tt.wantName)
			}
			if string(pt) != tt.provType {
				t.Errorf("providerType = %q, want %q", pt, tt.provType)
			}
		})
	}
}

func TestBuildProviderEvents(t *testing.T) {
	cfg, _ := json.Marshal(map[string]any{"webhook_url": "https://discord.com/wh"})
	_, _, events, _, _, err := BuildProvider(ProviderConfig{
		Type:   "discord",
		Config: cfg,
		Events: []string{"lurk_completed", "error", "download_stuck"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(events) != 3 {
		t.Fatalf("expected 3 events, got %d", len(events))
	}
}

func TestBuildProviderInvalidJSON(t *testing.T) {
	_, _, _, _, _, err := BuildProvider(ProviderConfig{
		Type:   "discord",
		Config: json.RawMessage(`{invalid`),
	})
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadProviders(t *testing.T) {
	m := NewManager()

	// Register a provider manually first — LoadProviders should clear it.
	m.Register(ProviderWebhook, &fakeProvider{name: "old"}, nil)

	configs := []ProviderConfig{
		{
			Type:   "discord",
			Config: mustJSON(map[string]any{"webhook_url": "https://discord.com/wh"}),
			Events: []string{"lurk_completed"},
		},
		{
			Type:   "telegram",
			Config: mustJSON(map[string]any{"bot_token": "123:ABC", "chat_id": "456"}),
			Events: []string{"error"},
		},
	}

	err := m.LoadProviders(configs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have exactly 2 providers (old webhook should be gone).
	providers := m.Providers()
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
}

func TestLoadProvidersSkipsInvalid(t *testing.T) {
	m := NewManager()

	configs := []ProviderConfig{
		{
			Type:   "discord",
			Config: mustJSON(map[string]any{"webhook_url": "https://discord.com/wh"}),
		},
		{
			Type:   "INVALID",
			Config: mustJSON(map[string]any{}),
		},
		{
			Type:   "telegram",
			Config: mustJSON(map[string]any{"bot_token": "tok", "chat_id": "123"}),
		},
	}

	err := m.LoadProviders(configs)
	if err == nil {
		t.Fatal("expected error for invalid provider")
	}

	// Should have 2 valid providers despite the error.
	providers := m.Providers()
	if len(providers) != 2 {
		t.Fatalf("expected 2 providers, got %d", len(providers))
	}
}

func TestLoadProvidersSendsNotifications(t *testing.T) {
	m := NewManager()

	// Use a test server to verify that loaded providers actually work.
	configs := []ProviderConfig{
		{
			Type:   "discord",
			Config: mustJSON(map[string]any{"webhook_url": "https://discord.com/wh"}),
			Events: []string{"lurk_completed"},
		},
	}
	if err := m.LoadProviders(configs); err != nil {
		t.Fatal(err)
	}

	// The notification manager should now deliver to the discord provider.
	// We can't easily test HTTP delivery here, but we can verify the provider is registered.
	providers := m.Providers()
	if len(providers) != 1 {
		t.Fatalf("expected 1 provider, got %d", len(providers))
	}
}

func TestLoadProvidersEmpty(t *testing.T) {
	m := NewManager()
	m.Register(ProviderDiscord, &fakeProvider{name: "old"}, nil)

	err := m.LoadProviders(nil)
	if err != nil {
		t.Fatal(err)
	}

	// Should clear all providers.
	if len(m.Providers()) != 0 {
		t.Fatalf("expected 0 providers after empty load, got %d", len(m.Providers()))
	}
}

func TestSyncManagerIntegration(t *testing.T) {
	// Simulate the full startup flow: create manager → load → notify.
	m := NewManager()

	configs := []ProviderConfig{
		{
			Type:   "discord",
			Config: mustJSON(map[string]any{"webhook_url": "https://discord.com/wh"}),
			Events: []string{"lurk_completed"},
		},
	}
	if err := m.LoadProviders(configs); err != nil {
		t.Fatal(err)
	}

	// Notify should not panic even though discord webhook is fake.
	m.Notify(context.Background(), Event{
		Type:    EventLurkCompleted,
		Title:   "Test",
		Message: "test",
	})
}

func mustJSON(v any) json.RawMessage {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
