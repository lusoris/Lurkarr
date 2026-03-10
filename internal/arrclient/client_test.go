package arrclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	c := NewClient("http://localhost:8989", "testkey", 30*time.Second, true)
	if c.BaseURL != "http://localhost:8989" {
		t.Errorf("BaseURL = %q, want %q", c.BaseURL, "http://localhost:8989")
	}
	if c.APIKey != "testkey" {
		t.Errorf("APIKey = %q, want %q", c.APIKey, "testkey")
	}
}

func TestNewClientTrimsTrailingSlash(t *testing.T) {
	c := NewClient("http://localhost:8989/", "key", 30*time.Second, true)
	if c.BaseURL != "http://localhost:8989" {
		t.Errorf("BaseURL = %q, want trailing slash trimmed", c.BaseURL)
	}
}

func TestClientGet(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Api-Key") != "testkey" {
			t.Error("missing X-Api-Key header")
		}
		if r.Method != http.MethodGet {
			t.Errorf("method = %s, want GET", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "testkey", 5*time.Second, true)
	var result map[string]string
	err := c.get(context.Background(), "/test", &result)
	if err != nil {
		t.Fatalf("get error: %v", err)
	}
	if result["status"] != "ok" {
		t.Errorf("result = %v", result)
	}
}

func TestClientGetErrorStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	}))
	defer server.Close()

	c := NewClient(server.URL, "testkey", 5*time.Second, true)
	var result map[string]string
	err := c.get(context.Background(), "/missing", &result)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

func TestClientPost(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"id": 42})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	var result map[string]int
	err := c.post(context.Background(), "/command", nil, &result)
	if err != nil {
		t.Fatalf("post error: %v", err)
	}
	if result["id"] != 42 {
		t.Errorf("result = %v", result)
	}
}

func TestTestConnection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v3/system/status" {
			t.Errorf("path = %s, want /api/v3/system/status", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SystemStatus{AppName: "Sonarr", Version: "4.0.0"})
	}))
	defer server.Close()

	c := NewClient(server.URL, "key", 5*time.Second, true)
	status, err := c.TestConnection(context.Background(), "v3")
	if err != nil {
		t.Fatalf("TestConnection error: %v", err)
	}
	if status.AppName != "Sonarr" {
		t.Errorf("AppName = %q, want %q", status.AppName, "Sonarr")
	}
	if status.Version != "4.0.0" {
		t.Errorf("Version = %q, want %q", status.Version, "4.0.0")
	}
}

func TestIsPrivateIP(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		want    bool
		wantErr bool
	}{
		{"loopback", "http://127.0.0.1:8989", true, false},
		{"private 10.x", "http://10.0.0.1:8989", true, false},
		{"private 192.168.x", "http://192.168.1.1:8989", true, false},
		{"private 172.16.x", "http://172.16.0.1:8989", true, false},
		{"public IP", "http://8.8.8.8:8989", false, false},
		{"localhost", "http://localhost:8989", true, false},
		{"invalid url", "://invalid", false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := IsPrivateIP(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("IsPrivateIP(%q) error = %v, wantErr %v", tt.url, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("IsPrivateIP(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
