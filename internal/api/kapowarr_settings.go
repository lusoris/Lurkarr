package api

import (
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// HandleGetSettings returns current Kapowarr settings.
func (h *KapowarrHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetKapowarrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	settings.APIKey = settings.MaskedAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateSettings updates Kapowarr settings.
func (h *KapowarrHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	update, ok := decodeJSON[database.KapowarrSettings](w, r)
	if !ok {
		return
	}
	if update.APIKey == "" || (len(update.APIKey) >= 4 && update.APIKey[:4] == "****") {
		existing, err := h.DB.GetKapowarrSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
			return
		}
		update.APIKey = existing.APIKey
	}
	if update.URL != "" {
		if err := validateAPIURL(update.URL); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
	}
	if err := h.DB.UpdateKapowarrSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	update.APIKey = update.MaskedAPIKey()
	writeJSON(w, http.StatusOK, update)
}
