package api

import (
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/lusoris/lurkarr/internal/database"
)

// ActivityEvent is a normalised event for the unified activity feed.
type ActivityEvent struct {
	ID        string    `json:"id"`
	Source    string    `json:"source"`
	AppType   string    `json:"app_type,omitempty"`
	Title     string    `json:"title"`
	Action    string    `json:"action"`
	Detail    string    `json:"detail,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// ActivityHandler serves the unified activity feed.
type ActivityHandler struct {
	DB Store
}

// HandleGetActivity handles GET /api/activity?limit=50.
// Merges lurk history, blocklist log, auto-import log, cross-instance actions,
// and schedule executions into a single chronological feed.
func (h *ActivityHandler) HandleGetActivity(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	// We over-fetch from each source so the merge has enough data.
	perSource := limit

	var events []ActivityEvent

	// 1. Lurk history (all app types).
	if items, _, err := h.DB.ListLurkHistory(r.Context(), database.HistoryQuery{Limit: perSource}); err == nil {
		for _, it := range items {
			events = append(events, ActivityEvent{
				ID:        fmt.Sprintf("lurk-%d", it.ID),
				Source:    "lurk",
				AppType:   string(it.AppType),
				Title:     it.MediaTitle,
				Action:    it.Operation,
				Detail:    it.InstanceName,
				Timestamp: it.CreatedAt,
			})
		}
	}

	// 2. Cross-instance actions.
	if items, err := h.DB.ListCrossInstanceActions(r.Context(), perSource); err == nil {
		for _, it := range items {
			events = append(events, ActivityEvent{
				ID:        fmt.Sprintf("xarr-%s", it.ID),
				Source:    "cross_instance",
				Title:     it.Title,
				Action:    it.Action,
				Detail:    it.Reason,
				Timestamp: it.ExecutedAt,
			})
		}
	}

	// 3. Blocklist log (all app types).
	for _, app := range database.AllAppTypes() {
		if items, err := h.DB.GetBlocklistLog(r.Context(), app, perSource); err == nil {
			for _, it := range items {
				events = append(events, ActivityEvent{
					ID:        fmt.Sprintf("block-%d", it.ID),
					Source:    "blocklist",
					AppType:   string(it.AppType),
					Title:     it.Title,
					Action:    "blocklisted",
					Detail:    it.Reason,
					Timestamp: it.BlocklistedAt,
				})
			}
		}
	}

	// 4. Auto-import log (all app types).
	for _, app := range database.AllAppTypes() {
		if items, err := h.DB.GetAutoImportLog(r.Context(), app, perSource); err == nil {
			for _, it := range items {
				events = append(events, ActivityEvent{
					ID:        fmt.Sprintf("import-%d", it.ID),
					Source:    "auto_import",
					AppType:   string(it.AppType),
					Title:     it.MediaTitle,
					Action:    it.Action,
					Detail:    it.Reason,
					Timestamp: it.CreatedAt,
				})
			}
		}
	}

	// 5. Schedule executions.
	if items, err := h.DB.ListScheduleExecutions(r.Context(), perSource); err == nil {
		for _, it := range items {
			result := ""
			if it.Result != nil {
				result = *it.Result
			}
			events = append(events, ActivityEvent{
				ID:        fmt.Sprintf("sched-%d", it.ID),
				Source:    "schedule",
				Title:     fmt.Sprintf("Schedule %s", it.ScheduleID),
				Action:    "executed",
				Detail:    result,
				Timestamp: it.ExecutedAt,
			})
		}
	}

	// 6. Strike log (all app types).
	for _, app := range database.AllAppTypes() {
		if items, err := h.DB.GetStrikeLog(r.Context(), app, perSource); err == nil {
			for _, it := range items {
				events = append(events, ActivityEvent{
					ID:        fmt.Sprintf("strike-%d", it.ID),
					Source:    "strike",
					AppType:   string(it.AppType),
					Title:     it.Title,
					Action:    "strike",
					Detail:    it.Reason,
					Timestamp: it.StruckAt,
				})
			}
		}
	}

	// Sort all events by timestamp descending, then truncate.
	sort.Slice(events, func(i, j int) bool {
		return events[i].Timestamp.After(events[j].Timestamp)
	})
	if len(events) > limit {
		events = events[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"items": events,
		"total": len(events),
	})
}
