package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
	"github.com/lusoris/lurkarr/internal/mocks"
)

// =============================================================================
// HandleList tests
// =============================================================================

func TestPasskeyHandleList(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	user := &database.User{ID: userID, Username: "alice"}

	creds := []database.WebAuthnCredential{
		{ID: uuid.New(), UserID: userID, Name: "YubiKey"},
		{ID: uuid.New(), UserID: userID, Name: "TouchID"},
	}
	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return(creds, nil)

	r := httptest.NewRequest("GET", "/api/passkeys", http.NoBody)
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleList(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var result []database.WebAuthnCredential
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	if result[0].Name != "YubiKey" {
		t.Errorf("name = %q, want %q", result[0].Name, "YubiKey")
	}
}

func TestPasskeyHandleList_EmptyReturnsArray(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	user := &database.User{ID: userID}

	// Return nil to exercise the nil→empty slice guard
	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return(nil, nil)

	r := httptest.NewRequest("GET", "/api/passkeys", http.NoBody)
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleList(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	// Should return [] not null
	var result []database.WebAuthnCredential
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result == nil {
		t.Error("expected empty array, got null")
	}
}

func TestPasskeyHandleList_NoUser(t *testing.T) {
	h := &PasskeyHandler{}
	r := httptest.NewRequest("GET", "/api/passkeys", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleList(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPasskeyHandleList_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	user := &database.User{ID: userID}
	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return(nil, errors.New("db fail"))

	r := httptest.NewRequest("GET", "/api/passkeys", http.NoBody)
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleList(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// =============================================================================
// HandleDelete tests
// =============================================================================

func TestPasskeyHandleDelete(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{
		{ID: credID, UserID: userID, Name: "YubiKey"},
	}, nil)
	store.EXPECT().DeleteWebAuthnCredential(gomock.Any(), credID).Return(nil)

	r := reqWithPathValue("DELETE", "/api/passkeys/"+credID.String(), nil, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleDelete(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestPasskeyHandleDelete_NoUser(t *testing.T) {
	h := &PasskeyHandler{}
	r := reqWithPathValue("DELETE", "/api/passkeys/abc", nil, "id", uuid.New().String())
	w := httptest.NewRecorder()
	h.HandleDelete(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPasskeyHandleDelete_InvalidID(t *testing.T) {
	ctrl := gomock.NewController(t)
	_ = NewMockStore(ctrl)
	h := &PasskeyHandler{}

	userID := uuid.New()
	user := &database.User{ID: userID}

	r := reqWithPathValue("DELETE", "/api/passkeys/not-a-uuid", nil, "id", "not-a-uuid")
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleDelete(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPasskeyHandleDelete_NotOwned(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	// Return empty list — credential doesn't belong to this user
	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{}, nil)

	r := reqWithPathValue("DELETE", "/api/passkeys/"+credID.String(), nil, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleDelete(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPasskeyHandleDelete_ListError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}
	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return(nil, errors.New("db fail"))

	r := reqWithPathValue("DELETE", "/api/passkeys/"+credID.String(), nil, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleDelete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestPasskeyHandleDelete_DeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{
		{ID: credID, UserID: userID},
	}, nil)
	store.EXPECT().DeleteWebAuthnCredential(gomock.Any(), credID).Return(errors.New("db fail"))

	r := reqWithPathValue("DELETE", "/api/passkeys/"+credID.String(), nil, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleDelete(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

// =============================================================================
// HandleRename tests
// =============================================================================

func TestPasskeyHandleRename(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{
		{ID: credID, UserID: userID, Name: "Old Name"},
	}, nil)
	store.EXPECT().RenameWebAuthnCredential(gomock.Any(), credID, "New Key").Return(nil)

	body, _ := json.Marshal(map[string]string{"name": "New Key"})
	r := reqWithPathValue("POST", "/api/passkeys/"+credID.String()+"/rename", body, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestPasskeyHandleRename_NoUser(t *testing.T) {
	h := &PasskeyHandler{}
	r := reqWithPathValue("POST", "/api/passkeys/abc/rename", nil, "id", uuid.New().String())
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPasskeyHandleRename_InvalidID(t *testing.T) {
	h := &PasskeyHandler{}
	userID := uuid.New()
	user := &database.User{ID: userID}

	r := reqWithPathValue("POST", "/api/passkeys/bad/rename", nil, "id", "bad")
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPasskeyHandleRename_EmptyName(t *testing.T) {
	h := &PasskeyHandler{}
	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	body, _ := json.Marshal(map[string]string{"name": ""})
	r := reqWithPathValue("POST", "/api/passkeys/"+credID.String()+"/rename", body, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestPasskeyHandleRename_NotOwned(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{}, nil)

	body, _ := json.Marshal(map[string]string{"name": "New Name"})
	r := reqWithPathValue("POST", "/api/passkeys/"+credID.String()+"/rename", body, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

func TestPasskeyHandleRename_RenameError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{
		{ID: credID, UserID: userID},
	}, nil)
	store.EXPECT().RenameWebAuthnCredential(gomock.Any(), credID, "New Name").Return(errors.New("db fail"))

	body, _ := json.Marshal(map[string]string{"name": "New Name"})
	r := reqWithPathValue("POST", "/api/passkeys/"+credID.String()+"/rename", body, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestPasskeyHandleRename_TruncatesLongName(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	h := &PasskeyHandler{DB: store}

	userID := uuid.New()
	credID := uuid.New()
	user := &database.User{ID: userID}

	// Name longer than 64 characters
	longName := ""
	for i := 0; i < 70; i++ {
		longName += "x"
	}
	truncated := longName[:64]

	store.EXPECT().ListWebAuthnCredentials(gomock.Any(), userID).Return([]database.WebAuthnCredential{
		{ID: credID, UserID: userID},
	}, nil)
	store.EXPECT().RenameWebAuthnCredential(gomock.Any(), credID, truncated).Return(nil)

	body, _ := json.Marshal(map[string]string{"name": longName})
	r := reqWithPathValue("POST", "/api/passkeys/"+credID.String()+"/rename", body, "id", credID.String())
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleRename(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

// =============================================================================
// HandleBeginRegistration tests (auth guards only — requires real WebAuthn for full test)
// =============================================================================

func TestPasskeyHandleBeginRegistration_NoUser(t *testing.T) {
	h := &PasskeyHandler{}
	r := httptest.NewRequest("POST", "/api/passkeys/register/begin", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleBeginRegistration(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPasskeyHandleFinishRegistration_NoUser(t *testing.T) {
	h := &PasskeyHandler{}
	r := httptest.NewRequest("POST", "/api/passkeys/register/finish", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleFinishRegistration(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestPasskeyHandleFinishRegistration_NoPendingSession(t *testing.T) {
	ctrl := gomock.NewController(t)
	_ = NewMockStore(ctrl)
	h := &PasskeyHandler{
		sessions: make(map[string]*sessionEntry),
	}

	userID := uuid.New()
	user := &database.User{ID: userID}

	r := httptest.NewRequest("POST", "/api/passkeys/register/finish", bytes.NewReader([]byte("{}")))
	r = r.WithContext(auth.ContextWithUser(r.Context(), user))
	w := httptest.NewRecorder()
	h.HandleFinishRegistration(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "no pending registration" {
		t.Errorf("error = %q, want %q", resp["error"], "no pending registration")
	}
}

// =============================================================================
// HandleBeginLogin tests
// =============================================================================

func TestPasskeyHandleBeginLogin_NilWebAuthn(t *testing.T) {
	// WebAuthn is nil → should panic (no nil check in handler)
	h := &PasskeyHandler{
		sessions: make(map[string]*sessionEntry),
	}
	r := httptest.NewRequest("POST", "/api/auth/passkey/login/begin", http.NoBody)
	w := httptest.NewRecorder()

	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic with nil WebAuthn, got none")
		}
	}()
	h.HandleBeginLogin(w, r)
}

// =============================================================================
// Session store/pop tests
// =============================================================================

func TestPasskeySessionStoreAndPop(t *testing.T) {
	h := &PasskeyHandler{
		sessions: make(map[string]*sessionEntry),
	}

	sd := &webauthn.SessionData{Challenge: "test-challenge"}

	// Store and retrieve
	h.storeSession("test-key", sd)
	got := h.popSession("test-key")
	if got == nil {
		t.Fatal("expected session, got nil")
	}
	if got.Challenge != "test-challenge" {
		t.Errorf("challenge = %q, want %q", got.Challenge, "test-challenge")
	}

	// Pop again should return nil (already consumed)
	got = h.popSession("test-key")
	if got != nil {
		t.Fatal("expected nil, got session (should be consumed)")
	}
}

func TestPasskeySessionPopNonExistent(t *testing.T) {
	h := &PasskeyHandler{
		sessions: make(map[string]*sessionEntry),
	}

	got := h.popSession("nonexistent")
	if got != nil {
		t.Fatal("expected nil for nonexistent session")
	}
}

// =============================================================================
// HandleFinishLogin tests
// =============================================================================

// parsedResponseWithChallenge returns a fake ParsedCredentialAssertionData with
// the given challenge string embedded in CollectedClientData.
func parsedResponseWithChallenge(challenge string) *protocol.ParsedCredentialAssertionData {
	return &protocol.ParsedCredentialAssertionData{
		ParsedPublicKeyCredential: protocol.ParsedPublicKeyCredential{
			ParsedCredential: protocol.ParsedCredential{
				ID:   "test-cred",
				Type: "public-key",
			},
			RawID:                  []byte("test-cred"),
			ClientExtensionResults: map[string]any{},
		},
		Response: protocol.ParsedAssertionResponse{
			CollectedClientData: protocol.CollectedClientData{
				Type:      "webauthn.get",
				Challenge: challenge,
				Origin:    "http://localhost",
			},
		},
	}
}

func TestHandleFinishLogin_ParseError(t *testing.T) {
	h := &PasskeyHandler{
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return nil, errors.New("parse failed")
		},
	}

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", bytes.NewReader([]byte("garbage")))
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "invalid passkey response" {
		t.Errorf("error = %q, want %q", resp["error"], "invalid passkey response")
	}
}

func TestHandleFinishLogin_NoPendingSession(t *testing.T) {
	h := &PasskeyHandler{
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return parsedResponseWithChallenge("unknown-challenge"), nil
		},
	}

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "no pending login" {
		t.Errorf("error = %q, want %q", resp["error"], "no pending login")
	}
}

func TestHandleFinishLogin_ValidationFails(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	waProvider := NewMockWebAuthnProvider(ctrl)

	challenge := "test-challenge-abc"
	h := &PasskeyHandler{
		DB:       store,
		WebAuthn: waProvider,
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return parsedResponseWithChallenge(challenge), nil
		},
	}
	h.storeSession("login:"+challenge, &webauthn.SessionData{Challenge: challenge})

	waProvider.EXPECT().ValidatePasskeyLogin(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, nil, errors.New("invalid signature"))

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["error"] != "passkey authentication failed" {
		t.Errorf("error = %q, want %q", resp["error"], "passkey authentication failed")
	}
}

func TestHandleFinishLogin_SessionCookieError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	waProvider := NewMockWebAuthnProvider(ctrl)
	authStore := mocks.NewMockAuthStore(ctrl)
	mw := &auth.Middleware{DB: authStore}

	challenge := "test-challenge-def"
	userID := uuid.New()
	user := &database.User{ID: userID, Username: "bob"}

	h := &PasskeyHandler{
		DB:       store,
		Auth:     mw,
		WebAuthn: waProvider,
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return parsedResponseWithChallenge(challenge), nil
		},
	}
	h.storeSession("login:"+challenge, &webauthn.SessionData{Challenge: challenge})

	waUser := &webAuthnUser{user: user, creds: nil}
	cred := &webauthn.Credential{ID: []byte("cred-123"), Authenticator: webauthn.Authenticator{SignCount: 5}}

	waProvider.EXPECT().ValidatePasskeyLogin(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(waUser, cred, nil)
	store.EXPECT().UpdateWebAuthnSignCount(gomock.Any(), []byte("cred-123"), int64(5)).Return(nil)
	authStore.EXPECT().CreateSessionWithMeta(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(nil, errors.New("session db fail"))

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}

func TestHandleFinishLogin_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	waProvider := NewMockWebAuthnProvider(ctrl)
	authStore := mocks.NewMockAuthStore(ctrl)
	mw := &auth.Middleware{DB: authStore}

	challenge := "test-challenge-ghi"
	userID := uuid.New()
	sessionID := uuid.New()
	user := &database.User{ID: userID, Username: "alice"}

	h := &PasskeyHandler{
		DB:       store,
		Auth:     mw,
		WebAuthn: waProvider,
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return parsedResponseWithChallenge(challenge), nil
		},
	}
	h.storeSession("login:"+challenge, &webauthn.SessionData{Challenge: challenge})

	waUser := &webAuthnUser{user: user, creds: nil}
	cred := &webauthn.Credential{ID: []byte("cred-456"), Authenticator: webauthn.Authenticator{SignCount: 10}}

	waProvider.EXPECT().ValidatePasskeyLogin(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(waUser, cred, nil)
	store.EXPECT().UpdateWebAuthnSignCount(gomock.Any(), []byte("cred-456"), int64(10)).Return(nil)
	authStore.EXPECT().CreateSessionWithMeta(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&database.Session{ID: sessionID, UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, nil)

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp map[string]any
	json.NewDecoder(w.Body).Decode(&resp)
	userMap, ok := resp["user"].(map[string]any)
	if !ok {
		t.Fatal("expected 'user' in response")
	}
	if userMap["username"] != "alice" {
		t.Errorf("username = %q, want %q", userMap["username"], "alice")
	}

	// Verify session cookie was set.
	cookies := w.Result().Cookies()
	found := false
	for _, c := range cookies {
		if c.Name == "lurkarr_session" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected lurkarr_session cookie to be set")
	}
}

func TestHandleFinishLogin_SessionRotation(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	waProvider := NewMockWebAuthnProvider(ctrl)
	authStore := mocks.NewMockAuthStore(ctrl)
	mw := &auth.Middleware{DB: authStore}

	challenge := "test-challenge-jkl"
	userID := uuid.New()
	oldSessionID := uuid.New()
	newSessionID := uuid.New()
	user := &database.User{ID: userID, Username: "charlie"}

	h := &PasskeyHandler{
		DB:       store,
		Auth:     mw,
		WebAuthn: waProvider,
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return parsedResponseWithChallenge(challenge), nil
		},
	}
	h.storeSession("login:"+challenge, &webauthn.SessionData{Challenge: challenge})

	waUser := &webAuthnUser{user: user, creds: nil}
	cred := &webauthn.Credential{ID: []byte("cred-789"), Authenticator: webauthn.Authenticator{SignCount: 1}}

	waProvider.EXPECT().ValidatePasskeyLogin(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(waUser, cred, nil)
	store.EXPECT().UpdateWebAuthnSignCount(gomock.Any(), []byte("cred-789"), int64(1)).Return(nil)
	// Old session should be deleted.
	store.EXPECT().DeleteSession(gomock.Any(), oldSessionID).Return(nil)
	authStore.EXPECT().CreateSessionWithMeta(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&database.Session{ID: newSessionID, UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, nil)

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", http.NoBody)
	r.AddCookie(&http.Cookie{Name: "lurkarr_session", Value: oldSessionID.String()})
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestHandleFinishLogin_SignCountUpdateError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	waProvider := NewMockWebAuthnProvider(ctrl)
	authStore := mocks.NewMockAuthStore(ctrl)
	mw := &auth.Middleware{DB: authStore}

	challenge := "test-challenge-mno"
	userID := uuid.New()
	sessionID := uuid.New()
	user := &database.User{ID: userID, Username: "dave"}

	h := &PasskeyHandler{
		DB:       store,
		Auth:     mw,
		WebAuthn: waProvider,
		sessions: make(map[string]*sessionEntry),
		parseCredentialRequest: func(r *http.Request) (*protocol.ParsedCredentialAssertionData, error) {
			return parsedResponseWithChallenge(challenge), nil
		},
	}
	h.storeSession("login:"+challenge, &webauthn.SessionData{Challenge: challenge})

	waUser := &webAuthnUser{user: user, creds: nil}
	cred := &webauthn.Credential{ID: []byte("cred-err"), Authenticator: webauthn.Authenticator{SignCount: 2}}

	waProvider.EXPECT().ValidatePasskeyLogin(gomock.Any(), gomock.Any(), gomock.Any()).
		Return(waUser, cred, nil)
	// Sign count update fails — should log but continue (non-fatal).
	store.EXPECT().UpdateWebAuthnSignCount(gomock.Any(), []byte("cred-err"), int64(2)).Return(errors.New("db fail"))
	authStore.EXPECT().CreateSessionWithMeta(gomock.Any(), userID, gomock.Any(), gomock.Any(), gomock.Any()).
		Return(&database.Session{ID: sessionID, UserID: userID, ExpiresAt: time.Now().Add(time.Hour)}, nil)

	r := httptest.NewRequest("POST", "/api/auth/passkey/login/finish", http.NoBody)
	w := httptest.NewRecorder()
	h.HandleFinishLogin(w, r)

	// Should still succeed — sign count update is non-fatal.
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (sign count error should be non-fatal)", w.Code, http.StatusOK)
	}
}
