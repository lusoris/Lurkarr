package api

import (
	"net/http"
	"time"

	"github.com/lusoris/lurkarr/internal/downloadclients/usenet/sabnzbd"
)

// SABnzbdHandler handles SABnzbd-related API endpoints.
type SABnzbdHandler struct {
	DB Store
}

// HandleGetQueue returns the SABnzbd download queue.
func (h *SABnzbdHandler) HandleGetQueue(w http.ResponseWriter, r *http.Request) {
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

// HandleGetHistory returns the SABnzbd download history.
func (h *SABnzbdHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	history, err := client.GetHistory(r.Context(), 100)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, history)
}

// HandleGetStats returns SABnzbd server statistics.
func (h *SABnzbdHandler) HandleGetStats(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	stats, err := client.GetServerStats(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, stats)
}

// HandlePause pauses SABnzbd downloads.
func (h *SABnzbdHandler) HandlePause(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	if err := client.Pause(r.Context()); err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "paused"})
}

// HandleResume resumes SABnzbd downloads.
func (h *SABnzbdHandler) HandleResume(w http.ResponseWriter, r *http.Request) {
	client, err := h.getClient(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}
	if err := client.Resume(r.Context()); err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "resumed"})
}

// HandleTestConnection tests the SABnzbd connection.
func (h *SABnzbdHandler) HandleTestConnection(w http.ResponseWriter, r *http.Request) {
	body, ok := decodeJSON[struct {
		URL    string `json:"url"`
		APIKey string `json:"api_key"`
	}](w, r)
	if !ok {
		return
	}
	if body.URL == "" || body.APIKey == "" || (len(body.APIKey) >= 4 && body.APIKey[:4] == "****") {
		existing, err := h.DB.GetSABnzbdSettings(r.Context())
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("url and api_key required"))
			return
		}
		if body.APIKey == "" || (len(body.APIKey) >= 4 && body.APIKey[:4] == "****") {
			body.APIKey = existing.APIKey
		}
		if body.URL == "" {
			body.URL = existing.URL
		}
	}
	if body.URL == "" || body.APIKey == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("url and api_key required"))
		return
	}

	client := sabnzbd.NewClient(body.URL, body.APIKey, 15*time.Second)
	version, err := client.TestConnection(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, errorResponse(err.Error()))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "ok",
		"version": version,
	})
}

func (h *SABnzbdHandler) getClient(r *http.Request) (*sabnzbd.Client, error) {
	settings, err := h.DB.GetSABnzbdSettings(r.Context())
	if err != nil {
		return nil, err
	}
	return sabnzbd.NewClient(settings.URL, settings.APIKey, time.Duration(settings.Timeout)*time.Second), nil
}
