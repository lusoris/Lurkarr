package api

import (
	"encoding/json"
	"net/http"
)

const maxRequestBodySize = 1 << 20 // 1 MB

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func errorResponse(msg string) map[string]string {
	return map[string]string{"error": msg}
}

// limitBody wraps the request body with MaxBytesReader.
func limitBody(r *http.Request) {
	r.Body = http.MaxBytesReader(nil, r.Body, maxRequestBodySize)
}
