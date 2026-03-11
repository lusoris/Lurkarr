package notifications

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
)

// ─── Manager tests ───────────────────────────────────────────────────────────

type fakeProvider struct {
	name string
	sent []Event
	mu   sync.Mutex
	err  error
}

func (f *fakeProvider) Name() string { return f.name }
func (f *fakeProvider) Send(_ context.Context, e Event) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.sent = append(f.sent, e)
	return f.err
}
func (f *fakeProvider) Test(ctx context.Context) error {
	return f.Send(ctx, Event{Type: EventTestNotification, Title: "test"})
}

func TestManagerNotify(t *testing.T) {
	m := NewManager()
	p := &fakeProvider{name: "test"}
	m.Register(ProviderDiscord, p, nil)

	m.Notify(context.Background(), Event{
		Type:    EventHuntCompleted,
		Title:   "Hunt Done",
		Message: "Found 5 missing items",
	})

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.sent) != 1 {
		t.Fatalf("expected 1 sent, got %d", len(p.sent))
	}
	if p.sent[0].Title != "Hunt Done" {
		t.Errorf("expected title 'Hunt Done', got %q", p.sent[0].Title)
	}
}

func TestManagerEventFiltering(t *testing.T) {
	m := NewManager()
	p := &fakeProvider{name: "filtered"}
	m.Register(ProviderTelegram, p, []EventType{EventError})

	// This should NOT reach the provider.
	m.Notify(context.Background(), Event{Type: EventHuntCompleted, Title: "ignored"})

	// This should reach the provider.
	m.Notify(context.Background(), Event{Type: EventError, Title: "alert"})

	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.sent) != 1 {
		t.Fatalf("expected 1 (only error), got %d", len(p.sent))
	}
	if p.sent[0].Title != "alert" {
		t.Errorf("expected 'alert', got %q", p.sent[0].Title)
	}
}

func TestManagerTestProvider(t *testing.T) {
	m := NewManager()
	p := &fakeProvider{name: "testable"}
	m.Register(ProviderWebhook, p, nil)

	if err := m.TestProvider(context.Background(), ProviderWebhook); err != nil {
		t.Fatal(err)
	}

	// Unregistered provider should error.
	if err := m.TestProvider(context.Background(), ProviderDiscord); err == nil {
		t.Error("expected error for unregistered provider")
	}
}

func TestManagerUnregister(t *testing.T) {
	m := NewManager()
	p := &fakeProvider{name: "removable"}
	m.Register(ProviderNtfy, p, nil)
	m.Unregister(ProviderNtfy)

	if len(m.Providers()) != 0 {
		t.Error("expected 0 providers after unregister")
	}
}

// ─── Discord tests ───────────────────────────────────────────────────────────

func TestDiscordSend(t *testing.T) {
	var received map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	d := NewDiscord(ts.URL, "Lurkarr", "")
	err := d.Send(context.Background(), Event{
		Type:     EventHuntCompleted,
		Title:    "Hunt Complete",
		Message:  "Found 3 items",
		AppType:  "sonarr",
		Instance: "main",
		Fields:   map[string]string{"Missing": "3"},
	})
	if err != nil {
		t.Fatal(err)
	}

	embeds, ok := received["embeds"].([]any)
	if !ok || len(embeds) == 0 {
		t.Fatal("expected embeds in payload")
	}
	embed := embeds[0].(map[string]any)
	if embed["title"] != "Hunt Complete" {
		t.Errorf("expected title 'Hunt Complete', got %v", embed["title"])
	}
}

func TestDiscordErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	d := NewDiscord(ts.URL, "", "")
	err := d.Send(context.Background(), Event{Type: EventTestNotification, Title: "test"})
	if err == nil {
		t.Error("expected error for bad status")
	}
}

