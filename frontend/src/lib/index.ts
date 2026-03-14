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
	rtorrent: 'rTorrent',
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
	qbittorrent: 'text-blue-400',
	transmission: 'text-red-400',
	deluge: 'text-cyan-400',
	rtorrent: 'text-teal-400',
	nzbget: 'text-lime-400',
	seerr: 'text-purple-400'
};

export function appColor(appType: string): string {
	return colors[appType] ?? 'text-surface-300';
}

// Subtle accent border class per app — used on content areas to indicate the active app.
const accentBorders: Record<string, string> = {
	sonarr: 'border-l-sky-400',
	radarr: 'border-l-amber-400',
	lidarr: 'border-l-emerald-400',
	readarr: 'border-l-rose-400',
	whisparr: 'border-l-pink-400',
	eros: 'border-l-purple-400',
	prowlarr: 'border-l-orange-400',
	sabnzbd: 'border-l-yellow-400',
	qbittorrent: 'border-l-blue-400',
	transmission: 'border-l-red-400',
	deluge: 'border-l-cyan-400',
	rtorrent: 'border-l-teal-400',
	nzbget: 'border-l-lime-400',
	seerr: 'border-l-purple-400'
};

export function appAccentBorder(appType: string): string {
	return accentBorders[appType] ?? '';
}

// Background color class per app — used on selected buttons/tabs.
const bgColors: Record<string, string> = {
	sonarr: 'bg-sky-500',
	radarr: 'bg-amber-500',
	lidarr: 'bg-emerald-500',
	readarr: 'bg-rose-500',
	whisparr: 'bg-pink-500',
	eros: 'bg-purple-500',
	prowlarr: 'bg-orange-500',
	sabnzbd: 'bg-yellow-500',
	qbittorrent: 'bg-blue-500',
	transmission: 'bg-red-500',
	deluge: 'bg-cyan-500',
	rtorrent: 'bg-teal-500',
	nzbget: 'bg-lime-500',
	seerr: 'bg-purple-500'
};

export function appBgColor(appType: string): string {
	return bgColors[appType] ?? 'bg-primary';
}

// Hover background color per app — for button hover states.
const hoverBgColors: Record<string, string> = {
	sonarr: 'hover:bg-sky-600',
	radarr: 'hover:bg-amber-600',
	lidarr: 'hover:bg-emerald-600',
	readarr: 'hover:bg-rose-600',
	whisparr: 'hover:bg-pink-600',
	eros: 'hover:bg-purple-600',
	prowlarr: 'hover:bg-orange-600',
	sabnzbd: 'hover:bg-yellow-600',
	qbittorrent: 'hover:bg-blue-600',
	transmission: 'hover:bg-red-600',
	deluge: 'hover:bg-cyan-600',
	rtorrent: 'hover:bg-teal-600',
	nzbget: 'hover:bg-lime-600',
	seerr: 'hover:bg-purple-600'
};

// Full button class for an app-colored action button (save, submit, etc.)
export function appButtonClass(appType: string): string {
	const bg = bgColors[appType] ?? 'bg-primary';
	const hover = hoverBgColors[appType] ?? 'hover:bg-primary/90';
	return `${bg} ${hover} text-white shadow-sm`;
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
	rtorrent: '/logos/rtorrent.svg',
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
	rtorrent: 'https://rakshasa.github.io/rtorrent/',
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
