// Shared TypeScript interfaces used across multiple pages.
// Import from '$lib/types' instead of declaring inline.

// ─── Arr Instances ──────────────────────────────────────────
export interface AppInstance {
	id: string;
	app_type: string;
	name: string;
	api_url: string;
	api_key: string;
	enabled: boolean;
}

// ─── Download Clients ───────────────────────────────────────
export interface DownloadClientInstance {
	id: string;
	name: string;
	client_type: string;
	url: string;
	api_key: string;
	username: string;
	password: string;
	category: string;
	enabled: boolean;
	timeout: number;
}

// ─── Health ─────────────────────────────────────────────────
export interface HealthInfo {
	status: string;
	version?: string;
}

// ─── Service Settings (Prowlarr / Seerr / SABnzbd) ─────────
export interface ServiceSettings {
	url: string;
	api_key: string;
	enabled: boolean;
}

export interface ProwlarrSettings extends ServiceSettings {
	sync_indexers: boolean;
	timeout: number;
}

export interface SeerrSettings extends ServiceSettings {
	id: string;
	sync_interval_minutes: number;
	auto_approve: boolean;
	cleanup_enabled: boolean;
	cleanup_after_days: number;
}

export interface BazarrSettings extends ServiceSettings {
	id: number;
	timeout: number;
}

export interface KapowarrSettings extends ServiceSettings {
	id: number;
	timeout: number;
}

export interface ShokoSettings extends ServiceSettings {
	id: number;
	timeout: number;
}

export interface SABnzbdSettings extends ServiceSettings {
	id: number;
	timeout: number;
	category: string;
}

// ─── Dashboard ──────────────────────────────────────────────
export interface Stats {
	app_type: string;
	instance_id: string;
	lurked: number;
	upgraded: number;
	updated_at: string;
}

export interface HourlyCap {
	app_type: string;
	instance_id: string;
	hour_bucket: string;
	api_hits: number;
}

// ─── State ──────────────────────────────────────────────────
export interface StateEntry {
	app_type: string;
	instance_id: string;
	name: string;
	last_reset: string | null;
}

// ─── History ────────────────────────────────────────────────
export interface HistoryItem {
	id: number;
	app_type: string;
	instance_name: string;
	media_title: string;
	operation: string;
	created_at: string;
}

export interface BlocklistEntry {
	id: number;
	app_type: string;
	instance_id: string;
	download_id: string;
	title: string;
	reason: string;
	blocklisted_at: string;
}

export interface ImportEntry {
	id: number;
	app_type: string;
	instance_id: string;
	media_id: number;
	media_title: string;
	queue_item_id: number;
	action: string;
	reason: string;
	created_at: string;
}

export interface StrikeEntry {
	id: number;
	app_type: string;
	instance_id: string;
	download_id: string;
	title: string;
	reason: string;
	struck_at: string;
}

// ─── Downloads ──────────────────────────────────────────────
export interface ClientStatus {
	version: string;
	download_speed: number;
	upload_speed: number;
	paused: boolean;
	item_count: number;
}

export interface DownloadItem {
	id: string;
	name: string;
	status: string;
	total_size: number;
	remaining_size: number;
	progress: number;
	download_speed: number;
	upload_speed: number;
	eta: number;
	category: string;
}

export interface SABnzbdQueueSlot {
	nzo_id: string;
	filename: string;
	status: string;
	mb: string;
	mbleft: string;
	percentage: string;
	timeleft: string;
	cat: string;
}

export interface SABnzbdQueue {
	status: string;
	speed: string;
	sizeleft: string;
	noofslots: number;
	slots: SABnzbdQueueSlot[];
	paused: boolean;
}

export interface SABnzbdStats {
	total: string;
	day: string;
	week: string;
	month: string;
}

// ─── Activity ───────────────────────────────────────────────
export interface ActivityEvent {
	id: string;
	source: string;
	app_type?: string;
	title: string;
	action: string;
	detail?: string;
	timestamp: string;
}

// ─── Scheduling ─────────────────────────────────────────────
export interface Schedule {
	id: string;
	app_type: string;
	action: string;
	days: string[];
	hour: number;
	minute: number;
	enabled: boolean;
}

