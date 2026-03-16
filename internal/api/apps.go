package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// AppsHandler handles app instance CRUD endpoints.
type AppsHandler struct {
	DB Store
}

// HandleListAllInstances handles GET /api/instances (no app filter).
func (h *AppsHandler) HandleListAllInstances(w http.ResponseWriter, r *http.Request) {
	all, err := h.DB.ListAllInstances(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list instances"))
		return
	}
	if all == nil {
		all = []database.AppInstance{}
	}
	grouped := make(map[string][]database.AppInstance)
	for i := range all {
		all[i].APIKey = all[i].MaskedAPIKey()
		grouped[string(all[i].AppType)] = append(grouped[string(all[i].AppType)], all[i])
	}
	writeJSON(w, http.StatusOK, grouped)
}

// HandleListInstances handles GET /api/instances/{app}.
func (h *AppsHandler) HandleListInstances(w http.ResponseWriter, r *http.Request) {
	appType, ok := validAppTypeParam(w, r)
	if !ok {
		return
	}

	instances, err := h.DB.ListInstances(r.Context(), appType)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list instances"))
		return
	}
	if instances == nil {
		instances = []database.AppInstance{}
	}

	// Mask API keys in response
	for i := range instances {
		instances[i].APIKey = instances[i].MaskedAPIKey()
	}

	writeJSON(w, http.StatusOK, instances)
}

// HandleCreateInstance handles POST /api/instances/{app}.
func (h *AppsHandler) HandleCreateInstance(w http.ResponseWriter, r *http.Request) {
	appType, ok := validAppTypeParam(w, r)
	if !ok {
		return
	}

	req, ok := decodeJSON[struct {
		Name   string `json:"name"`
		APIURL string `json:"api_url"`
		APIKey string `json:"api_key"`
	}](w, r)
	if !ok {
		return
	}
	if req.Name == "" || req.APIURL == "" || req.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name, api_url, and api_key required"))
		return
	}

	if err := validateAPIURL(req.APIURL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	inst, err := h.DB.CreateInstance(r.Context(), appType, req.Name, req.APIURL, req.APIKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create instance"))
		return
	}

	inst.APIKey = inst.MaskedAPIKey()
	writeJSON(w, http.StatusCreated, inst)
}

// HandleUpdateInstance handles PUT /api/instances/{id}.
func (h *AppsHandler) HandleUpdateInstance(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeJSON[struct {
		Name    string `json:"name"`
		APIURL  string `json:"api_url"`
		APIKey  string `json:"api_key"`
		Enabled bool   `json:"enabled"`
	}](w, r)
	if !ok {
		return
	}

	// If API key is masked, keep the existing one
	if req.APIKey == "" || (len(req.APIKey) >= 4 && req.APIKey[:4] == "****") {
		existing, err := h.DB.GetInstance(r.Context(), id)
		if err != nil {
			writeJSON(w, http.StatusNotFound, errorResponse("instance not found"))
			return
		}
		req.APIKey = existing.APIKey
	}

	if err := validateAPIURL(req.APIURL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.DB.UpdateInstance(r.Context(), id, req.Name, req.APIURL, req.APIKey, req.Enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update instance"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteInstance handles DELETE /api/instances/{id}.
func (h *AppsHandler) HandleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	if err := h.DB.DeleteInstance(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete instance"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleHealthCheckInstance handles GET /api/instances/{id}/health.
// It tests the connection to a stored instance using its saved credentials.
func (h *AppsHandler) HandleHealthCheckInstance(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	inst, err := h.DB.GetInstance(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("instance not found"))
		return
	}

	genSettings, err := h.DB.GetGeneralSettings(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load settings"))
		return
	}

	timeout := time.Duration(genSettings.APITimeout) * time.Second
	if timeout == 0 {
		timeout = 15 * time.Second
	}

	client := arrclient.NewClient(inst.APIURL, inst.APIKey, timeout, genSettings.SSLVerify)
	apiVersion := arrclient.APIVersionFor(string(inst.AppType))

	status, err := client.TestConnection(r.Context(), apiVersion)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  "offline",
			"version": "",
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"app":     status.AppName,
		"version": status.Version,
	})
}

// HandleTestConnection handles POST /api/instances/test.
// When editing an existing instance, the frontend sends the instance ID so the
// backend can fall back to stored credentials for empty/masked API keys.
func (h *AppsHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	req, ok := decodeJSON[struct {
		ID      string `json:"id"`
		APIURL  string `json:"api_url"`
		APIKey  string `json:"api_key"`
		AppType string `json:"app_type"`
	}](w, r)
	if !ok {
		return
	}

	// If API key is empty or masked, resolve from stored instance.
	if req.ID != "" && (req.APIKey == "" || (len(req.APIKey) >= 4 && req.APIKey[:4] == "****")) {
		id, err := uuid.Parse(req.ID)
		if err == nil {
			if existing, err := h.DB.GetInstance(r.Context(), id); err == nil {
				req.APIKey = existing.APIKey
				if req.APIURL == "" {
					req.APIURL = existing.APIURL
				}
				if req.AppType == "" {
					req.AppType = string(existing.AppType)
				}
			}
		}
	}

	if req.APIURL == "" || req.APIKey == "" || req.AppType == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("api_url, api_key, and app_type required"))
		return
	}

	if err := validateAPIURL(req.APIURL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if !database.ValidAppType(req.AppType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	genSettings, genErr := h.DB.GetGeneralSettings(r.Context())
	if genErr != nil {
		slog.Warn("failed to load general settings, using defaults", "error", genErr)
	}
	sslVerify := true
	if genSettings != nil {
		sslVerify = genSettings.SSLVerify
	}

	client := arrclient.NewClient(req.APIURL, req.APIKey, 15*time.Second, sslVerify)
	status, err := client.TestConnection(r.Context(), arrclient.APIVersionFor(req.AppType))
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse("connection failed: "+err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":  "ok",
		"app":     status.AppName,
		"version": status.Version,
	})
}
