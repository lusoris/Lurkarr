package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

// =============================================================================
// BlocklistHandler — Sources
// =============================================================================

func TestBlocklistListSources(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListBlocklistSources(gomock.Any()).Return([]database.BlocklistSource{
		{ID: uuid.New(), Name: "test", URL: "https://example.com/list.txt"},
	}, nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListSources(w, httptest.NewRequest("GET", "/api/blocklist/sources", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var sources []database.BlocklistSource
	json.NewDecoder(w.Body).Decode(&sources)
	if len(sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(sources))
	}
}

func TestBlocklistListSources_NilSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListBlocklistSources(gomock.Any()).Return(nil, nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListSources(w, httptest.NewRequest("GET", "/api/blocklist/sources", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	// Should be [] not null
	if w.Body.String() == "null\n" {
		t.Fatal("expected empty array, got null")
	}
}

func TestBlocklistListSources_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListBlocklistSources(gomock.Any()).Return(nil, errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListSources(w, httptest.NewRequest("GET", "/api/blocklist/sources", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBlocklistGetSource(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetBlocklistSource(gomock.Any(), id).Return(&database.BlocklistSource{ID: id, Name: "test"}, nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/blocklist/sources/"+id.String(), nil, "id", id.String())
	h.HandleGetSource(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBlocklistGetSource_InvalidID(t *testing.T) {
	h := &BlocklistHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/blocklist/sources/bad", nil, "id", "bad")
	h.HandleGetSource(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistGetSource_NotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().GetBlocklistSource(gomock.Any(), id).Return(nil, errors.New("not found"))
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("GET", "/api/blocklist/sources/"+id.String(), nil, "id", id.String())
	h.HandleGetSource(w, r)
	if w.Code != 404 {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestBlocklistCreateSource(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistSource(gomock.Any(), gomock.Any()).Return(nil)
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "test", "url": "https://example.com/list.txt"})
	w := httptest.NewRecorder()
	h.HandleCreateSource(w, httptest.NewRequest("POST", "/api/blocklist/sources", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestBlocklistCreateSource_MissingFields(t *testing.T) {
	h := &BlocklistHandler{}
	body, _ := json.Marshal(map[string]string{"name": "test"})
	w := httptest.NewRecorder()
	h.HandleCreateSource(w, httptest.NewRequest("POST", "/api/blocklist/sources", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistCreateSource_BadBody(t *testing.T) {
	h := &BlocklistHandler{}
	w := httptest.NewRecorder()
	h.HandleCreateSource(w, httptest.NewRequest("POST", "/api/blocklist/sources", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistCreateSource_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistSource(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "test", "url": "https://example.com/list.txt"})
	w := httptest.NewRecorder()
	h.HandleCreateSource(w, httptest.NewRequest("POST", "/api/blocklist/sources", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBlocklistUpdateSource(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateBlocklistSource(gomock.Any(), gomock.Any()).Return(nil)
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "updated", "url": "https://example.com/list2.txt"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/blocklist/sources/"+id.String(), body, "id", id.String())
	h.HandleUpdateSource(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBlocklistUpdateSource_InvalidID(t *testing.T) {
	h := &BlocklistHandler{}
	body, _ := json.Marshal(map[string]string{"name": "x", "url": "y"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/blocklist/sources/bad", body, "id", "bad")
	h.HandleUpdateSource(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistUpdateSource_MissingFields(t *testing.T) {
	id := uuid.New()
	h := &BlocklistHandler{}
	body, _ := json.Marshal(map[string]string{"name": ""})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/blocklist/sources/"+id.String(), body, "id", id.String())
	h.HandleUpdateSource(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistUpdateSource_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().UpdateBlocklistSource(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"name": "x", "url": "y"})
	w := httptest.NewRecorder()
	r := reqWithPathValue("PUT", "/api/blocklist/sources/"+id.String(), body, "id", id.String())
	h.HandleUpdateSource(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBlocklistDeleteSource(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteBlocklistRulesBySource(gomock.Any(), id).Return(nil)
	store.EXPECT().DeleteBlocklistSource(gomock.Any(), id).Return(nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/sources/"+id.String(), nil, "id", id.String())
	h.HandleDeleteSource(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBlocklistDeleteSource_InvalidID(t *testing.T) {
	h := &BlocklistHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/sources/bad", nil, "id", "bad")
	h.HandleDeleteSource(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistDeleteSource_RulesDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteBlocklistRulesBySource(gomock.Any(), id).Return(errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/sources/"+id.String(), nil, "id", id.String())
	h.HandleDeleteSource(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBlocklistDeleteSource_SourceDeleteError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteBlocklistRulesBySource(gomock.Any(), id).Return(nil)
	store.EXPECT().DeleteBlocklistSource(gomock.Any(), id).Return(errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/sources/"+id.String(), nil, "id", id.String())
	h.HandleDeleteSource(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

// =============================================================================
// BlocklistHandler — Rules
// =============================================================================

func TestBlocklistListRules(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListBlocklistRules(gomock.Any()).Return([]database.BlocklistRule{
		{ID: uuid.New(), Pattern: "test", PatternType: "release_group"},
	}, nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListRules(w, httptest.NewRequest("GET", "/api/blocklist/rules", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBlocklistListRules_NilSlice(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListBlocklistRules(gomock.Any()).Return(nil, nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListRules(w, httptest.NewRequest("GET", "/api/blocklist/rules", http.NoBody))
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if w.Body.String() == "null\n" {
		t.Fatal("expected empty array, got null")
	}
}

func TestBlocklistListRules_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().ListBlocklistRules(gomock.Any()).Return(nil, errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	h.HandleListRules(w, httptest.NewRequest("GET", "/api/blocklist/rules", http.NoBody))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBlocklistCreateRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistRule(gomock.Any(), gomock.Any()).Return(nil)
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"pattern": "YTS", "pattern_type": "release_group"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_TitleContains(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistRule(gomock.Any(), gomock.Any()).Return(nil)
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"pattern": "CAM", "pattern_type": "title_contains"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_TitleRegex(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistRule(gomock.Any(), gomock.Any()).Return(nil)
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"pattern": "^CAM.*720p", "pattern_type": "title_regex"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_InvalidRegex(t *testing.T) {
	h := &BlocklistHandler{}
	body, _ := json.Marshal(map[string]string{"pattern": "[invalid", "pattern_type": "title_regex"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_RegexTooLong(t *testing.T) {
	h := &BlocklistHandler{}
	longPattern := strings.Repeat("a", 1025)
	body, _ := json.Marshal(map[string]string{"pattern": longPattern, "pattern_type": "title_regex"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBlocklistCreateRule_Indexer(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistRule(gomock.Any(), gomock.Any()).Return(nil)
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"pattern": "badindexer", "pattern_type": "indexer"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 201 {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_BadBody(t *testing.T) {
	h := &BlocklistHandler{}
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader([]byte("bad"))))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_EmptyPattern(t *testing.T) {
	h := &BlocklistHandler{}
	body, _ := json.Marshal(map[string]string{"pattern": "", "pattern_type": "release_group"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_InvalidPatternType(t *testing.T) {
	h := &BlocklistHandler{}
	body, _ := json.Marshal(map[string]string{"pattern": "test", "pattern_type": "invalid"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistCreateRule_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	store.EXPECT().CreateBlocklistRule(gomock.Any(), gomock.Any()).Return(errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	body, _ := json.Marshal(map[string]string{"pattern": "test", "pattern_type": "release_group"})
	w := httptest.NewRecorder()
	h.HandleCreateRule(w, httptest.NewRequest("POST", "/api/blocklist/rules", bytes.NewReader(body)))
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestBlocklistDeleteRule(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteBlocklistRule(gomock.Any(), id).Return(nil)
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/rules/"+id.String(), nil, "id", id.String())
	h.HandleDeleteRule(w, r)
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBlocklistDeleteRule_InvalidID(t *testing.T) {
	h := &BlocklistHandler{}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/rules/bad", nil, "id", "bad")
	h.HandleDeleteRule(w, r)
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestBlocklistDeleteRule_DBError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := NewMockStore(ctrl)
	id := uuid.New()
	store.EXPECT().DeleteBlocklistRule(gomock.Any(), id).Return(errors.New("fail"))
	h := &BlocklistHandler{DB: store}
	w := httptest.NewRecorder()
	r := reqWithPathValue("DELETE", "/api/blocklist/rules/"+id.String(), nil, "id", id.String())
	h.HandleDeleteRule(w, r)
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
