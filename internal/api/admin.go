package api

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/auth"
)

// AdminHandler handles admin-only user management endpoints.
type AdminHandler struct {
	DB Store
}

// HandleListUsers handles GET /api/admin/users.
func (h *AdminHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		writeJSON(w, http.StatusForbidden, errorResponse("admin access required"))
		return
	}

	users, err := h.DB.ListUsers(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list users"))
		return
	}

	type userResponse struct {
		ID           uuid.UUID `json:"id"`
		Username     string    `json:"username"`
		AuthProvider string    `json:"auth_provider"`
		IsAdmin      bool      `json:"is_admin"`
		Has2FA       bool      `json:"has_2fa"`
		CreatedAt    string    `json:"created_at"`
	}

	result := make([]userResponse, len(users))
	for i, u := range users {
		result[i] = userResponse{
			ID:           u.ID,
			Username:     u.Username,
			AuthProvider: u.AuthProvider,
			IsAdmin:      u.IsAdmin,
			Has2FA:       u.TOTPSecret != nil && *u.TOTPSecret != "",
			CreatedAt:    u.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	writeJSON(w, http.StatusOK, result)
}

// HandleCreateUser handles POST /api/admin/users.
func (h *AdminHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		writeJSON(w, http.StatusForbidden, errorResponse("admin access required"))
		return
	}

	limitBody(w, r)
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		IsAdmin  bool   `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("username and password required"))
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to hash password"))
		return
	}

	newUser, err := h.DB.CreateUser(r.Context(), req.Username, hash)
	if err != nil {
		writeJSON(w, http.StatusConflict, errorResponse("username already exists"))
		return
	}

	if req.IsAdmin {
		if err := h.DB.UpdateUserAdmin(r.Context(), newUser.ID, true); err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("failed to set admin flag"))
			return
		}
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":       newUser.ID,
		"username": newUser.Username,
		"is_admin": req.IsAdmin,
	})
}

// HandleDeleteUser handles DELETE /api/admin/users/{id}.
func (h *AdminHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		writeJSON(w, http.StatusForbidden, errorResponse("admin access required"))
		return
	}

	targetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid user ID"))
		return
	}

	// Prevent self-deletion
	if targetID == user.ID {
		writeJSON(w, http.StatusBadRequest, errorResponse("cannot delete yourself"))
		return
	}

	if err := h.DB.DeleteUser(r.Context(), targetID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete user"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleResetUserPassword handles POST /api/admin/users/{id}/reset-password.
func (h *AdminHandler) HandleResetUserPassword(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		writeJSON(w, http.StatusForbidden, errorResponse("admin access required"))
		return
	}

	targetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid user ID"))
		return
	}

	limitBody(w, r)
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("password required"))
		return
	}
	if err := auth.ValidatePassword(req.Password); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to hash password"))
		return
	}

	if err := h.DB.UpdatePassword(r.Context(), targetID, hash); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to reset password"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleToggleAdmin handles POST /api/admin/users/{id}/toggle-admin.
func (h *AdminHandler) HandleToggleAdmin(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil || !user.IsAdmin {
		writeJSON(w, http.StatusForbidden, errorResponse("admin access required"))
		return
	}

	targetID, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid user ID"))
		return
	}

	// Prevent self-demotion
	if targetID == user.ID {
		writeJSON(w, http.StatusBadRequest, errorResponse("cannot change your own admin status"))
		return
	}

	limitBody(w, r)
	var req struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if err := h.DB.UpdateUserAdmin(r.Context(), targetID, req.IsAdmin); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to update admin status"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
