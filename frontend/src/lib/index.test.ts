import { describe, it, expect } from 'vitest';
import {
	appTypes,
	visibleAppTypes,
	appDisplayName,
	appTabLabel,
	appColor,
	appAccentBorder,
	appBgColor,
	appButtonClass,
	appLogo,
	appWebsite,
	appPlaceholderUrl
} from '$lib/index';

describe('appTypes', () => {
	it('contains the 6 canonical types', () => {
		expect(appTypes).toEqual(['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros']);
	});

	it('visibleAppTypes excludes eros', () => {
		expect(visibleAppTypes).toEqual(['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr']);
		expect(visibleAppTypes).not.toContain('eros');
	});
});

describe('appDisplayName', () => {
	it('returns correct display names', () => {
		expect(appDisplayName('sonarr')).toBe('Sonarr');
		expect(appDisplayName('radarr')).toBe('Radarr');
		expect(appDisplayName('lidarr')).toBe('Lidarr');
		expect(appDisplayName('readarr')).toBe('Readarr');
		expect(appDisplayName('whisparr')).toBe('Whisparr');
		expect(appDisplayName('eros')).toBe('Whisparr');
		expect(appDisplayName('prowlarr')).toBe('Prowlarr');
		expect(appDisplayName('sabnzbd')).toBe('SABnzbd');
		expect(appDisplayName('seerr')).toBe('Seerr');
	});

	it('falls back to raw string for unknown type', () => {
		expect(appDisplayName('unknown')).toBe('unknown');
	});
});

describe('appTabLabel', () => {
	it('distinguishes whisparr v2 and eros v3', () => {
		expect(appTabLabel('whisparr')).toBe('Whisparr v2');
		expect(appTabLabel('eros')).toBe('Whisparr v3');
	});

	it('falls back for unknown', () => {
		expect(appTabLabel('xyz')).toBe('xyz');
	});
});

describe('appColor', () => {
	it('returns tailwind classes for known types', () => {
		expect(appColor('sonarr')).toBe('text-sky-400');
		expect(appColor('radarr')).toBe('text-amber-400');
	});

	it('returns default for unknown type', () => {
		expect(appColor('unknown')).toBe('text-surface-300');
	});
});

describe('appLogo', () => {
	it('returns logo paths for known types', () => {
		expect(appLogo('sonarr')).toBe('/logos/sonarr.png');
		expect(appLogo('sabnzbd')).toBe('/logos/sabnzbd.svg');
	});

	it('returns undefined for unknown type', () => {
		expect(appLogo('unknown')).toBeUndefined();
	});
});

describe('appWebsite', () => {
	it('returns URLs for known types', () => {
		expect(appWebsite('sonarr')).toBe('https://sonarr.tv');
		expect(appWebsite('prowlarr')).toBe('https://prowlarr.com');
	});

	it('returns undefined for unknown type', () => {
		expect(appWebsite('unknown')).toBeUndefined();
	});
});

describe('appPlaceholderUrl', () => {
	it('returns default URLs for arr types', () => {
		expect(appPlaceholderUrl('sonarr')).toBe('http://sonarr:8989');
		expect(appPlaceholderUrl('radarr')).toBe('http://radarr:7878');
		expect(appPlaceholderUrl('prowlarr')).toBe('http://prowlarr:9696');
	});

	it('generates fallback for unknown type', () => {
		expect(appPlaceholderUrl('myapp')).toBe('http://myapp:8080');
	});
});

describe('appAccentBorder', () => {
	it('returns border class for known types', () => {
		expect(appAccentBorder('sonarr')).toBe('border-l-sky-400');
		expect(appAccentBorder('radarr')).toBe('border-l-amber-400');
		expect(appAccentBorder('lidarr')).toBe('border-l-emerald-400');
		expect(appAccentBorder('whisparr')).toBe('border-l-pink-400');
		expect(appAccentBorder('eros')).toBe('border-l-purple-400');
	});

	it('returns empty string for unknown type', () => {
		expect(appAccentBorder('unknown')).toBe('');
	});
});

describe('appBgColor', () => {
	it('returns bg class for known types', () => {
		expect(appBgColor('sonarr')).toBe('bg-sky-500');
		expect(appBgColor('radarr')).toBe('bg-amber-500');
		expect(appBgColor('seerr')).toBe('bg-purple-500');
	});

	it('returns default bg-primary for unknown type', () => {
		expect(appBgColor('unknown')).toBe('bg-primary');
	});
});

describe('appButtonClass', () => {
	it('returns full button class for known app', () => {
		const sonarr = appButtonClass('sonarr');
		expect(sonarr).toContain('bg-sky-500');
		expect(sonarr).toContain('hover:bg-sky-600');
		expect(sonarr).toContain('text-white');
		expect(sonarr).toContain('shadow-sm');
	});

	it('returns primary classes for unknown app', () => {
		const result = appButtonClass('unknown');
		expect(result).toContain('bg-primary');
		expect(result).toContain('hover:bg-primary/90');
		expect(result).toContain('text-white');
	});

	it('returns correct classes for all arr types', () => {
		expect(appButtonClass('radarr')).toContain('bg-amber-500');
		expect(appButtonClass('lidarr')).toContain('bg-emerald-500');
		expect(appButtonClass('readarr')).toContain('bg-rose-500');
		expect(appButtonClass('whisparr')).toContain('bg-pink-500');
		expect(appButtonClass('eros')).toContain('bg-purple-500');
	});
});
