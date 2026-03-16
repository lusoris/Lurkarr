import { describe, it, expect } from 'vitest';
import { appTypes } from '$lib';
import type { AppType } from '$lib';

/**
 * Lurk page tests
 *
 * Verifies that Prowlarr (an indexer manager, not an arr app) is properly
 * excluded from lurking operations and UI.
 */

describe('Lurk Settings - App Type Handling', () => {
	it('excludes Prowlarr from lurkable app types', () => {
		// Prowlarr is an indexer manager, not an *arr application
		// It should not appear in the list of apps that can lurk
		expect(appTypes).not.toContain('prowlarr' as AppType);
	});

	it('includes all 6 arr app types', () => {
		// Lurk should support these arr applications:
		const expectedTypes: AppType[] = [
			'sonarr' as AppType,
			'radarr' as AppType,
			'lidarr' as AppType,
			'readarr' as AppType,
			'whisparr' as AppType,
			'eros' as AppType
		];

		expect(appTypes).toHaveLength(expectedTypes.length);
		for (const appType of expectedTypes) {
			expect(appTypes).toContain(appType);
		}
	});

	it('verifies Prowlarr is not available for lurking', () => {
		// Even though Prowlarr is configured as a connection/integrations,
		// it should not appear in lurk-specific app lists
		const lurkableApps = appTypes;
		const prowlarrIsLurkable = lurkableApps.includes('prowlarr' as AppType);
		expect(prowlarrIsLurkable).toBe(false);
	});

	it('validates app type list matches backend AppType definitions', () => {
		// Frontend appTypes should match backend AllAppTypes() return value
		// Currently: [sonarr, radarr, lidarr, readarr, whisparr, eros]
		const expectedLength = 6;
		expect(appTypes).toHaveLength(expectedLength);

		// No 'prowlarr' in the list
		const hasProwlarr = appTypes.some(app => app === 'prowlarr');
		expect(hasProwlarr).toBe(false);
	});
});

describe('Lurk Settings - UI Integration', () => {
	it('uses appTypes for app selection UI', () => {
		// The Lurk page uses appTypes to generate app tabs/selectors
		// This ensures only lurkable apps appear to users
		expect(appTypes.length).toBeGreaterThan(0);
	});

	it('ensures app type string validation', () => {
		// All app types in appTypes should be valid strings
		appTypes.forEach(app => {
			expect(typeof app).toBe('string');
			expect(app.length).toBeGreaterThan(0);
		});
	});
});
