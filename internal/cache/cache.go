package cache

import (
	"context"
	"time"

	"github.com/lusoris/lurkarr/internal/database"
	"github.com/maypok86/otter/v2"
)

const defaultTTL = 30 * time.Second

type settingsLoader interface {
	GetAppSettings(ctx context.Context, appType database.AppType) (*database.AppSettings, error)
	GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error)
}

// Cache provides a TTL-based in-memory cache for frequently-read settings.
type Cache struct {
	appSettings     *otter.Cache[database.AppType, *database.AppSettings]
	generalSettings *otter.Cache[string, *database.GeneralSettings]
	db              settingsLoader
}

// New creates a Cache backed by the given DB.
func New(db settingsLoader) *Cache {
	appCache := otter.Must(&otter.Options[database.AppType, *database.AppSettings]{
		MaximumSize:      64,
		ExpiryCalculator: otter.ExpiryWriting[database.AppType, *database.AppSettings](defaultTTL),
	})
	genCache := otter.Must(&otter.Options[string, *database.GeneralSettings]{
		MaximumSize:      4,
		ExpiryCalculator: otter.ExpiryWriting[string, *database.GeneralSettings](defaultTTL),
	})
	return &Cache{
		appSettings:     appCache,
		generalSettings: genCache,
		db:              db,
	}
}

// GetAppSettings returns cached app settings, loading from DB on miss.
func (c *Cache) GetAppSettings(ctx context.Context, appType database.AppType) (*database.AppSettings, error) {
	if v, ok := c.appSettings.GetIfPresent(appType); ok {
		return v, nil
	}
	s, err := c.db.GetAppSettings(ctx, appType)
	if err != nil {
		return nil, err
	}
	c.appSettings.Set(appType, s)
	return s, nil
}

// GetGeneralSettings returns cached general settings, loading from DB on miss.
func (c *Cache) GetGeneralSettings(ctx context.Context) (*database.GeneralSettings, error) {
	const key = "general"
	if v, ok := c.generalSettings.GetIfPresent(key); ok {
		return v, nil
	}
	s, err := c.db.GetGeneralSettings(ctx)
	if err != nil {
		return nil, err
	}
	c.generalSettings.Set(key, s)
	return s, nil
}

// InvalidateAppSettings removes a specific app type from cache.
func (c *Cache) InvalidateAppSettings(appType database.AppType) {
	c.appSettings.Invalidate(appType)
}

// InvalidateGeneralSettings removes general settings from cache.
func (c *Cache) InvalidateGeneralSettings() {
	c.generalSettings.Invalidate("general")
}

// InvalidateAll clears the entire cache.
func (c *Cache) InvalidateAll() {
	c.appSettings.InvalidateAll()
	c.generalSettings.InvalidateAll()
}
