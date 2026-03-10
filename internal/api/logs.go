package api

import (
	"net/http"
	"strconv"

	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/logging"
)

// LogsHandler handles log endpoints.
type LogsHandler struct {
	DB  *database.DB
	Hub *logging.Hub
}

// HandleGetLogs handles GET /api/logs.
func (h *LogsHandler) HandleGetLogs(w http.ResponseWriter, r *http.Request) {
	q := database.LogQuery{
		AppType: r.URL.Query().Get("app"),
		Level:   r.URL.Query().Get("level"),
		Limit:   100,
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 500 {
			q.Limit = n
		}
	}
	if v := r.URL.Query().Get("before"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			q.Before = n
		}
	}

	entries, err := h.DB.QueryLogs(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load logs"))
		return
	}

	writeJSON(w, http.StatusOK, entries)
}

// HandleWebSocketLogs handles WS /ws/logs.
func (h *LogsHandler) HandleWebSocketLogs(w http.ResponseWriter, r *http.Request) {
	h.Hub.HandleWebSocket(w, r)
}
