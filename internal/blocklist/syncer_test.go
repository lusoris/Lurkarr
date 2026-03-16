package blocklist

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// mockSyncStore implements SyncStore for testing.
type mockSyncStore struct {
	sources       []database.BlocklistSource
	listErr       error
	replacedSrcID uuid.UUID
	replacedRules []database.BlocklistRule
	replaceErr    error
	updatedIDs    []uuid.UUID
	updatedETags  []string
	updateErr     error
}

func (m *mockSyncStore) ListBlocklistSources(_ context.Context) ([]database.BlocklistSource, error) {
	return m.sources, m.listErr
}
func (m *mockSyncStore) ReplaceBlocklistRulesForSource(_ context.Context, sourceID uuid.UUID, rules []database.BlocklistRule) error {
	m.replacedSrcID = sourceID
	m.replacedRules = rules
	return m.replaceErr
}
func (m *mockSyncStore) UpdateBlocklistSourceSync(_ context.Context, id uuid.UUID, etag string) error {
	m.updatedIDs = append(m.updatedIDs, id)
	m.updatedETags = append(m.updatedETags, etag)
	return m.updateErr
}

func TestSyncSource_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("ETag", `"abc123"`)
		fmt.Fprintln(w, "# Test blocklist")
		fmt.Fprintln(w, "EVO")
		fmt.Fprintln(w, "group:SPARKS")
		fmt.Fprintln(w, "regex:(?i)\\bCAM\\b")
	}))
	defer srv.Close()

	store := &mockSyncStore{}
	syncer := NewSyncer(store)
	srcID := uuid.New()
	src := database.BlocklistSource{
		ID:      srcID,
		Name:    "test-source",
		URL:     srv.URL,
		Enabled: true,
	}

	err := syncer.SyncSource(context.Background(), src)
	if err != nil {
		t.Fatalf("SyncSource error: %v", err)
	}

	if store.replacedSrcID != srcID {
		t.Errorf("expected rules replaced for source %s, got %s", srcID, store.replacedSrcID)
	}
	if len(store.replacedRules) != 3 {
		t.Fatalf("expected 3 rules replaced, got %d", len(store.replacedRules))
	}
	if store.replacedRules[0].Pattern != "EVO" {
		t.Errorf("rule 0 pattern = %q, want EVO", store.replacedRules[0].Pattern)
	}
	if store.replacedRules[0].SourceID == nil || *store.replacedRules[0].SourceID != srcID {
		t.Error("rule 0 should have source ID set")
	}
	if !store.replacedRules[0].Enabled {
		t.Error("rule 0 should be enabled")
	}
	if len(store.updatedETags) != 1 || store.updatedETags[0] != `"abc123"` {
		t.Errorf("expected ETag update, got %v", store.updatedETags)
	}
}

func TestSyncSource_NotModified(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("If-None-Match") == `"cached"` {
			w.WriteHeader(http.StatusNotModified)
			return
		}
		t.Error("expected ETag conditional request")
	}))
	defer srv.Close()

	store := &mockSyncStore{}
	syncer := NewSyncer(store)
	src := database.BlocklistSource{
		ID:      uuid.New(),
		Name:    "cached-source",
		URL:     srv.URL,
		Enabled: true,
		ETag:    `"cached"`,
	}

	err := syncer.SyncSource(context.Background(), src)
	if err != nil {
		t.Fatalf("SyncSource error: %v", err)
	}

	if len(store.replacedRules) != 0 {
		t.Error("no rules should be replaced for 304")
	}
}

func TestSyncSource_ServerError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	store := &mockSyncStore{}
	syncer := NewSyncer(store)
	err := syncer.SyncSource(context.Background(), database.BlocklistSource{
		ID:  uuid.New(),
		URL: srv.URL,
	})

	if err == nil {
		t.Fatal("expected error for 500 status")
	}
}

func TestSyncSource_ReplaceError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "EVO")
	}))
	defer srv.Close()

	store := &mockSyncStore{replaceErr: errors.New("db error")}
	syncer := NewSyncer(store)
	err := syncer.SyncSource(context.Background(), database.BlocklistSource{
		ID:  uuid.New(),
		URL: srv.URL,
	})

	if err == nil {
		t.Fatal("expected error when delete fails")
	}
}

func TestSyncSource_UpdateSyncError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "EVO")
	}))
	defer srv.Close()

	store := &mockSyncStore{updateErr: errors.New("db error")}
	syncer := NewSyncer(store)
	err := syncer.SyncSource(context.Background(), database.BlocklistSource{
		ID:  uuid.New(),
		URL: srv.URL,
	})

	if err == nil {
		t.Fatal("expected error when update sync fails")
	}
}

func TestSyncAll_Success(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		fmt.Fprintln(w, "EVO")
	}))
	defer srv.Close()

	store := &mockSyncStore{
		sources: []database.BlocklistSource{
			{ID: uuid.New(), Name: "s1", URL: srv.URL, Enabled: true},
			{ID: uuid.New(), Name: "s2", URL: srv.URL, Enabled: false}, // disabled
			{ID: uuid.New(), Name: "s3", URL: srv.URL, Enabled: true},
		},
	}

	syncer := NewSyncer(store)
	syncer.SyncAll(context.Background())

	if callCount != 2 {
		t.Errorf("expected 2 HTTP requests (skipping disabled), got %d", callCount)
	}
}

func TestSyncAll_ListError(t *testing.T) {
	store := &mockSyncStore{listErr: errors.New("db error")}
	syncer := NewSyncer(store)
	// Should not panic.
	syncer.SyncAll(context.Background())
}

func TestSyncSource_InvalidURL(t *testing.T) {
	store := &mockSyncStore{}
	syncer := NewSyncer(store)
	err := syncer.SyncSource(context.Background(), database.BlocklistSource{
		ID:  uuid.New(),
		URL: "://invalid",
	})
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestSyncSource_LargeBody(t *testing.T) {
	// Verify the 5MB limit doesn't cause issues with normal content.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := 0; i < 100; i++ {
			fmt.Fprintf(w, "group:G%d\n", i)
		}
	}))
	defer srv.Close()

	store := &mockSyncStore{}
	syncer := NewSyncer(store)
	err := syncer.SyncSource(context.Background(), database.BlocklistSource{
		ID:  uuid.New(),
		URL: srv.URL,
	})
	if err != nil {
		t.Fatalf("SyncSource error: %v", err)
	}
	if len(store.replacedRules) != 100 {
		t.Errorf("expected 100 rules, got %d", len(store.replacedRules))
	}
}

func TestNewSyncer(t *testing.T) {
	store := &mockSyncStore{}
	syncer := NewSyncer(store)
	if syncer == nil {
		t.Fatal("NewSyncer returned nil")
	}
	if syncer.client == nil {
		t.Error("NewSyncer should initialize HTTP client")
	}
}
