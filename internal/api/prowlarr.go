package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
)

// ProwlarrHandler handles Prowlarr-related API endpoints.
type ProwlarrHandler struct {
	DB Store
}

// HandleGetIndexers returns all Prowlarr indexers.
func (h *ProwlarrHandler) HandleGetIndexers(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	indexers, err := client.ProwlarrGetIndexers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, indexers)
}

// HandleGetIndexerStats returns Prowlarr indexer statistics.
func (h *ProwlarrHandler) HandleGetIndexerStats(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	stats, err := client.ProwlarrGetIndexerStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// HandleTestConnection tests the Prowlarr connection.
// If a masked API key (starts with "****") or empty key is sent, the stored key is used instead.
func (h *ProwlarrHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}](w, r)
	if !ok {
		return
	}

	// If key is empty or masked, resolve from stored settings.
	if body.APIKey == "" || (len(body.APIKey) >= 4 && body.APIKey[:4] == "****") {
		existing, err := h.DB.GetProwlarrSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("no stored prowlarr settings"))
			return
		}
		body.APIKey = existing.APIKey
		if body.URL == "" {
			body.URL = existing.URL
		}
	}

	if body.URL == "" || body.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("url and api_key required"))
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

	client := arrclient.NewClient(body.URL, body.APIKey, 15*time.Second, sslVerify)
	status, err := client.ProwlarrTestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": status.Version,
	})
}

func (h *ProwlarrHandler) getClient(r *http.Request) (*arrclient.Client, error) {
	settings, err := h.DB.GetProwlarrSettings(r.Context())
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(settings.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	genSettings, genErr := h.DB.GetGeneralSettings(r.Context())
	if genErr != nil {
		slog.Warn("failed to load general settings, using defaults", "error", genErr)
	}
	sslVerify := true
	if genSettings != nil {
		sslVerify = genSettings.SSLVerify
	}
	return arrclient.NewClient(settings.URL, settings.APIKey, timeout, sslVerify), nil
}
