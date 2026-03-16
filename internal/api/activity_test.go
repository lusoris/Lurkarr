package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

func TestActivityFeed_MergesAndSorts(t *testing.T) {
	te := newTestEnv(t)

	now := time.Now()

	te.Store.EXPECT().ListLurkHistory(gomock.Any(), gomock.Any()).Return([]database.LurkHistory{
		{ID: 1, AppType: "radarr", MediaTitle: "Movie A", Operation: "searched", InstanceName: "radarr-1", CreatedAt: now.Add(-1 * time.Minute)},
		{ID: 2, AppType: "sonarr", MediaTitle: "Show B", Operation: "upgraded", InstanceName: "sonarr-1", CreatedAt: now.Add(-5 * time.Minute)},
	}, 2, nil)

	te.Store.EXPECT().ListCrossInstanceActions(gomock.Any(), gomock.Any()).Return([]database.CrossInstanceAction{
		{ID: uuid.New(), Title: "Decline dup", Action: "decline", Reason: "already in 4K", ExecutedAt: now.Add(-2 * time.Minute)},
	}, nil)

	// Blocklist + auto-import: return empty for all app types.
	te.Store.EXPECT().GetBlocklistLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().GetAutoImportLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().GetStrikeLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().ListNotificationHistory(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()

	te.Store.EXPECT().ListScheduleExecutions(gomock.Any(), gomock.Any()).Return(nil, nil)

	h := &ActivityHandler{DB: te.Store}
	w := te.recorder()
	r := httptest.NewRequest("GET", "/api/activity?limit=10", http.NoBody)

	h.HandleGetActivity(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}

	var resp struct {
		Items []ActivityEvent `json:"items"`
		Total int             `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Total != 3 {
		t.Fatalf("total = %d, want 3", resp.Total)
	}

	// Verify sorted by timestamp descending.
	if resp.Items[0].Source != "lurk" || resp.Items[0].Title != "Movie A" {
		t.Errorf("first item = %v, want lurk/Movie A (most recent)", resp.Items[0])
	}
	if resp.Items[1].Source != "cross_instance" {
		t.Errorf("second item source = %q, want cross_instance", resp.Items[1].Source)
	}
	if resp.Items[2].Source != "lurk" || resp.Items[2].Title != "Show B" {
		t.Errorf("third item = %v, want lurk/Show B (oldest)", resp.Items[2])
	}
}

func TestActivityFeed_DefaultLimit(t *testing.T) {
	te := newTestEnv(t)

	te.Store.EXPECT().ListLurkHistory(gomock.Any(), gomock.Any()).Return(nil, 0, nil)
	te.Store.EXPECT().ListCrossInstanceActions(gomock.Any(), gomock.Any()).Return(nil, nil)
	te.Store.EXPECT().GetBlocklistLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().GetAutoImportLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().GetStrikeLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().ListNotificationHistory(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().ListScheduleExecutions(gomock.Any(), gomock.Any()).Return(nil, nil)

	h := &ActivityHandler{DB: te.Store}
	w := te.recorder()
	r := httptest.NewRequest("GET", "/api/activity", http.NoBody)

	h.HandleGetActivity(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	var resp struct {
		Items []ActivityEvent `json:"items"`
		Total int             `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Total != 0 {
		t.Errorf("total = %d, want 0", resp.Total)
	}
}

func TestActivityFeed_LimitTruncation(t *testing.T) {
	te := newTestEnv(t)

	now := time.Now()
	// Generate more items than requested limit.
	history := make([]database.LurkHistory, 5)
	for i := range history {
		history[i] = database.LurkHistory{
			ID: int64(i + 1), AppType: "radarr", MediaTitle: "Movie",
			Operation: "searched", CreatedAt: now.Add(-time.Duration(i) * time.Minute),
		}
	}

	te.Store.EXPECT().ListLurkHistory(gomock.Any(), gomock.Any()).Return(history, 5, nil)
	te.Store.EXPECT().ListCrossInstanceActions(gomock.Any(), gomock.Any()).Return(nil, nil)
	te.Store.EXPECT().GetBlocklistLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().GetAutoImportLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().GetStrikeLog(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().ListNotificationHistory(gomock.Any(), gomock.Any()).Return(nil, nil).AnyTimes()
	te.Store.EXPECT().ListScheduleExecutions(gomock.Any(), gomock.Any()).Return(nil, nil)

	h := &ActivityHandler{DB: te.Store}
	w := te.recorder()
	r := httptest.NewRequest("GET", "/api/activity?limit=3", http.NoBody)

	h.HandleGetActivity(w, r)

	var resp struct {
		Items []ActivityEvent `json:"items"`
		Total int             `json:"total"`
	}
	json.NewDecoder(w.Body).Decode(&resp)

	if resp.Total != 3 {
		t.Errorf("total = %d, want 3 (truncated)", resp.Total)
	}
}
