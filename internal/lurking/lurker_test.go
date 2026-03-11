package lurking

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

func TestSonarrLurkerGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 2,
			"records": []arrclient.SonarrEpisode{
				{ID: 1, Title: "Ep 1"},
				{ID: 2, Title: "Ep 2"},
			},
		})
	})

	h := LurkerFor(database.AppSonarr)
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

func TestSonarrLurkerGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.SonarrEpisode{{ID: 5, Title: "Upgrade"}},
		})
	})

	h := LurkerFor(database.AppSonarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Upgrade" {
		t.Errorf("unexpected items: %+v", items)
	}
}

func TestSonarrLurkerSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "EpisodeSearch"})
	})

	h := LurkerFor(database.AppSonarr)
	err := h.Search(context.Background(), client, 42)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestSonarrLurkerGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 3})
	})

	h := LurkerFor(database.AppSonarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 3 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestRadarrLurkerGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.RadarrMovie{{ID: 10, Title: "Missing Movie"}},
		})
	})

	h := LurkerFor(database.AppRadarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing Movie" {
		t.Errorf("unexpected: %+v", items)
	}
}

func TestRadarrLurkerGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.RadarrMovie{{ID: 20, Title: "Upgrade Movie"}},
		})
	})

	h := LurkerFor(database.AppRadarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestRadarrLurkerSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "MoviesSearch"})
	})

	h := LurkerFor(database.AppRadarr)
	if err := h.Search(context.Background(), client, 10); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestRadarrLurkerGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 5})
	})

	h := LurkerFor(database.AppRadarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 5 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestLidarrLurkerGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.LidarrAlbum{{ID: 30, Title: "Missing Album"}},
		})
	})

	h := LurkerFor(database.AppLidarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing Album" {
		t.Errorf("unexpected: %+v", items)
	}
}

func TestLidarrLurkerGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.LidarrAlbum{{ID: 31, Title: "Upgrade Album"}},
		})
	})

	h := LurkerFor(database.AppLidarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestLidarrLurkerSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "AlbumSearch"})
	})

	h := LurkerFor(database.AppLidarr)
	if err := h.Search(context.Background(), client, 30); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestLidarrLurkerGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 1})
	})

	h := LurkerFor(database.AppLidarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 1 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestReadarrLurkerGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.ReadarrBook{{ID: 40, Title: "Missing Book"}},
		})
	})

	h := LurkerFor(database.AppReadarr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestReadarrLurkerGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.ReadarrBook{{ID: 41, Title: "Upgrade Book"}},
		})
	})

	h := LurkerFor(database.AppReadarr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestReadarrLurkerSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "BookSearch"})
	})

	h := LurkerFor(database.AppReadarr)
	if err := h.Search(context.Background(), client, 40); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestReadarrLurkerGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 0})
	})

	h := LurkerFor(database.AppReadarr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 0 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestWhisparrLurkerGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records": []arrclient.WhisparrEpisode{
				{ID: 51, SeriesID: 1, Title: "Missing", Monitored: true, HasFile: false},
			},
		})
	})

	h := LurkerFor(database.AppWhisparr)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing" {
		t.Errorf("expected 1 missing, got %+v", items)
	}
}

func TestWhisparrLurkerGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"totalRecords": 1,
			"records":      []arrclient.WhisparrEpisode{{ID: 52, Title: "Upgrade"}},
		})
	})

	h := LurkerFor(database.AppWhisparr)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("got %d, want 1", len(items))
	}
}

func TestWhisparrLurkerSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "EpisodeSearch"})
	})

	h := LurkerFor(database.AppWhisparr)
	if err := h.Search(context.Background(), client, 50); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestWhisparrLurkerGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 2})
	})

	h := LurkerFor(database.AppWhisparr)
	queue, err := h.GetQueue(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if queue.TotalRecords != 2 {
		t.Errorf("TotalRecords = %d", queue.TotalRecords)
	}
}

func TestErosLurkerGetMissing(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode([]arrclient.ErosMovie{
			{ID: 60, Title: "Has File", Monitored: true, HasFile: true},
			{ID: 61, Title: "Missing Eros", Monitored: true, HasFile: false},
			{ID: 62, Title: "Unmonitored", Monitored: false, HasFile: false},
		})
	})

	h := LurkerFor(database.AppEros)
	items, err := h.GetMissing(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Missing Eros" {
		t.Errorf("expected 1 missing, got %+v", items)
	}
}

func TestErosLurkerGetUpgrades(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		// Eros has no cutoff endpoint, should not be called
		t.Error("unexpected HTTP request")
	})

	h := LurkerFor(database.AppEros)
	items, err := h.GetUpgrades(context.Background(), client)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	if len(items) != 0 {
		t.Fatalf("got %d, want 0 (Eros has no cutoff endpoint)", len(items))
	}
}

func TestErosLurkerSearch(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.CommandResponse{ID: 1, Name: "MoviesSearch"})
	})

	h := LurkerFor(database.AppEros)
	if err := h.Search(context.Background(), client, 60); err != nil {
		t.Fatalf("error: %v", err)
	}
}

func TestErosLurkerGetQueue(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(arrclient.QueueResponse{TotalRecords: 4})
	})

	h := LurkerFor(database.AppEros)
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
