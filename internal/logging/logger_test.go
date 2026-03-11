package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// entryCollector captures InsertLogs calls in a concurrency-safe way for verification.
type entryCollector struct {
	mu      sync.Mutex
	entries []database.LogEntry
}

func (c *entryCollector) collect(_ context.Context, entries []database.LogEntry) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.entries = append(c.entries, entries...)
	return nil
}

func (c *entryCollector) get() []database.LogEntry {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]database.LogEntry, len(c.entries))
	copy(cp, c.entries)
	return cp
}

func TestNewHub(t *testing.T) {
	hub := NewHub()
	if hub == nil {
		t.Fatal("NewHub() returned nil")
	}
	if hub.ClientCount() != 0 {
		t.Errorf("ClientCount() = %d, want 0", hub.ClientCount())
	}
}

func TestHubBroadcastNoClients(t *testing.T) {
	hub := NewHub()
	// Should not panic
	hub.Broadcast(database.LogEntry{
		AppType: "sonarr",
		Level:   "INFO",
		Message: "test message",
	})
}

func TestHubClientCount(t *testing.T) {
	hub := NewHub()
	if hub.ClientCount() != 0 {
		t.Errorf("expected 0 clients, got %d", hub.ClientCount())
	}
}

func TestAppHandlerEnabled(t *testing.T) {
	hub := NewHub()
	// Create a Logger that won't write to DB (nil db)
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	h := &appHandler{logger: l, appType: "sonarr"}
	if !h.Enabled(nil, slog.LevelInfo) {
		t.Error("Enabled() = false, want true")
	}
	if !h.Enabled(nil, slog.LevelDebug) {
		t.Error("Enabled(Debug) = false, want true")
	}
}

func TestAppHandlerHandle(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	h := &appHandler{logger: l, appType: "radarr"}
	r := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)
	r.AddAttrs(slog.String("key", "val"))

	if err := h.Handle(nil, r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	// Check that buffer received the entry
	select {
	case entry := <-l.buffer:
		if entry.AppType != "radarr" {
			t.Errorf("AppType = %q", entry.AppType)
		}
		if entry.Level != "INFO" {
			t.Errorf("Level = %q", entry.Level)
		}
	default:
		t.Error("expected entry in buffer")
	}
}

func TestAppHandlerWithAttrs(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	h := &appHandler{logger: l, appType: "sonarr"}
	h2 := h.WithAttrs([]slog.Attr{slog.String("instance", "main")})

	ah, ok := h2.(*appHandler)
	if !ok {
		t.Fatal("WithAttrs did not return *appHandler")
	}
	if len(ah.attrs) != 1 {
		t.Fatalf("expected 1 attr, got %d", len(ah.attrs))
	}
	if ah.attrs[0].Key != "instance" {
		t.Errorf("attr key = %q", ah.attrs[0].Key)
	}
}

func TestAppHandlerWithGroup(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	h := &appHandler{logger: l, appType: "sonarr"}
	h2 := h.WithGroup("group")
	if h2 != h {
		t.Error("WithGroup should return the same handler (no-op)")
	}
}

func TestAppHandlerWithAttrsAndHandle(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	h := &appHandler{logger: l, appType: "sonarr"}
	h2 := h.WithAttrs([]slog.Attr{slog.String("instance", "main")}).(*appHandler)

	r := slog.NewRecord(time.Now(), slog.LevelWarn, "warning msg", 0)
	if err := h2.Handle(nil, r); err != nil {
		t.Fatalf("Handle() error: %v", err)
	}

	select {
	case entry := <-l.buffer:
		if entry.Level != "WARN" {
			t.Errorf("Level = %q", entry.Level)
		}
		// Message should contain the pre-set attr
		if entry.Message == "warning msg" {
			t.Error("expected message to include attrs")
		}
	default:
		t.Error("expected entry in buffer")
	}
}

func TestLoggerLog(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	l.Log("sonarr", "INFO", "test log message")

	select {
	case entry := <-l.buffer:
		if entry.AppType != "sonarr" {
			t.Errorf("AppType = %q", entry.AppType)
		}
		if entry.Message != "test log message" {
			t.Errorf("Message = %q", entry.Message)
		}
	default:
		t.Error("expected entry in buffer")
	}
}

