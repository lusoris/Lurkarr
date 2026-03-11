package api

import (
	"encoding/json"
	"net/http"

	"github.com/lusoris/lurkarr/internal/auth"
)

// UserHandler handles user profile endpoints.
type UserHandler struct {
	DB Store
}

// HandleGetUser handles GET /api/user.
func (h *UserHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":         user.ID,
		"username":   user.Username,
		"has_2fa":    user.TOTPSecret != nil && *user.TOTPSecret != "",
		"created_at": user.CreatedAt,
	})
}

// HandleUpdateUsername handles POST /api/user/username.
func (h *UserHandler) HandleUpdateUsername(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	limitBody(w, r)
	var req struct {
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Username == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("username required"))
		return
	}

	if err := h.DB.UpdateUsername(r.Context(), user.ID, req.Username); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update username"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleUpdatePassword handles POST /api/user/password.
func (h *UserHandler) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	limitBody(w, r)
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if !auth.CheckPassword(req.CurrentPassword, user.Password) {
		writeJSON(w, http.StatusUnauthorized, errorResponse("current password incorrect"))
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to hash password"))
		return
	}

	if err := h.DB.UpdatePassword(r.Context(), user.ID, hash); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update password"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
