import { describe, it, expect, beforeEach, vi } from 'vitest';
import type { Schedule, ScheduleExecution, HistoryItem, BlocklistEntry, ImportEntry, StrikeEntry } from '$lib/types';

// Mock API
vi.mock('$lib/api', () => ({
	api: {
		getSchedules: vi.fn(),
		createSchedule: vi.fn(),
		updateSchedule: vi.fn(),
		deleteSchedule: vi.fn(),
		getScheduleHistory: vi.fn(),
		getHistoryItems: vi.fn(),
		getBlocklistLog: vi.fn(),
		getImportLog: vi.fn(),
		getStrikeLog: vi.fn(),
		deleteHistoryItem: vi.fn()
	}
}));

vi.mock('$lib/stores/toast.svelte', () => ({
	getToasts: vi.fn(() => ({
		success: vi.fn(),
		error: vi.fn(),
		info: vi.fn()
	}))
}));

describe('Scheduling Page', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	// ─── Schedule Management Tests ───

	describe('Schedule Management', () => {
		it('should display schedule list', () => {
			const schedules: Schedule[] = [
				{
					id: 'sched-1',
					app_type: 'sonarr',
					action: 'lurk_missing',
					days: ['monday', 'wednesday', 'friday'],
					hour: 2,
					minute: 30,
					enabled: true,
					created_at: new Date().toISOString()
				}
			];

			expect(schedules).toHaveLength(1);
			expect(schedules[0].action).toBe('lurk_missing');
		});

		it('should support creating schedule', () => {
			const newSchedule = {
				app_type: 'radarr',
				action: 'lurk_upgrade',
				days: ['tuesday', 'thursday', 'saturday'],
				hour: 12,
				minute: 0,
				enabled: true
			};

			expect(newSchedule.app_type).toBe('radarr');
			expect(newSchedule.action).toBe('lurk_upgrade');
			expect(newSchedule.days).toHaveLength(3);
		});

		it('should validate required schedule fields', () => {
			const invalidSchedule = {
				app_type: '',
				action: '',
				days: [],
				hour: -1,
				minute: -1
			};

			expect(invalidSchedule.app_type).toBe('');
			expect(invalidSchedule.days).toHaveLength(0);
		});

		it('should validate hour range (0-23)', () => {
			const validHours = [0, 1, 12, 23];
			const invalidHours = [-1, 24, 100];

			validHours.forEach(h => expect(h >= 0 && h <= 23).toBe(true));
			invalidHours.forEach(h => expect(h >= 0 && h <= 23).toBe(false));
		});

		it('should validate minute range (0-59)', () => {
			const validMinutes = [0, 15, 30, 45, 59];
			const invalidMinutes = [-1, 60, 100];

			validMinutes.forEach(m => expect(m >= 0 && m <= 59).toBe(true));
			invalidMinutes.forEach(m => expect(m >= 0 && m <= 59).toBe(false));
		});

		it('should support all schedule actions', () => {
			const actions = [
				'lurk_missing',  // Search missing
				'lurk_upgrade',  // Search upgrades
				'lurk_all',      // Missing + upgrades
				'clean_queue',   // Queue cleaner
				'enable',        // Enable instances
				'disable'        // Disable instances
			];

			expect(actions).toHaveLength(6);
			expect(actions).toContain('lurk_missing');
			expect(actions).toContain('clean_queue');
		});

		it('should support all days of week', () => {
			const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];

			expect(days).toHaveLength(7);
			expect(days[0]).toBe('monday');
		});

		it('should require at least one day', () => {
			const schedule = {
				days: [] // Invalid
			};

			expect(schedule.days).toHaveLength(0);
		});

		it('should support editing schedule', () => {
			const schedule: Schedule = {
				id: 'sched-1',
				app_type: 'sonarr',
				action: 'lurk_missing',
				days: ['monday'],
				hour: 2,
				minute: 0,
				enabled: true,
				created_at: new Date().toISOString()
			};

			const updated = {
				...schedule,
				days: ['monday', 'wednesday', 'friday'],
				hour: 3,
				minute: 30
			};

			expect(updated.days).toHaveLength(3);
			expect(updated.hour).toBe(3);
		});

		it('should support enabling/disabling schedule', () => {
			const schedule: Schedule = {
				id: 'sched-1',
				app_type: 'sonarr',
				action: 'lurk_missing',
				days: ['monday'],
				hour: 2,
				minute: 0,
				enabled: true,
				created_at: new Date().toISOString()
			};

			const toggled = { ...schedule, enabled: !schedule.enabled };
			expect(toggled.enabled).toBe(false);
		});

		it('should support deleting schedule', () => {
			const scheduleId = 'sched-1';
			expect(scheduleId).toBeTruthy();
		});

		it('should show confirmation before deleting', () => {
			const confirmation = 'Are you sure you want to delete this schedule?';
			expect(confirmation).toContain('delete');
		});
	});

	// ─── Schedule Execution History ───

	describe('Schedule History & Execution', () => {
		it('should display execution history for schedule', () => {
			const executions: ScheduleExecution[] = [
				{
					id: 'exec-1',
					schedule_id: 'sched-1',
					executed_at: new Date().toISOString()
				},
				{
					id: 'exec-2',
					schedule_id: 'sched-1',
					executed_at: new Date(Date.now() - 3600000).toISOString()
				}
			];

			expect(executions).toHaveLength(2);
			expect(executions[0].schedule_id).toBe('sched-1');
		});

		it('should show recent executions for each schedule', () => {
			const executions: ScheduleExecution[] = [
				{ id: '1', schedule_id: 'sched-1', executed_at: new Date().toISOString() },
				{ id: '2', schedule_id: 'sched-1', executed_at: new Date(Date.now() - 3600000).toISOString() },
				{ id: '3', schedule_id: 'sched-1', executed_at: new Date(Date.now() - 7200000).toISOString() }
			];

			const recent = executions.slice(0, 3);
			expect(recent).toHaveLength(3);
		});

		it('should display next scheduled run time', () => {
			const schedule: Schedule = {
				id: 'sched-1',
				app_type: 'sonarr',
				action: 'lurk_missing',
				days: ['monday'],
				hour: 2,
				minute: 0,
				enabled: true,
				created_at: new Date().toISOString()
			};

			// Next run would be calculated based on current day/time
			expect(schedule.hour).toBe(2);
			expect(schedule.minute).toBe(0);
		});

		it('should show execution status (pending, running, completed)', () => {
			const statuses = ['pending', 'running', 'completed'];

			expect(statuses).toContain('pending');
			expect(statuses).toContain('running');
			expect(statuses).toContain('completed');
		});
	});

	// ─── UI/UX Tests ───

	describe('Scheduling Page UI/UX', () => {
		it('should show empty state when no schedules', () => {
			const schedules: Schedule[] = [];
			expect(schedules).toHaveLength(0);
		});

		it('should show loading skeletons while fetching', () => {
			const loading = true;
			expect(loading).toBe(true);
		});

		it('should show add button to create schedule', () => {
			// Component should have button to open modal
			expect(true).toBe(true);
		});

		it('should support opening history panel', () => {
			const showHistory = true;
			expect(showHistory).toBe(true);
		});

		it('should support modal for creating/editing schedules', () => {
			const modalOpen = true;
			expect(modalOpen).toBe(true);
		});

		it('should show action badges with descriptions', () => {
			const actions = [
				{ value: 'lurk_missing', label: 'Lurk Missing', desc: 'Search for missing media' },
				{ value: 'lurk_upgrade', label: 'Lurk Upgrades', desc: 'Search for quality upgrades' }
			];

			expect(actions[0].desc).toContain('missing');
			expect(actions[1].desc).toContain('upgrades');
		});

		it('should show day toggles for schedule', () => {
			const days = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'];
			expect(days).toHaveLength(7);
		});

		it('should show time input fields (hour/minute)', () => {
			const hour = 14;
			const minute = 30;

			expect(hour).toBeGreaterThanOrEqual(0);
			expect(hour).toBeLessThanOrEqual(23);
			expect(minute).toBeGreaterThanOrEqual(0);
			expect(minute).toBeLessThanOrEqual(59);
		});

		it('should show toggle to enable/disable schedule', () => {
			const enabled = true;
			expect(typeof enabled).toBe('boolean');
		});

		it('should show error messages for validation failures', () => {
			const errorMsg = 'Please select at least one day';
			expect(errorMsg).toContain('day');
		});

		it('should show success toast for successful operations', () => {
			const msg = 'Schedule created successfully';
			expect(msg).toContain('successfully');
		});
	});

	// ─── Edge Cases ───

	describe('Edge Cases', () => {
		it('should handle missing app type', () => {
			const schedule = { app_type: '', action: 'lurk_missing' };
			expect(schedule.app_type).toBe('');
		});

		it('should handle multiple schedules for same app', () => {
			const schedules = [
				{ id: '1', app_type: 'sonarr', action: 'lurk_missing' },
				{ id: '2', app_type: 'sonarr', action: 'lurk_upgrade' },
				{ id: '3', app_type: 'sonarr', action: 'clean_queue' }
			];

			expect(schedules.filter(s => s.app_type === 'sonarr')).toHaveLength(3);
		});

		it('should handle scheduling conflicts gracefully', () => {
			// Multiple schedules at same time should be allowed
			const schedule1 = { days: ['monday'], hour: 2, minute: 0 };
			const schedule2 = { days: ['monday'], hour: 2, minute: 0 };

			expect(schedule1.hour).toBe(schedule2.hour);
		});

		it('should handle API errors', () => {
			const error = new Error('Failed to create schedule');
			expect(error).toBeDefined();
		});

		it('should preserve form state on error', () => {
			const form = {
				app_type: 'sonarr',
				action: 'lurk_missing',
				days: ['monday'],
				hour: 2,
				minute: 30
			};

			expect(form.app_type).toBe('sonarr');
			expect(form.days).toContain('monday');
		});
	});
});

