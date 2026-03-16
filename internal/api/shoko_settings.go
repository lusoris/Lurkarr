package api

import (
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// HandleGetSettings returns current Shoko settings.
func (h *ShokoHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetShokoSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	settings.APIKey = settings.MaskedAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateSettings updates Shoko settings.
func (h *ShokoHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	update, ok := decodeJSON[database.ShokoSettings](w, r)
	if !ok {
		return
	}
	if update.APIKey == "" || (len(update.APIKey) >= 4 && update.APIKey[:4] == "****") {
		existing, err := h.DB.GetShokoSettings(r.Context())
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
	if err := h.DB.UpdateShokoSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	update.APIKey = update.MaskedAPIKey()
	writeJSON(w, http.StatusOK, update)
}
