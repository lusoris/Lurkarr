package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// HandleGetProwlarrSettings returns current Prowlarr settings.
func (h *ProwlarrHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetProwlarrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	settings.APIKey = settings.MaskedAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateProwlarrSettings updates Prowlarr settings.
func (h *ProwlarrHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	limitBody(r)
	var update database.ProwlarrSettings
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	// If masked key sent back, preserve existing
	if update.APIKey == "" || update.APIKey[:4] == "****" {
		existing, err := h.DB.GetProwlarrSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
			return
		}
		update.APIKey = existing.APIKey
	}
	if err := h.DB.UpdateProwlarrSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	update.APIKey = update.MaskedAPIKey()
	writeJSON(w, http.StatusOK, update)
}
