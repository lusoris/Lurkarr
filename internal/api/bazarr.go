package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/bazarrclient"
)

// BazarrHandler handles Bazarr-related API endpoints.
type BazarrHandler struct {
	DB Store
}

// HandleTestConnection tests the Bazarr connection.
func (h *BazarrHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}](w, r)
	if !ok {
		return
	}

	// If key is empty or masked, resolve from stored settings.
	if body.APIKey == "" || (len(body.APIKey) >= 4 && body.APIKey[:4] == "****") {
		existing, err := h.DB.GetBazarrSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("no stored bazarr settings"))
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

	client := bazarrclient.NewClient(body.URL, body.APIKey, 15*time.Second, sslVerify)
	status, err := client.TestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": status.Version,
	})
}

// HandleGetWanted returns counts and items of missing subtitles.
func (h *BazarrHandler) HandleGetWanted(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	episodes, epErr := client.GetWantedEpisodes(r.Context())
	movies, mvErr := client.GetWantedMovies(r.Context())

	resp := map[string]any{
		"episodes_total": 0,
		"movies_total":   0,
	}
	if epErr == nil {
		resp["episodes_total"] = episodes.Total
		resp["episodes"] = episodes.Data
	}
	if mvErr == nil {
		resp["movies_total"] = movies.Total
		resp["movies"] = movies.Data
	}
	if epErr != nil && mvErr != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(epErr.Error()))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

// HandleGetHealth returns Bazarr health check results.
func (h *BazarrHandler) HandleGetHealth(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	items, err := client.GetHealth(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, items)
}

// HandleGetHistory returns recent Bazarr subtitle download history.
func (h *BazarrHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	episodes, epErr := client.GetEpisodeHistory(r.Context())
	movies, mvErr := client.GetMovieHistory(r.Context())

	resp := map[string]any{
		"episodes_total": 0,
		"movies_total":   0,
	}
	if epErr == nil {
		resp["episodes_total"] = episodes.Total
		resp["episodes"] = episodes.Data
	}
	if mvErr == nil {
		resp["movies_total"] = movies.Total
		resp["movies"] = movies.Data
	}
	if epErr != nil && mvErr != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(epErr.Error()))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *BazarrHandler) getClient(r *http.Request) (*bazarrclient.Client, error) {
	settings, err := h.DB.GetBazarrSettings(r.Context())
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
	return bazarrclient.NewClient(settings.URL, settings.APIKey, timeout, sslVerify), nil
}
