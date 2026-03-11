package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

// ProwlarrHandler handles Prowlarr-related API endpoints.
type ProwlarrHandler struct {
	DB *database.DB
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
func (h *ProwlarrHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	limitBody(r)
	var body struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	if body.URL == "" || body.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("url and api_key required"))
		return
	}

	// SSRF protection
	isPrivate, err := arrclient.IsPrivateIP(body.URL)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid URL"))
		return
	}
	if isPrivate {
		writeJSON(w, http.StatusForbidden, errorResponse("private/internal URLs are not allowed"))
		return
	}

	client := arrclient.NewClient(body.URL, body.APIKey, 15*time.Second, true)
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
	return arrclient.NewClient(settings.URL, settings.APIKey, time.Duration(settings.Timeout)*time.Second, true), nil
}
