<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appTabLabel } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';

	const toasts = getToasts();

	interface QueueCleanerSettings {
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
	}

	interface ScoringProfile {
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
	}

	interface BlocklistLog {
		id: number;
		app_type: string;
		instance_id: string;
		download_id: string;
		title: string;
		reason: string;
		blocklisted_at: string;
	}

	interface AutoImportLog {
		id: number;
		app_type: string;
		instance_id: string;
		media_title: string;
		action: string;
		reason: string;
		created_at: string;
	}

	interface BlocklistLog {
		id: number;
		app_type: string;
		instance_id: string;
		download_id: string;
		title: string;
		reason: string;
		blocklisted_at: string;
	}

	type Tab = 'cleaner' | 'scoring' | 'blocklist' | 'imports';

	let activeTab = $state<Tab>('cleaner');
	let selectedApp = $state<string>('sonarr');
	let cleanerSettings = $state<Record<string, QueueCleanerSettings>>({});
	let scoringProfiles = $state<Record<string, ScoringProfile>>({});
	let blocklist = $state<BlocklistLog[]>([]);
	let imports = $state<AutoImportLog[]>([]);
	let saving = $state(false);
	let loading = $state(false);
	let loadedCleaners = $state<Set<string>>(new Set());
	let loadedScoring = $state<Set<string>>(new Set());

	// --- Global Blocklist Management ---
	interface BlocklistSource {
		id: string;
		name: string;
		url: string;
		enabled: boolean;
		sync_interval_hours: number;
		last_synced_at: string | null;
		created_at: string;
	}

	interface BlocklistRule {
		id: string;
		source_id: string | null;
		pattern: string;
		pattern_type: string;
		reason: string;
		enabled: boolean;
		created_at: string;
	}

	let sources = $state<BlocklistSource[]>([]);
	let rules = $state<BlocklistRule[]>([]);
	let sourcesLoaded = $state(false);
	let rulesLoaded = $state(false);
	let showAddSource = $state(false);
	let showAddRule = $state(false);
	let editingSource = $state<BlocklistSource | null>(null);
	let newSource = $state({ name: '', url: '', enabled: true, sync_interval_hours: 24 });
	let newRule = $state({ pattern: '', pattern_type: 'title_contains', reason: '' });
	let savingSource = $state(false);
	let savingRule = $state(false);
	let confirmDeleteSource = $state<string | null>(null);
	let confirmDeleteRule = $state<string | null>(null);

	async function loadSources() {
		try {
			sources = await api.get<BlocklistSource[]>('/blocklist/sources');
		} catch { sources = []; }
		sourcesLoaded = true;
	}

	async function loadRules() {
		try {
			rules = await api.get<BlocklistRule[]>('/blocklist/rules');
		} catch { rules = []; }
		rulesLoaded = true;
	}

	async function createSource() {
		savingSource = true;
		try {
			await api.post('/blocklist/sources', newSource);
			toasts.success('Source added');
			newSource = { name: '', url: '', enabled: true, sync_interval_hours: 24 };
			showAddSource = false;
			await loadSources();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to create source');
		}
		savingSource = false;
	}

	async function updateSource(src: BlocklistSource) {
		try {
			await api.put(`/blocklist/sources/${src.id}`, src);
			toasts.success('Source updated');
			editingSource = null;
			await loadSources();
		} catch {
			toasts.error('Failed to update source');
		}
	}

	async function deleteSource(id: string) {
		try {
			await api.del(`/blocklist/sources/${id}`);
			toasts.success('Source deleted');
			await Promise.all([loadSources(), loadRules()]);
		} catch {
			toasts.error('Failed to delete source');
		}
	}

	async function createRule() {
		savingRule = true;
		try {
			await api.post('/blocklist/rules', { ...newRule, enabled: true });
			toasts.success('Rule added');
			newRule = { pattern: '', pattern_type: 'title_contains', reason: '' };
			showAddRule = false;
			await loadRules();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to create rule');
		}
		savingRule = false;
	}

	async function deleteRule(id: string) {
		try {
			await api.del(`/blocklist/rules/${id}`);
			toasts.success('Rule deleted');
			await loadRules();
		} catch {
			toasts.error('Failed to delete rule');
		}
	}

	// Load global blocklist data once on mount
	$effect(() => {
		if (!sourcesLoaded) loadSources();
		if (!rulesLoaded) loadRules();
	});

	async function loadCleaner(app: string) {
		try {
			cleanerSettings[app] = await api.get<QueueCleanerSettings>(`/queue/settings/${app}`);
		} catch { /* first load may 404 */ }
		loadedCleaners = new Set([...loadedCleaners, app]);
	}

	async function loadScoring(app: string) {
		try {
			scoringProfiles[app] = await api.get<ScoringProfile>(`/queue/scoring/${app}`);
		} catch { /* handled */ }
		loadedScoring = new Set([...loadedScoring, app]);
	}

	async function loadBlocklist(app: string) {
		loading = true;
		try {
			blocklist = await api.get<BlocklistLog[]>(`/queue/blocklist/${app}`);
		} catch {
			blocklist = [];
		}
		loading = false;
	}

	async function loadImports(app: string) {
		loading = true;
		try {
			imports = await api.get<AutoImportLog[]>(`/queue/imports/${app}`);
		} catch {
			imports = [];
		}
		loading = false;
	}

	async function saveCleaner() {
		const settings = cleanerSettings[selectedApp];
		if (!settings) return;
		saving = true;
		try {
			await api.put(`/queue/settings/${selectedApp}`, settings);
			toasts.success('Queue cleaner settings saved');
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
		saving = false;
	}

	async function saveScoring() {
		const profile = scoringProfiles[selectedApp];
		if (!profile) return;
		saving = true;
		try {
			await api.put(`/queue/scoring/${selectedApp}`, profile);
			toasts.success('Scoring profile saved');
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
		saving = false;
	}

	function loadTabData() {
		if (activeTab === 'cleaner') loadCleaner(selectedApp);
		else if (activeTab === 'scoring') loadScoring(selectedApp);
		else if (activeTab === 'blocklist') loadBlocklist(selectedApp);
		else if (activeTab === 'imports') loadImports(selectedApp);
	}

	$effect(() => { loadTabData(); });

	function formatBytes(bytes: number): string {
		if (bytes === 0) return '0 B';
		const k = 1024;
		const sizes = ['B', 'KB', 'MB', 'GB'];
		const i = Math.floor(Math.log(bytes) / Math.log(k));
		return `${(bytes / Math.pow(k, i)).toFixed(1)} ${sizes[i]}`;
	}

	const tabs: { id: Tab; label: string }[] = [
		{ id: 'cleaner', label: 'Queue Cleaner' },
		{ id: 'scoring', label: 'Scoring' },
		{ id: 'blocklist', label: 'Blocklist' },
		{ id: 'imports', label: 'Import Log' }
	];
</script>

<svelte:head><title>Queue Management - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Queue Management</h1>

	<!-- App selector -->
	<div class="flex gap-1 rounded-lg bg-surface-900 border border-surface-800 p-1 overflow-x-auto">
		{#each appTypes as app}
			<button
				onclick={() => { selectedApp = app; loadTabData(); }}
				class="shrink-0 rounded-md px-2 py-1.5 text-xs font-medium transition-colors
					{selectedApp === app ? 'bg-lurk-600 text-white' : 'text-surface-400 hover:text-surface-200 hover:bg-surface-800'}"
			>{appTabLabel(app)}</button>
		{/each}
	</div>

	<!-- Tab navigation -->
	<div class="flex gap-1 rounded-lg bg-surface-800/50 p-1 overflow-x-auto">
		{#each tabs as tab}
			<button
				onclick={() => { activeTab = tab.id; loadTabData(); }}
				class="shrink-0 rounded-md px-3 py-2 text-sm font-medium transition-colors
					{activeTab === tab.id ? 'bg-surface-700 text-surface-100' : 'text-surface-400 hover:text-surface-200'}"
			>{tab.label}</button>
		{/each}
	</div>

	<!-- Queue Cleaner Settings -->
	{#if activeTab === 'cleaner'}
		{@const settings = cleanerSettings[selectedApp]}
		{#if settings}
			<Card>
				<div class="space-y-4">
					<Toggle bind:checked={settings.enabled} label="Enable Queue Cleaner" hint="Automatically manage stalled, slow, and failed downloads" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Stall Detection</h3>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input bind:value={settings.stalled_threshold_minutes} type="number" label="Stalled Threshold (min)" hint="Minutes with no progress before a download is considered stalled" />
						<Input bind:value={settings.slow_threshold_bytes_per_sec} type="number" label="Slow Threshold (bytes/s)" hint="Downloads below this speed are flagged as slow" />
					</div>
					<Input bind:value={settings.slow_ignore_above_bytes} type="number" label="Ignore Slow Above (bytes, 0 = disabled)" hint="Skip slow detection for downloads larger than this size" />
					<Input bind:value={settings.metadata_stuck_minutes} type="number" label="Metadata Stuck (min, 0 = disabled)" hint="Remove downloads stuck importing metadata after this long" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Strike System</h3>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input bind:value={settings.max_strikes} type="number" label="Max Strikes" hint="Strikes before a download is removed and blocklisted" />
						<Input bind:value={settings.strike_window_hours} type="number" label="Strike Window (hours)" hint="Strikes expire after this many hours" />
					</div>
					<Toggle bind:checked={settings.strike_public} label="Strike Public Trackers" hint="Apply strike system to public tracker downloads" />
					<Toggle bind:checked={settings.strike_private} label="Strike Private Trackers" hint="Apply strike system to private tracker downloads" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Actions</h3>
					<Input bind:value={settings.check_interval_seconds} type="number" label="Check Interval (seconds)" hint="How often the queue cleaner scans for issues" />
					<Toggle bind:checked={settings.remove_from_client} label="Remove from Download Client" hint="Also remove the affected download from the client (qBit, SABnzbd, etc.)" />
					<Toggle bind:checked={settings.blocklist_on_remove} label="Blocklist on Remove" hint="Add removed releases to the blocklist to prevent re-download" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Failed Imports</h3>
					<Toggle bind:checked={settings.failed_import_remove} label="Remove Failed Imports" hint="Automatically remove downloads that failed to import" />
					<Toggle bind:checked={settings.failed_import_blocklist} label="Blocklist Failed Imports" hint="Blocklist releases that fail to import" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Seeding Rules</h3>
					<Toggle bind:checked={settings.seeding_enabled} label="Enable Seeding Enforcement" hint="Automatically manage completed downloads based on seeding rules" />
					{#if settings.seeding_enabled}
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
							<Input bind:value={settings.seeding_max_ratio} type="number" label="Max Ratio (0 = disabled)" />
							<Input bind:value={settings.seeding_max_hours} type="number" label="Max Hours (0 = disabled)" />
						</div>
						<label class="block">
							<span class="block text-sm font-medium text-surface-300 mb-1.5">Mode</span>
							<select bind:value={settings.seeding_mode} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
								<option value="or">Either condition (OR)</option>
								<option value="and">Both conditions (AND)</option>
							</select>
						</label>
						<Toggle bind:checked={settings.seeding_delete_files} label="Delete Files on Seeding Removal" />
						<Toggle bind:checked={settings.seeding_skip_private} label="Skip Private Trackers" />
					{/if}

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Orphan Cleanup</h3>
					<Toggle bind:checked={settings.orphan_enabled} label="Enable Orphan Detection" />
					{#if settings.orphan_enabled}
						<Input bind:value={settings.orphan_grace_minutes} type="number" label="Grace Period (minutes)" />
						<Toggle bind:checked={settings.orphan_delete_files} label="Delete Orphan Files" />
						<Input bind:value={settings.orphan_excluded_categories} label="Excluded Categories (comma-separated)" />
					{/if}

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Advanced</h3>
					<Toggle bind:checked={settings.hardlink_protection} label="Hardlink Protection" hint="Prevent removal of downloads that have active hardlinks" />
					<Toggle bind:checked={settings.skip_cross_seeds} label="Skip Cross-Seeded Torrents" hint="Exclude torrents detected as cross-seeds from cleanup" />
					<Toggle bind:checked={settings.cross_arr_sync} label="Cross-Arr Blocklist Sync" hint="Sync blocklist entries across all connected arr apps" />

					<Button onclick={saveCleaner} loading={saving}>Save Cleaner Settings</Button>
				</div>
			</Card>
		{:else if loadedCleaners.has(selectedApp)}
			<Card>
				<p class="text-sm text-surface-500 text-center py-4">No cleaner settings configured for {appDisplayName(selectedApp)}.</p>
			</Card>
		{:else}
			<Card>
				<p class="text-sm text-surface-500 text-center py-4">Loading cleaner settings...</p>
			</Card>
		{/if}
	{/if}

	<!-- Scoring Profile -->
	{#if activeTab === 'scoring'}
		{@const profile = scoringProfiles[selectedApp]}
		{#if profile}
			<Card>
				<div class="space-y-4">
					<Input bind:value={profile.name} label="Profile Name" />
					<label class="block">
						<span class="block text-sm font-medium text-surface-300 mb-1.5">Strategy</span>
						<select bind:value={profile.strategy} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
							<option value="highest">Highest Score</option>
							<option value="adequate">Adequate Threshold</option>
						</select>
					</label>
					{#if profile.strategy === 'adequate'}
						<Input bind:value={profile.adequate_threshold} type="number" label="Adequate Threshold" />
					{/if}

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Preferences</h3>
					<Toggle bind:checked={profile.prefer_higher_quality} label="Prefer Higher Quality" />
					<Toggle bind:checked={profile.prefer_larger_size} label="Prefer Larger Size" />
					<Toggle bind:checked={profile.prefer_indexer_flags} label="Prefer Indexer Flags" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Weights</h3>
					<div class="grid grid-cols-2 gap-4">
						<Input bind:value={profile.custom_format_weight} type="number" label="Custom Format Weight" />
						<Input bind:value={profile.size_weight} type="number" label="Size Weight" />
						<Input bind:value={profile.age_weight} type="number" label="Age Weight" />
						<Input bind:value={profile.seeders_weight} type="number" label="Seeders Weight" />
					</div>

					<Button onclick={saveScoring} loading={saving}>Save Scoring Profile</Button>
				</div>
			</Card>
		{:else if loadedScoring.has(selectedApp)}
			<Card>
				<p class="text-sm text-surface-500 text-center py-4">No scoring profile configured for {appDisplayName(selectedApp)}.</p>
			</Card>
		{:else}
			<Card>
				<div class="space-y-3 py-2">
					{#each Array(4) as _}
						<div class="h-10 rounded-lg bg-surface-800/50 animate-pulse"></div>
					{/each}
				</div>
			</Card>
		{/if}
	{/if}

	<!-- Blocklist Log -->
	{#if activeTab === 'blocklist'}
		{#if loading}
			<Card>
				<div class="space-y-3 py-2">
					{#each Array(4) as _}
						<div class="h-10 rounded-lg bg-surface-800/50 animate-pulse"></div>
					{/each}
				</div>
			</Card>
		{:else if blocklist.length === 0}
			<Card><p class="text-sm text-surface-500 text-center py-4">No blocklist entries</p></Card>
		{:else}
			<div class="rounded-xl border border-surface-800 overflow-hidden">
				<table class="w-full text-sm">
					<thead class="bg-surface-900 text-surface-400 text-xs uppercase">
						<tr>
							<th class="px-4 py-3 text-left">Title</th>
							<th class="px-4 py-3 text-left">Reason</th>
							<th class="px-4 py-3 text-left">Date</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-surface-800">
						{#each blocklist as entry}
							<tr class="hover:bg-surface-800/30 transition-colors">
								<td class="px-4 py-3 text-surface-100 max-w-xs truncate">{entry.title}</td>
								<td class="px-4 py-3"><Badge variant="error">{entry.reason}</Badge></td>
								<td class="px-4 py-3 text-surface-500 text-xs">{new Date(entry.blocklisted_at).toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}

	<!-- Auto-Import Log -->
	{#if activeTab === 'imports'}
		{#if loading}
			<Card>
				<div class="space-y-3 py-2">
					{#each Array(4) as _}
						<div class="h-10 rounded-lg bg-surface-800/50 animate-pulse"></div>
					{/each}
				</div>
			</Card>
		{:else if imports.length === 0}
			<Card><p class="text-sm text-surface-500 text-center py-4">No import entries</p></Card>
		{:else}
			<div class="rounded-xl border border-surface-800 overflow-hidden">
				<table class="w-full text-sm">
					<thead class="bg-surface-900 text-surface-400 text-xs uppercase">
						<tr>
							<th class="px-4 py-3 text-left">Media</th>
							<th class="px-4 py-3 text-left">Action</th>
							<th class="px-4 py-3 text-left">Reason</th>
							<th class="px-4 py-3 text-left">Date</th>
						</tr>
					</thead>
					<tbody class="divide-y divide-surface-800">
						{#each imports as entry}
							<tr class="hover:bg-surface-800/30 transition-colors">
								<td class="px-4 py-3 text-surface-100">{entry.media_title}</td>
								<td class="px-4 py-3"><Badge variant="info">{entry.action}</Badge></td>
								<td class="px-4 py-3 text-surface-400 max-w-xs truncate">{entry.reason}</td>
								<td class="px-4 py-3 text-surface-500 text-xs">{new Date(entry.created_at).toLocaleString()}</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		{/if}
	{/if}

	<!-- ──────────────────────────────────────────── -->
	<!-- Global Blocklist Management                  -->
	<!-- ──────────────────────────────────────────── -->
	<div class="border-t border-surface-800 pt-8 mt-8 space-y-6">
		<h2 class="text-lg font-semibold text-surface-200">Global Blocklist Management</h2>
		<p class="text-sm text-surface-400">Manage community blocklist sources and custom rules that apply across all apps.</p>

		<!-- Sources -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-semibold text-surface-300">Blocklist Sources</h3>
				<Button size="sm" onclick={() => { showAddSource = !showAddSource; }}>
					{showAddSource ? 'Cancel' : '+ Add Source'}
				</Button>
			</div>

			{#if showAddSource}
				<div class="mb-4 p-4 rounded-lg bg-surface-800/50 border border-surface-700 space-y-3">
					<Input bind:value={newSource.name} label="Name" placeholder="e.g. Trash Guides blocklist" />
					<Input bind:value={newSource.url} label="URL" placeholder="https://example.com/blocklist.txt" />
					<Input bind:value={newSource.sync_interval_hours} type="number" label="Sync Interval (hours)" />
					<div class="flex justify-end">
						<Button size="sm" onclick={createSource} loading={savingSource}>Add Source</Button>
					</div>
				</div>
			{/if}

			{#if !sourcesLoaded}
				<div class="space-y-2 py-2">
					{#each Array(3) as _}
						<div class="h-16 rounded-lg bg-surface-800/50 animate-pulse"></div>
					{/each}
				</div>
			{:else if sources.length === 0}
				<p class="text-sm text-surface-500 py-4 text-center">No blocklist sources configured</p>
			{:else}
				<div class="space-y-2">
					{#each sources as src}
						{#if editingSource?.id === src.id}
							<div class="p-3 rounded-lg bg-surface-800/50 border border-surface-700 space-y-3">
								<Input bind:value={editingSource.name} label="Name" />
								<Input bind:value={editingSource.url} label="URL" />
								<Input bind:value={editingSource.sync_interval_hours} type="number" label="Sync Interval (hours)" />
								<Toggle bind:checked={editingSource.enabled} label="Enabled" />
								<div class="flex justify-end gap-2">
									<Button size="sm" variant="ghost" onclick={() => { editingSource = null; }}>Cancel</Button>
									<Button size="sm" onclick={() => updateSource(editingSource!)}>Save</Button>
								</div>
							</div>
						{:else}
							<div class="flex items-center justify-between p-3 rounded-lg bg-surface-900 border border-surface-800">
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<span class="text-sm font-medium text-surface-100 truncate">{src.name}</span>
										{#if !src.enabled}
											<Badge variant="default">Disabled</Badge>
										{/if}
									</div>
									<p class="text-xs text-surface-500 mt-0.5 truncate">{src.url}</p>
									{#if src.last_synced_at}
										<p class="text-xs text-surface-600 mt-0.5">Last synced: {new Date(src.last_synced_at).toLocaleString()}</p>
									{/if}
								</div>
								<div class="flex items-center gap-1.5 ml-3 shrink-0">
									<button onclick={() => { editingSource = { ...src }; }} class="text-surface-400 hover:text-surface-200 transition-colors" title="Edit">
										<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"/></svg>
									</button>
									{#if confirmDeleteSource === src.id}
										<span class="flex items-center gap-1">
											<button onclick={() => { deleteSource(src.id); confirmDeleteSource = null; }} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
											<button onclick={() => confirmDeleteSource = null} class="rounded px-1.5 py-0.5 bg-surface-700 text-surface-300 text-[10px] hover:bg-surface-600">No</button>
										</span>
							{:else}
										<button onclick={() => confirmDeleteSource = src.id} class="text-surface-400 hover:text-red-400 transition-colors" title="Delete">
											<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
										</button>
							{/if}
								</div>
							</div>
						{/if}
					{/each}
				</div>
			{/if}
		</Card>

		<!-- Rules -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-semibold text-surface-300">Custom Rules</h3>
				<Button size="sm" onclick={() => { showAddRule = !showAddRule; }}>
					{showAddRule ? 'Cancel' : '+ Add Rule'}
				</Button>
			</div>

			{#if showAddRule}
				<div class="mb-4 p-4 rounded-lg bg-surface-800/50 border border-surface-700 space-y-3">
					<label class="block">
						<span class="block text-sm font-medium text-surface-300 mb-1.5">Pattern Type</span>
						<select bind:value={newRule.pattern_type} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
							<option value="title_contains">Title Contains</option>
							<option value="title_regex">Title Regex</option>
							<option value="release_group">Release Group</option>
							<option value="indexer">Indexer</option>
						</select>
					</label>
					<Input bind:value={newRule.pattern} label="Pattern" placeholder={newRule.pattern_type === 'title_regex' ? '.*YIFY.*' : newRule.pattern_type === 'release_group' ? 'YTS' : 'pattern'} />
					<Input bind:value={newRule.reason} label="Reason (optional)" placeholder="Why this pattern is blocked" />
					<div class="flex justify-end">
						<Button size="sm" onclick={createRule} loading={savingRule}>Add Rule</Button>
					</div>
				</div>
			{/if}

			{#if !rulesLoaded}
				<div class="space-y-2 py-2">
					{#each Array(3) as _}
						<div class="h-10 rounded-lg bg-surface-800/50 animate-pulse"></div>
					{/each}
				</div>
			{:else if rules.length === 0}
				<p class="text-sm text-surface-500 py-4 text-center">No custom rules</p>
			{:else}
				<div class="rounded-xl border border-surface-800 overflow-hidden">
					<table class="w-full text-sm">
						<thead class="bg-surface-900 text-surface-400 text-xs uppercase">
							<tr>
								<th class="px-4 py-3 text-left">Pattern</th>
								<th class="px-4 py-3 text-left">Type</th>
								<th class="px-4 py-3 text-left">Reason</th>
								<th class="px-4 py-3 text-left">Source</th>
								<th class="px-4 py-3 text-right">Actions</th>
							</tr>
						</thead>
						<tbody class="divide-y divide-surface-800">
							{#each rules as rule}
								<tr class="hover:bg-surface-800/30 transition-colors">
									<td class="px-4 py-3 text-surface-100 font-mono text-xs max-w-xs truncate">{rule.pattern}</td>
									<td class="px-4 py-3"><Badge variant="default">{rule.pattern_type.replace('_', ' ')}</Badge></td>
									<td class="px-4 py-3 text-surface-400 text-xs max-w-xs truncate">{rule.reason || '—'}</td>
									<td class="px-4 py-3 text-surface-500 text-xs">
										{#if rule.source_id}
											{@const srcName = sources.find(s => s.id === rule.source_id)?.name}
											<Badge variant="info">{srcName ?? 'synced'}</Badge>
										{:else}
											<Badge variant="warning">manual</Badge>
										{/if}
									</td>
									<td class="px-4 py-3 text-right">
										{#if !rule.source_id}
											{#if confirmDeleteRule === rule.id}
												<span class="flex items-center gap-1">
													<button onclick={() => { deleteRule(rule.id); confirmDeleteRule = null; }} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
													<button onclick={() => confirmDeleteRule = null} class="rounded px-1.5 py-0.5 bg-surface-700 text-surface-300 text-[10px] hover:bg-surface-600">No</button>
												</span>
											{:else}
												<button onclick={() => confirmDeleteRule = rule.id} class="text-surface-400 hover:text-red-400 transition-colors" title="Delete">
													<svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
												</button>
											{/if}
										{:else}
											<span class="text-surface-600 text-xs">synced</span>
										{/if}
									</td>
								</tr>
							{/each}
						</tbody>
					</table>
				</div>
			{/if}
		</Card>
	</div>
</div>
