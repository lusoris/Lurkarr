package downloadclient

import "testing"

func TestFilterCompleted(t *testing.T) {
	items := []DownloadItem{
		{ID: "1", Name: "complete", Progress: 1.0},
		{ID: "2", Name: "half", Progress: 0.5},
		{ID: "3", Name: "done", Progress: 1.0},
		{ID: "4", Name: "zero", Progress: 0.0},
	}

	completed := filterCompleted(items)
	if len(completed) != 2 {
		t.Fatalf("expected 2 completed items, got %d", len(completed))
	}
	if completed[0].ID != "1" || completed[1].ID != "3" {
		t.Fatalf("unexpected completed items: %+v", completed)
	}
}

func TestFilterCompleted_Empty(t *testing.T) {
	completed := filterCompleted(nil)
	if len(completed) != 0 {
		t.Fatalf("expected 0 completed items, got %d", len(completed))
	}
}

func TestFilterCompleted_NoneComplete(t *testing.T) {
	items := []DownloadItem{
		{ID: "1", Progress: 0.5},
		{ID: "2", Progress: 0.99},
	}
	completed := filterCompleted(items)
	if len(completed) != 0 {
		t.Fatalf("expected 0 completed items, got %d", len(completed))
	}
}

func TestDownloadItem_GetProgress(t *testing.T) {
	item := DownloadItem{Progress: 0.75}
	if item.GetProgress() != 0.75 {
		t.Fatalf("expected 0.75, got %f", item.GetProgress())
	}
}

func TestClientType_Constants(t *testing.T) {
	types := []ClientType{
		TypeSABnzbd, TypeNZBGet, TypeQBittorrent,
		TypeTransmission, TypeDeluge, TypeRTorrent, TypeUTorrent,
	}
	seen := make(map[ClientType]bool)
	for _, ct := range types {
		if seen[ct] {
			t.Fatalf("duplicate client type: %s", ct)
		}
		seen[ct] = true
		if ct == "" {
			t.Fatal("empty client type constant")
		}
	}
}
