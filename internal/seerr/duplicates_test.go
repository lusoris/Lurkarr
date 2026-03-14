package seerr

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

func TestScanForDuplicates_NilDB(t *testing.T) {
	r := &RequestRouter{}
	result, err := r.ScanForDuplicates(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalScanned != 0 {
		t.Fatalf("expected 0 scanned, got %d", result.TotalScanned)
	}
}

func TestScanForDuplicates_NoDuplicates(t *testing.T) {
	// Mock Seerr server returning one request with no cross-instance match.
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RequestsResponse{
			PageInfo: PageInfo{Results: 1, Pages: 1, PageSize: 50, Page: 1},
			Results: []MediaRequest{
				{ID: 1, Type: "movie", Media: Media{TmdbID: 999}, RequestedBy: User{DisplayName: "test"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	store := &mockRoutingStore{} // No presences → approve
	router := &RequestRouter{DB: store}
	client := NewClient(ts.URL, "test-key", 0)

	result, err := router.ScanForDuplicates(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalScanned != 1 {
		t.Fatalf("expected 1 scanned, got %d", result.TotalScanned)
	}
	if len(result.Duplicates) != 0 {
		t.Fatalf("expected 0 duplicates, got %d", len(result.Duplicates))
	}
}

func TestScanForDuplicates_FindsDuplicate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := RequestsResponse{
			PageInfo: PageInfo{Results: 2, Pages: 1, PageSize: 50, Page: 1},
			Results: []MediaRequest{
				{ID: 10, Type: "movie", Is4K: false, Media: Media{TmdbID: 42}, RequestedBy: User{DisplayName: "alice"}},
				{ID: 11, Type: "movie", Is4K: false, Media: Media{TmdbID: 99}, RequestedBy: User{DisplayName: "bob"}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer ts.Close()

	gid := uuid.New()
	store := &mockRoutingStore{
		presenceByID: map[string][]database.MediaPresenceResult{
			"tmdb:42": {{
				GroupID:    gid,
				GroupMode:  "quality_hierarchy",
				ExternalID: "tmdb:42",
				Title:      "Test Movie",
				Instances: []database.PresenceInstance{
					{Name: "radarr-4k", QualityRank: 1, HasFile: true},
				},
			}},
		},
	}
	router := &RequestRouter{DB: store}
	client := NewClient(ts.URL, "test-key", 0)

	result, err := router.ScanForDuplicates(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalScanned != 2 {
		t.Fatalf("expected 2 scanned, got %d", result.TotalScanned)
	}
	// tmdb:42 should be flagged (rank 1 has file, non-4K request)
	// tmdb:99 should NOT be flagged (no presence data)
	if len(result.Duplicates) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(result.Duplicates))
	}
	if result.Duplicates[0].RequestID != 10 {
		t.Fatalf("expected request 10 flagged, got %d", result.Duplicates[0].RequestID)
	}
}

func TestScanForDuplicates_Pagination(t *testing.T) {
	page := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var results []MediaRequest
		total := 75
		skip := page * 50
		end := skip + 50
		if end > total {
			end = total
		}
		for i := skip; i < end; i++ {
			results = append(results, MediaRequest{
				ID:    i + 1,
				Type:  "movie",
				Media: Media{TmdbID: i + 100},
			})
		}
		resp := RequestsResponse{
			PageInfo: PageInfo{Results: total, Pages: 2, PageSize: 50, Page: page + 1},
			Results:  results,
		}
		json.NewEncoder(w).Encode(resp)
		page++
	}))
	defer ts.Close()

	store := &mockRoutingStore{}
	router := &RequestRouter{DB: store}
	client := NewClient(ts.URL, "test-key", 0)

	result, err := router.ScanForDuplicates(context.Background(), client)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalScanned != 75 {
		t.Fatalf("expected 75 scanned, got %d", result.TotalScanned)
	}
}

func TestScanForDuplicates_ClientError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		fmt.Fprint(w, "internal error")
	}))
	defer ts.Close()

	store := &mockRoutingStore{}
	router := &RequestRouter{DB: store}
	client := NewClient(ts.URL, "test-key", 0)

	_, err := router.ScanForDuplicates(context.Background(), client)
	if err == nil {
		t.Fatal("expected error from failed client")
	}
}
