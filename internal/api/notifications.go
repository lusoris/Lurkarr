package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/notifications"
)

// NotificationHandler handles notification provider API endpoints.
type NotificationHandler struct {
	DB      Store
	Manager *notifications.Manager
}

// syncManager reloads all enabled providers from the DB into the Manager.
func (h *NotificationHandler) syncManager(r *http.Request) {
	providers, err := h.DB.ListEnabledNotificationProviders(r.Context())
	if err != nil {
		slog.Error("failed to reload notification providers", "error", err)
		return
	}
	configs := make([]notifications.ProviderConfig, len(providers))
	for i, np := range providers {
		configs[i] = notifications.ProviderConfig{
			Type:   np.Type,
			Config: np.Config,
			Events: np.Events,
		}
	}
	if err := h.Manager.LoadProviders(configs); err != nil {
		slog.Error("failed to sync notification providers", "error", err)
	}
}

// HandleListProviders handles GET /api/notifications/providers.
func (h *NotificationHandler) HandleListProviders(w http.ResponseWriter, r *http.Request) {
	providers, err := h.DB.ListNotificationProviders(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list providers"))
		return
	}
	if providers == nil {
		providers = []database.NotificationProvider{}
	}
	writeJSON(w, http.StatusOK, providers)
}

// HandleGetProvider handles GET /api/notifications/providers/{id}.
func (h *NotificationHandler) HandleGetProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider ID"))
		return
	}

	provider, err := h.DB.GetNotificationProvider(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("provider not found"))
		return
	}
	writeJSON(w, http.StatusOK, provider)
}

// HandleCreateProvider handles POST /api/notifications/providers.
func (h *NotificationHandler) HandleCreateProvider(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)

	var provider database.NotificationProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if !validProviderType(provider.Type) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider type"))
		return
	}

	if err := h.DB.CreateNotificationProvider(r.Context(), &provider); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create provider"))
		return
	}

	h.syncManager(r)
	writeJSON(w, http.StatusCreated, provider)
}

// HandleUpdateProvider handles PUT /api/notifications/providers/{id}.
func (h *NotificationHandler) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider ID"))
		return
	}

	limitBody(w, r)

	var provider database.NotificationProvider
	if err := json.NewDecoder(r.Body).Decode(&provider); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	provider.ID = id

	if !validProviderType(provider.Type) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider type"))
		return
	}

	if err := h.DB.UpdateNotificationProvider(r.Context(), &provider); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update provider"))
		return
	}

	h.syncManager(r)
	writeJSON(w, http.StatusOK, provider)
}

// HandleDeleteProvider handles DELETE /api/notifications/providers/{id}.
func (h *NotificationHandler) HandleDeleteProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider ID"))
		return
	}

	if err := h.DB.DeleteNotificationProvider(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete provider"))
		return
	}

	h.syncManager(r)
	writeJSON(w, http.StatusNoContent, nil)
}

// HandleTestProvider handles POST /api/notifications/providers/{id}/test.
func (h *NotificationHandler) HandleTestProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider ID"))
		return
	}

	provider, err := h.DB.GetNotificationProvider(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("provider not found"))
		return
	}

	p, _, _, err := notifications.BuildProvider(notifications.ProviderConfig{
		Type:   provider.Type,
		Config: provider.Config,
		Events: provider.Events,
	})
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	if err := p.Test(r.Context()); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{
			"error":   "test notification failed",
			"details": err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func validProviderType(t string) bool {
	switch notifications.ProviderType(t) {
	case notifications.ProviderDiscord, notifications.ProviderTelegram,
		notifications.ProviderPushover, notifications.ProviderGotify,
		notifications.ProviderNtfy, notifications.ProviderApprise,
		notifications.ProviderEmail, notifications.ProviderWebhook:
		return true
	}
	return false
}

// HandleGetNotificationHistory handles GET /api/notifications/history.
func (h *NotificationHandler) HandleGetNotificationHistory(w http.ResponseWriter, r *http.Request) {
	history, err := h.DB.ListNotificationHistory(r.Context(), 200)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list notification history"))
		return
	}
	if history == nil {
		history = []database.NotificationHistory{}
	}
	writeJSON(w, http.StatusOK, history)
}
