package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteJSON(t *testing.T) {
	rec := httptest.NewRecorder()
	data := map[string]string{"key": "value"}
	writeJSON(rec, http.StatusOK, data)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var result map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&result); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if result["key"] != "value" {
		t.Errorf("result[key] = %q, want %q", result["key"], "value")
	}
}

func TestWriteJSONStatus(t *testing.T) {
	tests := []struct {
		status int
	}{
		{http.StatusOK},
		{http.StatusCreated},
		{http.StatusBadRequest},
		{http.StatusNotFound},
		{http.StatusInternalServerError},
	}
	for _, tt := range tests {
		rec := httptest.NewRecorder()
		writeJSON(rec, tt.status, map[string]string{"ok": "true"})
		if rec.Code != tt.status {
			t.Errorf("writeJSON status = %d, want %d", rec.Code, tt.status)
		}
	}
}

func TestErrorResponse(t *testing.T) {
	result := errorResponse("something went wrong")
	if result["error"] != "something went wrong" {
		t.Errorf("errorResponse = %v, want error=%q", result, "something went wrong")
	}
}

func TestErrorResponseEmpty(t *testing.T) {
	result := errorResponse("")
	if result["error"] != "" {
		t.Errorf("errorResponse empty = %v, want empty error", result)
	}
}

func TestLimitBodyRejectsOversized(t *testing.T) {
	// Create a body larger than maxRequestBodySize (1 MB).
	bigBody := strings.NewReader(strings.Repeat("x", maxRequestBodySize+1))
	r := httptest.NewRequest("POST", "/", bigBody)
	w := httptest.NewRecorder()

	limitBody(w, r)

	// Reading the body should fail with MaxBytesError.
	buf := make([]byte, maxRequestBodySize+2)
	_, err := r.Body.Read(buf)
	if err == nil {
		t.Fatal("expected error reading oversized body")
	}
}

func TestWriteJSONUnmarshalableLogsError(t *testing.T) {
	rec := httptest.NewRecorder()
	// A channel is not JSON-marshalable — should log error, not panic.
	writeJSON(rec, http.StatusOK, make(chan int))
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestValidateAPIURL(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{"valid http", "http://sonarr.local:8989", false},
		{"valid https", "https://radarr.example.com", false},
		{"valid with path", "http://192.168.1.100:7878/radarr", false},
		{"empty", "", true},
		{"no scheme", "sonarr.local:8989", true},
		{"file scheme", "file:///etc/passwd", true},
		{"ftp scheme", "ftp://example.com", true},
		{"gopher scheme", "gopher://evil.com", true},
		{"no host", "http://", true},
		{"embedded credentials", "http://user:pass@sonarr.local:8989", true},
		{"embedded user only", "http://admin@sonarr.local:8989", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateAPIURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateAPIURL(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
			}
		})
	}
}
