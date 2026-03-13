package api

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/auth"
)

// SessionHandler handles session management endpoints.
type SessionHandler struct {
	DB Store
}

// HandleListSessions handles GET /api/sessions.
func (h *SessionHandler) HandleListSessions(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	sessions, err := h.DB.ListUserSessions(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list sessions"))
		return
	}

	// Identify current session
	currentSessionID := ""
	if cookie, err := r.Cookie("lurkarr_session"); err == nil {
		currentSessionID = cookie.Value
	}

	type sessionResponse struct {
		ID        uuid.UUID `json:"id"`
		CreatedAt string    `json:"created_at"`
		ExpiresAt string    `json:"expires_at"`
		IPAddress string    `json:"ip_address"`
		UserAgent string    `json:"user_agent"`
		Current   bool      `json:"current"`
	}

	result := make([]sessionResponse, len(sessions))
	for i, s := range sessions {
		result[i] = sessionResponse{
			ID:        s.ID,
			CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
			ExpiresAt: s.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			IPAddress: s.IPAddress,
			UserAgent: s.UserAgent,
			Current:   s.ID.String() == currentSessionID,
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// HandleRevokeSession handles DELETE /api/sessions/{id}.
func (h *SessionHandler) HandleRevokeSession(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	sessionID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid session ID"))
		return
	}

	// Verify the session belongs to this user
	sessions, err := h.DB.ListUserSessions(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to verify session"))
		return
	}

	found := false
	for _, s := range sessions {
		if s.ID == sessionID {
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusNotFound, errorResponse("session not found"))
		return
	}

	if err := h.DB.DeleteSession(r.Context(), sessionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to revoke session"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleRevokeAllSessions handles DELETE /api/sessions.
func (h *SessionHandler) HandleRevokeAllSessions(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	// Get current session to preserve it
	currentSessionID := uuid.Nil
	if cookie, err := r.Cookie("lurkarr_session"); err == nil {
		if id, err := uuid.Parse(cookie.Value); err == nil {
			currentSessionID = id
		}
	}

	// Delete all sessions except the current one
	if err := h.DB.DeleteUserSessionsExcept(r.Context(), user.ID, currentSessionID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to revoke sessions"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
