package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/seerr"
)

// SeerrHandler handles Seerr API endpoints.
type SeerrHandler struct {
	DB Store
}

// HandleGetSettings handles GET /api/seerr/settings.
func (h *SeerrHandler) HandleGetSettings(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load seerr settings"))
		return
	}
	settings.APIKey = settings.MaskedSeerrAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleUpdateSettings handles PUT /api/seerr/settings.
func (h *SeerrHandler) HandleUpdateSettings(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)

	var req struct {
		URL                 string `json:"url"`
		APIKey              string `json:"api_key"`
		Enabled             bool   `json:"enabled"`
		SyncIntervalMinutes int    `json:"sync_interval_minutes"`
		AutoApprove         bool   `json:"auto_approve"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load seerr settings"))
		return
	}

	settings.URL = req.URL
	settings.Enabled = req.Enabled
	settings.SyncIntervalMinutes = req.SyncIntervalMinutes
	settings.AutoApprove = req.AutoApprove

	// Only update API key if a non-masked value is provided.
	if req.APIKey != "" && req.APIKey != settings.MaskedSeerrAPIKey() {
		settings.APIKey = req.APIKey
	}

	if err := h.DB.UpdateSeerrSettings(r.Context(), settings); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update seerr settings"))
		return
	}

	settings.APIKey = settings.MaskedSeerrAPIKey()
	writeJSON(w, http.StatusOK, settings)
}

// HandleTestConnection handles POST /api/seerr/test.
func (h *SeerrHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load seerr settings"))
		return
	}

	if settings.URL == "" || settings.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("seerr URL and API key are required"))
		return
	}

	client := seerr.NewClient(settings.URL, settings.APIKey, 10*time.Second)
	info, err := client.GetAbout(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error":   "connection failed",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"version": info.Version,
	})
}

// HandleGetRequests handles GET /api/seerr/requests.
func (h *SeerrHandler) HandleGetRequests(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil || settings.URL == "" || settings.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("seerr not configured"))
		return
	}

	client := seerr.NewClient(settings.URL, settings.APIKey, 10*time.Second)
	filter := r.URL.Query().Get("filter")
	resp, err := client.ListRequests(r.Context(), filter, 50, 0)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse("failed to fetch requests"))
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

// HandleGetRequestCount handles GET /api/seerr/requests/count.
func (h *SeerrHandler) HandleGetRequestCount(w http.ResponseWriter, r *http.Request) {
	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil || settings.URL == "" || settings.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("seerr not configured"))
		return
	}

	client := seerr.NewClient(settings.URL, settings.APIKey, 10*time.Second)
	count, err := client.GetRequestCount(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse("failed to fetch request count"))
		return
	}

	writeJSON(w, http.StatusOK, count)
}
