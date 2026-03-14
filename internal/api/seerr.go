package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/lusoris/lurkarr/internal/seerr"
)

// SeerrHandler handles Seerr API endpoints.
type SeerrHandler struct {
	DB     Store
	Router *seerr.RequestRouter
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

	if req.URL != "" {
		if err := validateAPIURL(req.URL); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
			return
		}
	}

	// Only update API key if a non-masked value is provided.
	if req.APIKey != "" && (len(req.APIKey) < 4 || req.APIKey[:4] != "****") {
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
// Accepts optional url/api_key in body; falls back to stored settings.
func (h *SeerrHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var body struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	// Ignore decode errors — body is optional, we fall back to stored settings.
	_ = json.NewDecoder(r.Body).Decode(&body)

	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load seerr settings"))
		return
	}

	url := body.URL
	if url == "" {
		url = settings.URL
	}
	apiKey := body.APIKey
	if apiKey == "" || (len(apiKey) >= 4 && apiKey[:4] == "****") {
		apiKey = settings.APIKey
	}

	if url == "" || apiKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("seerr URL and API key are required"))
		return
	}

	if err := validateAPIURL(url); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	client := seerr.NewClient(url, apiKey, 10*time.Second)
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

// HandleScanDuplicates scans all Seerr requests against cross-instance data
// and returns flags for potential duplicates.
func (h *SeerrHandler) HandleScanDuplicates(w http.ResponseWriter, r *http.Request) {
	if h.Router == nil {
		writeJSON(w, http.StatusServiceUnavailable, errorResponse("request router not configured"))
		return
	}

	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil || settings.URL == "" || settings.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("seerr not configured"))
		return
	}

	client := seerr.NewClient(settings.URL, settings.APIKey, 30*time.Second)
	result, err := h.Router.ScanForDuplicates(r.Context(), client)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse("duplicate scan failed"))
		return
	}

	writeJSON(w, http.StatusOK, result)
}

// HandleReassignRequest handles POST /api/seerr/requests/{id}/reassign.
// It modifies a Seerr request to target a different server/quality profile.
func (h *SeerrHandler) HandleReassignRequest(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)

	idStr := r.PathValue("id")
	requestID, err := strconv.Atoi(idStr)
	if err != nil || requestID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request ID"))
		return
	}

	var body struct {
		ServerID   int    `json:"server_id"`
		ProfileID  int    `json:"profile_id"`
		RootFolder string `json:"root_folder"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	if body.ServerID <= 0 || body.ProfileID <= 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse("server_id and profile_id are required"))
		return
	}

	settings, err := h.DB.GetSeerrSettings(r.Context())
	if err != nil || settings.URL == "" || settings.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("seerr not configured"))
		return
	}

	client := seerr.NewClient(settings.URL, settings.APIKey, 15*time.Second)

	// Apply the reassignment via the Seerr API.
	updated, err := client.ModifyRequest(r.Context(), requestID, body.ServerID, body.ProfileID, body.RootFolder)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(fmt.Sprintf("failed to reassign request: %v", err)))
		return
	}

	writeJSON(w, http.StatusOK, updated)
}
