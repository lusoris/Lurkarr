package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// =============================================================================
// InstanceGroupsHandler tests
// =============================================================================

func TestHandleListGroups(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstanceGroups(gomock.Any(), database.AppType("sonarr")).
		Return([]database.InstanceGroup{{Name: "HD Group"}}, nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListGroups(w, reqWithPathValue("GET", "/api/instance-groups/sonarr", nil, "app", "sonarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleListGroups_InvalidApp(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	h.HandleListGroups(w, reqWithPathValue("GET", "/api/instance-groups/bad", nil, "app", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleListGroups_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstanceGroups(gomock.Any(), database.AppType("sonarr")).
		Return(nil, errors.New("db fail"))
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListGroups(w, reqWithPathValue("GET", "/api/instance-groups/sonarr", nil, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleListGroups_Nil(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListInstanceGroups(gomock.Any(), database.AppType("radarr")).Return(nil, nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListGroups(w, reqWithPathValue("GET", "/api/instance-groups/radarr", nil, "app", "radarr"))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Should return [] not null
	var out []database.InstanceGroup
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatal(err)
	}
	if out == nil || len(out) != 0 {
		t.Fatalf("expected empty slice, got %v", out)
	}
}

func TestHandleCreateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().CreateInstanceGroup(gomock.Any(), database.AppType("sonarr"), "HD Group").
		Return(&database.InstanceGroup{ID: id, AppType: "sonarr", Name: "HD Group", CreatedAt: time.Now()}, nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"name": "HD Group"})
	h.HandleCreateGroup(w, reqWithPathValue("POST", "/api/instance-groups/sonarr", body, "app", "sonarr"))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestHandleCreateGroup_InvalidApp(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"name": "test"})
	h.HandleCreateGroup(w, reqWithPathValue("POST", "/api/instance-groups/bad", body, "app", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateGroup_EmptyName(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"name": ""})
	h.HandleCreateGroup(w, reqWithPathValue("POST", "/api/instance-groups/sonarr", body, "app", "sonarr"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleCreateGroup_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateInstanceGroup(gomock.Any(), database.AppType("sonarr"), "x").
		Return(nil, errors.New("dup"))
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"name": "x"})
	h.HandleCreateGroup(w, reqWithPathValue("POST", "/api/instance-groups/sonarr", body, "app", "sonarr"))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleGetGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetInstanceGroup(gomock.Any(), id).
		Return(&database.InstanceGroup{ID: id, Name: "test"}, nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetGroup(w, reqWithPathValue("GET", "/api/instance-groups/by-id/"+id.String(), nil, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetGroup_BadID(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	h.HandleGetGroup(w, reqWithPathValue("GET", "/api/instance-groups/by-id/bad", nil, "id", "bad"))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleGetGroup_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetInstanceGroup(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleGetGroup(w, reqWithPathValue("GET", "/api/instance-groups/by-id/"+id.String(), nil, "id", id.String()))
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestHandleUpdateGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateInstanceGroup(gomock.Any(), id, "new name").Return(nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]string{"name": "new name"})
	h.HandleUpdateGroup(w, reqWithPathValue("PUT", "/api/instance-groups/by-id/"+id.String(), body, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleUpdateGroup_EmptyName(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	id := uuid.New()
	body, _ := json.Marshal(map[string]string{"name": ""})
	h.HandleUpdateGroup(w, reqWithPathValue("PUT", "/api/instance-groups/by-id/"+id.String(), body, "id", id.String()))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleDeleteGroup(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteInstanceGroup(gomock.Any(), id).Return(nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteGroup(w, reqWithPathValue("DELETE", "/api/instance-groups/by-id/"+id.String(), nil, "id", id.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDeleteGroup_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteInstanceGroup(gomock.Any(), id).Return(errors.New("fail"))
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleDeleteGroup(w, reqWithPathValue("DELETE", "/api/instance-groups/by-id/"+id.String(), nil, "id", id.String()))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleSetMembers(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	groupID := uuid.New()
	instID := uuid.New()

	store.EXPECT().SetGroupMembers(gomock.Any(), groupID, gomock.Any()).Return(nil)
	store.EXPECT().GetInstanceGroup(gomock.Any(), groupID).
		Return(&database.InstanceGroup{
			ID:   groupID,
			Name: "test",
			Members: []database.InstanceGroupMember{
				{GroupID: groupID, InstanceID: instID, InstanceName: "Sonarr 4K", QualityRank: 1},
			},
		}, nil)
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{
		"members": []map[string]any{
			{"instance_id": instID.String(), "quality_rank": 1},
		},
	})
	h.HandleSetMembers(w, reqWithPathValue("PUT", "/api/instance-groups/by-id/"+groupID.String()+"/members", body, "id", groupID.String()))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSetMembers_BadRank(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	groupID := uuid.New()
	instID := uuid.New()
	body, _ := json.Marshal(map[string]any{
		"members": []map[string]any{
			{"instance_id": instID.String(), "quality_rank": 0},
		},
	})
	h.HandleSetMembers(w, reqWithPathValue("PUT", "/api/instance-groups/by-id/"+groupID.String()+"/members", body, "id", groupID.String()))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSetMembers_BadBody(t *testing.T) {
	h := &InstanceGroupsHandler{}
	w := httptest.NewRecorder()
	groupID := uuid.New()
	h.HandleSetMembers(w, reqWithPathValue("PUT", "/api/instance-groups/by-id/"+groupID.String()+"/members", []byte("bad json"), "id", groupID.String()))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSetMembers_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	groupID := uuid.New()
	instID := uuid.New()
	store.EXPECT().SetGroupMembers(gomock.Any(), groupID, gomock.Any()).Return(errors.New("fail"))
	h := &InstanceGroupsHandler{DB: store}
	w := httptest.NewRecorder()
	body, _ := json.Marshal(map[string]any{
		"members": []map[string]any{
			{"instance_id": instID.String(), "quality_rank": 1},
		},
	})
	h.HandleSetMembers(w, reqWithPathValue("PUT", "/api/instance-groups/by-id/"+groupID.String()+"/members", body, "id", groupID.String()))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
