package database

import (
	"context"
	"fmt"
	"github.com/lusoris/lurkarr/internal/config"
)

// AutoProvision seeds the database from environment variables (via config.Load())
// if the settings are currently empty. This permits a working dashboard on first
// run in Docker compose stacks without manual SQL fixes.
func (db *DB) AutoProvision(ctx context.Context, cfg *config.Config) error {
	// Multi-instance Arr Apps
	arrs := []struct {
		Name    string
		AppType string
		URL     string
		APIKey  string
	}{
		{"Sonarr", "sonarr", "http://sonarr:8989", cfg.SonarrAPIKey},
		{"Radarr", "radarr", "http://radarr:7878", cfg.RadarrAPIKey},
		{"Lidarr", "lidarr", "http://lidarr:8686", cfg.LidarrAPIKey},
		{"Readarr", "readarr", "http://readarr:8787", cfg.ReadarrAPIKey},
	}

	for _, arr := range arrs {
		if arr.APIKey == "" {
			continue
		}
		// Avoid inserting if instance already exists
		_, err := db.Pool.Exec(ctx,
			`INSERT INTO app_instances (api_url, api_key, app_type, name, enabled)
			 VALUES ($1, $2, $3, $4, true)
			 ON CONFLICT (app_type, name) DO NOTHING`,
			arr.URL, arr.APIKey, arr.AppType, arr.Name,
		)
		if err != nil {
			return fmt.Errorf("auto-provision %s: %w", arr.Name, err)
		}
	}

	// Singleton services (DB row ID=1 seeded by migrations)
	singletons := []struct {
		Table  string
		APIKey string
		URL    string
	}{
		{"prowlarr_settings", cfg.ProwlarrAPIKey, "http://prowlarr:9696"},
		{"bazarr_settings", cfg.BazarrAPIKey, "http://bazarr:6767"},
		{"kapowarr_settings", cfg.KapowarrAPIKey, "http://kapowarr:5601"},
		{"seerr_settings", cfg.SeerrAPIKey, "http://seerr:5055"},
	}

	for _, s := range singletons {
		if s.APIKey == "" {
			continue
		}
		// Only update if current key is empty to avoid overwriting user edits after first boot
		query := fmt.Sprintf("UPDATE %s SET url = $1, api_key = $2, enabled = true WHERE id = 1 AND (api_key = '' OR api_key IS NULL)", s.Table)
		_, err := db.Pool.Exec(ctx, query, s.URL, s.APIKey)
		if err != nil {
			return fmt.Errorf("auto-provision %s: %w", s.Table, err)
		}
	}

	return nil
}
