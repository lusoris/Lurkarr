// Canonical app types used by the backend.
export const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;
export type AppType = (typeof appTypes)[number];

// App types shown in the UI — whisparr and eros are merged into one "Whisparr" section.
export const visibleAppTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr'] as const;

// Service types for non-arr connections.
export type ServiceType = 'prowlarr' | 'sabnzbd' | 'seerr';

// Human-readable display name for each backend app type.
const displayNames: Record<string, string> = {
	sonarr: 'Sonarr',
	radarr: 'Radarr',
	lidarr: 'Lidarr',
	readarr: 'Readarr',
	whisparr: 'Whisparr',
	eros: 'Whisparr',
	prowlarr: 'Prowlarr',
	sabnzbd: 'SABnzbd',
	seerr: 'Seerr',
	qbittorrent: 'qBittorrent',
	transmission: 'Transmission',
	deluge: 'Deluge',
	nzbget: 'NZBGet'
};

// Short label for per-app tabs where whisparr and eros are listed separately.
const tabLabels: Record<string, string> = {
	...displayNames,
	whisparr: 'Whisparr v2',
	eros: 'Whisparr v3'
};

export function appDisplayName(appType: string): string {
	return displayNames[appType] ?? appType;
}

export function appTabLabel(appType: string): string {
	return tabLabels[appType] ?? appType;
}

// Consistent colors per app type.
const colors: Record<string, string> = {
	sonarr: 'text-sky-400',
	radarr: 'text-amber-400',
	lidarr: 'text-emerald-400',
	readarr: 'text-rose-400',
	whisparr: 'text-pink-400',
	eros: 'text-purple-400',
	prowlarr: 'text-orange-400',
	sabnzbd: 'text-yellow-400',
	seerr: 'text-purple-400'
};

export function appColor(appType: string): string {
	return colors[appType] ?? 'text-surface-300';
}

// Logo paths for each app type (served from /logos/).
const logos: Record<string, string> = {
	sonarr: '/logos/sonarr.png',
	radarr: '/logos/radarr.png',
	lidarr: '/logos/lidarr.png',
	readarr: '/logos/readarr.png',
	whisparr: '/logos/whisparr.png',
	eros: '/logos/eros.png',
	prowlarr: '/logos/prowlarr.png',
	sabnzbd: '/logos/sabnzbd.svg',
	seerr: '/logos/seerr.png',
	qbittorrent: '/logos/qbittorrent.svg',
	transmission: '/logos/transmission.png',
	deluge: '/logos/deluge.png',
	nzbget: '/logos/nzbget.png'
};

export function appLogo(appType: string): string | undefined {
	return logos[appType];
}

// Official website URLs for each app.
const websites: Record<string, string> = {
	sonarr: 'https://sonarr.tv',
	radarr: 'https://radarr.video',
	lidarr: 'https://lidarr.audio',
	readarr: 'https://readarr.com',
	whisparr: 'https://whisparr.com',
	eros: 'https://github.com/Whisparr/Whisparr-Eros',
	prowlarr: 'https://prowlarr.com',
	sabnzbd: 'https://sabnzbd.org',
	seerr: 'https://seerr.dev',
	qbittorrent: 'https://www.qbittorrent.org',
	transmission: 'https://transmissionbt.com',
	deluge: 'https://deluge-torrent.org',
	nzbget: 'https://nzbget.com'
};

export function appWebsite(appType: string): string | undefined {
	return websites[appType];
}

const defaultUrls: Record<string, string> = {
	sonarr: 'http://sonarr:8989',
	radarr: 'http://radarr:7878',
	lidarr: 'http://lidarr:8686',
	readarr: 'http://readarr:8787',
	whisparr: 'http://whisparr:6969',
	eros: 'http://whisparr:6969',
	prowlarr: 'http://prowlarr:9696'
};

export function appPlaceholderUrl(appType: string): string {
	return defaultUrls[appType] ?? `http://${appType}:8080`;
}
