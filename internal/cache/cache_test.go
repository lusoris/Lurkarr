package cache

import (
	"context"
	"testing"

	"github.com/lusoris/lurkarr/internal/database"
)

// mockDB implements settingsLoader for testing.
type mockDB struct {
	appCalls int
	genCalls int
}

func (m *mockDB) GetAppSettings(_ context.Context, appType database.AppType) (*database.AppSettings, error) {
	m.appCalls++
	return &database.AppSettings{AppType: appType, HourlyCap: 10}, nil
}

func (m *mockDB) GetGeneralSettings(_ context.Context) (*database.GeneralSettings, error) {
	m.genCalls++
	return &database.GeneralSettings{APITimeout: 30, SSLVerify: true}, nil
}

func TestGetAppSettingsCachesResult(t *testing.T) {
	db := &mockDB{}
	c := New(db)

	ctx := context.Background()

	// First call should hit DB
	s1, err := c.GetAppSettings(ctx, database.AppSonarr)
	if err != nil {
		t.Fatal(err)
	}
	if db.appCalls != 1 {
		t.Fatalf("expected 1 DB call, got %d", db.appCalls)
	}
	if s1.HourlyCap != 10 {
		t.Fatalf("expected HourlyCap=10, got %d", s1.HourlyCap)
	}

	// Second call should be cached
	s2, err := c.GetAppSettings(ctx, database.AppSonarr)
	if err != nil {
		t.Fatal(err)
	}
	if db.appCalls != 1 {
		t.Fatalf("expected still 1 DB call after cache hit, got %d", db.appCalls)
	}
	if s2.HourlyCap != 10 {
		t.Fatalf("expected HourlyCap=10, got %d", s2.HourlyCap)
	}

	// Different app type should hit DB again
	_, err = c.GetAppSettings(ctx, database.AppRadarr)
	if err != nil {
		t.Fatal(err)
	}
	if db.appCalls != 2 {
		t.Fatalf("expected 2 DB calls for different app type, got %d", db.appCalls)
	}
}

func TestGetGeneralSettingsCachesResult(t *testing.T) {
	db := &mockDB{}
	c := New(db)

	ctx := context.Background()

	_, err := c.GetGeneralSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if db.genCalls != 1 {
		t.Fatalf("expected 1 DB call, got %d", db.genCalls)
	}

	// Cached
	_, err = c.GetGeneralSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if db.genCalls != 1 {
		t.Fatalf("expected still 1 DB call, got %d", db.genCalls)
	}
}

func TestInvalidateAppSettings(t *testing.T) {
	db := &mockDB{}
	c := New(db)

	ctx := context.Background()

	// Populate cache
	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
	if db.appCalls != 1 {
		t.Fatalf("expected 1 call, got %d", db.appCalls)
	}

	// Invalidate
	c.InvalidateAppSettings(database.AppSonarr)

	// Should hit DB again
	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
	if db.appCalls != 2 {
		t.Fatalf("expected 2 calls after invalidate, got %d", db.appCalls)
	}
}

func TestInvalidateAll(t *testing.T) {
	db := &mockDB{}
	c := New(db)

	ctx := context.Background()

	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
	_, _ = c.GetGeneralSettings(ctx)

	c.InvalidateAll()

	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
	_, _ = c.GetGeneralSettings(ctx)

	if db.appCalls != 2 {
		t.Fatalf("expected 2 app calls after InvalidateAll, got %d", db.appCalls)
	}
	if db.genCalls != 2 {
		t.Fatalf("expected 2 gen calls after InvalidateAll, got %d", db.genCalls)
	}
}
