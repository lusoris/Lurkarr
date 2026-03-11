package nzbget

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func newTestServer(t *testing.T, handler func(method string, params json.RawMessage) interface{}) (*httptest.Server, *Client) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req jsonRPCRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad json", http.StatusBadRequest)
			return
		}
		paramsRaw, _ := json.Marshal(req.Params)
		result := handler(req.Method, json.RawMessage(paramsRaw))
		resultJSON, _ := json.Marshal(result)
		json.NewEncoder(w).Encode(jsonRPCResponse{
			ID:     req.ID,
			Result: json.RawMessage(resultJSON),
		})
	}))
	client := NewClient(server.URL, "nzbget", "secret", 5*time.Second)
	return server, client
}

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:6789/", "nzbget", "pass", 30*time.Second)
	if c.BaseURL != "http://localhost:6789" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", c.BaseURL)
	}
	if c.Username != "nzbget" {
		t.Errorf("Username = %q", c.Username)
	}
}

func TestGetQueue(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "listgroups" {
			t.Errorf("method = %q, want listgroups", method)
		}
		return []QueueItem{
			{NZBID: 1, NZBName: "file1.nzb", Status: "DOWNLOADING"},
			{NZBID: 2, NZBName: "file2.nzb", Status: "QUEUED"},
		}
	})
	defer server.Close()

	items, err := client.GetQueue(context.Background())
	if err != nil {
		t.Fatalf("GetQueue error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].NZBName != "file1.nzb" {
		t.Errorf("NZBName = %q", items[0].NZBName)
	}
}

func TestGetHistory(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "history" {
			t.Errorf("method = %q, want history", method)
		}
		return []HistoryItem{
			{NZBID: 10, NZBName: "completed.nzb", Status: "SUCCESS"},
		}
	})
	defer server.Close()

	items, err := client.GetHistory(context.Background())
	if err != nil {
		t.Fatalf("GetHistory error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d items, want 1", len(items))
	}
}

func TestGetStatus(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "status" {
			t.Errorf("method = %q, want status", method)
		}
		return StatusInfo{
			DownloadRate:   1048576,
			DownloadPaused: false,
			ThreadCount:    5,
			UpTimeSec:      3600,
		}
	})
	defer server.Close()

	status, err := client.GetStatus(context.Background())
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if status.DownloadRate != 1048576 {
		t.Errorf("DownloadRate = %d", status.DownloadRate)
	}
	if status.DownloadPaused {
		t.Error("expected not paused")
	}
}

func TestPause(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "pausedownload" {
			t.Errorf("method = %q, want pausedownload", method)
		}
		return true
	})
	defer server.Close()

	if err := client.Pause(context.Background()); err != nil {
		t.Fatalf("Pause error: %v", err)
	}
}

func TestResume(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "resumedownload" {
			t.Errorf("method = %q, want resumedownload", method)
		}
		return true
	})
	defer server.Close()

	if err := client.Resume(context.Background()); err != nil {
		t.Fatalf("Resume error: %v", err)
	}
}

func TestDeleteItem(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "editqueue" {
			t.Errorf("method = %q, want editqueue", method)
		}
		return true
	})
	defer server.Close()

	if err := client.DeleteItem(context.Background(), 42); err != nil {
		t.Fatalf("DeleteItem error: %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		if method != "version" {
			t.Errorf("method = %q, want version", method)
		}
		return "21.1"
	})
	defer server.Close()

	ver, err := client.GetVersion(context.Background())
	if err != nil {
		t.Fatalf("GetVersion error: %v", err)
	}
	if ver != "21.1" {
		t.Errorf("version = %q", ver)
	}
}

func TestTestConnection(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		return "21.1"
	})
	defer server.Close()

	ver, err := client.TestConnection(context.Background())
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if ver != "21.1" {
		t.Errorf("version = %q", ver)
	}
}

func TestRPCError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req jsonRPCRequest
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(jsonRPCResponse{
			ID:    req.ID,
			Error: &jsonRPCError{Code: -1, Message: "unknown method"},
		})
	}))
	defer server.Close()

	client := NewClient(server.URL, "nzbget", "pass", 5*time.Second)
	_, err := client.GetQueue(context.Background())
	if err == nil {
		t.Fatal("expected error for RPC error response")
	}
}

func TestAuthFailure(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewClient(server.URL, "wrong", "creds", 5*time.Second)
	_, err := client.GetQueue(context.Background())
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
}

func TestPauseFailed(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		return false
	})
	defer server.Close()

	err := client.Pause(context.Background())
	if err == nil {
		t.Fatal("expected error when pause returns false")
	}
}

func TestResumeFailed(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		return false
	})
	defer server.Close()

	err := client.Resume(context.Background())
	if err == nil {
		t.Fatal("expected error when resume returns false")
	}
}

func TestDeleteItemFailed(t *testing.T) {
	server, client := newTestServer(t, func(method string, params json.RawMessage) interface{} {
		return false
	})
	defer server.Close()

	err := client.DeleteItem(context.Background(), 42)
	if err == nil {
		t.Fatal("expected error when delete returns false")
	}
}

func TestHTTPServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClient(server.URL, "nzbget", "pass", 5*time.Second)
	_, err := client.GetVersion(context.Background())
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}
