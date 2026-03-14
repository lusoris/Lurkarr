package seerr

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

type mockRoutingStore struct {
	presences    []database.MediaPresenceResult
	presenceByID map[string][]database.MediaPresenceResult
	findErr      error
	actions      []database.CrossInstanceAction
	createErr    error
}

func (m *mockRoutingStore) FindMediaPresenceByExternalID(_ context.Context, externalID string) ([]database.MediaPresenceResult, error) {
	if m.presenceByID != nil {
		return m.presenceByID[externalID], m.findErr
	}
	return m.presences, m.findErr
}

func (m *mockRoutingStore) CreateCrossInstanceAction(_ context.Context, action database.CrossInstanceAction) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.actions = append(m.actions, action)
	return nil
}

func TestEvaluate_NilRouter(t *testing.T) {
	r := &RequestRouter{}
	d := r.Evaluate(context.Background(), MediaRequest{})
	if d.Action != "approve" {
		t.Fatalf("expected approve, got %s", d.Action)
	}
}

func TestEvaluate_NoExternalID(t *testing.T) {
	r := &RequestRouter{DB: &mockRoutingStore{}}
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Media: Media{TmdbID: 0}})
	if d.Action != "approve" {
		t.Fatalf("expected approve, got %s", d.Action)
	}
}

func TestEvaluate_LookupError(t *testing.T) {
	r := &RequestRouter{DB: &mockRoutingStore{findErr: errors.New("db fail")}}
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Media: Media{TmdbID: 123}})
	if d.Action != "approve" {
		t.Fatalf("expected approve on error, got %s", d.Action)
	}
}

func TestEvaluate_NoCrossData(t *testing.T) {
	r := &RequestRouter{DB: &mockRoutingStore{}}
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Media: Media{TmdbID: 123}})
	if d.Action != "approve" {
		t.Fatalf("expected approve, got %s", d.Action)
	}
}

func TestEvaluate_HigherQualityExists_Decline(t *testing.T) {
	gid := uuid.New()
	r := &RequestRouter{DB: &mockRoutingStore{
		presences: []database.MediaPresenceResult{{
			GroupID:    gid,
			GroupMode:  "quality_hierarchy",
			ExternalID: "tmdb:123",
			Title:      "Test Movie",
			Instances: []database.PresenceInstance{
				{InstanceID: uuid.New(), Name: "radarr-4k", QualityRank: 1, HasFile: true, Monitored: true},
				{InstanceID: uuid.New(), Name: "radarr-hd", QualityRank: 2, HasFile: false, Monitored: true},
			},
		}},
	}}
	// Non-4K request for movie that already exists in rank 1 (4K)
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Is4K: false, Media: Media{TmdbID: 123}})
	if d.Action != "decline" {
		t.Fatalf("expected decline, got %s: %s", d.Action, d.Reason)
	}
	if d.GroupID == nil || *d.GroupID != gid {
		t.Fatal("expected group ID in decision")
	}
}

func TestEvaluate_4KRequest_NoDecline(t *testing.T) {
	r := &RequestRouter{DB: &mockRoutingStore{
		presences: []database.MediaPresenceResult{{
			GroupID:    uuid.New(),
			GroupMode:  "quality_hierarchy",
			ExternalID: "tmdb:123",
			Title:      "Test Movie",
			Instances: []database.PresenceInstance{
				{InstanceID: uuid.New(), Name: "radarr-4k", QualityRank: 1, HasFile: true, Monitored: true},
			},
		}},
	}}
	// 4K request is not declined even if rank 1 already has the file
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Is4K: true, Media: Media{TmdbID: 123}})
	if d.Action != "approve" {
		t.Fatalf("expected approve for 4K request, got %s: %s", d.Action, d.Reason)
	}
}

