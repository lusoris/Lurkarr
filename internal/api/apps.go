package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// AppsHandler handles app instance CRUD endpoints.
type AppsHandler struct {
	DB Store
}

// HandleListInstances handles GET /api/instances/{app}.
func (h *AppsHandler) HandleListInstances(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	instances, err := h.DB.ListInstances(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list instances"))
		return
	}

	// Mask API keys in response
	for i := range instances {
		instances[i].APIKey = instances[i].MaskedAPIKey()
	}

	writeJSON(w, http.StatusOK, instances)
}

// HandleCreateInstance handles POST /api/instances/{app}.
func (h *AppsHandler) HandleCreateInstance(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	limitBody(r)
	var req struct {
		Name   string `json:"name"`
		APIURL string `json:"api_url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" || req.APIURL == "" || req.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name, api_url, and api_key required"))
		return
	}

	inst, err := h.DB.CreateInstance(r.Context(), database.AppType(appType), req.Name, req.APIURL, req.APIKey)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create instance"))
		return
	}

	inst.APIKey = inst.MaskedAPIKey()
	writeJSON(w, http.StatusCreated, inst)
}

// HandleUpdateInstance handles PUT /api/instances/{id}.
func (h *AppsHandler) HandleUpdateInstance(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid instance ID"))
		return
	}

	limitBody(r)
	var req struct {
		Name    string `json:"name"`
		APIURL  string `json:"api_url"`
		APIKey  string `json:"api_key"`
		Enabled bool   `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
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

	if err := h.DB.UpdateInstance(r.Context(), id, req.Name, req.APIURL, req.APIKey, req.Enabled); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update instance"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteInstance handles DELETE /api/instances/{id}.
func (h *AppsHandler) HandleDeleteInstance(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid instance ID"))
		return
	}

	if err := h.DB.DeleteInstance(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete instance"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleTestConnection handles POST /api/instances/test.
func (h *AppsHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	limitBody(r)
	var req struct {
		APIURL  string `json:"api_url"`
		APIKey  string `json:"api_key"`
		AppType string `json:"app_type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.APIURL == "" || req.APIKey == "" || req.AppType == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("api_url, api_key, and app_type required"))
		return
	}

	if !database.ValidAppType(req.AppType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	// SSRF protection
	isPrivate, err := arrclient.IsPrivateIP(req.APIURL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid URL"))
		return
	}
	if isPrivate {
		writeJSON(w, http.StatusForbidden, errorResponse("private/internal URLs are not allowed"))
		return
	}

	client := arrclient.NewClient(req.APIURL, req.APIKey, 15_000_000_000, true)

	apiVersion := "v3"
	if req.AppType == "lidarr" || req.AppType == "readarr" {
		apiVersion = "v1"
	}

	status, err := client.TestConnection(r.Context(), apiVersion)
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
