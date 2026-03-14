package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/database"
)

// SettingsHandler handles settings endpoints.
type SettingsHandler struct {
	DB Store
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

	limitBody(w, r)
	var settings database.AppSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	settings.AppType = database.AppType(appType)

	// Input validation
	if settings.HourlyCap < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("hourly_cap cannot be negative"))
		return
	}
	if settings.LurkMissingCount < 0 || settings.LurkUpgradeCount < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("lurk counts cannot be negative"))
		return
	}
	if settings.SleepDuration < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("sleep_duration cannot be negative"))
		return
	}

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

	limitBody(w, r)
	var update database.GeneralSettings
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	// Preserve secret key — never from client
	update.SecretKey = current.SecretKey

	// Input validation
	if update.APITimeout < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse("api_timeout must be at least 1"))
		return
	}
	if update.StatefulResetHours < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("stateful_reset_hours cannot be negative"))
		return
	}
	if update.CommandWaitDelay < 0 || update.CommandWaitAttempts < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("command_wait values cannot be negative"))
		return
	}
	if update.MinDownloadQueueSize < 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("min_download_queue_size cannot be negative"))
		return
	}
	if update.AutoImportIntervalMinutes < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse("auto_import_interval_minutes must be at least 1"))
		return
	}

	if err := h.DB.UpsertGeneralSettings(r.Context(), &update); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save settings"))
		return
	}

	update.SecretKey = "****"
	writeJSON(w, http.StatusOK, update)
}
