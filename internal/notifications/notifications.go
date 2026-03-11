// Package notifications provides a unified interface for sending notifications
// through multiple providers (Discord, Telegram, Pushover, etc.).
package notifications

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// EventType enumerates the types of events that can trigger notifications.
type EventType string

const (
	EventLurkCompleted    EventType = "lurk_completed"
	EventQueueItemRemoved EventType = "queue_item_removed"
	EventDownloadStuck    EventType = "download_stuck"
	EventSchedulerAction  EventType = "scheduler_action"
	EventError            EventType = "error"
	EventTestNotification EventType = "test"
)

// Event represents a notification event.
type Event struct {
	Type     EventType
	Title    string
	Message  string
	AppType  string // e.g. "sonarr", "radarr"
	Instance string // instance name
	Fields   map[string]string
}

// Provider is the interface that all notification providers must implement.
type Provider interface {
	// Name returns the provider's display name (e.g. "Discord").
	Name() string

	// Send delivers a notification event. Must be safe for concurrent use.
	Send(ctx context.Context, event Event) error

	// Test sends a test notification to verify the provider is configured correctly.
	Test(ctx context.Context) error
}

// ProviderType identifies a notification provider.
type ProviderType string

const (
	ProviderDiscord  ProviderType = "discord"
	ProviderTelegram ProviderType = "telegram"
	ProviderPushover ProviderType = "pushover"
	ProviderGotify   ProviderType = "gotify"
	ProviderNtfy     ProviderType = "ntfy"
	ProviderApprise  ProviderType = "apprise"
	ProviderEmail    ProviderType = "email"
	ProviderWebhook  ProviderType = "webhook"
)

// Manager coordinates sending events to all enabled notification providers.
// It also implements the Notifier interface.
type Manager struct {
	mu        sync.RWMutex
	providers map[ProviderType]providerEntry
}

// Notifier is a minimal interface for sending notification events.
// Used by other packages to avoid a hard dependency on the full Manager.
type Notifier interface {
	Notify(ctx context.Context, event Event)
}

type providerEntry struct {
	provider Provider
	events   map[EventType]bool // which events this provider subscribes to
}

// NewManager creates a notification manager.
func NewManager() *Manager {
	return &Manager{
		providers: make(map[ProviderType]providerEntry),
	}
}

// Register adds a provider that will receive the specified event types.
// If events is nil, the provider receives all events.
func (m *Manager) Register(pt ProviderType, p Provider, events []EventType) {
	m.mu.Lock()
	defer m.mu.Unlock()

	evSet := make(map[EventType]bool)
	if events == nil {
		// Subscribe to all events.
		for _, e := range []EventType{EventLurkCompleted, EventQueueItemRemoved, EventDownloadStuck, EventSchedulerAction, EventError} {
			evSet[e] = true
		}
	} else {
		for _, e := range events {
			evSet[e] = true
		}
	}
	// Test event is always allowed.
	evSet[EventTestNotification] = true

	m.providers[pt] = providerEntry{provider: p, events: evSet}
}

// Unregister removes a provider.
func (m *Manager) Unregister(pt ProviderType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.providers, pt)
}

// Notify sends an event to all providers subscribed to that event type.
// Errors are logged but do not halt delivery to other providers.
func (m *Manager) Notify(ctx context.Context, event Event) {
	m.mu.RLock()
	entries := make([]providerEntry, 0, len(m.providers))
	for _, e := range m.providers {
		if e.events[event.Type] {
			entries = append(entries, e)
		}
	}
	m.mu.RUnlock()

	var wg sync.WaitGroup
	for _, entry := range entries {
		wg.Add(1)
		go func(e providerEntry) {
			defer wg.Done()
			if err := e.provider.Send(ctx, event); err != nil {
				slog.Error("notification delivery failed",
					"provider", e.provider.Name(),
					"event", event.Type,
					"error", err,
				)
			}
		}(entry)
	}
	wg.Wait()
}

// TestProvider sends a test notification to a specific provider.
func (m *Manager) TestProvider(ctx context.Context, pt ProviderType) error {
	m.mu.RLock()
	entry, ok := m.providers[pt]
	m.mu.RUnlock()
	if !ok {
		return fmt.Errorf("provider %q not registered", pt)
	}
	return entry.provider.Test(ctx)
}

// Providers returns the list of registered provider types.
func (m *Manager) Providers() []ProviderType {
	m.mu.RLock()
	defer m.mu.RUnlock()
	types := make([]ProviderType, 0, len(m.providers))
	for pt := range m.providers {
		types = append(types, pt)
	}
	return types
}
