package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
