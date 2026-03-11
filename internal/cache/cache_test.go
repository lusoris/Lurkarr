package cache

import (
	"context"
	"testing"

	"go.uber.org/mock/gomock"

	"github.com/lusoris/lurkarr/internal/database"
)

func TestGetAppSettingsCachesResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := NewMockSettingsLoader(ctrl)
	// Expect exactly one call for sonarr and one for radarr
	db.EXPECT().GetAppSettings(gomock.Any(), database.AppSonarr).
		Return(&database.AppSettings{AppType: database.AppSonarr, HourlyCap: 10}, nil).Times(1)
	db.EXPECT().GetAppSettings(gomock.Any(), database.AppRadarr).
		Return(&database.AppSettings{AppType: database.AppRadarr, HourlyCap: 10}, nil).Times(1)

	c := New(db)
	ctx := context.Background()

	// First call should hit DB
	s1, err := c.GetAppSettings(ctx, database.AppSonarr)
	if err != nil {
		t.Fatal(err)
	}
	if s1.HourlyCap != 10 {
		t.Fatalf("expected HourlyCap=10, got %d", s1.HourlyCap)
	}

	// Second call should be cached (no additional DB call)
	s2, err := c.GetAppSettings(ctx, database.AppSonarr)
	if err != nil {
		t.Fatal(err)
	}
	if s2.HourlyCap != 10 {
		t.Fatalf("expected HourlyCap=10, got %d", s2.HourlyCap)
	}

	// Different app type should hit DB again
	_, err = c.GetAppSettings(ctx, database.AppRadarr)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetGeneralSettingsCachesResult(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := NewMockSettingsLoader(ctrl)
	db.EXPECT().GetGeneralSettings(gomock.Any()).
		Return(&database.GeneralSettings{APITimeout: 30, SSLVerify: true}, nil).Times(1)

	c := New(db)
	ctx := context.Background()

	_, err := c.GetGeneralSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}

	// Cached — no additional DB call
	_, err = c.GetGeneralSettings(ctx)
	if err != nil {
		t.Fatal(err)
	}
}

func TestInvalidateAppSettings(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := NewMockSettingsLoader(ctrl)
	// First call + second call after invalidation = 2
	db.EXPECT().GetAppSettings(gomock.Any(), database.AppSonarr).
		Return(&database.AppSettings{AppType: database.AppSonarr, HourlyCap: 10}, nil).Times(2)

	c := New(db)
	ctx := context.Background()

	// Populate cache
	_, _ = c.GetAppSettings(ctx, database.AppSonarr)

	// Invalidate
	c.InvalidateAppSettings(database.AppSonarr)

	// Should hit DB again
	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
}

func TestInvalidateAll(t *testing.T) {
	ctrl := gomock.NewController(t)
	db := NewMockSettingsLoader(ctrl)
	// 2 calls each: initial + after invalidation
	db.EXPECT().GetAppSettings(gomock.Any(), database.AppSonarr).
		Return(&database.AppSettings{AppType: database.AppSonarr, HourlyCap: 10}, nil).Times(2)
	db.EXPECT().GetGeneralSettings(gomock.Any()).
		Return(&database.GeneralSettings{APITimeout: 30, SSLVerify: true}, nil).Times(2)

	c := New(db)
	ctx := context.Background()

	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
	_, _ = c.GetGeneralSettings(ctx)

	c.InvalidateAll()

	_, _ = c.GetAppSettings(ctx, database.AppSonarr)
	_, _ = c.GetGeneralSettings(ctx)
}
