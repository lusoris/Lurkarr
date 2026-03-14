package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// InstanceGroupsHandler handles instance group CRUD and member management.
type InstanceGroupsHandler struct {
	DB Store
}

// HandleListGroups returns all groups for a given app type.
func (h *InstanceGroupsHandler) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	groups, err := h.DB.ListInstanceGroups(r.Context(), database.AppType(appType))
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list instance groups"))
		return
	}
	if groups == nil {
		groups = []database.InstanceGroup{}
	}
	writeJSON(w, http.StatusOK, groups)
}

// HandleCreateGroup creates a new instance group.
func (h *InstanceGroupsHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return
	}

	limitBody(w, r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name is required"))
		return
	}

	group, err := h.DB.CreateInstanceGroup(r.Context(), database.AppType(appType), req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create instance group"))
		return
	}
	writeJSON(w, http.StatusCreated, group)
}

// HandleGetGroup returns a single group with its members.
func (h *InstanceGroupsHandler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid group ID"))
		return
	}

	group, err := h.DB.GetInstanceGroup(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("instance group not found"))
		return
	}
	writeJSON(w, http.StatusOK, group)
}

// HandleUpdateGroup renames an instance group.
func (h *InstanceGroupsHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid group ID"))
		return
	}

	limitBody(w, r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name is required"))
		return
	}

	if err := h.DB.UpdateInstanceGroup(r.Context(), id, req.Name); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update instance group"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteGroup deletes an instance group and its member associations.
func (h *InstanceGroupsHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid group ID"))
		return
	}

	if err := h.DB.DeleteInstanceGroup(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete instance group"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleSetMembers replaces all members of a group.
func (h *InstanceGroupsHandler) HandleSetMembers(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid group ID"))
		return
	}

	limitBody(w, r)
	var req struct {
		Members []struct {
			InstanceID  uuid.UUID `json:"instance_id"`
			QualityRank int       `json:"quality_rank"`
		} `json:"members"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	members := make([]database.InstanceGroupMember, len(req.Members))
	for i, m := range req.Members {
		if m.InstanceID == uuid.Nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("instance_id is required for each member"))
			return
		}
		if m.QualityRank < 1 {
			writeJSON(w, http.StatusBadRequest, errorResponse("quality_rank must be >= 1"))
			return
		}
		members[i] = database.InstanceGroupMember{
			GroupID:     id,
			InstanceID:  m.InstanceID,
			QualityRank: m.QualityRank,
		}
	}

	if err := h.DB.SetGroupMembers(r.Context(), id, members); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to set group members"))
		return
	}

	// Return the updated group
	group, err := h.DB.GetInstanceGroup(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("group updated but failed to reload"))
		return
	}
	writeJSON(w, http.StatusOK, group)
}
