package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/gorilla/csrf"
	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
)

// webAuthnUser adapts database.User + credentials for the webauthn library.
type webAuthnUser struct {
	user  *database.User
	creds []webauthn.Credential
}

func (u *webAuthnUser) WebAuthnID() []byte                         { return u.user.ID[:] }
func (u *webAuthnUser) WebAuthnName() string                       { return u.user.Username }
func (u *webAuthnUser) WebAuthnDisplayName() string                { return u.user.Username }
func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.creds }

// dbCredToWebAuthn converts a database credential to the webauthn library type.
func dbCredToWebAuthn(c database.WebAuthnCredential) webauthn.Credential {
	transport := make([]protocol.AuthenticatorTransport, len(c.Transport))
	for i, t := range c.Transport {
		transport[i] = protocol.AuthenticatorTransport(t)
	}
	var aaguid [16]byte
	if len(c.AAGUID) == 16 {
		copy(aaguid[:], c.AAGUID)
	}
	return webauthn.Credential{
		ID:              c.CredentialID,
		PublicKey:       c.PublicKey,
		AttestationType: c.AttestationType,
		Transport:       transport,
		Authenticator: webauthn.Authenticator{
			AAGUID:    aaguid[:],
			SignCount: uint32(c.SignCount),
		},
	}
}

// PasskeyHandler handles WebAuthn/passkey endpoints.
type PasskeyHandler struct {
	DB       Store
	Auth     *auth.Middleware
	WebAuthn *webauthn.WebAuthn

	// In-memory session storage (challenge → session data).
	// Entries expire after 5 minutes.
	mu       sync.Mutex
	sessions map[string]*sessionEntry
}

type sessionEntry struct {
	data      *webauthn.SessionData
	expiresAt time.Time
}

// NewPasskeyHandler creates a PasskeyHandler with the given WebAuthn configuration.
func NewPasskeyHandler(db Store, authMw *auth.Middleware, wa *webauthn.WebAuthn) *PasskeyHandler {
	h := &PasskeyHandler{
		DB:       db,
		Auth:     authMw,
		WebAuthn: wa,
		sessions: make(map[string]*sessionEntry),
	}
	// Background cleanup of expired challenge sessions.
	go func() {
		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		for range ticker.C {
			h.mu.Lock()
			now := time.Now()
			for k, v := range h.sessions {
				if now.After(v.expiresAt) {
					delete(h.sessions, k)
				}
			}
			h.mu.Unlock()
		}
	}()
	return h
}

func (h *PasskeyHandler) storeSession(key string, s *webauthn.SessionData) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.sessions[key] = &sessionEntry{data: s, expiresAt: time.Now().Add(5 * time.Minute)}
}

func (h *PasskeyHandler) popSession(key string) *webauthn.SessionData {
	h.mu.Lock()
	defer h.mu.Unlock()
	e, ok := h.sessions[key]
	if !ok || time.Now().After(e.expiresAt) {
		delete(h.sessions, key)
		return nil
	}
	delete(h.sessions, key)
	return e.data
}

// HandleBeginRegistration starts passkey registration for the authenticated user.
// POST /api/passkeys/register/begin
func (h *PasskeyHandler) HandleBeginRegistration(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	// Build webauthn user with existing credentials to exclude.
	dbCreds, _ := h.DB.ListWebAuthnCredentials(r.Context(), user.ID)
	waCreds := make([]webauthn.Credential, len(dbCreds))
	for i, c := range dbCreds {
		waCreds[i] = dbCredToWebAuthn(c)
	}
	waUser := &webAuthnUser{user: user, creds: waCreds}

	// Exclude existing credentials and require resident key for discoverable login.
	excludeList := make([]protocol.CredentialDescriptor, len(waCreds))
	for i, c := range waCreds {
		excludeList[i] = protocol.CredentialDescriptor{
			Type:            protocol.PublicKeyCredentialType,
			CredentialID:    c.ID,
			Transport:       c.Transport,
		}
	}

	creation, session, err := h.WebAuthn.BeginRegistration(waUser,
		webauthn.WithExclusions(excludeList),
		webauthn.WithResidentKeyRequirement(protocol.ResidentKeyRequirementRequired),
		webauthn.WithConveyancePreference(protocol.PreferNoAttestation),
	)
	if err != nil {
		slog.Error("passkey begin registration failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to start passkey registration"))
		return
	}

	h.storeSession("reg:"+user.ID.String(), session)
	writeJSON(w, http.StatusOK, creation)
}

// HandleFinishRegistration completes passkey registration.
// POST /api/passkeys/register/finish
func (h *PasskeyHandler) HandleFinishRegistration(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	session := h.popSession("reg:" + user.ID.String())
	if session == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("no pending registration"))
		return
	}

	dbCreds, _ := h.DB.ListWebAuthnCredentials(r.Context(), user.ID)
	waCreds := make([]webauthn.Credential, len(dbCreds))
	for i, c := range dbCreds {
		waCreds[i] = dbCredToWebAuthn(c)
	}
	waUser := &webAuthnUser{user: user, creds: waCreds}

	cred, err := h.WebAuthn.FinishRegistration(waUser, *session, r)
	if err != nil {
		slog.Error("passkey finish registration failed", "error", err)
		writeJSON(w, http.StatusBadRequest, errorResponse("passkey registration failed"))
		return
	}

	// Read optional name from query.
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "Passkey"
	}
	if len(name) > 64 {
		name = name[:64]
	}

	transport := make([]string, len(cred.Transport))
	for i, t := range cred.Transport {
		transport[i] = string(t)
	}

	dbCred := &database.WebAuthnCredential{
		UserID:          user.ID,
		Name:            name,
		CredentialID:    cred.ID,
		PublicKey:       cred.PublicKey,
		AttestationType: cred.AttestationType,
		Transport:       transport,
		AAGUID:          cred.Authenticator.AAGUID,
		SignCount:       int64(cred.Authenticator.SignCount),
	}

	if err := h.DB.CreateWebAuthnCredential(r.Context(), dbCred); err != nil {
		slog.Error("passkey save failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to save passkey"))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
}