func TestEvaluate_OverlapDetectMode_NoDecline(t *testing.T) {
	r := &RequestRouter{DB: &mockRoutingStore{
		presences: []database.MediaPresenceResult{{
			GroupID:    uuid.New(),
			GroupMode:  "overlap_detect",
			ExternalID: "tmdb:123",
			Title:      "Test Movie",
			Instances: []database.PresenceInstance{
				{InstanceID: uuid.New(), Name: "radarr-4k", QualityRank: 1, HasFile: true, Monitored: true},
			},
		}},
	}}
	// overlap_detect mode doesn't decline
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Is4K: false, Media: Media{TmdbID: 123}})
	if d.Action != "approve" {
		t.Fatalf("expected approve in overlap_detect mode, got %s", d.Action)
	}
}

func TestEvaluate_MultiInstanceDuplicate_Decline(t *testing.T) {
	gid := uuid.New()
	r := &RequestRouter{DB: &mockRoutingStore{
		presences: []database.MediaPresenceResult{{
			GroupID:    gid,
			GroupMode:  "quality_hierarchy",
			ExternalID: "tmdb:456",
			Title:      "Dup Movie",
			Instances: []database.PresenceInstance{
				{InstanceID: uuid.New(), Name: "radarr-a", QualityRank: 2, HasFile: true, Monitored: true},
				{InstanceID: uuid.New(), Name: "radarr-b", QualityRank: 3, HasFile: true, Monitored: true},
			},
		}},
	}}
	// Media in 2+ instances with files → decline
	d := r.Evaluate(context.Background(), MediaRequest{Type: "movie", Is4K: true, Media: Media{TmdbID: 456}})
	if d.Action != "decline" {
		t.Fatalf("expected decline for multi-instance dup, got %s: %s", d.Action, d.Reason)
	}
}

func TestBuildExternalID_Movie(t *testing.T) {
	id := buildExternalID(MediaRequest{Type: "movie", Media: Media{TmdbID: 42}})
	if id != "tmdb:42" {
		t.Fatalf("expected tmdb:42, got %s", id)
	}
}

func TestBuildExternalID_TV_TVDB(t *testing.T) {
	tvdbID := 99
	id := buildExternalID(MediaRequest{Type: "tv", Media: Media{TvdbID: &tvdbID}})
	if id != "tvdb:99" {
		t.Fatalf("expected tvdb:99, got %s", id)
	}
}

func TestBuildExternalID_TV_FallbackTMDB(t *testing.T) {
	id := buildExternalID(MediaRequest{Type: "tv", Media: Media{TmdbID: 77}})
	if id != "tmdb:77" {
		t.Fatalf("expected tmdb:77, got %s", id)
	}
}

func TestBuildExternalID_Unknown(t *testing.T) {
	id := buildExternalID(MediaRequest{Type: "music"})
	if id != "" {
		t.Fatalf("expected empty, got %s", id)
	}
}

func TestLogAction_NilDB(t *testing.T) {
	r := &RequestRouter{}
	// Should not panic
	r.LogAction(context.Background(), MediaRequest{}, RoutingDecision{})
}

func TestLogAction_NilGroupID(t *testing.T) {
	r := &RequestRouter{DB: &mockRoutingStore{}}
	// No GroupID → should not log
	r.LogAction(context.Background(), MediaRequest{}, RoutingDecision{Action: "approve"})
}

func TestLogAction_Success(t *testing.T) {
	store := &mockRoutingStore{}
	r := &RequestRouter{DB: store}
	gid := uuid.New()
	r.LogAction(context.Background(), MediaRequest{
		ID:    42,
		Type:  "movie",
		Media: Media{TmdbID: 123},
	}, RoutingDecision{
		Action:  "decline",
		Reason:  "higher quality exists",
		GroupID: &gid,
	})
	if len(store.actions) != 1 {
		t.Fatalf("expected 1 action logged, got %d", len(store.actions))
	}
	if store.actions[0].Action != "decline" {
		t.Fatalf("expected decline action, got %s", store.actions[0].Action)
	}
}
