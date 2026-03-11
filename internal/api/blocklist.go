package api

import (
	"encoding/json"
	"net/http"
	"regexp"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// BlocklistHandler handles blocklist source and rule endpoints.
type BlocklistHandler struct {
	DB Store
}

// --- Sources ---

// HandleListSources handles GET /api/blocklist/sources.
func (h *BlocklistHandler) HandleListSources(w http.ResponseWriter, r *http.Request) {
	sources, err := h.DB.ListBlocklistSources(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list blocklist sources"))
		return
	}
	if sources == nil {
		sources = []database.BlocklistSource{}
	}
	writeJSON(w, http.StatusOK, sources)
}

// HandleGetSource handles GET /api/blocklist/sources/{id}.
func (h *BlocklistHandler) HandleGetSource(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid source id"))
		return
	}

	source, err := h.DB.GetBlocklistSource(r.Context(), id)
	if err != nil {
		writeJSON(w, http.StatusNotFound, errorResponse("source not found"))
		return
	}

	writeJSON(w, http.StatusOK, source)
}

// HandleCreateSource handles POST /api/blocklist/sources.
func (h *BlocklistHandler) HandleCreateSource(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var s database.BlocklistSource
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if s.Name == "" || s.URL == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name and url are required"))
		return
	}

	if err := h.DB.CreateBlocklistSource(r.Context(), &s); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create source"))
		return
	}

	writeJSON(w, http.StatusCreated, s)
}

// HandleUpdateSource handles PUT /api/blocklist/sources/{id}.
func (h *BlocklistHandler) HandleUpdateSource(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid source id"))
		return
	}

	limitBody(w, r)
	var s database.BlocklistSource
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	s.ID = id

	if s.Name == "" || s.URL == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name and url are required"))
		return
	}

	if err := h.DB.UpdateBlocklistSource(r.Context(), &s); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update source"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleDeleteSource handles DELETE /api/blocklist/sources/{id}.
func (h *BlocklistHandler) HandleDeleteSource(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid source id"))
		return
	}

	// Delete associated rules first (FK cascade should handle this, but be explicit).
	if err := h.DB.DeleteBlocklistRulesBySource(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete source rules"))
		return
	}

	if err := h.DB.DeleteBlocklistSource(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete source"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// --- Rules ---

// HandleListRules handles GET /api/blocklist/rules.
func (h *BlocklistHandler) HandleListRules(w http.ResponseWriter, r *http.Request) {
	rules, err := h.DB.ListBlocklistRules(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list blocklist rules"))
		return
	}
	if rules == nil {
		rules = []database.BlocklistRule{}
	}
	writeJSON(w, http.StatusOK, rules)
}

var validPatternTypes = map[string]bool{
	"release_group":  true,
	"title_contains": true,
	"title_regex":    true,
	"indexer":        true,
}

// HandleCreateRule handles POST /api/blocklist/rules.
func (h *BlocklistHandler) HandleCreateRule(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
	var rule database.BlocklistRule
	if err := json.NewDecoder(r.Body).Decode(&rule); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if rule.Pattern == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("pattern is required"))
		return
	}
	if !validPatternTypes[rule.PatternType] {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid pattern_type: must be release_group, title_contains, title_regex, or indexer"))
		return
	}
	if rule.PatternType == "title_regex" {
		if _, err := regexp.Compile(rule.Pattern); err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("invalid regex pattern"))
			return
		}
	}

	if err := h.DB.CreateBlocklistRule(r.Context(), &rule); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create rule"))
		return
	}

	writeJSON(w, http.StatusCreated, rule)
}

// HandleDeleteRule handles DELETE /api/blocklist/rules/{id}.
func (h *BlocklistHandler) HandleDeleteRule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid rule id"))
		return
	}

	if err := h.DB.DeleteBlocklistRule(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete rule"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
