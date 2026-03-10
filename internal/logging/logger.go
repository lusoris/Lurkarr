package logging

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/lusoris/lurkarr/internal/database"
)

const (
	ringBufferSize = 10_000
	flushInterval  = 500 * time.Millisecond
	flushBatchSize = 100
)

// Logger wraps slog with async DB writes and WebSocket broadcast.
type Logger struct {
	db     *database.DB
	hub    *Hub
	buffer chan database.LogEntry
	done   chan struct{}
	wg     sync.WaitGroup
}

// New creates a new Logger that writes to DB asynchronously and broadcasts via WebSocket.
func New(db *database.DB, hub *Hub) *Logger {
	l := &Logger{
		db:     db,
		hub:    hub,
		buffer: make(chan database.LogEntry, ringBufferSize),
		done:   make(chan struct{}),
	}
	l.wg.Add(1)
	go l.flusher()
	return l
}

// Log writes a log entry to the ring buffer (non-blocking).
func (l *Logger) Log(appType, level, message string) {
	entry := database.LogEntry{
		AppType:   appType,
		Level:     level,
		Message:   message,
		CreatedAt: time.Now(),
	}
	// Non-blocking send — if buffer is full, drop oldest entry
	select {
	case l.buffer <- entry:
	default:
		select {
		case <-l.buffer:
		default:
		}
		select {
		case l.buffer <- entry:
		default:
		}
	}
	// Broadcast to WebSocket clients immediately
	l.hub.Broadcast(entry)
}

// ForApp returns an slog.Logger scoped to a specific app type.
func (l *Logger) ForApp(appType string) *slog.Logger {
	return slog.New(&appHandler{logger: l, appType: appType})
}

// Close flushes remaining entries and stops the flusher.
func (l *Logger) Close() {
	close(l.done)
	l.wg.Wait()
}

func (l *Logger) flusher() {
	defer l.wg.Done()
	ticker := time.NewTicker(flushInterval)
	defer ticker.Stop()
	var batch []database.LogEntry

	for {
		select {
		case entry := <-l.buffer:
			batch = append(batch, entry)
			if len(batch) >= flushBatchSize {
				l.flush(batch)
				batch = nil
			}
		case <-ticker.C:
			if len(batch) > 0 {
				l.flush(batch)
				batch = nil
			}
		case <-l.done:
			// Drain remaining
			for {
				select {
				case entry := <-l.buffer:
					batch = append(batch, entry)
				default:
					if len(batch) > 0 {
						l.flush(batch)
					}
					return
				}
			}
		}
	}
}

func (l *Logger) flush(entries []database.LogEntry) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := l.db.InsertLogs(ctx, entries); err != nil {
		slog.Error("failed to flush logs to database", "error", err, "count", len(entries))
	}
}

// appHandler is an slog.Handler that routes logs through our Logger.
type appHandler struct {
	logger  *Logger
	appType string
	attrs   []slog.Attr
}

func (h *appHandler) Enabled(_ context.Context, _ slog.Level) bool {
	return true
}

func (h *appHandler) Handle(_ context.Context, r slog.Record) error {
	h.logger.Log(h.appType, r.Level.String(), r.Message)
	return nil
}

func (h *appHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &appHandler{
		logger:  h.logger,
		appType: h.appType,
		attrs:   append(h.attrs, attrs...),
	}
}

func (h *appHandler) WithGroup(_ string) slog.Handler {
	return h
}
