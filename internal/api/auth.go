package api

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/csrf"
	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
)

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	DB   *database.DB
	Auth *auth.Middleware
}

// HandleLogin handles POST /api/auth/login.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	user, err := h.DB.GetUserByUsername(r.Context(), req.Username)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("invalid credentials"))
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		writeJSON(w, http.StatusUnauthorized, errorResponse("invalid credentials"))
		return
	}

	// Check TOTP if enabled
	if user.TOTPSecret != nil && *user.TOTPSecret != "" {
		if req.TOTPCode == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error":         "totp_required",
				"totp_required": true,
			})
			return
		}
		if !auth.ValidateTOTP(req.TOTPCode, *user.TOTPSecret) {
			writeJSON(w, http.StatusUnauthorized, errorResponse("invalid TOTP code"))
			return
		}
	}

	if err := h.Auth.SetSessionCookie(r.Context(), w, user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create session"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":       user,
		"csrf_token": csrf.Token(r),
	})
}

// HandleLogout handles POST /api/auth/logout.
func (h *AuthHandler) HandleLogout(w http.ResponseWriter, r *http.Request) {
	h.Auth.ClearSessionCookie(r.Context(), w, r)
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleSetup handles POST /api/auth/setup (first-run).
func (h *AuthHandler) HandleSetup(w http.ResponseWriter, r *http.Request) {
	count, err := h.DB.UserCount(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("database error"))
		return
	}
	if count > 0 {
		writeJSON(w, http.StatusConflict, errorResponse("setup already completed"))
		return
	}

	var req setupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}
	if req.Username == "" || req.Password == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("username and password required"))
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to hash password"))
		return
	}

	user, err := h.DB.CreateUser(r.Context(), req.Username, hash)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create user"))
		return
	}

	// Generate and store secret key
	secretKey, err := auth.GenerateSecretKey(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to generate secret key"))
		return
	}

	settings := &database.GeneralSettings{
		SecretKey:            secretKey,
		SSLVerify:            true,
		APITimeout:           120,
		StatefulResetHours:   168,
		CommandWaitDelay:     1,
		CommandWaitAttempts:  600,
		MinDownloadQueueSize: -1,
	}
	if err := h.DB.UpsertGeneralSettings(r.Context(), settings); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save settings"))
		return
	}

	if err := h.Auth.SetSessionCookie(r.Context(), w, user.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create session"))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{"user": user})
}

// Handle2FAEnable handles POST /api/auth/2fa/enable.
func (h *AuthHandler) Handle2FAEnable(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	secret, qrBase64, err := auth.GenerateTOTP(user.Username, "Lurkarr")
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to generate TOTP"))
		return
	}

	// Store secret temporarily — will be confirmed on verify
	if err := h.DB.SetTOTPSecret(r.Context(), user.ID, &secret); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save TOTP secret"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"secret":    secret,
		"qr_base64": qrBase64,
	})
}

// Handle2FADisable handles POST /api/auth/2fa/disable.
func (h *AuthHandler) Handle2FADisable(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	if err := h.DB.SetTOTPSecret(r.Context(), user.ID, nil); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to disable 2FA"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// Handle2FAVerify handles POST /api/auth/2fa/verify.
func (h *AuthHandler) Handle2FAVerify(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return
	}

	if user.TOTPSecret == nil || *user.TOTPSecret == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("2FA not enabled"))
		return
	}

	if !auth.ValidateTOTP(req.Code, *user.TOTPSecret) {
		writeJSON(w, http.StatusUnauthorized, errorResponse("invalid TOTP code"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "verified"})
}