export interface ScheduleExecution {
	id: number;
	schedule_id: string;
	executed_at: string;
	result: string | null;
}

// ─── Notifications ──────────────────────────────────────────
export interface NotificationProvider {
	id: string;
	type: string;
	name: string;
	enabled: boolean;
	config: Record<string, string>;
	events: string[];
	created_at: string;
	updated_at: string;
}

export interface NotificationHistoryEntry {
	id: string;
	provider_type: string;
	provider_name: string;
	event_type: string;
	title: string;
	message: string;
	app_type: string;
	instance: string;
	status: string;
	error: string;
	duration_ms: number;
	created_at: string;
}

// ─── Queue Cleaner Logs ─────────────────────────────────────
export interface AutoImportLog {
	id: number;
	app_type: string;
	instance_id: string;
	media_title: string;
	action: string;
	reason: string;
	created_at: string;
}

// ─── Queue Cleaner ──────────────────────────────────────────
export interface QueueCleanerSettings {
	app_type: string;
	enabled: boolean;
	stalled_threshold_minutes: number;
	slow_threshold_bytes_per_sec: number;
	max_strikes: number;
	strike_window_hours: number;
	check_interval_seconds: number;
	remove_from_client: boolean;
	blocklist_on_remove: boolean;
	strike_public: boolean;
	strike_private: boolean;
	slow_ignore_above_bytes: number;
	failed_import_remove: boolean;
	failed_import_blocklist: boolean;
	metadata_stuck_minutes: number;
	seeding_enabled: boolean;
	seeding_max_ratio: number;
	seeding_max_hours: number;
	seeding_mode: string;
	seeding_delete_files: boolean;
	seeding_skip_private: boolean;
	orphan_enabled: boolean;
	orphan_grace_minutes: number;
	orphan_delete_files: boolean;
	orphan_excluded_categories: string;
	hardlink_protection: boolean;
	skip_cross_seeds: boolean;
	cross_arr_sync: boolean;
	dry_run: boolean;
	protected_tags: string;
	search_on_remove: boolean;
	ignored_indexers: string;
	bandwidth_limit_bytes_per_sec: number;
	max_strikes_stalled: number;
	max_strikes_slow: number;
	max_strikes_metadata: number;
	max_strikes_paused: number;
	ignore_above_bytes: number;
	tag_instead_of_delete: boolean;
	obsolete_tag_label: string;
	failed_import_patterns: string;
	strike_queued: boolean;
	max_strikes_queued: number;
	search_cooldown_hours: number;
	max_searches_per_run: number;
	max_search_failures: number;
	blocklist_stalled: boolean;
	blocklist_slow: boolean;
	blocklist_metadata: boolean;
	blocklist_duplicate: boolean;
	blocklist_unregistered: boolean;
	ignored_download_clients: string;
	deletion_detection_enabled: boolean;
	unmonitored_cleanup_enabled: boolean;
	unregistered_enabled: boolean;
	max_strikes_unregistered: number;
	recheck_paused_enabled: boolean;
	recycle_bin_enabled: boolean;
	recycle_bin_path: string;
	ignored_release_groups: string;
	public_tracker_list: string;
	mismatch_enabled: boolean;
	max_strikes_mismatch: number;
	blocklist_mismatch: boolean;
	keep_archives: boolean;
	custom_unregistered_keywords: string;
	custom_mismatch_keywords: string;
}

export interface ScoringProfile {
	id: string;
	app_type: string;
	name: string;
	strategy: string;
	adequate_threshold: number;
	prefer_higher_quality: boolean;
	prefer_larger_size: boolean;
	prefer_indexer_flags: boolean;
	custom_format_weight: number;
	size_weight: number;
	age_weight: number;
	seeders_weight: number;
	resolution_weight: number;
	source_weight: number;
	hdr_weight: number;
	audio_weight: number;
	revision_bonus: number;
}

