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
	ID                  uuid.UUID  `json:"id"`
	Username            string     `json:"username"`
	Password            string     `json:"-"`
	TOTPSecret          *string    `json:"-"`
	RecoveryCodes       []string   `json:"-"`
	AuthProvider        string     `json:"auth_provider"` // "local", "oidc", "proxy"
	ExternalID          string     `json:"external_id,omitempty"`
	IsAdmin             bool       `json:"is_admin"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
	FailedLoginAttempts int        `json:"failed_login_attempts"`
	LockedUntil         *time.Time `json:"locked_until,omitempty"`
}

type Session struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IPAddress string    `json:"ip_address"`
	UserAgent string    `json:"user_agent"`
}

type WebAuthnCredential struct {
	ID              uuid.UUID `json:"id"`
	UserID          uuid.UUID `json:"user_id"`
	Name            string    `json:"name"`
	CredentialID    []byte    `json:"-"`
	PublicKey       []byte    `json:"-"`
	AttestationType string    `json:"-"`
	Transport       []string  `json:"-"`
	AAGUID          []byte    `json:"-"`
	SignCount       int64     `json:"-"`
	CreatedAt       time.Time `json:"created_at"`
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

// InstanceGroup represents a named collection of instances with quality tiers.
type InstanceGroup struct {
	ID        uuid.UUID             `json:"id"`
	AppType   AppType               `json:"app_type"`
	Name      string                `json:"name"`
	Mode      string                `json:"mode"`
	CreatedAt time.Time             `json:"created_at"`
	Members   []InstanceGroupMember `json:"members,omitempty"`
}

// InstanceGroupMember links an instance to a group with a quality rank.
// Lower rank means higher quality (e.g., 1 = 4K, 2 = 1080p, 3 = 720p).
type InstanceGroupMember struct {
	GroupID       uuid.UUID `json:"group_id"`
	InstanceID    uuid.UUID `json:"instance_id"`
	InstanceName  string    `json:"instance_name,omitempty"`
	QualityRank   int       `json:"quality_rank"`
	IsIndependent bool      `json:"is_independent"`
}

// CrossInstanceMedia represents a media item detected across multiple instances.
type CrossInstanceMedia struct {
	ID         uuid.UUID               `json:"id"`
	GroupID    uuid.UUID               `json:"group_id"`
	ExternalID string                  `json:"external_id"`
	Title      string                  `json:"title"`
	DetectedAt time.Time               `json:"detected_at"`
	Presence   []CrossInstancePresence `json:"presence,omitempty"`
}

// CrossInstancePresence records that a media item exists in a specific instance.
type CrossInstancePresence struct {
	MediaID      uuid.UUID `json:"media_id"`
	InstanceID   uuid.UUID `json:"instance_id"`
	InstanceName string    `json:"instance_name,omitempty"`
	Monitored    bool      `json:"monitored"`
	HasFile      bool      `json:"has_file"`
}

// CrossInstanceAction records a routing or dedup action taken by Lurkarr.
type CrossInstanceAction struct {
	ID               uuid.UUID  `json:"id"`
	GroupID          uuid.UUID  `json:"group_id"`
	ExternalID       string     `json:"external_id"`
	Title            string     `json:"title"`
	Action           string     `json:"action"`
	Reason           string     `json:"reason"`
	SeerrRequestID   *int       `json:"seerr_request_id,omitempty"`
	SourceInstanceID *uuid.UUID `json:"source_instance_id,omitempty"`
	TargetInstanceID *uuid.UUID `json:"target_instance_id,omitempty"`
	ExecutedAt       time.Time  `json:"executed_at"`
}

// SplitSeasonRule assigns a season range of a series to a specific instance.
type SplitSeasonRule struct {
	ID         uuid.UUID `json:"id"`
	GroupID    uuid.UUID `json:"group_id"`
	ExternalID string    `json:"external_id"`
	Title      string    `json:"title"`
	InstanceID uuid.UUID `json:"instance_id"`
	SeasonFrom int       `json:"season_from"`
	SeasonTo   *int      `json:"season_to,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