func TestLoggerLogDropsWhenFull(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 1), // tiny buffer
		done:   make(chan struct{}),
	}

	// First should succeed
	l.Log("sonarr", "INFO", "first")
	// Second should be dropped (buffer full, non-blocking)
	l.Log("sonarr", "INFO", "second")

	select {
	case entry := <-l.buffer:
		if entry.Message != "first" {
			t.Errorf("expected first message, got %q", entry.Message)
		}
	default:
		t.Error("expected at least one entry")
	}
}

func TestLoggerForApp(t *testing.T) {
	hub := NewHub()
	l := &Logger{
		hub:    hub,
		buffer: make(chan database.LogEntry, 100),
		done:   make(chan struct{}),
	}

	slogger := l.ForApp("lidarr")
	if slogger == nil {
		t.Fatal("ForApp() returned nil")
	}

	slogger.Info("test from slog")

	select {
	case entry := <-l.buffer:
		if entry.AppType != "lidarr" {
			t.Errorf("AppType = %q, want lidarr", entry.AppType)
		}
	default:
		t.Error("expected entry in buffer")
	}
}

func TestHubBroadcastWithFilters(t *testing.T) {
	hub := NewHub()

	// Add a mock client with filters
	client := &wsClient{
		send:      make(chan []byte, 10),
		appFilter: "sonarr",
		lvlFilter: "INFO",
	}
	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()

	// Matching broadcast
	hub.Broadcast(database.LogEntry{AppType: "sonarr", Level: "INFO", Message: "match"})
	select {
	case <-client.send:
		// OK
	default:
		t.Error("expected matching message to be sent")
	}

	// Non-matching app
	hub.Broadcast(database.LogEntry{AppType: "radarr", Level: "INFO", Message: "no match"})
	select {
	case <-client.send:
		t.Error("expected radarr message to be filtered out")
	default:
		// OK
	}

	// Non-matching level
	hub.Broadcast(database.LogEntry{AppType: "sonarr", Level: "WARN", Message: "no match"})
	select {
	case <-client.send:
		t.Error("expected WARN message to be filtered out")
	default:
		// OK
	}

	// Clean up
	hub.mu.Lock()
	delete(hub.clients, client)
	hub.mu.Unlock()
}

func TestHubBroadcastNoFilters(t *testing.T) {
	hub := NewHub()

	client := &wsClient{
		send: make(chan []byte, 10),
	}
	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()

	hub.Broadcast(database.LogEntry{AppType: "anything", Level: "DEBUG", Message: "should receive"})
	select {
	case <-client.send:
		// OK
	default:
		t.Error("expected unfiltered message to be sent")
	}

	hub.mu.Lock()
	delete(hub.clients, client)
	hub.mu.Unlock()
}

func TestHubBroadcastSlowClient(t *testing.T) {
	hub := NewHub()

	// Client with full send buffer
	client := &wsClient{
		send: make(chan []byte, 0), // zero-capacity, always full
	}
	hub.mu.Lock()
	hub.clients[client] = struct{}{}
	hub.mu.Unlock()

	// Should not block
	hub.Broadcast(database.LogEntry{AppType: "sonarr", Level: "INFO", Message: "dropped"})

	hub.mu.Lock()
	delete(hub.clients, client)
	hub.mu.Unlock()
}

func TestNewLogger(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockLogStore(ctrl)
	store.EXPECT().InsertLogs(gomock.Any(), gomock.Any()).Return(nil).AnyTimes()
	hub := NewHub()
	l := New(store, hub)
	if l == nil {
		t.Fatal("New() returned nil")
	}
	l.Close()
}

func TestLoggerFlushOnClose(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockLogStore(ctrl)
	col := &entryCollector{}
	store.EXPECT().InsertLogs(gomock.Any(), gomock.Any()).DoAndReturn(col.collect).AnyTimes()
	hub := NewHub()
	l := New(store, hub)

	l.Log("sonarr", "INFO", "msg1")
	l.Log("radarr", "WARN", "msg2")
	l.Close()

	entries := col.get()
	if len(entries) != 2 {
		t.Fatalf("expected 2 flushed entries, got %d", len(entries))
	}
	if entries[0].AppType != "sonarr" {
		t.Errorf("entry[0].AppType = %q", entries[0].AppType)
	}
	if entries[1].AppType != "radarr" {
		t.Errorf("entry[1].AppType = %q", entries[1].AppType)
	}
}