// HandleList returns the user's registered passkeys.
// GET /api/passkeys
func (h *PasskeyHandler) HandleList(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	creds, err := h.DB.ListWebAuthnCredentials(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to list passkeys"))
		return
	}
	if creds == nil {
		creds = []database.WebAuthnCredential{}
	}

	writeJSON(w, http.StatusOK, creds)
}

// HandleDelete deletes a passkey.
// DELETE /api/passkeys/{id}
func (h *PasskeyHandler) HandleDelete(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	// Verify the passkey belongs to this user.
	creds, err := h.DB.ListWebAuthnCredentials(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("database error"))
		return
	}
	found := false
	for _, c := range creds {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusNotFound, errorResponse("passkey not found"))
		return
	}

	if err := h.DB.DeleteWebAuthnCredential(r.Context(), id); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to delete passkey"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleRename renames a passkey.
// POST /api/passkeys/{id}/rename
func (h *PasskeyHandler) HandleRename(w http.ResponseWriter, r *http.Request) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		writeJSON(w, http.StatusUnauthorized, errorResponse("unauthorized"))
		return
	}

	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid id"))
		return
	}

	limitBody(w, r)
	var req struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
		writeJSON(w, http.StatusBadRequest, errorResponse("name required"))
		return
	}
	if len(req.Name) > 64 {
		req.Name = req.Name[:64]
	}

	// Verify ownership.
	creds, err := h.DB.ListWebAuthnCredentials(r.Context(), user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("database error"))
		return
	}
	found := false
	for _, c := range creds {
		if c.ID == id {
			found = true
			break
		}
	}
	if !found {
		writeJSON(w, http.StatusNotFound, errorResponse("passkey not found"))
		return
	}

	if err := h.DB.RenameWebAuthnCredential(r.Context(), id, req.Name); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to rename passkey"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

// HandleBeginLogin starts passkey (discoverable) login.
// POST /api/auth/passkey/login/begin
func (h *PasskeyHandler) HandleBeginLogin(w http.ResponseWriter, r *http.Request) {
	assertion, session, err := h.WebAuthn.BeginDiscoverableLogin()
	if err != nil {
		slog.Error("passkey begin login failed", "error", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to start passkey login"))
		return
	}

	h.storeSession("login:"+session.Challenge, session)
	writeJSON(w, http.StatusOK, assertion)
}

// HandleFinishLogin completes passkey login and creates a session.
// POST /api/auth/passkey/login/finish
func (h *PasskeyHandler) HandleFinishLogin(w http.ResponseWriter, r *http.Request) {
	// Parse the credential assertion to get the challenge.
	parsedResponse, err := protocol.ParseCredentialRequestResponse(r)
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid passkey response"))
		return
	}

	challengeStr := parsedResponse.Response.CollectedClientData.Challenge
	session := h.popSession("login:" + challengeStr)
	if session == nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("no pending login"))
		return
	}

	// Discoverable user handler: look up user by the credential's user handle.
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		uid, err := uuid.FromBytes(userHandle)
		if err != nil {
			return nil, err
		}
		user, err := h.DB.GetUserByID(r.Context(), uid)
		if err != nil {
			return nil, err
		}
		dbCreds, err := h.DB.ListWebAuthnCredentials(r.Context(), user.ID)
		if err != nil {
			return nil, err
		}
		waCreds := make([]webauthn.Credential, len(dbCreds))
		for i, c := range dbCreds {
			waCreds[i] = dbCredToWebAuthn(c)
		}
		return &webAuthnUser{user: user, creds: waCreds}, nil
	}

	waUser, cred, err := h.WebAuthn.ValidatePasskeyLogin(handler, *session, parsedResponse)
	if err != nil {
		slog.Error("passkey login validation failed", "error", err)
		writeJSON(w, http.StatusUnauthorized, errorResponse("passkey authentication failed"))
		return
	}

	// Update sign count.
	_ = h.DB.UpdateWebAuthnSignCount(r.Context(), cred.ID, int64(cred.Authenticator.SignCount))

	// Get the actual database user from the webauthn user.
	dbUser := waUser.(*webAuthnUser).user

	// Session rotation: invalidate any existing session.
	if cookie, cookieErr := r.Cookie("lurkarr_session"); cookieErr == nil {
		if oldID, parseErr := uuid.Parse(cookie.Value); parseErr == nil {
			_ = h.DB.DeleteSession(r.Context(), oldID)
		}
	}

	if err := h.Auth.SetSessionCookie(r.Context(), w, r, dbUser.ID); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse("failed to create session"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"user":       dbUser,
		"csrf_token": csrf.Token(r),
	})
}
