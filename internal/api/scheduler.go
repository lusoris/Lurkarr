package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/scheduler"
)

// SchedulerHandler handles scheduling endpoints.
type SchedulerHandler struct {
	DB        *database.DB
	Scheduler *scheduler.Scheduler
}

// HandleListSchedules handles GET /api/schedules.
func (h *SchedulerHandler) HandleListSchedules(w http.ResponseWriter, r *http.Request) {
	schedules, err := h.DB.ListSchedules(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list schedules"))
		return
	}

	writeJSON(w, http.StatusOK, schedules)
}

// HandleCreateSchedule handles POST /api/schedules.
func (h *SchedulerHandler) HandleCreateSchedule(w http.ResponseWriter, r *http.Request) {
	var sched database.Schedule
	if err := json.NewDecoder(r.Body).Decode(&sched); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if sched.AppType == "" || sched.Action == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("app_type and action required"))
		return
	}

	if err := h.DB.CreateSchedule(r.Context(), &sched); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create schedule"))
		return
	}

	// Reload scheduler
	_ = h.Scheduler.Reload(r.Context())

	writeJSON(w, http.StatusCreated, sched)
}

// HandleUpdateSchedule handles PUT /api/schedules/{id}.
func (h *SchedulerHandler) HandleUpdateSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid schedule ID"))
		return
	}

	var sched database.Schedule
	if err := json.NewDecoder(r.Body).Decode(&sched); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	sched.ID = id

	if err := h.DB.UpdateSchedule(r.Context(), &sched); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update schedule"))
		return
	}

	_ = h.Scheduler.Reload(r.Context())

	writeJSON(w, http.StatusOK, sched)
}

// HandleDeleteSchedule handles DELETE /api/schedules/{id}.
func (h *SchedulerHandler) HandleDeleteSchedule(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid schedule ID"))
		return
	}

	if err := h.DB.DeleteSchedule(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete schedule"))
		return
	}

	_ = h.Scheduler.Reload(r.Context())

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleScheduleHistory handles GET /api/schedules/history.
func (h *SchedulerHandler) HandleScheduleHistory(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 && n <= 200 {
			limit = n
		}
	}

	executions, err := h.DB.ListScheduleExecutions(r.Context(), limit)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to load schedule history"))
		return
	}

	writeJSON(w, http.StatusOK, executions)
}
