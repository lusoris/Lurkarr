package lurking

import (
	"testing"
	"time"

	"github.com/lusoris/lurkarr/internal/database"
)

func TestBackoff(t *testing.T) {
	tests := []struct {
		errors   int
		expected time.Duration
	}{
		{0, 1 * time.Second},    // 1<<0 = 1s
		{1, 2 * time.Second},    // 1<<1 = 2s
		{2, 4 * time.Second},    // 1<<2 = 4s
		{3, 8 * time.Second},    // 1<<3 = 8s
		{8, 256 * time.Second},  // 1<<8 = 256s (< 5min, clamped by min(errors,8))
		{10, 256 * time.Second}, // min(10,8)=8, 1<<8 = 256s
		{20, 256 * time.Second}, // min(20,8)=8, 1<<8 = 256s
	}
	for _, tt := range tests {
		got := backoff(tt.errors)
		if got != tt.expected {
			t.Errorf("backoff(%d) = %v, want %v", tt.errors, got, tt.expected)
		}
	}
}

func TestSelectItemsAll(t *testing.T) {
	items := []lurkableItem{
		{ID: 1, Title: "a"},
		{ID: 2, Title: "b"},
		{ID: 3, Title: "c"},
	}
	// Request more than available
	selected := selectItems(items, 10, "oldest", nil)
	if len(selected) != 3 {
		t.Fatalf("expected 3, got %d", len(selected))
	}
}

func TestSelectItemsLimited(t *testing.T) {
	items := []lurkableItem{
		{ID: 1, Title: "a"},
		{ID: 2, Title: "b"},
		{ID: 3, Title: "c"},
	}
	selected := selectItems(items, 2, "oldest", nil)
	if len(selected) != 2 {
		t.Fatalf("expected 2, got %d", len(selected))
	}
	// Non-random: oldest mode with zero dates preserves order
	if selected[0].ID != 1 || selected[1].ID != 2 {
		t.Fatalf("expected first 2 items in order, got %v", selected)
	}
}

func TestSelectItemsRandom(t *testing.T) {
	items := make([]lurkableItem, 100)
	for i := range items {
		items[i] = lurkableItem{ID: i, Title: "item"}
	}
	selected := selectItems(items, 5, "random", nil)
	if len(selected) != 5 {
		t.Fatalf("expected 5, got %d", len(selected))
	}
	// With 100 items and random selection, the first selected is unlikely to be ID 0
	// (probabilistically — but we just check count and type)
}

func TestLurkerForKnownTypes(t *testing.T) {
	// Verify Prowlarr is NOT in AllAppTypes (it's an indexer manager, not an arr)
	for _, appType := range database.AllAppTypes() {
		if appType == database.AppProwlarr {
			t.Errorf("Prowlarr should not be in AllAppTypes, but found it")
		}
	}

	// Verify all arr types have lurkers
	for _, appType := range database.AllAppTypes() {
		h := LurkerFor(appType)
		if h == nil {
			t.Errorf("expected lurker for %s, got nil", appType)
		}
	}
}

func TestLurkerForUnknown(t *testing.T) {
	h := LurkerFor(database.AppType("nonexistent"))
	if h != nil {
		t.Errorf("expected nil for unknown type, got %v", h)
	}
}

func TestProwlarrExcludedFromLurking(t *testing.T) {
	// Verify Prowlarr constant still exists (for connections/config)
	if database.AppProwlarr != "prowlarr" {
		t.Errorf("AppProwlarr constant changed unexpectedly")
	}

	// But Prowlarr should not be in AllAppTypes (used by lurking engine)
	allTypes := database.AllAppTypes()
	for _, appType := range allTypes {
		if appType == database.AppProwlarr {
			t.Fatal("Prowlarr must not be in AllAppTypes() - lurking engine must skip indexer apps")
		}
	}

	// And LurkerFor(Prowlarr) should return nil
	if LurkerFor(database.AppProwlarr) != nil {
		t.Error("LurkerFor(AppProwlarr) must return nil - Prowlarr is an indexer, not an arr app")
	}
}
