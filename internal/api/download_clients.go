package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	downloadclient "github.com/lusoris/lurkarr/internal/downloadclients"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/deluge"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/qbittorrent"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/rtorrent"
	"github.com/lusoris/lurkarr/internal/downloadclients/torrent/transmission"
	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/nzbget"
	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/sabnzbd"
)

// DownloadClientHandler handles download client instance CRUD endpoints.
type DownloadClientHandler struct {
	DB Store
}

var validClientTypes = map[string]bool{
	"qbittorrent":  true,
	"transmission": true,
	"deluge":       true,
	"rtorrent":     true,
	"sabnzbd":      true,
	"nzbget":       true,
}

// HandleListDownloadClients handles GET /api/download-clients.
func (h *DownloadClientHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	instances, err := h.DB.ListDownloadClientInstances(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list download clients"))
		return
	}
	if instances == nil {
		instances = []database.DownloadClientInstance{}
	}
	for i := range instances {
		instances[i].APIKey = instances[i].MaskedAPIKey()
		instances[i].Password = instances[i].MaskedPassword()
	}
	writeJSON(w, http.StatusOK, instances)
}

// HandleCreateDownloadClient handles POST /api/download-clients.
func (h *DownloadClientHandler) HandleCreate(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var req struct {
		Name       string `json:"name"`
		ClientType string `json:"client_type"`
		URL        string `json:"url"`
		APIKey     string `json:"api_key"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Category   string `json:"category"`
		Timeout    int    `json:"timeout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	if req.Name == "" || req.URL == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name and url are required"))
		return
	}
	if err := validateAPIURL(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	if !validClientTypes[req.ClientType] {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid client_type"))
		return
	}
	if req.Timeout <= 0 {
		req.Timeout = 30
	}

	inst, err := h.DB.CreateDownloadClientInstance(r.Context(), &database.DownloadClientInstance{
		Name:       req.Name,
		ClientType: req.ClientType,
		URL:        req.URL,
		APIKey:     req.APIKey,
		Username:   req.Username,
		Password:   req.Password,
		Category:   req.Category,
		Enabled:    true,
		Timeout:    req.Timeout,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create download client"))
		return
	}
	inst.APIKey = inst.MaskedAPIKey()
	inst.Password = inst.MaskedPassword()
	writeJSON(w, http.StatusCreated, inst)
}

// HandleUpdateDownloadClient handles PUT /api/download-clients/{id}.
func (h *DownloadClientHandler) HandleUpdate(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	limitBody(w, r)
	var req struct {
		Name       string `json:"name"`
		ClientType string `json:"client_type"`
		URL        string `json:"url"`
		APIKey     string `json:"api_key"`
		Username   string `json:"username"`
		Password   string `json:"password"`
		Category   string `json:"category"`
		Enabled    bool   `json:"enabled"`
		Timeout    int    `json:"timeout"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if !validClientTypes[req.ClientType] {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid client_type"))
		return
	}

	existing, err := h.DB.GetDownloadClientInstance(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("download client not found"))
		return
	}

	// Preserve masked secrets
	apiKey := req.APIKey
	if apiKey == "" || (len(apiKey) >= 4 && apiKey[:4] == "****") {
		apiKey = existing.APIKey
	}
	password := req.Password
	if password == "" || password == "****" {
		password = existing.Password
	}

	if err := validateAPIURL(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := h.DB.UpdateDownloadClientInstance(r.Context(), &database.DownloadClientInstance{
		ID:         id,
		Name:       req.Name,
		ClientType: req.ClientType,
		URL:        req.URL,
		APIKey:     apiKey,
		Username:   req.Username,
		Password:   password,
		Category:   req.Category,
		Enabled:    req.Enabled,
		Timeout:    req.Timeout,
	}); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update download client"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteDownloadClient handles DELETE /api/download-clients/{id}.
func (h *DownloadClientHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}
	if err := h.DB.DeleteDownloadClientInstance(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete download client"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleTestDownloadClient handles POST /api/download-clients/test.
// When editing an existing client, the frontend sends the client ID so the
// backend can fall back to stored credentials for empty/masked fields.
func (h *DownloadClientHandler) HandleTest(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var req struct {
		ID         string `json:"id"`
		ClientType string `json:"client_type"`
		URL        string `json:"url"`
		APIKey     string `json:"api_key"`
		Username   string `json:"username"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	// If editing an existing client, resolve empty/masked credentials from stored record.
	if req.ID != "" {
		id, err := uuid.Parse(req.ID)
		if err == nil {
			if existing, err := h.DB.GetDownloadClientInstance(r.Context(), id); err == nil {
				if req.APIKey == "" || req.APIKey == "keep" || (len(req.APIKey) >= 4 && req.APIKey[:4] == "****") {
					req.APIKey = existing.APIKey
				}
				if req.Password == "" || req.Password == "keep" || req.Password == "****" {
					req.Password = existing.Password
				}
				if req.URL == "" {
					req.URL = existing.URL
				}
				if req.ClientType == "" {
					req.ClientType = existing.ClientType
				}
				if req.Username == "" {
					req.Username = existing.Username
				}
			}
		}
	}

	if req.URL == "" || !validClientTypes[req.ClientType] {
		writeJSON(w, http.StatusBadRequest, errorResponse("url and valid client_type required"))
		return
	}
	if err := validateAPIURL(req.URL); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	client := buildClient(req.ClientType, req.URL, req.APIKey, req.Username, req.Password, 15*time.Second)
	if client == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("unsupported client type"))
		return
	}

	version, err := client.TestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "version": version})
}