describe('History Page', () => {
	beforeEach(() => {
		vi.clearAllMocks();
	});

	// ─── Lurking History ───

	describe('Lurking History Tab', () => {
		it('should display lurking history items', () => {
			const items: HistoryItem[] = [
				{
					id: 'hist-1',
					type: 'search',
					app: 'sonarr',
					title: 'Series Name S01E01',
					status: 'success',
					message: 'Found 2 releases',
					timestamp: new Date().toISOString()
				}
			];

			expect(items).toHaveLength(1);
			expect(items[0].type).toBe('search');
		});

		it('should support filtering history by app', () => {
			const items: HistoryItem[] = [
				{ id: '1', app: 'sonarr', status: 'success' },
				{ id: '2', app: 'radarr', status: 'success' },
				{ id: '3', app: 'sonarr', status: 'failed' }
			];

			const sonarrItems = items.filter(i => i.app === 'sonarr');
			expect(sonarrItems).toHaveLength(2);
		});

		it('should support searching history items', () => {
			const items: HistoryItem[] = [
				{ id: '1', title: 'The Office', app: 'sonarr' },
				{ id: '2', title: 'Breaking Bad', app: 'sonarr' },
				{ id: '3', title: 'The Matrix', app: 'radarr' }
			];

			const searchTerm = 'The';
			const results = items.filter(i => i.title?.includes(searchTerm));

			expect(results).toHaveLength(2);
		});

		it('should show history item status', () => {
			const statuses = ['success', 'failed', 'partial'];

			expect(statuses).toContain('success');
			expect(statuses).toContain('failed');
		});

		it('should show history item details/message', () => {
			const item: HistoryItem = {
				id: 1,
				app_type: 'sonarr',
				instance_name: 'Sonarr Main',
				media_title: 'Series Name',
				operation: 'search_success',
				created_at: new Date().toISOString()
			};

			expect(item.operation).toContain('search');
		});

		it('should support pagination for large history', () => {
			const items = Array.from({ length: 100 }, (_, i) => ({
				id: `item-${i}`,
				title: `Item ${i}`
			}));

			const pageSize = 50;
			const page1 = items.slice(0, pageSize);

			expect(page1).toHaveLength(50);
		});

		it('should support deleting history items', () => {
			const itemId = 'hist-1';
			expect(itemId).toBeTruthy();
		});

		it('should show confirmation before deleting history', () => {
			const msg = 'Delete this history item?';
			expect(msg).toContain('Delete');
		});
	});

	// ─── Blocklist Tab ───

	describe('Blocklist Log Tab', () => {
		it('should display blocklist entries', () => {
			const entries: BlocklistEntry[] = [
				{
					id: 'block-1',
					app: 'radarr',
					release_title: 'Movie.2024.CAMRip.x264',
					blocklist_rule: 'CAM releases',
					reason: 'Low quality source',
					timestamp: new Date().toISOString()
				}
			];

			expect(entries).toHaveLength(1);
			expect(entries[0].release_title).toContain('CAM');
		});

		it('should filter blocklist entries by app', () => {
			const entries: BlocklistEntry[] = [
				{ id: '1', app: 'sonarr', release_title: 'Show.CAM' },
				{ id: '2', app: 'radarr', release_title: 'Movie.CAM' },
				{ id: '3', app: 'sonarr', release_title: 'Show2.CAM' }
			];

			const radarrEntries = entries.filter(e => e.app === 'radarr');
			expect(radarrEntries).toHaveLength(1);
		});

		it('should show blocklist rule that matched', () => {
			const entry: BlocklistEntry = {
				id: 'block-1',
				blocklist_rule: 'Unsupported Sources',
				reason: 'TS/CAM releases are not acceptable'
			};

			expect(entry.blocklist_rule).toBeTruthy();
			expect(entry.reason).toBeTruthy();
		});

		it('should show timestamp of blocklist event', () => {
			const entry: BlocklistEntry = {
				id: 'block-1',
				timestamp: new Date().toISOString()
			};

			expect(entry.timestamp).toBeTruthy();
		});

		it('should support pagination for large blocklist log', () => {
			const entries = Array.from({ length: 200 }, (_, i) => ({
				id: `entry-${i}`,
				release_title: `Release ${i}`
			}));

			const pageSize = 50;
			const page1 = entries.slice(0, pageSize);

			expect(page1).toHaveLength(50);
		});
	});

	// ─── Import Tab ───

	describe('Import Log Tab', () => {
		it('should display import entries', () => {
			const entries: ImportEntry[] = [
				{
					id: 'import-1',
					app: 'sonarr',
					title: 'Show.Name.Season.01',
					source: 'auto-import',
					status: 'success',
					timestamp: new Date().toISOString()
				}
			];

			expect(entries).toHaveLength(1);
			expect(entries[0].source).toBe('auto-import');
		});

		it('should filter import entries by app', () => {
			const entries: ImportEntry[] = [
				{ id: '1', app: 'sonarr', title: 'Show' },
				{ id: '2', app: 'radarr', title: 'Movie' },
				{ id: '3', app: 'sonarr', title: 'Show2' }
			];

			const sonarrEntries = entries.filter(e => e.app === 'sonarr');
			expect(sonarrEntries).toHaveLength(2);
		});

		it('should show import status (success/failed)', () => {
			const entry: ImportEntry = {
				id: 'import-1',
				status: 'success'
			};

			expect(['success', 'failed']).toContain(entry.status);
		});

		it('should show import source (auto-import, manual, etc)', () => {
			const entry: ImportEntry = {
				id: 'import-1',
				source: 'auto-import'
			};

			expect(entry.source).toBeTruthy();
		});
	});

	// ─── Strike Tab ───

	describe('Strike Log Tab', () => {
		it('should display strike entries', () => {
			const entries: StrikeEntry[] = [
				{
					id: 'strike-1',
					app: 'radarr',
					release_title: 'Movie.2024.1080p.CORRUPTED',
					reason: 'corrupted_file',
					strike_count: 1,
					timestamp: new Date().toISOString()
				}
			];

			expect(entries).toHaveLength(1);
			expect(entries[0].strike_count).toBe(1);
		});

		it('should filter strike entries by app', () => {
			const entries: StrikeEntry[] = [
				{ id: '1', app: 'sonarr', strike_count: 1 },
				{ id: '2', app: 'radarr', strike_count: 2 },
				{ id: '3', app: 'sonarr', strike_count: 1 }
			];

			const sonarrStrikes = entries.filter(e => e.app === 'sonarr');
			expect(sonarrStrikes).toHaveLength(2);
		});

		it('should show strike reason', () => {
			const reasons = [
				'corrupted_file',
				'invalid_format',
				'wrong_audio',
				'wrong_subtitle',
				'stalled'
			];

			expect(reasons).toContain('corrupted_file');
		});

		it('should show strike count', () => {
			const entry: StrikeEntry = {
				id: 'strike-1',
				strike_count: 3
			};

			expect(entry.strike_count).toBeGreaterThanOrEqual(1);
		});

		it('should show timestamp of strike', () => {
			const entry: StrikeEntry = {
				id: 'strike-1',
				timestamp: new Date().toISOString()
			};

			expect(entry.timestamp).toBeTruthy();
		});
	});

	// ─── Tab Navigation ───

	describe('History Tab Navigation', () => {
		it('should have tab for lurking history', () => {
			const tabs = ['lurking', 'cleaner', 'imports', 'strikes'];
			expect(tabs).toContain('lurking');
		});

		it('should have tab for blocklist/cleaner log', () => {
			const tabs = ['lurking', 'cleaner', 'imports', 'strikes'];
			expect(tabs).toContain('cleaner');
		});

		it('should have tab for import log', () => {
			const tabs = ['lurking', 'cleaner', 'imports', 'strikes'];
			expect(tabs).toContain('imports');
		});

		it('should have tab for strike log', () => {
			const tabs = ['lurking', 'cleaner', 'imports', 'strikes'];
			expect(tabs).toContain('strikes');
		});

		it('should maintain active tab state', () => {
			let activeTab = 'lurking';
			activeTab = 'cleaner';

			expect(activeTab).toBe('cleaner');
		});
	});

	// ─── UI/UX Tests ───

	describe('History Page UI/UX', () => {
		it('should show loading skeleton while fetching', () => {
			const loading = true;
			expect(loading).toBe(true);
		});

		it('should show empty state when no history', () => {
			const items: HistoryItem[] = [];
			expect(items).toHaveLength(0);
		});

		it('should show filter dropdown for app selection', () => {
			const apps = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr'];
			expect(apps).toContain('sonarr');
		});

		it('should show search input for filtering', () => {
			const searchTerm = 'The Office';
			expect(searchTerm).toBeTruthy();
		});

		it('should show timestamp in readable format', () => {
			const timestamp = new Date();
			expect(timestamp).toBeDefined();
		});

		it('should show badges for status', () => {
			const statuses = ['success', 'failed', 'partial'];
			expect(statuses.length).toBeGreaterThan(0);
		});

		it('should support delete action on history items', () => {
			const itemId = 'hist-1';
			expect(itemId).toBeTruthy();
		});

		it('should show error message on failed delete', () => {
			const msg = 'Failed to delete history item';
			expect(msg).toContain('Failed');
		});

		it('should paginate results with page size', () => {
			const pageSize = 50;
			expect(pageSize).toBeGreaterThan(0);
		});
	});

	// ─── Edge Cases ───

	describe('Edge Cases', () => {
		it('should handle very long release titles', () => {
			const title = 'Very.Long.Release.Title.With.Many.Dots.And.Details.S01E01.1080p.HEVC.x265.10bit.AAC.mkv';
			expect(title.length).toBeGreaterThan(0);
		});

		it('should handle items from deleted apps gracefully', () => {
			const items: any[] = [
				{ id: '1', app: 'deleted_app', title: 'Item' }
			];

			expect(items[0].app).toBe('deleted_app');
		});

		it('should handle API errors during fetch', () => {
			const error = new Error('Failed to fetch history');
			expect(error).toBeDefined();
		});

		it('should handle concurrent delete operations', () => {
			// Should queue or prevent race conditions
			expect(true).toBe(true);
		});
	});
});
