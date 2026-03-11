package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
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