// HandleHealthCheckDownloadClient handles GET /api/download-clients/{id}/health.
func (h *DownloadClientHandler) HandleHealthCheck(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	inst, err := h.DB.GetDownloadClientInstance(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("download client not found"))
		return
	}

	timeout := time.Duration(inst.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := buildClient(inst.ClientType, inst.URL, inst.APIKey, inst.Username, inst.Password, timeout)
	if client == nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "offline", "client_type": inst.ClientType, "version": ""})
		return
	}

	version, err := client.TestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "offline", "client_type": inst.ClientType, "version": ""})
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok", "client_type": inst.ClientType, "version": version})
}

// HandleStatus handles GET /api/download-clients/{id}/status.
func (h *DownloadClientHandler) HandleStatus(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	inst, err := h.DB.GetDownloadClientInstance(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("download client not found"))
		return
	}

	timeout := time.Duration(inst.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := buildClient(inst.ClientType, inst.URL, inst.APIKey, inst.Username, inst.Password, timeout)
	if client == nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "offline"})
		return
	}

	status, err := client.GetStatus(r.Context())
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]any{"status": "offline", "error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, status)
}

// HandleItems handles GET /api/download-clients/{id}/items.
func (h *DownloadClientHandler) HandleItems(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	inst, err := h.DB.GetDownloadClientInstance(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("download client not found"))
		return
	}

	timeout := time.Duration(inst.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	client := buildClient(inst.ClientType, inst.URL, inst.APIKey, inst.Username, inst.Password, timeout)
	if client == nil {
		writeJSON(w, http.StatusOK, []any{})
		return
	}

	items, err := client.GetItems(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// buildClient constructs the appropriate download client based on type.
func buildClient(clientType, url, apiKey, username, password string, timeout time.Duration) downloadclient.Client {
	switch downloadclient.ClientType(clientType) {
	case downloadclient.TypeQBittorrent:
		return downloadclient.NewQBittorrentAdapter(qbittorrent.NewClient(url, username, password, timeout))
	case downloadclient.TypeTransmission:
		return downloadclient.NewTransmissionAdapter(transmission.NewClient(url, username, password, timeout))
	case downloadclient.TypeDeluge:
		return downloadclient.NewDelugeAdapter(deluge.NewClient(url, password, timeout))
	case downloadclient.TypeSABnzbd:
		return downloadclient.NewSABnzbdAdapter(sabnzbd.NewClient(url, apiKey, timeout))
	case downloadclient.TypeNZBGet:
		return downloadclient.NewNZBGetAdapter(nzbget.NewClient(url, username, password, timeout))
	case downloadclient.TypeRTorrent:
		return downloadclient.NewRTorrentAdapter(rtorrent.NewClient(url, username, password, timeout))
	default:
		return nil
	}
}
