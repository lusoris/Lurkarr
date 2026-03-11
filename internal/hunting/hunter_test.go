package hunting

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/lusoris/lurkarr/internal/arrclient"
	"github.com/lusoris/lurkarr/internal/database"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *arrclient.Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return arrclient.NewClient(server.URL, "key", 5*time.Second, true)
}

func TestSonarrHunterGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 2,
			"records": []arrclient.SonarrEpisode{
				{ID: 1, Title: "Ep 1"},
				{ID: 2, Title: "Ep 2"},
			},
		})
	})

	h := HunterFor(database.AppSonarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d items, want 2", len(items))
	}
	if items[0].ID != 1 || items[0].Title != "Ep 1" {
		t.Errorf("items[0] = %+v", items[0])
	}
}

func TestSonarrHunterGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.SonarrEpisode{{ID: 5, Title: "Upgrade"}},
		})
	})

	h := HunterFor(database.AppSonarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Upgrade" {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestSonarrHunterSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "EpisodeSearch"})
	})

	h := HunterFor(database.AppSonarr)
	err := h.Search(context.Background(), client, 42)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestSonarrHunterGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 3})
	})

	h := HunterFor(database.AppSonarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestRadarrHunterGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.RadarrMovie{{ID: 10, Title: "Missing Movie"}},
		})
	})

	h := HunterFor(database.AppRadarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing Movie" {
		t.Errorf("unexpected: %+v", items)
	}
}

func TestRadarrHunterGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.RadarrMovie{{ID: 20, Title: "Upgrade Movie"}},
		})
	})

	h := HunterFor(database.AppRadarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestRadarrHunterSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "MoviesSearch"})
	})

	h := HunterFor(database.AppRadarr)
	if err := h.Search(context.Background(), client, 10); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestRadarrHunterGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 5})
	})

	h := HunterFor(database.AppRadarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 5 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestLidarrHunterGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.LidarrAlbum{{ID: 30, Title: "Missing Album"}},
		})
	})

	h := HunterFor(database.AppLidarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing Album" {
		t.Errorf("unexpected: %+v", items)
	}
}

func TestLidarrHunterGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.LidarrAlbum{{ID: 31, Title: "Upgrade Album"}},
		})
	})

	h := HunterFor(database.AppLidarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestLidarrHunterSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "AlbumSearch"})
	})

	h := HunterFor(database.AppLidarr)
	if err := h.Search(context.Background(), client, 30); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestLidarrHunterGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 1})
	})

	h := HunterFor(database.AppLidarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestReadarrHunterGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.ReadarrBook{{ID: 40, Title: "Missing Book"}},
		})
	})

	h := HunterFor(database.AppReadarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestReadarrHunterGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.ReadarrBook{{ID: 41, Title: "Upgrade Book"}},
		})
	})

	h := HunterFor(database.AppReadarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestReadarrHunterSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "BookSearch"})
	})

	h := HunterFor(database.AppReadarr)
	if err := h.Search(context.Background(), client, 40); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestReadarrHunterGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 0})
	})

	h := HunterFor(database.AppReadarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 0 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestWhisparrHunterGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]arrclient.WhisparrMovie{
			{ID: 50, Title: "Has File", Monitored: true, HasFile: true},
			{ID: 51, Title: "Missing", Monitored: true, HasFile: false},
		})
	})

	h := HunterFor(database.AppWhisparr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing" {
		t.Errorf("expected 1 missing, got %+v", items)
	}
}

func TestWhisparrHunterGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.WhisparrMovie{{ID: 52, Title: "Upgrade"}},
		})
	})

	h := HunterFor(database.AppWhisparr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestWhisparrHunterSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "MoviesSearch"})
	})

	h := HunterFor(database.AppWhisparr)
	if err := h.Search(context.Background(), client, 50); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestWhisparrHunterGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 2})
	})

	h := HunterFor(database.AppWhisparr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 2 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestErosHunterGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]arrclient.ErosMovie{
			{ID: 60, Title: "Has File", Monitored: true, HasFile: true},
			{ID: 61, Title: "Missing Eros", Monitored: true, HasFile: false},
			{ID: 62, Title: "Unmonitored", Monitored: false, HasFile: false},
		})
	})

	h := HunterFor(database.AppEros)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing Eros" {
		t.Errorf("expected 1 missing, got %+v", items)
	}
}

func TestErosHunterGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.ErosMovie{{ID: 63, Title: "Upgrade Eros"}},
		})
	})

	h := HunterFor(database.AppEros)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestErosHunterSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "MoviesSearch"})
	})

	h := HunterFor(database.AppEros)
	if err := h.Search(context.Background(), client, 60); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestErosHunterGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 4})
	})

	h := HunterFor(database.AppEros)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 4 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestNewEngine(t *testing.T) {
	e := New(nil, nil)
	if e == nil {
		t.Fatal("New() returned nil")
	}
}

func TestEngineStopNilCancel(t *testing.T) {
	e := New(nil, nil)
	// Should not panic
	e.Stop()
}

func TestSleepReturnsTrue(t *testing.T) {
	e := New(nil, nil)
	ok := e.sleep(context.Background(), 1*time.Millisecond)
	if !ok {
		t.Error("expected sleep to return true")
	}
}

func TestSleepReturnsFalseOnCancel(t *testing.T) {
	e := New(nil, nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ok := e.sleep(ctx, 1*time.Minute)
	if ok {
		t.Error("expected sleep to return false with cancelled context")
	}
}
