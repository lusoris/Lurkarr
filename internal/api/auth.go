package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
)

type loginRequest struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	TOTPCode     string `json:"totp_code,omitempty"`
	RecoveryCode string `json:"recovery_code,omitempty"`
}

type setupRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

const (
	maxFailedLogins = 5
	lockoutDuration = 15 * time.Minute
)

// AuthHandler handles authentication endpoints.
type AuthHandler struct {
	DB   Store
	Auth *auth.Middleware
}

// HandleLogin handles POST /api/auth/login.
func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	limitBody(w, r)
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

	// Check account lockout
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		writeJSON(w, http.StatusTooManyRequests, errorResponse("account temporarily locked"))
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		if err := h.DB.IncrementFailedLogins(r.Context(), user.ID, maxFailedLogins, lockoutDuration); err != nil {
			slog.Error("failed to increment failed logins", "error", err, "user_id", user.ID)
		}
		writeJSON(w, http.StatusUnauthorized, errorResponse("invalid credentials"))
		return
	}

	// Check TOTP if enabled
	if user.TOTPSecret != nil && *user.TOTPSecret != "" {
		if req.TOTPCode == "" && req.RecoveryCode == "" {
			writeJSON(w, http.StatusUnauthorized, map[string]any{
				"error":         "totp_required",
				"totp_required": true,
			})
			return
		}
		if req.RecoveryCode != "" {
			// Try recovery code
			idx := auth.ValidateRecoveryCode(req.RecoveryCode, user.RecoveryCodes)
			if idx < 0 {
				writeJSON(w, http.StatusUnauthorized, errorResponse("invalid recovery code"))
				return
			}
			// Consume the used recovery code
			remaining := make([]string, 0, len(user.RecoveryCodes)-1)
			remaining = append(remaining, user.RecoveryCodes[:idx]...)
			remaining = append(remaining, user.RecoveryCodes[idx+1:]...)
			if err := h.DB.SetRecoveryCodes(r.Context(), user.ID, remaining); err != nil {
				slog.Error("failed to consume recovery code", "error", err, "user", user.ID)
				writeJSON(w, http.StatusInternalServerError, errorResponse("failed to consume recovery code"))
				return
			}
		} else if !auth.ValidateTOTP(req.TOTPCode, *user.TOTPSecret) {
			writeJSON(w, http.StatusUnauthorized, errorResponse("invalid TOTP code"))
			return
		}
	}

	// Reset failed login counter on successful authentication
	if user.FailedLoginAttempts > 0 {
		if err := h.DB.ResetFailedLogins(r.Context(), user.ID); err != nil {
			slog.Error("failed to reset failed logins", "error", err, "user_id", user.ID)
		}
	}

	// Session rotation: invalidate any existing session before creating a new one
	if cookie, err := r.Cookie("lurkarr_session"); err == nil {
		if oldID, err := uuid.Parse(cookie.Value); err == nil {
			if err := h.DB.DeleteSession(r.Context(), oldID); err != nil {
				slog.Error("failed to delete old session during rotation", "error", err, "session_id", oldID)
			}
		}
	}

	if err := h.Auth.SetSessionCookie(r.Context(), w, r, user.ID); err != nil {
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

// HandleSetupCheck handles GET /api/auth/setup — returns whether initial setup is needed.
func (h *AuthHandler) HandleSetupCheck(w http.ResponseWriter, r *http.Request) {
	count, err := h.DB.UserCount(r.Context())
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("database error"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"needs_setup": count == 0})
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

	limitBody(w, r)
	var req setupRequest
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

	// Generate secret key for initial settings.
	secretKey, err := auth.GenerateSecretKey(32)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to generate secret key"))
		return
	}

	settings := &database.GeneralSettings{
		SecretKey:                 secretKey,
		SSLVerify:                 true,
		APITimeout:                120,
		StatefulResetHours:        168,
		CommandWaitDelay:          1,
		CommandWaitAttempts:       600,
		MaxDownloadQueueSize:      0,
		AutoImportIntervalMinutes: 5,
	}

	// Atomically create the first user and initial settings in one transaction.
	user, err := h.DB.SetupFirstUser(r.Context(), req.Username, hash, settings)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to complete setup"))
		return
	}

	if err := h.Auth.SetSessionCookie(r.Context(), w, r, user.ID); err != nil {
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

	// Generate recovery codes first — if this fails, TOTP stays disabled
	plainCodes, hashedCodes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to generate recovery codes"))
		return
	}
	if err := h.DB.SetRecoveryCodes(r.Context(), user.ID, hashedCodes); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save recovery codes"))
		return
	}

	// Now activate TOTP — recovery codes are already persisted
	if err := h.DB.SetTOTPSecret(r.Context(), user.ID, &secret); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save TOTP secret"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"secret":         secret,
		"qr_base64":      qrBase64,
		"recovery_codes": plainCodes,
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

	// Clear recovery codes
	if err := h.DB.SetRecoveryCodes(r.Context(), user.ID, nil); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to clear recovery codes"))
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

	limitBody(w, r)
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

// HandleRegenerateRecoveryCodes handles POST /api/auth/2fa/recovery-codes.
func (h *AuthHandler) HandleRegenerateRecoveryCodes(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	if user.TOTPSecret == nil || *user.TOTPSecret == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("2FA not enabled"))
		return
	}

	plainCodes, hashedCodes, err := auth.GenerateRecoveryCodes()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to generate recovery codes"))
		return
	}
	if err := h.DB.SetRecoveryCodes(r.Context(), user.ID, hashedCodes); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save recovery codes"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"recovery_codes": plainCodes,
	})
}
