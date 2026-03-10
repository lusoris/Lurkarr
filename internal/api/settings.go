package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// SettingsHandler handles settings endpoints.
type SettingsHandler struct {
	DB *database.DB
}

// HandleGetAppSettings handles GET /api/settings/{app}.
func (h *SettingsHandler) HandleGetAppSettings(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	settings, err := h.DB.GetAppSettings(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load settings"))
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateAppSettings handles PUT /api/settings/{app}.
func (h *SettingsHandler) HandleUpdateAppSettings(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	var settings database.AppSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	settings.AppType = database.AppType(appType)

	if err := h.DB.UpdateAppSettings(r.Context(), &settings); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save settings"))
		return
	}

	writeJSON(w, http.StatusOK, settings)
}

// HandleGetGeneralSettings handles GET /api/settings/general.
func (h *SettingsHandler) HandleGetGeneralSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetGeneralSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load settings"))
		return
	}

	// Mask the secret key
	settings.SecretKey = "****"
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateGeneralSettings handles PUT /api/settings/general.
func (h *SettingsHandler) HandleUpdateGeneralSettings(w http.ResponseWriter, r *http.Request) {
	current, err := h.DB.GetGeneralSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load current settings"))
		return
	}

	var update database.GeneralSettings
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	// Preserve secret key — never from client
	update.SecretKey = current.SecretKey

	if err := h.DB.UpsertGeneralSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save settings"))
		return
	}

	update.SecretKey = "****"
	writeJSON(w, http.StatusOK, update)
}
