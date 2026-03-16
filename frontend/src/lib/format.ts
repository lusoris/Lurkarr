/** Format a byte count as a human-readable string (e.g. "1.5 GB"). */
export function formatBytes(bytes: number): string {
	if (!bytes || !isFinite(bytes) || bytes <= 0) return '0 B';
	const k = 1024;
	const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
	const i = Math.floor(Math.log(bytes) / Math.log(k));
	return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
}

/** Format bytes-per-second as a speed string (e.g. "4.2 MB/s"). */
export function formatSpeed(bytesPerSec: number): string {
	return `${formatBytes(bytesPerSec)}/s`;
}

/** Format seconds as a compact duration (e.g. "2h 15m", "45s"). */
export function formatETA(seconds: number): string {
	if (seconds <= 0) return '—';
	const h = Math.floor(seconds / 3600);
	const m = Math.floor((seconds % 3600) / 60);
	const s = seconds % 60;
	if (h > 0) return `${h}h ${m}m`;
	if (m > 0) return `${m}m ${s}s`;
	return `${s}s`;
}

/** Format a timestamp as a relative time string (e.g. "5m ago", "2d ago"). */
export function timeAgo(ts: string): string {
	const d = new Date(ts);
	const now = new Date();
	const diffMs = now.getTime() - d.getTime();
	const diffMin = Math.floor(diffMs / 60000);
	if (diffMin < 1) return 'just now';
	if (diffMin < 60) return `${diffMin}m ago`;
	const diffHr = Math.floor(diffMin / 60);
	if (diffHr < 24) return `${diffHr}h ago`;
	const diffDay = Math.floor(diffHr / 24);
	if (diffDay < 7) return `${diffDay}d ago`;
	return d.toLocaleDateString();
}

/** Ensure a version string is prefixed with "v". */
export function fmtVersion(v?: string): string {
	if (!v) return '';
	return v.startsWith('v') ? v : `v${v}`;
}
