package api

import (
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// InstanceGroupsHandler handles instance group CRUD and member management.
type InstanceGroupsHandler struct {
	DB Store
}

// HandleListGroups returns all groups for a given app type.
func (h *InstanceGroupsHandler) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	appType, ok := validAppTypeParam(w, r)
	if !ok {
		return
	}

	groups, err := h.DB.ListInstanceGroups(r.Context(), appType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if groups == nil {
		groups = []database.InstanceGroup{}
	}
	writeJSON(w, http.StatusOK, groups)
}

// HandleCreateGroup creates a new instance group.
func (h *InstanceGroupsHandler) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	appType, ok := validAppTypeParam(w, r)
	if !ok {
		return
	}

	req, ok := decodeJSON[struct {
		Name string `json:"name"`
	}](w, r)
	if !ok {
		return
	}
	if req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name is required"))
		return
	}

	group, err := h.DB.CreateInstanceGroup(r.Context(), appType, req.Name)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create instance group"))
		return
	}
	writeJSON(w, http.StatusCreated, group)
}

// HandleGetGroup returns a single group with its members.
func (h *InstanceGroupsHandler) HandleGetGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	group, err := h.DB.GetInstanceGroup(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("instance group not found"))
		return
	}
	writeJSON(w, http.StatusOK, group)
}

// HandleUpdateGroup updates an instance group's name and/or mode.
func (h *InstanceGroupsHandler) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeJSON[struct {
		Name string `json:"name"`
		Mode string `json:"mode"`
	}](w, r)
	if !ok {
		return
	}

	if req.Name != "" {
		if err := h.DB.UpdateInstanceGroup(r.Context(), id, req.Name); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update instance group"))
			return
		}
	}
	if req.Mode != "" {
		validModes := map[string]bool{"quality_hierarchy": true, "overlap_detect": true, "split_season": true}
		if !validModes[req.Mode] {
			writeJSON(w, http.StatusBadRequest, errorResponse("mode must be quality_hierarchy, overlap_detect, or split_season"))
			return
		}
		if err := h.DB.UpdateInstanceGroupMode(r.Context(), id, req.Mode); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update group mode"))
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteGroup deletes an instance group and its member associations.
func (h *InstanceGroupsHandler) HandleDeleteGroup(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
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
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	req, ok := decodeJSON[struct {
		Members []struct {
			InstanceID    uuid.UUID `json:"instance_id"`
			QualityRank   int       `json:"quality_rank"`
			IsIndependent bool      `json:"is_independent"`
		} `json:"members"`
	}](w, r)
	if !ok {
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
			GroupID:       id,
			InstanceID:    m.InstanceID,
			QualityRank:   m.QualityRank,
			IsIndependent: m.IsIndependent,
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

// HandleListOverlaps returns all detected media overlaps for a group.
func (h *InstanceGroupsHandler) HandleListOverlaps(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}

	media, err := h.DB.ListCrossInstanceMedia(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list overlaps"))
		return
	}
	if media == nil {
		media = []database.CrossInstanceMedia{}
	}
	writeJSON(w, http.StatusOK, media)
}

// HandleListActions returns the recent cross-instance routing actions.
func (h *InstanceGroupsHandler) HandleListActions(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}
	actions, err := h.DB.ListCrossInstanceActions(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list actions"))
		return
	}
	if actions == nil {
		actions = []database.CrossInstanceAction{}
	}
	writeJSON(w, http.StatusOK, actions)
}

// HandleListSeasonRules returns split-season rules for a group.
func (h *InstanceGroupsHandler) HandleListSeasonRules(w http.ResponseWriter, r *http.Request) {
	id, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	rules, err := h.DB.ListSplitSeasonRules(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list season rules"))
		return
	}
	if rules == nil {
		rules = []database.SplitSeasonRule{}
	}
	writeJSON(w, http.StatusOK, rules)
}

// HandleCreateSeasonRule creates a split-season rule for a group.
func (h *InstanceGroupsHandler) HandleCreateSeasonRule(w http.ResponseWriter, r *http.Request) {
	groupID, ok := parseUUID(w, r, "id")
	if !ok {
		return
	}
	req, ok := decodeJSON[struct {
		ExternalID string    `json:"external_id"`
		Title      string    `json:"title"`
		InstanceID uuid.UUID `json:"instance_id"`
		SeasonFrom int       `json:"season_from"`
		SeasonTo   *int      `json:"season_to"`
	}](w, r)
	if !ok {
		return
	}
	if req.ExternalID == "" || req.InstanceID == uuid.Nil || req.SeasonFrom < 1 {
		writeJSON(w, http.StatusBadRequest, errorResponse("external_id, instance_id, and season_from (>= 1) are required"))
		return
	}
	if req.SeasonTo != nil && *req.SeasonTo < req.SeasonFrom {
		writeJSON(w, http.StatusBadRequest, errorResponse("season_to must be >= season_from"))
		return
	}

	rule, err := h.DB.CreateSplitSeasonRule(r.Context(), database.SplitSeasonRule{
		GroupID:    groupID,
		ExternalID: req.ExternalID,
		Title:      req.Title,
		InstanceID: req.InstanceID,
		SeasonFrom: req.SeasonFrom,
		SeasonTo:   req.SeasonTo,
	})
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create season rule"))
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

// HandleDeleteSeasonRule deletes a split-season rule.
func (h *InstanceGroupsHandler) HandleDeleteSeasonRule(w http.ResponseWriter, r *http.Request) {
	ruleID, ok := parseUUID(w, r, "ruleId")
	if !ok {
		return
	}
	if err := h.DB.DeleteSplitSeasonRule(r.Context(), ruleID); err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("season rule not found"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