func TestDiscordTest(t *testing.T) {
	var called atomic.Bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer ts.Close()

	d := NewDiscord(ts.URL, "", "")
	if err := d.Test(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !called.Load() {
		t.Error("webhook was not called")
	}
}

// ─── Telegram tests ──────────────────────────────────────────────────────────

func TestTelegramSend(t *testing.T) {
	var received string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		received = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	tg := &Telegram{BotToken: "fake", ChatID: "123", client: http.DefaultClient}
	// Override the API URL by using the test server directly.
	tg.client = ts.Client()

	// For testing, we need to intercept the request. Use the test server URL.
	// We'll test the message formatting instead.
	msg := formatTelegramMessage(Event{
		Type:     EventHuntCompleted,
		Title:    "Hunt Done",
		Message:  "5 items found",
		AppType:  "radarr",
		Instance: "main",
	})

	if msg == "" {
		t.Error("expected non-empty message")
	}
	_ = received
}

func TestTelegramEndpoint(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// We can't easily override Telegram's API URL without modifying the struct,
	// so we'll test the error case with a bad URL.
	tg := NewTelegram("faketoken", "123")
	tg.client = ts.Client()
	// The actual API call goes to api.telegram.org, but the client is the test server's.
	// This test verifies the request construction works.

	// Test with a broken context to verify error handling.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := tg.Send(ctx, Event{Type: EventTestNotification, Title: "test"})
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// ─── Pushover tests ──────────────────────────────────────────────────────────

func TestPushoverSend(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	// Replace the Pushover API URL for testing. We'll test by using a short-circuit.
	p := NewPushover("token", "user", "device", 0)
	p.client = ts.Client()
	// Pushover sends to pushover.net, not our test server, so test with cancelled context.
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	err := p.Send(ctx, Event{Type: EventError, Title: "Error", Message: "something broke"})
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// ─── Gotify tests ────────────────────────────────────────────────────────────

func TestGotifySend(t *testing.T) {
	var received map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Gotify-Key") != "testtoken" {
			t.Error("missing or wrong Gotify token")
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	g := NewGotify(ts.URL, "testtoken", 5)
	err := g.Send(context.Background(), Event{
		Type:    EventHuntCompleted,
		Title:   "Hunt Done",
		Message: "Found items",
	})
	if err != nil {
		t.Fatal(err)
	}
	if received["title"] != "Hunt Done" {
		t.Errorf("expected title 'Hunt Done', got %v", received["title"])
	}
}

func TestGotifyErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	g := NewGotify(ts.URL, "token", 5)
	err := g.Send(context.Background(), Event{Type: EventTestNotification, Title: "test"})
	if err == nil {
		t.Error("expected error for 500 status")
	}
}

// ─── Ntfy tests ──────────────────────────────────────────────────────────────

func TestNtfySend(t *testing.T) {
	var title, tags string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		title = r.Header.Get("Title")
		tags = r.Header.Get("Tags")
		if r.Header.Get("Authorization") != "Bearer testtoken" {
			t.Error("missing auth")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := NewNtfy(ts.URL, "lurkarr", "testtoken", 3)
	err := n.Send(context.Background(), Event{
		Type:    EventQueueItemRemoved,
		Title:   "Queue Cleaned",
		Message: "Removed stalled item",
	})
	if err != nil {
		t.Fatal(err)
	}
	if title != "Queue Cleaned" {
		t.Errorf("expected title 'Queue Cleaned', got %q", title)
	}
	if tags != "wastebasket" {
		t.Errorf("expected tags 'wastebasket', got %q", tags)
	}
}

func TestNtfyErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer ts.Close()

	n := NewNtfy(ts.URL, "test", "", 3)
	err := n.Send(context.Background(), Event{Type: EventTestNotification, Title: "test"})
	if err == nil {
		t.Error("expected error for 403 status")
	}
}

// ─── Apprise tests ───────────────────────────────────────────────────────────

func TestAppriseSend(t *testing.T) {
	var received map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	a := NewApprise(ts.URL, []string{"discord://webhook", "slack://token"}, "lurkarr")
	err := a.Send(context.Background(), Event{
		Type:    EventError,
		Title:   "Error Alert",
		Message: "Something went wrong",
	})
	if err != nil {
		t.Fatal(err)
	}
	if received["type"] != "failure" {
		t.Errorf("expected type 'failure', got %v", received["type"])
	}
	if received["tag"] != "lurkarr" {
		t.Errorf("expected tag 'lurkarr', got %v", received["tag"])
	}
}

func TestAppriseErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer ts.Close()

	a := NewApprise(ts.URL, []string{"test://url"}, "")
	err := a.Send(context.Background(), Event{Type: EventTestNotification, Title: "test"})
	if err == nil {
		t.Error("expected error for 502 status")
	}
}

// ─── Webhook tests ───────────────────────────────────────────────────────────

func TestWebhookSend(t *testing.T) {
	var received WebhookPayload
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("expected JSON content type")
		}
		if r.Header.Get("X-Custom") != "value" {
			t.Error("missing custom header")
		}
		body, _ := io.ReadAll(r.Body)
		json.Unmarshal(body, &received)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := NewWebhook(ts.URL, map[string]string{"X-Custom": "value"})
	err := wh.Send(context.Background(), Event{
		Type:     EventSchedulerAction,
		Title:    "Schedule Ran",
		Message:  "Disabled sonarr",
		AppType:  "sonarr",
		Instance: "main",
		Fields:   map[string]string{"action": "disable"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if received.Event != "scheduler_action" {
		t.Errorf("expected event 'scheduler_action', got %q", received.Event)
	}
	if received.Title != "Schedule Ran" {
		t.Errorf("expected title 'Schedule Ran', got %q", received.Title)
	}
}

func TestWebhookErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer ts.Close()

	wh := NewWebhook(ts.URL, nil)
	err := wh.Send(context.Background(), Event{Type: EventTestNotification, Title: "test"})
	if err == nil {
		t.Error("expected error for 503 status")
	}
}

func TestWebhookTest(t *testing.T) {
	var called atomic.Bool
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called.Store(true)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := NewWebhook(ts.URL, nil)
	if err := wh.Test(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !called.Load() {
		t.Error("webhook test was not called")
	}
}

// ─── Format tests ────────────────────────────────────────────────────────────

func TestFormatPlainMessage(t *testing.T) {
	msg := formatPlainMessage(Event{
		Type:     EventHuntCompleted,
		Title:    "Hunt Complete",
		Message:  "Found 5 items",
		AppType:  "sonarr",
		Instance: "main",
		Fields:   map[string]string{"Missing": "3", "Upgrades": "2"},
	})

	if msg == "" {
		t.Error("expected non-empty message")
	}
}

func TestFormatPlainMessageMinimal(t *testing.T) {
	msg := formatPlainMessage(Event{
		Type:    EventTestNotification,
		Title:   "Test",
		Message: "Hello",
	})
	if msg != "Hello" {
		t.Errorf("expected 'Hello', got %q", msg)
	}
}
