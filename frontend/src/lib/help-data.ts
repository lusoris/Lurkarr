// Contextual help data for each page. Used by both HelpDrawer
// and the global Help hub page.

export interface HelpTip {
	q: string;
	a: string;
}

export interface HelpSection {
	title: string;
	tips: HelpTip[];
}

export interface PageHelp {
	quickStart?: string[];
	sections: HelpSection[];
}

export const helpData: Record<string, PageHelp> = {
	dashboard: {
		quickStart: [
			'Check the connection summary to ensure all your apps are online.',
			'Hourly cap bars show how close you are to your search limits.',
			'Use the quick-action buttons to jump straight to configuration.',
		],
		sections: [
			{
				title: 'Overview',
				tips: [
					{ q: 'What is Lurkarr?', a: 'Lurkarr automates media management for the *arr stack — searching for missing/upgradeable media, cleaning download queues, deduplicating across instances, and forwarding Seerr requests.' },
					{ q: 'What do the cap bars show?', a: 'Each bar shows how many API searches you\'ve used this hour vs. your limit. Yellow (≥70%) means nearing the cap; red (≥90%) means nearly exhausted. Caps reset hourly.' },
					{ q: 'What is the activity feed?', a: 'Shows the 5 most recent actions — lurk operations, queue removals, blocklist additions, schedule runs, and notifications. Click "View All" for the full log.' },
				],
			},
		],
	},

	apps: {
		quickStart: [
			'Click "Add Connection" to register your first *arr app.',
			'Enter the URL and API key, then test the connection.',
			'Add download clients so Lurkarr can monitor active transfers.',
			'Create Instance Groups if you have multiple instances of the same app.',
		],
		sections: [
			{
				title: 'Arr Apps',
				tips: [
					{ q: 'How do I find my API key?', a: 'In each *arr app, go to Settings → General → Security. Copy the API Key field. For Prowlarr, it\'s in the same location.' },
					{ q: 'Can I have multiple instances of the same app?', a: 'Yes! For example, Sonarr + Sonarr Anime, or Radarr + Radarr 4K. Each gets its own lurk/queue settings and runs independently.' },
					{ q: 'What about Whisparr v2 vs v3?', a: 'Whisparr v2 is Sonarr-based (series). v3 "Eros" is Radarr-based (movies). Both are shown under Whisparr with a version selector.' },
				],
			},
			{
				title: 'Download Clients',
				tips: [
					{ q: 'Which download clients are supported?', a: 'Torrent: qBittorrent, Transmission, Deluge, rTorrent. Usenet: SABnzbd, NZBGet. Each needs a URL and either API key or username/password.' },
					{ q: 'What is the category field?', a: 'Sets the download category/label in your client. Useful for routing — e.g., "tv" for Sonarr, "movies" for Radarr. Leave blank to use the client\'s default.' },
				],
			},
			{
				title: 'Services',
				tips: [
					{ q: 'What is Prowlarr?', a: 'A centralized indexer manager. Lurkarr connects to monitor indexer health and optionally sync indexer configurations to your *arr apps.' },
					{ q: 'What is Bazarr?', a: 'A subtitle manager. Lurkarr connects to monitor wanted subtitles, subtitle health, and download history.' },
					{ q: 'What is Kapowarr?', a: 'A comic book library manager. Lurkarr connects to view library stats, download queue, and running tasks.' },
					{ q: 'What is Shoko?', a: 'An anime library manager (shokoanime.com). Lurkarr connects to monitor collection stats and series summaries.' },
					{ q: 'What is Seerr?', a: 'A request management tool (Overseerr/Jellyseerr). Lurkarr connects to sync and clean up media requests.' },
				],
			},
			{
				title: 'Instance Groups',
				tips: [
					{ q: 'What are Instance Groups?', a: 'Groups link multiple instances of the same app type for deduplication. For example, Radarr + Radarr 4K can be grouped so Lurkarr detects when the same movie exists in both.' },
					{ q: 'What modes are available?', a: 'Quality Hierarchy: rank instances by quality, keep the best copy. Overlap Detect: flag duplicates without auto-removal. Split Season: distribute seasons across instances.' },
					{ q: 'What is the independence flag?', a: 'Marking an instance as "independent" excludes it from automatic dedup actions while still tracking overlaps. Useful for archive instances.' },
				],
			},
		],
	},

	lurk: {
		quickStart: [
			'Select which app types to configure (Sonarr, Radarr, etc.).',
			'Set the search mode: Missing, Upgrade, or All.',
			'Adjust batch size and hourly caps to stay within indexer limits.',
			'Click Save to apply — schedules use these settings automatically.',
		],
		sections: [
			{
				title: 'Search Modes',
				tips: [
					{ q: 'What does "Lurk" mean?', a: 'Lurking is Lurkarr\'s term for automated searching. It queries your *arr instances for media that\'s missing or below your quality cutoff, then triggers indexer searches.' },
					{ q: 'Missing vs Upgrade vs All?', a: 'Missing: only media not yet downloaded. Upgrade: only media below your quality profile cutoff. All: both missing and upgradeable items.' },
				],
			},
			{
				title: 'Rate Limiting',
				tips: [
					{ q: 'What is the batch size?', a: 'How many items to search per run per instance. Smaller = gentler on indexers. Typical: 10–50. Set 0 for unlimited (not recommended).' },
					{ q: 'What are hourly caps?', a: 'Maximum total searches per hour. Prevents indexer bans. Monitor usage on the Monitoring page. Recommended: start at 100, increase as needed.' },
					{ q: 'What is the command delay?', a: 'Time in seconds between individual search commands sent to arr apps. Prevents overwhelming them with rapid-fire requests. 1–5 seconds is typical.' },
				],
			},
		],
	},

	queue: {
		quickStart: [
			'Enable the Queue Cleaner for each app type you want to manage.',
			'Configure stall detection thresholds (minutes, speed).',
			'Set the strike count — how many checks before removal.',
			'Enable "Dry Run" first to see what would be removed without taking action.',
		],
		sections: [
			{
				title: 'Stall Detection',
				tips: [
					{ q: 'How does stall detection work?', a: 'Lurkarr checks your queue periodically. If a download has no progress for the configured time (e.g., 30 minutes), it gets a strike. After enough strikes, it\'s removed.' },
					{ q: 'What about slow downloads?', a: 'Separately from stalls, you can set a minimum speed threshold. Downloads below this speed get strikes too. Useful for removing torrents stuck at 10 KB/s.' },
				],
			},
			{
				title: 'Strike System',
				tips: [
					{ q: 'Why strikes instead of immediate removal?', a: 'Temporary slowdowns happen. The strike system gives downloads multiple chances. A download might stall briefly then resume — strikes prevent premature removal.' },
					{ q: 'What happens at max strikes?', a: 'The download is removed from the queue. Optionally: blocklisted (preventing re-grab), re-searched (find a new version), and/or the media tagged for review.' },
				],
			},
			{
				title: 'Advanced',
				tips: [
					{ q: 'What are Seeding Rules?', a: 'Control when completed torrents are removed. Set max ratio (e.g., 2.0x) and/or max hours (e.g., 48h). Use Rule Groups to override per tracker, category, or tag.' },
					{ q: 'What is a Scoring Profile?', a: 'Ranks competing downloads by quality. Weights consider indexer type, seeders, age, and size. Use penalties for public trackers, old releases, or oversized files.' },
					{ q: 'What is "Tag on Removal"?', a: 'Tags the media item in your *arr app (e.g., "lurkarr-removed") instead of or in addition to blocklisting. Gives you a review trail.' },
				],
			},
		],
	},

	scheduling: {
		quickStart: [
			'Click "Add Schedule" to create a new automated task.',
			'Choose the app type and action (Lurk Missing, Lurk Upgrade, Clean Queue).',
			'Set time and days — leave days empty for every day.',
			'Schedules use your Lurk Settings for search parameters.',
		],
		sections: [
			{
				title: 'Schedules',
				tips: [
					{ q: 'What actions can I schedule?', a: 'Lurk Missing, Lurk Upgrade, Lurk All, and Clean Queue. Each targets a specific app type, so you can search Sonarr every 6 hours but Radarr once daily.' },
					{ q: 'How do I see past runs?', a: 'Click the clock icon on any schedule to see execution history — timestamps, success/failure, and error messages.' },
					{ q: 'What timezone is used?', a: 'Schedules use the server timezone. The displayed times reflect your browser timezone.' },
				],
			},
		],
	},

	downloads: {
		quickStart: [
			'Add download clients on the Connections page first.',
			'Active downloads from all clients appear here automatically.',
			'Progress bars show real-time status with auto-refresh.',
		],
		sections: [
			{
				title: 'Monitoring',
				tips: [
					{ q: 'What does this page show?', a: 'Real-time download progress across all configured clients — SABnzbd, qBittorrent, Transmission, etc. Shows file name, progress, speed, and ETA.' },
					{ q: 'Can I pause or remove downloads here?', a: 'This page is for monitoring. Use your download client\'s web UI for direct control, or configure Queue Cleaner rules for automated management.' },
				],
			},
		],
	},

	seerr: {
		quickStart: [
			'Configure Seerr connection on the Connections page first.',
			'Pending requests appear automatically once connected.',
			'Use "Cleanup Fulfilled" to remove completed requests.',
		],
		sections: [
			{
				title: 'Seerr Integration',
				tips: [
					{ q: 'What is Seerr?', a: 'A request management tool. Users submit media requests through Seerr, and Lurkarr can monitor, clean up fulfilled requests, and scan for duplicates.' },
					{ q: 'What does cleanup do?', a: 'Removes requests from Seerr that have been fulfilled (media downloaded and available). Configure delay (days after fulfillment) in the Seerr settings.' },
					{ q: 'What does duplicate scan do?', a: 'Compares Seerr requests against your *arr instances to find media existing in multiple places. Helps identify wasted storage.' },
				],
			},
		],
	},

	dedup: {
		sections: [
			{
				title: 'Deduplication',
				tips: [
					{ q: 'How do I set up Dedup?', a: 'First create Instance Groups on the Connections page. Set quality ranks per instance. Then come here and click "Scan for Overlaps" to find duplicates.' },
					{ q: 'What does "quality_rank" mean?', a: 'Higher rank = preferred quality. In Quality Hierarchy mode, the highest-ranked instance keeps the file; lower-ranked copies are flagged or removed.' },
					{ q: 'Is removal automatic?', a: 'Only in Quality Hierarchy mode with auto-remove enabled. Overlap Detect mode only flags — you decide what to do. Always test with dry-run first.' },
				],
			},
		],
	},

	notifications: {
		quickStart: [
			'Click "Add Provider" and select a notification service.',
			'Enter the service credentials (webhook URL, bot token, etc.).',
			'Select which events trigger notifications.',
			'Use "Test" to send a sample notification.',
		],
		sections: [
			{
				title: 'Providers',
				tips: [
					{ q: 'Which services are supported?', a: 'Discord (webhook), Telegram (bot), Gotify, Ntfy, Pushover, Email (SMTP), Apprise, and generic Webhook.' },
					{ q: 'How do templates work?', a: 'Notification text uses Go templates. Variables: {{.Title}}, {{.Message}}, {{.AppType}}, {{.Instance}}, {{.Type}}, {{index .Fields "key"}}. Leave blank for default format.' },
				],
			},
			{
				title: 'Events',
				tips: [
					{ q: 'What events are available?', a: 'Lurk completed/failed, queue item removed, item blocklisted, download stalled, import failed, schedule executed, health check failed, system errors.' },
					{ q: 'Can I filter per provider?', a: 'Yes! Each provider has its own event selection. Send critical alerts to Pushover but all events to Discord, for example.' },
				],
			},
		],
	},

	history: {
		sections: [
			{
				title: 'History',
				tips: [
					{ q: 'What does this page show?', a: 'All historical lurk operations — searches performed, results found, upgrades triggered. Filterable by app type, instance, and date range.' },
					{ q: 'Can I clear history?', a: 'Yes, use the "Clear History" button. This removes event records but does not affect your media or settings.' },
				],
			},
		],
	},

	monitoring: {
		sections: [
			{
				title: 'Monitoring',
				tips: [
					{ q: 'What is Lurk Statistics?', a: 'Shows total lurked and upgraded counts per instance. Reset to zero counters without affecting actual media.' },
					{ q: 'What are Hourly API Caps?', a: 'Tracks API hits per instance per hour. Helps ensure you stay within indexer rate limits.' },
					{ q: 'How do I set up Grafana?', a: 'Lurkarr exposes /metrics for Prometheus. Deploy files for Prometheus + Grafana dashboards are in the deploy/ directory. Run docker compose -f deploy/docker-compose.monitoring.yml up -d.' },
				],
			},
		],
	},

	settings: {
		quickStart: [
			'Configure lurking behaviour (auto-import, logging level).',
			'Set up API defaults and SSL verification.',
			'Optionally configure OIDC/SSO for centralized authentication.',
		],
		sections: [
			{
				title: 'General',
				tips: [
					{ q: 'What is Auto-Import?', a: 'Automatically discovers new *arr instances on your network. Set an interval and Lurkarr polls known endpoints for new instances to register.' },
					{ q: 'What does SSL Verify control?', a: 'When enabled, Lurkarr verifies TLS certificates for all outgoing API calls. Disable only if your services use self-signed certificates.' },
				],
			},
			{
				title: 'Authentication',
				tips: [
					{ q: 'How do I set up SSO?', a: 'Go to the Authentication tab. Enter your OIDC provider details (issuer URL, client ID, client secret). Supports Authentik, Authelia, Keycloak, and any OIDC-compatible provider.' },
					{ q: 'Can I disable local login?', a: 'Not recommended, but you can make OIDC the primary method. Local login serves as a fallback if your SSO provider is down.' },
				],
			},
		],
	},

	user: {
		quickStart: [
			'Review your profile and change your password.',
			'Enable Two-Factor Authentication for extra security.',
			'Register a Passkey for passwordless login.',
		],
		sections: [
			{
				title: 'Security',
				tips: [
					{ q: 'How do I enable 2FA?', a: 'Click "Enable 2FA" in the Security section. Scan the QR code with your authenticator app, enter the verification code, and save your recovery codes somewhere safe.' },
					{ q: 'What are recovery codes?', a: 'One-time-use backup codes for when you lose your authenticator device. Each code works once. Store them in a password manager or printed in a safe.' },
					{ q: 'How do Passkeys work?', a: 'Register a passkey (biometric, security key) on this page. At login, click "Sign in with Passkey" — no password needed. Works with fingerprint, Face ID, or hardware keys.' },
				],
			},
		],
	},
};
