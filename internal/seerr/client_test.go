package seerr

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetAbout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/settings/about" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("X-Api-Key") != "test-key" {
			t.Errorf("missing API key header")
		}
		json.NewEncoder(w).Encode(AboutInfo{
			Version:         "1.33.2",
			TotalMediaItems: 500,
			TotalRequests:   42,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "test-key", 0)
	info, err := c.GetAbout(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if info.Version != "1.33.2" {
		t.Errorf("version = %q, want %q", info.Version, "1.33.2")
	}
	if info.TotalRequests != 42 {
		t.Errorf("total_requests = %d, want 42", info.TotalRequests)
	}
}

func TestGetRequestCount(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/count" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(RequestCount{
			Total:   100,
			Movie:   60,
			TV:      40,
			Pending: 5,
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", 0)
	count, err := c.GetRequestCount(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if count.Total != 100 {
		t.Errorf("total = %d, want 100", count.Total)
	}
	if count.Pending != 5 {
		t.Errorf("pending = %d, want 5", count.Pending)
	}
}

func TestListRequests(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("filter") != "pending" {
			t.Errorf("filter = %q, want pending", r.URL.Query().Get("filter"))
		}
		if r.URL.Query().Get("take") != "20" {
			t.Errorf("take = %q, want 20", r.URL.Query().Get("take"))
		}
		json.NewEncoder(w).Encode(RequestsResponse{
			PageInfo: PageInfo{Pages: 1, PageSize: 20, Results: 2, Page: 1},
			Results: []MediaRequest{
				{ID: 1, Status: RequestPending, Type: "movie"},
				{ID: 2, Status: RequestPending, Type: "tv"},
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", 0)
	resp, err := c.ListRequests(context.Background(), "pending", 20, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(resp.Results) != 2 {
		t.Fatalf("len(results) = %d, want 2", len(resp.Results))
	}
	if resp.Results[0].Status != RequestPending {
		t.Errorf("status = %d, want %d", resp.Results[0].Status, RequestPending)
	}
}

func TestGetRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/request/42" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(MediaRequest{
			ID:     42,
			Status: RequestApproved,
			Type:   "movie",
			Media: Media{
				TmdbID:    998814,
				MediaType: "movie",
				Status:    MediaProcessing,
			},
		})
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", 0)
	req, err := c.GetRequest(context.Background(), 42)
	if err != nil {
		t.Fatal(err)
	}
	if req.Status != RequestApproved {
		t.Errorf("status = %d, want %d", req.Status, RequestApproved)
	}
	if req.Media.TmdbID != 998814 {
		t.Errorf("tmdbId = %d, want 998814", req.Media.TmdbID)
	}
}

func TestApproveRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/request/10/approve" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", 0)
	if err := c.ApproveRequest(context.Background(), 10); err != nil {
		t.Fatal(err)
	}
}

func TestDeclineRequest(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST", r.Method)
		}
		if r.URL.Path != "/api/v1/request/5/decline" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "key", 0)
	if err := c.DeclineRequest(context.Background(), 5); err != nil {
		t.Fatal(err)
	}
}

func TestErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message":"Invalid API key"}`))
	}))
	defer srv.Close()

	c := NewClient(srv.URL, "bad-key", 0)
	_, err := c.GetAbout(context.Background())
	if err == nil {
		t.Fatal("expected error for 401")
	}
}

func TestCancelledContext(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AboutInfo{})
	}))
	defer srv.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	c := NewClient(srv.URL, "key", 0)
	_, err := c.GetAbout(ctx)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
