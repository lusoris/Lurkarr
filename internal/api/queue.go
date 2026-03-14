package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/lusoris/lurkarr/internal/database"
)

// QueueHandler handles queue cleaner management endpoints.
type QueueHandler struct {
	DB Store
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

	limitBody(w, r)
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

	limitBody(w, r)
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

// HandleGetStrikeLog handles GET /api/queue/strikes/{app}.
func (h *QueueHandler) HandleGetStrikeLog(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	strikes, err := h.DB.GetStrikeLog(r.Context(), database.AppType(appType), 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get strike log"))
		return
	}

	if strikes == nil {
		strikes = []database.QueueStrike{}
	}
	writeJSON(w, http.StatusOK, strikes)
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

// HandleGetDownloadClientSettings handles GET /api/queue/download-client/{app}.
func (h *QueueHandler) HandleGetDownloadClientSettings(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	settings, err := h.DB.GetDownloadClientSettings(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to get download client settings"))
		return
	}

	settings.Password = settings.MaskedPassword()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateDownloadClientSettings handles PUT /api/queue/download-client/{app}.
func (h *QueueHandler) HandleUpdateDownloadClientSettings(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	limitBody(w, r)
	var s database.DownloadClientSettings
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	s.AppType = database.AppType(appType)

	// If masked password sent back, preserve existing.
	if s.Password == "" || s.Password == "****" {
		existing, err := h.DB.GetDownloadClientSettings(r.Context(), s.AppType)
		if err == nil {
			s.Password = existing.Password
		}
	}

	if err := h.DB.UpdateDownloadClientSettings(r.Context(), &s); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update download client settings"))
		return
	}

	s.Password = s.MaskedPassword()
	writeJSON(w, http.StatusOK, s)
}

// HandleListSeedingRuleGroups handles GET /api/queue/seeding-groups.
func (h *QueueHandler) HandleListSeedingRuleGroups(w http.ResponseWriter, r *http.Request) {
	groups, err := h.DB.ListSeedingRuleGroups(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list seeding rule groups"))
		return
	}
	if groups == nil {
		groups = []database.SeedingRuleGroup{}
	}
	writeJSON(w, http.StatusOK, groups)
}

// HandleCreateSeedingRuleGroup handles POST /api/queue/seeding-groups.
func (h *QueueHandler) HandleCreateSeedingRuleGroup(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var g database.SeedingRuleGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	created, err := h.DB.CreateSeedingRuleGroup(r.Context(), &g)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create seeding rule group"))
		return
	}
	writeJSON(w, http.StatusCreated, created)
}

// HandleUpdateSeedingRuleGroup handles PUT /api/queue/seeding-groups/{id}.
func (h *QueueHandler) HandleUpdateSeedingRuleGroup(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var g database.SeedingRuleGroup
	if err := json.NewDecoder(r.Body).Decode(&g); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid group id"))
		return
	}
	g.ID = id
	if err := h.DB.UpdateSeedingRuleGroup(r.Context(), &g); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update seeding rule group"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteSeedingRuleGroup handles DELETE /api/queue/seeding-groups/{id}.
func (h *QueueHandler) HandleDeleteSeedingRuleGroup(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid group id"))
		return
	}
	if err := h.DB.DeleteSeedingRuleGroup(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete seeding rule group"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
