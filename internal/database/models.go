package database

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// AppType enumerates supported *Arr application types.
type AppType string

const (
	AppSonarr   AppType = "sonarr"
	AppRadarr   AppType = "radarr"
	AppLidarr   AppType = "lidarr"
	AppReadarr  AppType = "readarr"
	AppWhisparr AppType = "whisparr"
	AppEros     AppType = "eros"
	AppProwlarr AppType = "prowlarr"
)

// AllAppTypes returns all supported app types.
func AllAppTypes() []AppType {
	return []AppType{AppSonarr, AppRadarr, AppLidarr, AppReadarr, AppWhisparr, AppEros, AppProwlarr}
}

// ValidAppType checks if a string is a valid app type.
func ValidAppType(s string) bool {
	for _, t := range AllAppTypes() {
		if string(t) == s {
			return true
		}
	}
	return false
}

type User struct {
	ID           uuid.UUID `json:"id"`
	Username     string    `json:"username"`
	Password     string    `json:"-"`
	TOTPSecret   *string   `json:"-"`
	AuthProvider string    `json:"auth_provider"` // "local", "oidc", "proxy"
	ExternalID   string    `json:"external_id,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

type AppInstance struct {
	ID        uuid.UUID `json:"id"`
	AppType   AppType   `json:"app_type"`
	Name      string    `json:"name"`
	APIURL    string    `json:"api_url"`
	APIKey    string    `json:"api_key"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// MaskedAPIKey returns the API key with all but the last 4 characters masked.
func (a *AppInstance) MaskedAPIKey() string {
	if len(a.APIKey) <= 4 {
		return "****"
	}
	return "****" + a.APIKey[len(a.APIKey)-4:]
}

type AppSettings struct {
	AppType          AppType `json:"app_type"`
	LurkMissingCount int     `json:"lurk_missing_count"`
	LurkUpgradeCount int     `json:"lurk_upgrade_count"`
	LurkMissingMode  string  `json:"lurk_missing_mode"`
	UpgradeMode      string  `json:"upgrade_mode"`
	SleepDuration    int     `json:"sleep_duration"`
	MonitoredOnly    bool    `json:"monitored_only"`
	SkipFuture       bool    `json:"skip_future"`
	HourlyCap        int     `json:"hourly_cap"`
	RandomSelection  bool    `json:"random_selection"`
	DebugMode        bool    `json:"debug_mode"`
}

type GeneralSettings struct {
	SecretKey            string `json:"secret_key"`
	ProxyAuthBypass      bool   `json:"proxy_auth_bypass"`
	SSLVerify            bool   `json:"ssl_verify"`
	APITimeout           int    `json:"api_timeout"`
	StatefulResetHours   int    `json:"stateful_reset_hours"`
	CommandWaitDelay     int    `json:"command_wait_delay"`
	CommandWaitAttempts  int    `json:"command_wait_attempts"`
	MinDownloadQueueSize int    `json:"min_download_queue_size"`
}

type ProcessedItem struct {
	ID          int64     `json:"id"`
	AppType     AppType   `json:"app_type"`
	InstanceID  uuid.UUID `json:"instance_id"`
	MediaID     int       `json:"media_id"`
	Operation   string    `json:"operation"`
	ProcessedAt time.Time `json:"processed_at"`
}

type LurkHistory struct {
	ID           int64      `json:"id"`
	AppType      AppType    `json:"app_type"`
	InstanceID   *uuid.UUID `json:"instance_id"`
	InstanceName string     `json:"instance_name"`
	MediaID      int        `json:"media_id"`
	MediaTitle   string     `json:"media_title"`
	Operation    string     `json:"operation"`
	CreatedAt    time.Time  `json:"created_at"`
}

type LurkStats struct {
	AppType    AppType   `json:"app_type"`
	InstanceID uuid.UUID `json:"instance_id"`
	Lurked     int64     `json:"lurked"`
	Upgraded   int64     `json:"upgraded"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type HourlyCap struct {
	AppType    AppType   `json:"app_type"`
	InstanceID uuid.UUID `json:"instance_id"`
	HourBucket time.Time `json:"hour_bucket"`
	APIHits    int       `json:"api_hits"`
}

type Schedule struct {
	ID        uuid.UUID `json:"id"`
	AppType   string    `json:"app_type"`
	Action    string    `json:"action"`
	Days      []string  `json:"days"`
	Hour      int       `json:"hour"`
	Minute    int       `json:"minute"`
	Enabled   bool      `json:"enabled"`
	CreatedAt time.Time `json:"created_at"`
}

// ScheduleExecution represents a single schedule execution log entry.
type ScheduleExecution struct {
	ID         int64     `json:"id"`
	ScheduleID uuid.UUID `json:"schedule_id"`
	ExecutedAt time.Time `json:"executed_at"`
	Result     *string   `json:"result"`
}

type LogEntry struct {
	ID        int64     `json:"id"`
	AppType   string    `json:"app_type"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

type ProwlarrSettings struct {
	ID           int       `json:"id"`
	URL          string    `json:"url"`
	APIKey       string    `json:"api_key"`
	Enabled      bool      `json:"enabled"`
	SyncIndexers bool      `json:"sync_indexers"`
	Timeout      int       `json:"timeout"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// MaskedAPIKey returns the Prowlarr API key masked.
func (p *ProwlarrSettings) MaskedAPIKey() string {
	if len(p.APIKey) <= 4 {
		return "****"
	}
	return "****" + p.APIKey[len(p.APIKey)-4:]
}

type SABnzbdSettings struct {
	ID        int       `json:"id"`
	URL       string    `json:"url"`
	APIKey    string    `json:"api_key"`
	Enabled   bool      `json:"enabled"`
	Timeout   int       `json:"timeout"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// MaskedAPIKey returns the SABnzbd API key masked.
func (s *SABnzbdSettings) MaskedAPIKey() string {
	if len(s.APIKey) <= 4 {
		return "****"
	}
	return "****" + s.APIKey[len(s.APIKey)-4:]
}

// QueueCleanerSettings holds per-app queue cleaning configuration.
type QueueCleanerSettings struct {
	AppType                  AppType `json:"app_type"`
	Enabled                  bool    `json:"enabled"`
	StalledThresholdMinutes  int     `json:"stalled_threshold_minutes"`
	SlowThresholdBytesPerSec int64   `json:"slow_threshold_bytes_per_sec"`
	MaxStrikes               int     `json:"max_strikes"`
	StrikeWindowHours        int     `json:"strike_window_hours"`
	CheckIntervalSeconds     int     `json:"check_interval_seconds"`
	RemoveFromClient         bool    `json:"remove_from_client"`
	BlocklistOnRemove        bool    `json:"blocklist_on_remove"`
	// Per-privacy type settings
	StrikePublic  bool `json:"strike_public"`  // Strike stalled public torrents
	StrikePrivate bool `json:"strike_private"` // Strike stalled private torrents
	// Slow detection exemptions
	SlowIgnoreAboveBytes int64 `json:"slow_ignore_above_bytes"` // Don't flag as slow if remaining > this
	// Failed import cleanup
	FailedImportRemove    bool `json:"failed_import_remove"`    // Auto-remove failed imports
	FailedImportBlocklist bool `json:"failed_import_blocklist"` // Blocklist failed imports on remove
	// Metadata stuck
	MetadataStuckMinutes int `json:"metadata_stuck_minutes"` // Minutes before metadata download is "stuck" (0=disabled)
	// Seeding rules (torrent clients)
	SeedingEnabled     bool    `json:"seeding_enabled"`      // Enable seeding rule enforcement
	SeedingMaxRatio    float64 `json:"seeding_max_ratio"`    // Remove after reaching this ratio (0=disabled)
	SeedingMaxHours    int     `json:"seeding_max_hours"`    // Remove after seeding this many hours (0=disabled)
	SeedingMode        string  `json:"seeding_mode"`         // "and" (both conditions) or "or" (either condition)
	SeedingDeleteFiles bool    `json:"seeding_delete_files"` // Delete downloaded files on seeding removal
	SeedingSkipPrivate bool    `json:"seeding_skip_private"` // Skip seeding rules for private trackers
	// Orphan cleanup (all download client types)
	OrphanEnabled            bool   `json:"orphan_enabled"`             // Enable orphan download detection
	OrphanGraceMinutes       int    `json:"orphan_grace_minutes"`       // Minutes to wait before considering a download orphaned
	OrphanDeleteFiles        bool   `json:"orphan_delete_files"`        // Delete downloaded files when removing orphans
	OrphanExcludedCategories string `json:"orphan_excluded_categories"` // Comma-separated categories to exclude from orphan detection
	// Hardlink protection
	HardlinkProtection bool `json:"hardlink_protection"` // Skip file deletion if files have hardlinks (nlink > 1)
	// Cross-seed awareness
	SkipCrossSeeds bool `json:"skip_cross_seeds"` // Skip removal if multiple torrents share the same content (save path + size)
}

// QueueStrike represents a strike against a problematic download.
type QueueStrike struct {
	ID         int64     `json:"id"`
	AppType    AppType   `json:"app_type"`
	InstanceID uuid.UUID `json:"instance_id"`
	DownloadID string    `json:"download_id"`
	Title      string    `json:"title"`
	Reason     string    `json:"reason"`
	StruckAt   time.Time `json:"struck_at"`
}

// AutoImportLog records auto-import actions.
type AutoImportLog struct {
	ID          int64     `json:"id"`
	AppType     AppType   `json:"app_type"`
	InstanceID  uuid.UUID `json:"instance_id"`
	MediaID     int       `json:"media_id"`
	MediaTitle  string    `json:"media_title"`
	QueueItemID int       `json:"queue_item_id"`
	Action      string    `json:"action"`
	Reason      string    `json:"reason"`
	CreatedAt   time.Time `json:"created_at"`
}

// ScoringProfile defines how to score queue items for deduplication.
type ScoringProfile struct {
	ID                  uuid.UUID `json:"id"`
	AppType             AppType   `json:"app_type"`
	Name                string    `json:"name"`
	Strategy            string    `json:"strategy"` // "highest" (keep best score) or "adequate" (keep first above threshold)
	AdequateThreshold   int       `json:"adequate_threshold"`
	PreferHigherQuality bool      `json:"prefer_higher_quality"`
	PreferLargerSize    bool      `json:"prefer_larger_size"`
	PreferIndexerFlags  bool      `json:"prefer_indexer_flags"`
	CustomFormatWeight  int       `json:"custom_format_weight"`
	SizeWeight          int       `json:"size_weight"`
	AgeWeight           int       `json:"age_weight"`
	SeedersWeight       int       `json:"seeders_weight"`
	CreatedAt           time.Time `json:"created_at"`
}

// BlocklistLog records blocklisted downloads.
type BlocklistLog struct {
	ID            int64     `json:"id"`
	AppType       AppType   `json:"app_type"`
	InstanceID    uuid.UUID `json:"instance_id"`
	DownloadID    string    `json:"download_id"`
	Title         string    `json:"title"`
	Reason        string    `json:"reason"`
	BlocklistedAt time.Time `json:"blocklisted_at"`
}

// BlocklistSource represents a community blocklist URL.
type BlocklistSource struct {
	ID                uuid.UUID  `json:"id"`
	Name              string     `json:"name"`
	URL               string     `json:"url"`
	Enabled           bool       `json:"enabled"`
	SyncIntervalHours int        `json:"sync_interval_hours"`
	LastSyncedAt      *time.Time `json:"last_synced_at"`
	ETag              string     `json:"etag,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
}

// BlocklistRule is a pattern for matching and blocking downloads.
type BlocklistRule struct {
	ID          uuid.UUID  `json:"id"`
	SourceID    *uuid.UUID `json:"source_id,omitempty"` // nil for manual rules
	Pattern     string     `json:"pattern"`
	PatternType string     `json:"pattern_type"` // release_group, title_contains, title_regex, indexer
	Reason      string     `json:"reason"`
	Enabled     bool       `json:"enabled"`
	CreatedAt   time.Time  `json:"created_at"`
}

// NotificationProvider stores configuration for a notification provider.
type NotificationProvider struct {
	ID        uuid.UUID       `json:"id"`
	Type      string          `json:"type"` // discord, telegram, pushover, etc.
	Name      string          `json:"name"` // user-friendly label
	Enabled   bool            `json:"enabled"`
	Config    json.RawMessage `json:"config"` // provider-specific JSON config
	Events    []string        `json:"events"` // event types to subscribe to
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// SeerrSettings holds Seerr integration configuration.
type SeerrSettings struct {
	ID                  uuid.UUID `json:"id"`
	URL                 string    `json:"url"`
	APIKey              string    `json:"api_key"`
	Enabled             bool      `json:"enabled"`
	SyncIntervalMinutes int       `json:"sync_interval_minutes"`
	AutoApprove         bool      `json:"auto_approve"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

// MaskedSeerrAPIKey returns the Seerr API key masked for display.
func (s *SeerrSettings) MaskedSeerrAPIKey() string {
	if len(s.APIKey) <= 4 {
		return "****"
	}
	return "****" + s.APIKey[len(s.APIKey)-4:]
}

// DownloadClientSettings holds per-app download client configuration.
type DownloadClientSettings struct {
	AppType    AppType `json:"app_type"`
	ClientType string  `json:"client_type"` // qbittorrent, transmission, deluge, sabnzbd, nzbget
	URL        string  `json:"url"`
	Username   string  `json:"username"`
	Password   string  `json:"password"`
	Enabled    bool    `json:"enabled"`
	Timeout    int     `json:"timeout"`
}

// MaskedPassword returns the password masked for display.
func (d *DownloadClientSettings) MaskedPassword() string {
	if d.Password == "" {
		return ""
	}
	return "****"
}