type AppSettings struct {
	AppType           AppType `json:"app_type"`
	LurkMissingCount  int     `json:"lurk_missing_count"`
	LurkUpgradeCount  int     `json:"lurk_upgrade_count"`
	LurkMissingMode   string  `json:"lurk_missing_mode"`
	UpgradeMode       string  `json:"upgrade_mode"`
	SleepDuration     int     `json:"sleep_duration"`
	MonitoredOnly     bool    `json:"monitored_only"`
	SkipFuture        bool    `json:"skip_future"`
	HourlyCap         int     `json:"hourly_cap"`
	SelectionMode     string  `json:"selection_mode"`
	MaxSearchFailures int     `json:"max_search_failures"`
	DebugMode         bool    `json:"debug_mode"`
}

type GeneralSettings struct {
	SecretKey                 string `json:"secret_key"`
	ProxyAuthBypass           bool   `json:"proxy_auth_bypass"`
	SSLVerify                 bool   `json:"ssl_verify"`
	APITimeout                int    `json:"api_timeout"`
	StatefulResetHours        int    `json:"stateful_reset_hours"`
	CommandWaitDelay          int    `json:"command_wait_delay"`
	CommandWaitAttempts       int    `json:"command_wait_attempts"`
	MinDownloadQueueSize      int    `json:"min_download_queue_size"`
	AutoImportIntervalMinutes int    `json:"auto_import_interval_minutes"`
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
	// Cross-Arr blocklist sync
	CrossArrSync bool `json:"cross_arr_sync"` // Propagate blocklist removals to all instances of the same app type
	// Safety
	DryRun          bool   `json:"dry_run"`          // Log all actions without executing deletions (safe preview mode)
	ProtectedTags   string `json:"protected_tags"`   // Comma-separated tag labels; tagged media is exempt from all cleanup
	SearchOnRemove  bool   `json:"search_on_remove"` // Trigger arr re-search for the media item after removing a queue entry
	IgnoredIndexers string `json:"ignored_indexers"` // Comma-separated indexer names; items from these indexers skip all cleanup
	// Adaptive speed detection
	BandwidthLimitBytesPerSec int64 `json:"bandwidth_limit_bytes_per_sec"` // Connection speed limit; skip slow detection when >80% saturated (0 = disabled)
	// Per-reason strike overrides (0 = use global max_strikes)
	MaxStrikesStalled  int `json:"max_strikes_stalled"`  // Override max_strikes for stalled items (0 = use global)
	MaxStrikesSlow     int `json:"max_strikes_slow"`     // Override max_strikes for slow items (0 = use global)
	MaxStrikesMetadata int `json:"max_strikes_metadata"` // Override max_strikes for metadata-stuck items (0 = use global)
	MaxStrikesPaused   int `json:"max_strikes_paused"`   // Override max_strikes for paused items (0 = use global)
	// Global size-based ignore
	IgnoreAboveBytes int64 `json:"ignore_above_bytes"` // Skip stalled/slow/metadata checks for items above this size (0 = disabled)
	// Tag instead of delete
	TagInsteadOfDelete bool   `json:"tag_instead_of_delete"` // Tag media with obsolete label instead of removing from queue
	ObsoleteTagLabel   string `json:"obsolete_tag_label"`    // Label for the tag applied to media on removal (e.g. "lurkarr-obsolete")
	// Failed import message patterns
	FailedImportPatterns string `json:"failed_import_patterns"` // Comma-separated substrings; only remove failed imports matching these (empty = all failures)
	// Queued item strikes
	StrikeQueued     bool `json:"strike_queued"`      // Strike items stuck in queued state (not downloading)
	MaxStrikesQueued int  `json:"max_strikes_queued"` // Override max_strikes for queued items (0 = use global)
	// Search cooldown
	SearchCooldownHours int `json:"search_cooldown_hours"` // Minimum hours between re-searches for the same media (0 = no cooldown)
	MaxSearchesPerRun   int `json:"max_searches_per_run"`  // Max re-searches per cleanup run per instance (0 = unlimited)
	MaxSearchFailures   int `json:"max_search_failures"`   // Max consecutive search failures before deprioritizing media (0 = no limit)
	// Deletion detection
	DeletionDetectionEnabled bool `json:"deletion_detection_enabled"` // Remove queue items for externally deleted media files
	// Unmonitored cleanup
	UnmonitoredCleanupEnabled bool `json:"unmonitored_cleanup_enabled"` // Remove queue items for unmonitored media
	// Unregistered torrent detection
	UnregisteredEnabled    bool `json:"unregistered_enabled"`     // Detect and strike torrents removed from tracker
	MaxStrikesUnregistered int  `json:"max_strikes_unregistered"` // Override max_strikes for unregistered items (0 = use global)
	// Recheck paused torrents
	RecheckPausedEnabled bool `json:"recheck_paused_enabled"` // Recheck paused torrents and auto-resume if complete
	// RecycleBin
	RecycleBinEnabled bool   `json:"recycle_bin_enabled"` // Move files to recycle folder instead of permanent deletion
	RecycleBinPath    string `json:"recycle_bin_path"`    // Absolute path to recycle bin folder
	// Per-reason blocklist toggles (override the global blocklist_on_remove)
	BlocklistStalled      bool `json:"blocklist_stalled"`      // Blocklist when removing stalled items
	BlocklistSlow         bool `json:"blocklist_slow"`         // Blocklist when removing slow items
	BlocklistMetadata     bool `json:"blocklist_metadata"`     // Blocklist when removing metadata-stuck items
	BlocklistDuplicate    bool `json:"blocklist_duplicate"`    // Blocklist when removing lower-quality duplicates
	BlocklistUnregistered bool `json:"blocklist_unregistered"` // Blocklist when removing unregistered torrents
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

// NotificationHistory records a single notification delivery attempt.
type NotificationHistory struct {
	ID           uuid.UUID  `json:"id"`
	ProviderID   *uuid.UUID `json:"provider_id,omitempty"`
	ProviderType string     `json:"provider_type"`
	ProviderName string     `json:"provider_name"`
	EventType    string     `json:"event_type"`
	Title        string     `json:"title"`
	Message      string     `json:"message"`
	AppType      string     `json:"app_type"`
	Instance     string     `json:"instance"`
	Status       string     `json:"status"` // "sent" | "failed"
	Error        string     `json:"error,omitempty"`
	DurationMs   int        `json:"duration_ms"`
	CreatedAt    time.Time  `json:"created_at"`
}

// SeerrSettings holds Seerr integration configuration.
type SeerrSettings struct {
	ID                  uuid.UUID `json:"id"`
	URL                 string    `json:"url"`
	APIKey              string    `json:"api_key"`
	Enabled             bool      `json:"enabled"`
	SyncIntervalMinutes int       `json:"sync_interval_minutes"`
	AutoApprove         bool      `json:"auto_approve"`
	CleanupEnabled      bool      `json:"cleanup_enabled"`
	CleanupAfterDays    int       `json:"cleanup_after_days"`
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

// DownloadClientInstance represents a configured download client (multi-instance).
type DownloadClientInstance struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	ClientType string    `json:"client_type"` // qbittorrent, transmission, deluge, sabnzbd, nzbget
	URL        string    `json:"url"`
	APIKey     string    `json:"api_key"`
	Username   string    `json:"username"`
	Password   string    `json:"password"`
	Category   string    `json:"category"`
	Enabled    bool      `json:"enabled"`
	Timeout    int       `json:"timeout"`
	CreatedAt  time.Time `json:"created_at"`
}

// MaskedAPIKey returns the download client API key masked.
func (d *DownloadClientInstance) MaskedAPIKey() string {
	if len(d.APIKey) <= 4 {
		return "****"
	}
	return "****" + d.APIKey[len(d.APIKey)-4:]
}

// MaskedPassword returns the download client password masked.
func (d *DownloadClientInstance) MaskedPassword() string {
	if d.Password == "" {
		return ""
	}
	return "****"
}

// OIDCSettings holds OIDC/SSO configuration.
type OIDCSettings struct {
	Enabled      bool      `json:"enabled"`
	IssuerURL    string    `json:"issuer_url"`
	ClientID     string    `json:"client_id"`
	ClientSecret string    `json:"client_secret"`
	RedirectURL  string    `json:"redirect_url"`
	Scopes       string    `json:"scopes"`
	AutoCreate   bool      `json:"auto_create"`
	AdminGroup   string    `json:"admin_group"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// PersistentCounter stores a cumulative metric value across restarts.
type PersistentCounter struct {
	MetricName string    `json:"metric_name"`
	LabelKey   string    `json:"label_key"`
	Value      int64     `json:"value"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// MaskedClientSecret returns the OIDC client secret masked for display.
func (o *OIDCSettings) MaskedClientSecret() string {
	if len(o.ClientSecret) <= 4 {
		return "****"
	}
	return "****" + o.ClientSecret[len(o.ClientSecret)-4:]
}