export interface SeedingRuleGroup {
	id: number;
	name: string;
	priority: number;
	match_type: string;
	match_pattern: string;
	max_ratio: number;
	max_hours: number;
	seeding_mode: string;
	skip_removal: boolean;
	delete_files: boolean;
}

export interface BlocklistSource {
	id: string;
	name: string;
	url: string;
	enabled: boolean;
	sync_interval_hours: number;
	last_synced_at: string | null;
	created_at: string;
}

export interface BlocklistRule {
	id: string;
	source_id: string | null;
	pattern: string;
	pattern_type: string;
	reason: string;
	enabled: boolean;
	created_at: string;
}

// ─── Dedup ──────────────────────────────────────────────────
export interface InstanceGroupMember {
	group_id: string;
	instance_id: string;
	instance_name?: string;
	quality_rank: number;
	is_independent: boolean;
}

export interface InstanceGroup {
	id: string;
	app_type: string;
	name: string;
	mode: string;
	created_at: string;
	members?: InstanceGroupMember[];
}

export interface CrossInstancePresence {
	media_id: string;
	instance_id: string;
	instance_name?: string;
	monitored: boolean;
	has_file: boolean;
}

export interface CrossInstanceMedia {
	id: string;
	group_id: string;
	external_id: string;
	title: string;
	detected_at: string;
	presence?: CrossInstancePresence[];
}

export interface CrossInstanceAction {
	id: string;
	group_id: string;
	external_id: string;
	title: string;
	action: string;
	reason: string;
	seerr_request_id?: number;
	source_instance_id?: string;
	target_instance_id?: string;
	executed_at: string;
}

export interface DuplicateFlag {
	request_id: number;
	media_title: string;
	external_id: string;
	request_type: string;
	is4k: boolean;
	requested_by: string;
	reason: string;
}

export interface DupScanResult {
	total_scanned: number;
	duplicates: DuplicateFlag[];
}

// ─── Seerr ──────────────────────────────────────────────────
export interface SeerrUser {
	id: number;
	displayName: string;
	email: string;
	avatar: string;
}

export interface SeerrMedia {
	id: number;
	mediaType: string;
	tmdbId: number;
	tvdbId: number | null;
	imdbId: string | null;
	status: number;
	status4k?: number;
	serviceUrl?: string | null;
}

export interface MediaRequest {
	id: number;
	status: number;
	type: string;
	is4k: boolean;
	isAutoRequest?: boolean;
	createdAt: string;
	updatedAt: string;
	requestedBy: SeerrUser;
	media: SeerrMedia;
}

export interface RequestCount {
	total: number;
	movie: number;
	tv: number;
	pending: number;
	approved: number;
	declined: number;
	processing: number;
	available: number;
}

// ─── Settings ───────────────────────────────────────────────
export interface GeneralSettings {
	secret_key: string;
	proxy_auth_bypass: boolean;
	ssl_verify: boolean;
	api_timeout: number;
	stateful_reset_hours: number;
	command_wait_delay: number;
	command_wait_attempts: number;
	max_download_queue_size: number;
	auto_import_interval_minutes: number;
}

export interface OIDCSettings {
	enabled: boolean;
	issuer_url: string;
	client_id: string;
	client_secret: string;
	redirect_url: string;
	scopes: string;
	auto_create: boolean;
	admin_group: string;
}

// ─── Users ──────────────────────────────────────────────────
export interface LurkarrUser {
	id: string;
	username: string;
	has_2fa: boolean;
	is_admin: boolean;
	auth_provider: string;
	created_at: string;
}

export interface UserSession {
	id: string;
	created_at: string;
	expires_at: string;
	ip_address: string;
	user_agent: string;
	current: boolean;
}

export interface Passkey {
	id: string;
	name: string;
	created_at: string;
}

// ─── Lurk Settings ──────────────────────────────────────────
export interface AppSettings {
	app_type: string;
	lurk_missing_count: number;
	lurk_upgrade_count: number;
	lurk_missing_mode: string;
	upgrade_mode: string;
	sleep_duration: number;
	monitored_only: boolean;
	skip_future: boolean;
	hourly_cap: number;
	selection_mode: string;
	max_search_failures: number;
	debug_mode: boolean;
}
