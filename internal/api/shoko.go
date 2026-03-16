package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/shokoclient"
)

// ShokoHandler handles Shoko-related API endpoints.
type ShokoHandler struct {
	DB Store
}

// HandleTestConnection tests the Shoko connection.
func (h *ShokoHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}](w, r)
	if !ok {
		return
	}

	if body.APIKey == "" || (len(body.APIKey) >= 4 && body.APIKey[:4] == "****") {
		existing, err := h.DB.GetShokoSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("no stored shoko settings"))
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

	sslVerify := true
	genSettings, genErr := h.DB.GetGeneralSettings(r.Context())
	if genErr != nil {
		slog.Warn("failed to load general settings, using defaults", "error", genErr)
	}
	if genSettings != nil {
		sslVerify = genSettings.SSLVerify
	}

	client := shokoclient.NewClient(body.URL, body.APIKey, 15*time.Second, sslVerify)
	version, err := client.TestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": version.Server.Version,
	})
}

// HandleGetStats returns Shoko collection statistics.
func (h *ShokoHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	stats, err := client.GetStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// HandleGetSeriesSummary returns the series type breakdown.
func (h *ShokoHandler) HandleGetSeriesSummary(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	summary, err := client.GetSeriesSummary(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (h *ShokoHandler) getClient(r *http.Request) (*shokoclient.Client, error) {
	settings, err := h.DB.GetShokoSettings(r.Context())
	if err != nil {
		return nil, err
	}
	timeout := time.Duration(settings.Timeout) * time.Second
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	sslVerify := true
	genSettings, genErr := h.DB.GetGeneralSettings(r.Context())
	if genErr != nil {
		slog.Warn("failed to load general settings, using defaults", "error", genErr)
	}
	if genSettings != nil {
		sslVerify = genSettings.SSLVerify
	}
	return shokoclient.NewClient(settings.URL, settings.APIKey, timeout, sslVerify), nil
}
