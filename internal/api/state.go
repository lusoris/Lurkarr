package api

import (
	"log/slog"
	"net/http"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// StateHandler handles state reset endpoints.
type StateHandler struct {
	DB Store
}

// HandleGetState handles GET /api/state.
func (h *StateHandler) HandleGetState(w http.ResponseWriter, r *http.Request) {
	appType := r.URL.Query().Get("app")
	if appType != "" && !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	// Return list of instances with their last reset times
	var result []map[string]any

	appTypes := database.AllAppTypes()
	if appType != "" {
		appTypes = []database.AppType{database.AppType(appType)}
	}

	for _, at := range appTypes {
		instances, err := h.DB.ListInstances(r.Context(), at)
		if err != nil {
			continue
		}
		for _, inst := range instances {
			lastReset, resetErr := h.DB.GetLastReset(r.Context(), at, inst.ID)
			if resetErr != nil {
				slog.Warn("failed to get last reset", "app_type", at, "instance", inst.ID, "error", resetErr)
			}
			result = append(result, map[string]any{
				"app_type":    at,
				"instance_id": inst.ID,
				"name":        inst.Name,
				"last_reset":  lastReset,
			})
		}
	}
	if result == nil {
		result = []map[string]any{}
	}

	writeJSON(w, http.StatusOK, result)
}

// HandleResetState handles POST /api/state/reset.
func (h *StateHandler) HandleResetState(w http.ResponseWriter, r *http.Request) {
	appType := r.URL.Query().Get("app")
	instanceID := r.URL.Query().Get("instance_id")

	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	id, err := uuid.Parse(instanceID)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid instance_id"))
		return
	}

	if err := h.DB.ResetState(r.Context(), database.AppType(appType), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to reset state"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