func TestLoggerFlushOnBatchSize(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockLogStore(ctrl)
	col := &entryCollector{}
	store.EXPECT().InsertLogs(gomock.Any(), gomock.Any()).DoAndReturn(col.collect).AnyTimes()
	hub := NewHub()
	l := New(store, hub)
	defer l.Close()

	// Send flushBatchSize entries to trigger batch flush
	for i := range flushBatchSize {
		l.Log("sonarr", "INFO", "batch msg "+string(rune('0'+i%10)))
	}

	// Wait for flush to complete
	time.Sleep(200 * time.Millisecond)

	entries := col.get()
	if len(entries) < flushBatchSize {
		t.Errorf("expected at least %d flushed entries, got %d", flushBatchSize, len(entries))
	}
}

func TestLoggerFlushOnTicker(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockLogStore(ctrl)
	col := &entryCollector{}
	store.EXPECT().InsertLogs(gomock.Any(), gomock.Any()).DoAndReturn(col.collect).AnyTimes()
	hub := NewHub()
	l := New(store, hub)
	defer l.Close()

	l.Log("sonarr", "INFO", "ticker flush test")

	// Wait longer than flushInterval (500ms)
	time.Sleep(700 * time.Millisecond)

	entries := col.get()
	if len(entries) != 1 {
		t.Fatalf("expected 1 flushed entry after ticker, got %d", len(entries))
	}
	if entries[0].Message != "ticker flush test" {
		t.Errorf("unexpected message: %q", entries[0].Message)
	}
}

func TestLoggerFlushDBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockLogStore(ctrl)
	store.EXPECT().InsertLogs(gomock.Any(), gomock.Any()).Return(context.DeadlineExceeded).AnyTimes()
	hub := NewHub()
	l := New(store, hub)

	l.Log("sonarr", "INFO", "will fail to flush")
	l.Close() // Should not panic even when InsertLogs returns error
}

func TestHandleWebSocket(t *testing.T) {
	hub := NewHub()
	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + srv.URL[4:] // http -> ws
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer conn.CloseNow()

	// Give the server goroutine time to register the client
	time.Sleep(50 * time.Millisecond)

	if hub.ClientCount() != 1 {
		t.Fatalf("expected 1 client, got %d", hub.ClientCount())
	}

	// Broadcast a message
	hub.Broadcast(database.LogEntry{AppType: "sonarr", Level: "INFO", Message: "ws test"})

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	var entry database.LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Message != "ws test" {
		t.Errorf("message = %q, want %q", entry.Message, "ws test")
	}

	conn.Close(websocket.StatusNormalClosure, "done")
}

func TestHandleWebSocketWithFilters(t *testing.T) {
	hub := NewHub()
	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	wsURL := "ws" + srv.URL[4:] + "?app=sonarr&level=INFO"
	conn, _, err := websocket.Dial(ctx, wsURL, nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer conn.CloseNow()

	time.Sleep(50 * time.Millisecond)

	// Should receive matching message
	hub.Broadcast(database.LogEntry{AppType: "sonarr", Level: "INFO", Message: "match"})
	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var entry database.LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Message != "match" {
		t.Errorf("message = %q, want %q", entry.Message, "match")
	}

	conn.Close(websocket.StatusNormalClosure, "done")
}

func TestHandleWebSocketFilterUpdate(t *testing.T) {
	hub := NewHub()
	srv := httptest.NewServer(http.HandlerFunc(hub.HandleWebSocket))
	defer srv.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, _, err := websocket.Dial(ctx, "ws"+srv.URL[4:], nil)
	if err != nil {
		t.Fatalf("websocket dial: %v", err)
	}
	defer conn.CloseNow()

	time.Sleep(50 * time.Millisecond)

	// Send a filter update message
	filterMsg := `{"app":"radarr","level":"WARN"}`
	err = conn.Write(ctx, websocket.MessageText, []byte(filterMsg))
	if err != nil {
		t.Fatalf("write filter: %v", err)
	}

	// Give server time to process filter
	time.Sleep(50 * time.Millisecond)

	// Broadcast matching entry
	hub.Broadcast(database.LogEntry{AppType: "radarr", Level: "WARN", Message: "filtered"})

	_, data, err := conn.Read(ctx)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	var entry database.LogEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if entry.Message != "filtered" {
		t.Errorf("message = %q, want %q", entry.Message, "filtered")
	}

	conn.Close(websocket.StatusNormalClosure, "done")
}
