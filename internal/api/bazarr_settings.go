package api

import (
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// HandleGetSettings returns current Bazarr settings.
func (h *BazarrHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetBazarrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	settings.APIKey = settings.MaskedAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateSettings updates Bazarr settings.
func (h *BazarrHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	update, ok := decodeJSON[database.BazarrSettings](w, r)
	if !ok {
		return
	}
	// If masked key sent back, preserve existing
	if update.APIKey == "" || (len(update.APIKey) >= 4 && update.APIKey[:4] == "****") {
		existing, err := h.DB.GetBazarrSettings(r.Context())
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
	if err := h.DB.UpdateBazarrSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse(err.Error()))
		return
	}
	update.APIKey = update.MaskedAPIKey()
	writeJSON(w, http.StatusOK, update)
}
