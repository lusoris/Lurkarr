package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/auth"
	"github.com/lusoris/lurkarr/internal/database"
)

// testEnv bundles the common test setup: a gomock controller, a mock store,
// and (optionally) a pre-built HTTP recorder.
type testEnv struct {
	Ctrl  *gomock.Controller
	Store *MockStore
}

// newTestEnv creates a testEnv with a fresh gomock controller and mock store.
func newTestEnv(t *testing.T) *testEnv {
	t.Helper()
	ctrl := gomock.NewController(t)
	return &testEnv{
		Ctrl:  ctrl,
		Store: NewMockStore(ctrl),
	}
}

// recorder returns a fresh httptest.ResponseRecorder.
func (te *testEnv) recorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}

// reqWithPathValue creates an *http.Request with a single path parameter set.
func reqWithPathValue(method, path string, body []byte, key, value string) *http.Request {
	var r *http.Request
	if body != nil {
		r = httptest.NewRequest(method, path, bytes.NewReader(body))
	} else {
		r = httptest.NewRequest(method, path, http.NoBody)
	}
	r.SetPathValue(key, value)
	return r
}

// reqWithUserCtx attaches a user to the request context.
func reqWithUserCtx(r *http.Request, user *database.User) *http.Request {
	ctx := auth.ContextWithUser(r.Context(), user)
	return r.WithContext(ctx)
}
