package api

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

const maxRequestBodySize = 1 << 20 // 1 MB

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		slog.Error("writeJSON: failed to encode response", "error", err)
	}
}

func errorResponse(msg string) map[string]string {
	return map[string]string{"error": msg}
}

// limitBody wraps the request body with MaxBytesReader.
func limitBody(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodySize)
}

// decodeJSON reads a JSON request body into the given type, returning false on error.
func decodeJSON[T any](w http.ResponseWriter, r *http.Request) (T, bool) {
	limitBody(w, r)
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid request body"))
		return v, false
	}
	return v, true
}

// parseUUID extracts and validates a UUID path parameter.
func parseUUID(w http.ResponseWriter, r *http.Request, param string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.PathValue(param))
	if err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid "+param))
		return uuid.UUID{}, false
	}
	return id, true
}

// validAppType extracts and validates an app type path parameter.
func validAppTypeParam(w http.ResponseWriter, r *http.Request) (database.AppType, bool) {
	appType := r.PathValue("app")
	if !database.ValidAppType(appType) {
		writeJSON(w, http.StatusBadRequest, errorResponse("invalid app type"))
		return "", false
	}
	return database.AppType(appType), true
}

// validateAPIURL checks that a URL is safe to use as an API endpoint.
// It ensures the scheme is http or https, the host is non-empty, and
// no embedded credentials (userinfo) are present.
func validateAPIURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("URL scheme must be http or https, got %q", u.Scheme)
	}
	if u.Host == "" {
		return fmt.Errorf("URL must have a host")
	}
	if u.User != nil {
		return fmt.Errorf("URL must not contain embedded credentials")
	}
	return nil
}
