package api

import (
	"encoding/json"
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
	limitBody(r)

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

	writeJSON(w, http.StatusCreated, provider)
}

// HandleUpdateProvider handles PUT /api/notifications/providers/{id}.
func (h *NotificationHandler) HandleUpdateProvider(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid provider ID"))
		return
	}

	limitBody(r)

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

	p, err := buildProvider(provider)
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

// buildProvider constructs a notification Provider from a database config.
func buildProvider(np *database.NotificationProvider) (notifications.Provider, error) {
	var cfg map[string]any
	if err := json.Unmarshal(np.Config, &cfg); err != nil {
		return nil, err
	}

	str := func(key string) string {
		v, _ := cfg[key].(string)
		return v
	}
	num := func(key string) int {
		v, _ := cfg[key].(float64)
		return int(v)
	}
	boolean := func(key string) bool {
		v, _ := cfg[key].(bool)
		return v
	}

	switch notifications.ProviderType(np.Type) {
	case notifications.ProviderDiscord:
		return notifications.NewDiscord(str("webhook_url"), str("username"), str("avatar_url")), nil
	case notifications.ProviderTelegram:
		return notifications.NewTelegram(str("bot_token"), str("chat_id")), nil
	case notifications.ProviderPushover:
		return notifications.NewPushover(str("api_token"), str("user_key"), str("device"), num("priority")), nil
	case notifications.ProviderGotify:
		return notifications.NewGotify(str("server_url"), str("app_token"), num("priority")), nil
	case notifications.ProviderNtfy:
		return notifications.NewNtfy(str("server_url"), str("topic"), str("token"), num("priority")), nil
	case notifications.ProviderApprise:
		var urls []string
		if rawURLs, ok := cfg["urls"].([]any); ok {
			for _, u := range rawURLs {
				if s, ok := u.(string); ok {
					urls = append(urls, s)
				}
			}
		}
		return notifications.NewApprise(str("server_url"), urls, str("tag")), nil
	case notifications.ProviderEmail:
		var to []string
		if rawTo, ok := cfg["to"].([]any); ok {
			for _, t := range rawTo {
				if s, ok := t.(string); ok {
					to = append(to, s)
				}
			}
		}
		return notifications.NewEmail(str("host"), num("port"), str("username"), str("password"), str("from"), to, boolean("starttls"), boolean("skip_verify")), nil
	case notifications.ProviderWebhook:
		headers := make(map[string]string)
		if rawHeaders, ok := cfg["headers"].(map[string]any); ok {
			for k, v := range rawHeaders {
				if s, ok := v.(string); ok {
					headers[k] = s
				}
			}
		}
		return notifications.NewWebhook(str("url"), headers), nil
	default:
		return nil, json.Unmarshal([]byte(`"unsupported provider type"`), new(string))
	}
}
