package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// QueueHandler handles queue cleaner management endpoints.
type QueueHandler struct {
	DB *database.DB
}

// HandleGetQueueCleanerSettings handles GET /api/queue/settings/{app}.
func (h *QueueHandler) HandleGetQueueCleanerSettings(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	settings, err := h.DB.GetQueueCleanerSettings(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get queue cleaner settings"))
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateQueueCleanerSettings handles PUT /api/queue/settings/{app}.
func (h *QueueHandler) HandleUpdateQueueCleanerSettings(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	limitBody(r)
	var s database.QueueCleanerSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	s.AppType = database.AppType(appType)

	if err := h.DB.UpdateQueueCleanerSettings(r.Context(), &s); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update queue cleaner settings"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleGetScoringProfile handles GET /api/queue/scoring/{app}.
func (h *QueueHandler) HandleGetScoringProfile(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	profile, err := h.DB.GetScoringProfile(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get scoring profile"))
		return
	}

	writeJSON(w, http.StatusOK, profile)
}

// HandleUpdateScoringProfile handles PUT /api/queue/scoring/{app}.
func (h *QueueHandler) HandleUpdateScoringProfile(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	limitBody(r)
	var p database.ScoringProfile
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	// Load existing to get the ID
	existing, err := h.DB.GetScoringProfile(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get existing profile"))
		return
	}
	p.ID = existing.ID

	if err := h.DB.UpdateScoringProfile(r.Context(), &p); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update scoring profile"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleGetBlocklistLog handles GET /api/queue/blocklist/{app}.
func (h *QueueHandler) HandleGetBlocklistLog(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	logs, err := h.DB.GetBlocklistLog(r.Context(), database.AppType(appType), 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get blocklist log"))
		return
	}

	if logs == nil {
		logs = []database.BlocklistLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}

// HandleGetAutoImportLog handles GET /api/queue/imports/{app}.
func (h *QueueHandler) HandleGetAutoImportLog(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	logs, err := h.DB.GetAutoImportLog(r.Context(), database.AppType(appType), 100)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get auto import log"))
		return
	}

	if logs == nil {
		logs = []database.AutoImportLog{}
	}
	writeJSON(w, http.StatusOK, logs)
}
