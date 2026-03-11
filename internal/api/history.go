package api

import (
	"net/http"
	"strconv"

	"github.com/lusoris/lurkarr/internal/database"
)

// HistoryHandler handles hunt history endpoints.
type HistoryHandler struct {
	DB Store
}

// HandleListHistory handles GET /api/history.
func (h *HistoryHandler) HandleListHistory(w http.ResponseWriter, r *http.Request) {
	q := database.HistoryQuery{
		AppType: r.URL.Query().Get("app"),
		Search:  r.URL.Query().Get("search"),
		Limit:   50,
		Offset:  0,
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			q.Limit = n
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			q.Offset = n
		}
	}

	items, total, err := h.DB.ListHuntHistory(r.Context(), q)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load history"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": items,
		"total": total,
	})
}

// HandleDeleteHistory handles DELETE /api/history/{app}.
func (h *HistoryHandler) HandleDeleteHistory(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	if err := h.DB.DeleteHistory(r.Context(), database.AppType(appType)); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete history"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
