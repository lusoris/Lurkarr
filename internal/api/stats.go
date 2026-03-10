package api

import (
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// StatsHandler handles stats endpoints.
type StatsHandler struct {
	DB *database.DB
}

// HandleGetStats handles GET /api/stats.
func (h *StatsHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.DB.GetAllStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load stats"))
		return
	}

	writeJSON(w, http.StatusOK, stats)
}

// HandleResetStats handles POST /api/stats/reset.
func (h *StatsHandler) HandleResetStats(w http.ResponseWriter, r *http.Request) {
	if err := h.DB.ResetStats(r.Context()); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to reset stats"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleGetHourlyCaps handles GET /api/stats/hourly-caps.
func (h *StatsHandler) HandleGetHourlyCaps(w http.ResponseWriter, r *http.Request) {
	caps, err := h.DB.GetAllHourlyCaps(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load hourly caps"))
		return
	}

	writeJSON(w, http.StatusOK, caps)
}
