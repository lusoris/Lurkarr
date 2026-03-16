package api

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/kapowarrclient"
)

// KapowarrHandler handles Kapowarr-related API endpoints.
type KapowarrHandler struct {
	DB Store
}

// HandleTestConnection tests the Kapowarr connection.
func (h *KapowarrHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}](w, r)
	if !ok {
		return
	}

	if body.APIKey == "" || (len(body.APIKey) >= 4 && body.APIKey[:4] == "****") {
		existing, err := h.DB.GetKapowarrSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("no stored kapowarr settings"))
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

	client := kapowarrclient.NewClient(body.URL, body.APIKey, 15*time.Second, sslVerify)
	about, err := client.TestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": about.Version,
	})
}

// HandleGetStats returns Kapowarr library statistics.
func (h *KapowarrHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	stats, err := client.GetVolumeStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// HandleGetQueue returns Kapowarr active downloads.
func (h *KapowarrHandler) HandleGetQueue(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	queue, err := client.GetQueue(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, queue)
}

// HandleGetTasks returns Kapowarr task queue.
func (h *KapowarrHandler) HandleGetTasks(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	tasks, err := client.GetTasks(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *KapowarrHandler) getClient(r *http.Request) (*kapowarrclient.Client, error) {
	settings, err := h.DB.GetKapowarrSettings(r.Context())
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
	return kapowarrclient.NewClient(settings.URL, settings.APIKey, timeout, sslVerify), nil
}
