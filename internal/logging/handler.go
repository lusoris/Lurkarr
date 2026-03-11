package logging

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/coder/websocket"
	"github.com/lusoris/lurkarr/internal/database"
)

// Hub manages WebSocket clients for log streaming.
type Hub struct {
	mu             sync.RWMutex
	clients        map[*wsClient]struct{}
	OriginPatterns []string
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte

	mu        sync.RWMutex
	appFilter string
	lvlFilter string
}

// NewHub creates a new WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		clients: make(map[*wsClient]struct{}),
	}
}

// Broadcast sends a log entry to all connected clients (with filtering).
func (h *Hub) Broadcast(entry database.LogEntry) {
	data, err := json.Marshal(entry)
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for c := range h.clients {
		c.mu.RLock()
		appF, lvlF := c.appFilter, c.lvlFilter
		c.mu.RUnlock()

		if appF != "" && appF != entry.AppType {
			continue
		}
		if lvlF != "" && lvlF != entry.Level {
			continue
		}
		// Non-blocking send with backpressure
		select {
		case c.send <- data:
		default:
			// Client too slow — skip this message
		}
	}
}

// HandleWebSocket upgrades HTTP to WebSocket and registers the client.
func (h *Hub) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	patterns := h.OriginPatterns
	if len(patterns) == 0 {
		patterns = []string{"*"}
	}
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		OriginPatterns: patterns,
	})
	if err != nil {
		slog.Error("websocket accept failed", "error", err)
		return
	}

	client := &wsClient{
		conn:      conn,
		send:      make(chan []byte, 256),
		appFilter: r.URL.Query().Get("app"),
		lvlFilter: r.URL.Query().Get("level"),
	}

	h.mu.Lock()
	h.clients[client] = struct{}{}
	h.mu.Unlock()

	ctx := r.Context()

	// Writer goroutine with ping/pong keepalive
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.clients, client)
			h.mu.Unlock()
			_ = conn.Close(websocket.StatusNormalClosure, "")
		}()

		pingTicker := time.NewTicker(30 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case msg, ok := <-client.send:
				if !ok {
					return
				}
				writeCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Write(writeCtx, websocket.MessageText, msg)
				cancel()
				if err != nil {
					return
				}
			case <-pingTicker.C:
				pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				err := conn.Ping(pingCtx)
				cancel()
				if err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Reader goroutine — handles filter updates from client
	for {
		_, msg, err := conn.Read(ctx)
		if err != nil {
			break
		}
		// Client can send filter updates: {"app":"sonarr","level":"INFO"}
		var filter struct {
			App   string `json:"app"`
			Level string `json:"level"`
		}
		if json.Unmarshal(msg, &filter) == nil {
			client.mu.Lock()
			client.appFilter = filter.App
			client.lvlFilter = filter.Level
			client.mu.Unlock()
		}
	}

	close(client.send)
}

// ClientCount returns the number of connected WebSocket clients.
func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
