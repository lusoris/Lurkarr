package crossarr

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/lusoris/lurkarr/internal/database"
)

// --- mockStore ---

type mockStore struct {
	groups       map[database.AppType][]database.InstanceGroup
	instances    map[uuid.UUID]*database.AppInstance
	upsertedMedia []database.CrossInstanceMedia
	upsertErr    error
	presenceMap  map[uuid.UUID][]database.CrossInstancePresence
	presenceErr  error
	deleted      []uuid.UUID
}

func newMockStore() *mockStore {
	return &mockStore{
		groups:      make(map[database.AppType][]database.InstanceGroup),
		instances:   make(map[uuid.UUID]*database.AppInstance),
		presenceMap: make(map[uuid.UUID][]database.CrossInstancePresence),
	}
}

func (m *mockStore) ListInstanceGroups(_ context.Context, appType database.AppType) ([]database.InstanceGroup, error) {
	return m.groups[appType], nil
}

func (m *mockStore) GetInstance(_ context.Context, id uuid.UUID) (*database.AppInstance, error) {
	inst, ok := m.instances[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return inst, nil
}

func (m *mockStore) UpsertCrossInstanceMedia(_ context.Context, groupID uuid.UUID, externalID, title string) (*database.CrossInstanceMedia, error) {
	if m.upsertErr != nil {
		return nil, m.upsertErr
	}
	media := database.CrossInstanceMedia{
		ID:         uuid.New(),
		GroupID:    groupID,
		ExternalID: externalID,
		Title:      title,
		DetectedAt: time.Now(),
	}
	m.upsertedMedia = append(m.upsertedMedia, media)
	return &media, nil
}

func (m *mockStore) SetCrossInstancePresence(_ context.Context, mediaID uuid.UUID, presence []database.CrossInstancePresence) error {
	if m.presenceErr != nil {
		return m.presenceErr
	}
	m.presenceMap[mediaID] = presence
	return nil
}

func (m *mockStore) DeleteCrossInstanceMediaByGroup(_ context.Context, groupID uuid.UUID) error {
	m.deleted = append(m.deleted, groupID)
	return nil
}

// --- Tests ---

func TestScanGroup_NoMembers(t *testing.T) {
	store := newMockStore()
	scanner := &Scanner{DB: store}
	group := &database.InstanceGroup{
		ID:   uuid.New(),
		Name: "empty",
	}
	result := scanner.ScanGroup(context.Background(), group)
	if result.Overlaps != 0 {
		t.Fatalf("expected 0 overlaps, got %d", result.Overlaps)
	}
}

func TestScanGroup_SingleMember(t *testing.T) {
	store := newMockStore()
	scanner := &Scanner{DB: store}
	group := &database.InstanceGroup{
		ID:   uuid.New(),
		Name: "single",
		Members: []database.InstanceGroupMember{
			{InstanceID: uuid.New(), QualityRank: 1},
		},
	}
	result := scanner.ScanGroup(context.Background(), group)
	if result.Overlaps != 0 {
		t.Fatalf("expected 0 overlaps with single member, got %d", result.Overlaps)
	}
}

func TestScanGroup_IndependentMembersSkipped(t *testing.T) {
	store := newMockStore()
	id1, id2 := uuid.New(), uuid.New()
	store.instances[id1] = &database.AppInstance{ID: id1, Enabled: true, APIURL: "http://a", APIKey: "k1"}
	store.instances[id2] = &database.AppInstance{ID: id2, Enabled: true, APIURL: "http://b", APIKey: "k2"}
	scanner := &Scanner{DB: store}
	group := &database.InstanceGroup{
		ID:      uuid.New(),
		Name:    "all-independent",
		AppType: "radarr",
		Members: []database.InstanceGroupMember{
			{InstanceID: id1, QualityRank: 1, IsIndependent: true},
			{InstanceID: id2, QualityRank: 2, IsIndependent: true},
		},
	}
	// Both are independent, so effective non-independent count < 2 -> skip
	result := scanner.ScanGroup(context.Background(), group)
	if result.Overlaps != 0 {
		t.Fatalf("expected 0 overlaps when all members independent, got %d", result.Overlaps)
	}
}

func TestScanGroup_DisabledInstanceSkipped(t *testing.T) {
	store := newMockStore()
	id1, id2 := uuid.New(), uuid.New()
	store.instances[id1] = &database.AppInstance{ID: id1, Enabled: false, APIURL: "http://a", APIKey: "k1"}
	store.instances[id2] = &database.AppInstance{ID: id2, Enabled: true, APIURL: "http://b", APIKey: "k2"}
	scanner := &Scanner{DB: store, ClientTimeout: time.Millisecond}
	group := &database.InstanceGroup{
		ID:      uuid.New(),
		Name:    "one-disabled",
		AppType: "radarr",
		Members: []database.InstanceGroupMember{
			{InstanceID: id1, QualityRank: 1},
			{InstanceID: id2, QualityRank: 2},
		},
	}
	// id1 is disabled, id2 will fail to connect (no real server)
	result := scanner.ScanGroup(context.Background(), group)
	// Should have errors from trying to connect to id2 but not crash
	if len(result.Errors) == 0 {
		t.Fatal("expected errors from unavailable instance, got none")
	}
}

func TestScanGroup_InstanceNotFound(t *testing.T) {
	store := newMockStore()
	scanner := &Scanner{DB: store}
	group := &database.InstanceGroup{
		ID:      uuid.New(),
		Name:    "missing-instance",
		AppType: "radarr",
		Members: []database.InstanceGroupMember{
			{InstanceID: uuid.New(), QualityRank: 1},
			{InstanceID: uuid.New(), QualityRank: 2},
		},
	}
	result := scanner.ScanGroup(context.Background(), group)
	if len(result.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(result.Errors), result.Errors)
	}
}

func TestFetchMedia_UnsupportedAppType(t *testing.T) {
	scanner := &Scanner{}
	_, err := scanner.fetchMedia(context.Background(), nil, "unknown")
	if err == nil {
		t.Fatal("expected error for unsupported app type")
	}
}
