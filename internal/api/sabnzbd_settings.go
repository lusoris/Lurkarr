package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// HandleGetSABnzbdSettings returns current SABnzbd settings.
func (h *SABnzbdHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetSABnzbdSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	settings.APIKey = settings.MaskedAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateSABnzbdSettings updates SABnzbd settings.
func (h *SABnzbdHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var update database.SABnzbdSettings
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	// If masked key sent back, preserve existing
	if update.APIKey == "" || (len(update.APIKey) >= 4 && update.APIKey[:4] == "****") {
		existing, err := h.DB.GetSABnzbdSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
			return
		}
		update.APIKey = existing.APIKey
	}
	if err := h.DB.UpdateSABnzbdSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	update.APIKey = update.MaskedAPIKey()
	writeJSON(w, http.StatusOK, update)
}
