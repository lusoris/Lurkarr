<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appTabLabel, appLogo, appColor, appAccentBorder, appBgColor, appButtonClass } from '$lib';
	import { SquarePen, Trash2 } from 'lucide-svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import InstanceSwitcher from '$lib/components/InstanceSwitcher.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';

	const toasts = getToasts();
	const store = getInstances();

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

	type Tab = 'cleaner' | 'scoring' | 'blocklist' | 'imports';

	let activeTab = $state<Tab>('cleaner');
	let selectedApp = $derived(store.selectedApp);
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

	// --- Column definitions ---

	const blocklistLogColumns: Column<BlocklistLog>[] = [
		{ key: 'title', header: 'Title', sortable: true },
		{ key: 'reason', header: 'Reason', sortable: true },
		{ key: 'blocklisted_at', header: 'Date', sortable: true }
	];

	const importLogColumns: Column<AutoImportLog>[] = [
		{ key: 'media_title', header: 'Media', sortable: true },
		{ key: 'action', header: 'Action', sortable: true },
		{ key: 'reason', header: 'Reason' },
		{ key: 'created_at', header: 'Date', sortable: true }
	];

	const ruleColumns: Column<BlocklistRule>[] = [
		{ key: 'pattern', header: 'Pattern', sortable: true },
		{ key: 'pattern_type', header: 'Type', sortable: true },
		{ key: 'reason', header: 'Reason' },
		{ key: 'source_id', header: 'Source' },
		{ key: '_actions', header: 'Actions', headerClass: 'text-right' }
	];


</script>

<svelte:head><title>Queue Management - Lurkarr</title></svelte:head>

<div class="space-y-4">
	<PageHeader title="Queue Management" description="Queue cleaner, scoring profiles, blocklist, and import history." />

	<!-- App selector -->
	<InstanceSwitcher showInstances={false} onchange={loadTabData} />

	<!-- Tab navigation -->
	<div class="flex gap-0.5 rounded-lg bg-muted/50 p-0.5 overflow-x-auto">
		{#each tabs as tab}
			<button
				onclick={() => { activeTab = tab.id; }}
				class="shrink-0 rounded-md px-2.5 py-1.5 text-xs font-medium transition-colors
					{activeTab === tab.id ? 'bg-secondary text-foreground' : 'text-muted-foreground hover:text-foreground'}"
			>{tab.label}</button>
		{/each}
	</div>

	<!-- Queue Cleaner Settings -->
	{#if activeTab === 'cleaner'}
		{@const settings = cleanerSettings[selectedApp]}
		{#if settings}
			<div class="space-y-3 border-l-2 {appAccentBorder(selectedApp)} pl-4">
				<Card>
					<Toggle bind:checked={settings.enabled} label="Enable Queue Cleaner" hint="Automatically manage stalled, slow, and failed downloads" />
					{#if settings.enabled}
						<div class="mt-2">
							<Toggle bind:checked={settings.dry_run} label="Dry-Run Mode" hint="Preview what would be removed without actually deleting anything — check your logs" />
						</div>
						<div class="mt-2">
							<Input bind:value={settings.protected_tags} type="text" label="Protected Tags" hint="Comma-separated tag names — items with these tags are never removed" />
						</div>
						<div class="mt-2">
							<Input bind:value={settings.ignored_indexers} type="text" label="Ignored Indexers" hint="Comma-separated indexer names — items from these indexers skip all cleanup" />
						</div>
					{/if}
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Stall Detection</h3>
					<div class="grid grid-cols-2 gap-3">
						<Input bind:value={settings.stalled_threshold_minutes} type="number" label="Stalled (min)" hint="No progress threshold" />
						<Input bind:value={settings.slow_threshold_bytes_per_sec} type="number" label="Slow (bytes/s)" hint="Below this = slow" />
						<Input bind:value={settings.slow_ignore_above_bytes} type="number" label="Ignore Slow Above" hint="0 = disabled" />
						<Input bind:value={settings.metadata_stuck_minutes} type="number" label="Metadata Stuck (min)" hint="0 = disabled" />
						<Input bind:value={settings.bandwidth_limit_bytes_per_sec} type="number" label="Bandwidth Limit (bytes/s)" hint="Skip slow detection when >80% saturated (0 = disabled)" />
					</div>
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Strike System</h3>
					<div class="grid grid-cols-2 gap-3 mb-3">
						<Input bind:value={settings.max_strikes} type="number" label="Max Strikes (global)" hint="Default for all reasons" />
						<Input bind:value={settings.strike_window_hours} type="number" label="Window (hours)" hint="Expiry time" />
					</div>
					<p class="text-xs text-muted-foreground mb-2">Per-reason overrides (0 = use global)</p>
					<div class="grid grid-cols-2 gap-3 mb-3">
						<Input bind:value={settings.max_strikes_stalled} type="number" label="Stalled" hint="Stalled torrents" />
						<Input bind:value={settings.max_strikes_slow} type="number" label="Slow" hint="Below speed threshold" />
						<Input bind:value={settings.max_strikes_metadata} type="number" label="Metadata Stuck" hint="No size info" />
						<Input bind:value={settings.max_strikes_paused} type="number" label="Paused" hint="Paused in SABnzbd" />
						<Input bind:value={settings.max_strikes_queued} type="number" label="Queued" hint="Stuck in queue" />
					</div>
					<Input bind:value={settings.ignore_above_bytes} type="number" label="Ignore Above (bytes)" hint="Skip stalled/slow/metadata for items above this size (0 = disabled)" class="mb-3" />
					<div class="space-y-2">
						<Toggle bind:checked={settings.strike_public} label="Strike Public Trackers" />
						<Toggle bind:checked={settings.strike_private} label="Strike Private Trackers" />
						<Toggle bind:checked={settings.strike_queued} label="Strike Queued Items" hint="Flag items stuck in queued state" />
					</div>
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Actions</h3>
					<div class="space-y-2">
						<Input bind:value={settings.check_interval_seconds} type="number" label="Check Interval (seconds)" />
						<Toggle bind:checked={settings.remove_from_client} label="Remove from Download Client" />
						<Toggle bind:checked={settings.blocklist_on_remove} label="Blocklist on Remove (global default)" hint="Fallback for reasons without a specific toggle below" />
						<details class="mt-1 pl-1 border-l-2 border-border">
							<summary class="text-xs text-muted-foreground cursor-pointer select-none py-1">Per-reason blocklist overrides</summary>
							<div class="space-y-2 pt-2 pl-2">
								<Toggle bind:checked={settings.blocklist_stalled} label="Blocklist Stalled" />
								<Toggle bind:checked={settings.blocklist_slow} label="Blocklist Slow" />
								<Toggle bind:checked={settings.blocklist_metadata} label="Blocklist Metadata Stuck" />
								<Toggle bind:checked={settings.blocklist_duplicate} label="Blocklist Duplicates" />
								<Toggle bind:checked={settings.blocklist_unregistered} label="Blocklist Unregistered" />
							</div>
						</details>
						<Toggle bind:checked={settings.search_on_remove} label="Re-search on Remove" hint="Trigger a new search when an item is removed (blocklist, stalled, failed import)" />
						{#if settings.search_on_remove}
							<Input bind:value={settings.search_cooldown_hours} type="number" label="Search Cooldown (hours)" hint="Min hours between re-searches for the same media (0 = no cooldown)" />
							<Input bind:value={settings.max_searches_per_run} type="number" label="Max Searches per Run" hint="Limit re-searches per cleanup cycle per instance (0 = unlimited)" />
							<Input bind:value={settings.max_search_failures} type="number" label="Max Search Failures" hint="Stop retrying items after this many consecutive failures (0 = no limit)" />
						{/if}
						<Toggle bind:checked={settings.tag_instead_of_delete} label="Tag Media on Removal" hint="Apply an obsolete tag to the media item when removing from queue" />
						{#if settings.tag_instead_of_delete}
							<Input bind:value={settings.obsolete_tag_label} label="Obsolete Tag Label" hint="Tag name applied in the *arr app (e.g. lurkarr-obsolete)" />
						{/if}
					</div>
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Failed Imports</h3>
					<div class="space-y-2">
						<Toggle bind:checked={settings.failed_import_remove} label="Remove Failed Imports" />
						<Toggle bind:checked={settings.failed_import_blocklist} label="Blocklist Failed Imports" />
						<Input bind:value={settings.failed_import_patterns} label="Message Patterns" hint="Comma-separated substrings to match (empty = built-in defaults: import failed, no files found, etc.)" />
					</div>
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Seeding Rules</h3>
					<Toggle bind:checked={settings.seeding_enabled} label="Enable Seeding Enforcement" />
					{#if settings.seeding_enabled}
						<div class="grid grid-cols-2 gap-3 mt-3">
							<Input bind:value={settings.seeding_max_ratio} type="number" label="Max Ratio" hint="0 = disabled" />
							<Input bind:value={settings.seeding_max_hours} type="number" label="Max Hours" hint="0 = disabled" />
						</div>
						<Select bind:value={settings.seeding_mode} label="Mode" class="mt-2">
							<option value="or">Either condition (OR)</option>
							<option value="and">Both conditions (AND)</option>
						</Select>
						<div class="space-y-2 mt-2">
							<Toggle bind:checked={settings.seeding_delete_files} label="Delete Files on Removal" />
							<Toggle bind:checked={settings.seeding_skip_private} label="Skip Private Trackers" />
						</div>
					{/if}
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Orphan Cleanup</h3>
					<Toggle bind:checked={settings.orphan_enabled} label="Enable Orphan Detection" />
					{#if settings.orphan_enabled}
						<div class="space-y-2 mt-3">
							<Input bind:value={settings.orphan_grace_minutes} type="number" label="Grace Period (minutes)" />
							<Toggle bind:checked={settings.orphan_delete_files} label="Delete Orphan Files" />
							<Input bind:value={settings.orphan_excluded_categories} label="Excluded Categories" hint="Comma-separated" />
						</div>
					{/if}
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Advanced</h3>
					<div class="space-y-2">
						<Toggle bind:checked={settings.hardlink_protection} label="Hardlink Protection" />
						<Toggle bind:checked={settings.skip_cross_seeds} label="Skip Cross-Seeded Torrents" />
						<Toggle bind:checked={settings.cross_arr_sync} label="Cross-Arr Blocklist Sync" />
					</div>
				</Card>

				<Button onclick={saveCleaner} loading={saving} class={appButtonClass(selectedApp)}>Save Cleaner Settings</Button>
			</div>
		{:else if loadedCleaners.has(selectedApp)}
			<Card>
				<p class="text-sm text-muted-foreground text-center py-4">No cleaner settings configured for {appDisplayName(selectedApp)}.</p>
			</Card>
		{:else}
			<Card>
				<p class="text-sm text-muted-foreground text-center py-4">Loading cleaner settings...</p>
			</Card>
		{/if}
	{:else if activeTab === 'scoring'}
		{@const profile = scoringProfiles[selectedApp]}
		{#if profile}
			<div class="space-y-3 border-l-2 {appAccentBorder(selectedApp)} pl-4">
				<Card>
					<div class="grid grid-cols-2 gap-3">
						<Input bind:value={profile.name} label="Profile Name" />
						<Select bind:value={profile.strategy} label="Strategy">
							<option value="highest">Highest Score</option>
							<option value="adequate">Adequate Threshold</option>
						</Select>
					</div>
					{#if profile.strategy === 'adequate'}
						<Input bind:value={profile.adequate_threshold} type="number" label="Adequate Threshold" />
					{/if}
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Preferences</h3>
					<div class="space-y-2">
						<Toggle bind:checked={profile.prefer_higher_quality} label="Prefer Higher Quality" />
						<Toggle bind:checked={profile.prefer_larger_size} label="Prefer Larger Size" />
						<Toggle bind:checked={profile.prefer_indexer_flags} label="Prefer Indexer Flags" />
					</div>
				</Card>

				<Card>
					<h3 class="text-xs font-semibold text-muted-foreground uppercase tracking-wider mb-3">Weights</h3>
					<div class="grid grid-cols-2 gap-3">
						<Input bind:value={profile.custom_format_weight} type="number" label="Custom Format" />
						<Input bind:value={profile.size_weight} type="number" label="Size" />
						<Input bind:value={profile.age_weight} type="number" label="Age" />
						<Input bind:value={profile.seeders_weight} type="number" label="Seeders" />
					</div>
				</Card>

				<Button onclick={saveScoring} loading={saving} class={appButtonClass(selectedApp)}>Save Scoring Profile</Button>
			</div>
		{:else if loadedScoring.has(selectedApp)}
			<Card>
				<p class="text-sm text-muted-foreground text-center py-4">No scoring profile configured for {appDisplayName(selectedApp)}.</p>
			</Card>
		{:else}
			<Skeleton rows={4} height="h-10" />
		{/if}
	{:else if activeTab === 'blocklist'}
		<div class="border-l-2 {appAccentBorder(selectedApp)} pl-4">
		{#if loading}
			<Skeleton rows={4} height="h-10" />
		{:else if blocklist.length === 0}
			<EmptyState title="No blocklist entries" description="Items removed from the queue will appear here." />
		{:else}
			<DataTable data={blocklist} columns={blocklistLogColumns} searchable pageSize={50} noun="entries">
				{#snippet row(entry)}
					<T.Row>
						<T.Cell class="text-foreground max-w-xs truncate">{entry.title}</T.Cell>
						<T.Cell><Badge variant="error">{entry.reason}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground text-xs">{new Date(entry.blocklisted_at).toLocaleString()}</T.Cell>
					</T.Row>
				{/snippet}
			</DataTable>
		{/if}
		</div>
	{:else if activeTab === 'imports'}
		<div class="border-l-2 {appAccentBorder(selectedApp)} pl-4">
		{#if loading}
			<Skeleton rows={4} height="h-10" />
		{:else if imports.length === 0}
			<EmptyState title="No import entries" description="Auto-imported items will appear here." />
		{:else}
			<DataTable data={imports} columns={importLogColumns} searchable pageSize={50} noun="imports">
				{#snippet row(entry)}
					<T.Row>
						<T.Cell class="text-foreground">{entry.media_title}</T.Cell>
						<T.Cell><Badge variant="info">{entry.action}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground max-w-xs truncate">{entry.reason}</T.Cell>
						<T.Cell class="text-muted-foreground text-xs">{new Date(entry.created_at).toLocaleString()}</T.Cell>
					</T.Row>
				{/snippet}
			</DataTable>
		{/if}
		</div>
	{/if}

	<!-- ──────────────────────────────────────────── -->
	<!-- Global Blocklist Management                  -->
	<!-- ──────────────────────────────────────────── -->
	{#if activeTab === 'blocklist'}
	<div class="border-t border-border pt-6 mt-6 space-y-4">
		<h2 class="text-base font-semibold text-foreground">Global Blocklist Management</h2>
		<p class="text-xs text-muted-foreground">Manage community blocklist sources and custom rules that apply across all apps.</p>

		<!-- Sources -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h3 class="text-sm font-semibold text-muted-foreground">Blocklist Sources</h3>
				<Button size="sm" onclick={() => { showAddSource = !showAddSource; }}>
					{showAddSource ? 'Cancel' : '+ Add Source'}
				</Button>
			</div>

			{#if showAddSource}
				<div class="mb-4 p-4 rounded-lg bg-muted/50 border border-border space-y-3">
					<Input bind:value={newSource.name} label="Name" placeholder="e.g. Trash Guides blocklist" />
					<Input bind:value={newSource.url} label="URL" placeholder="https://example.com/blocklist.txt" />
					<Input bind:value={newSource.sync_interval_hours} type="number" label="Sync Interval (hours)" />
					<div class="flex justify-end">
						<Button size="sm" onclick={createSource} loading={savingSource}>Add Source</Button>
					</div>
				</div>
			{/if}

			{#if !sourcesLoaded}
				<Skeleton rows={3} height="h-16" />
			{:else if sources.length === 0}
					<p class="text-sm text-muted-foreground py-4 text-center">No blocklist sources configured</p>
			{:else}
				<div class="space-y-2">
					{#each sources as src}
						{#if editingSource?.id === src.id}
							<div class="p-3 rounded-lg bg-muted/50 border border-border space-y-3">
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
							<div class="flex items-center justify-between p-3 rounded-lg bg-card border border-border">
								<div class="min-w-0 flex-1">
									<div class="flex items-center gap-2">
										<span class="text-sm font-medium text-foreground truncate">{src.name}</span>
										{#if !src.enabled}
											<Badge variant="default">Disabled</Badge>
										{/if}
									</div>
									<p class="text-xs text-muted-foreground mt-0.5 truncate">{src.url}</p>
									{#if src.last_synced_at}
										<p class="text-xs text-muted-foreground/50 mt-0.5">Last synced: {new Date(src.last_synced_at).toLocaleString()}</p>
									{/if}
								</div>
								<div class="flex items-center gap-1.5 ml-3 shrink-0">
									<button onclick={() => { editingSource = { ...src }; }} class="text-muted-foreground hover:text-foreground transition-colors" title="Edit">
									<SquarePen class="w-4 h-4" />
									</button>
									{#if confirmDeleteSource === src.id}
										<span class="flex items-center gap-1">
											<button onclick={() => { deleteSource(src.id); confirmDeleteSource = null; }} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
											<button onclick={() => confirmDeleteSource = null} class="rounded px-1.5 py-0.5 bg-secondary text-muted-foreground text-[10px] hover:bg-muted">No</button>
										</span>
							{:else}
										<button onclick={() => confirmDeleteSource = src.id} class="text-muted-foreground hover:text-red-400 transition-colors" title="Delete">
										<Trash2 class="w-4 h-4" />
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
				<h3 class="text-sm font-semibold text-muted-foreground">Custom Rules</h3>
				<Button size="sm" onclick={() => { showAddRule = !showAddRule; }}>
					{showAddRule ? 'Cancel' : '+ Add Rule'}
				</Button>
			</div>

			{#if showAddRule}
				<div class="mb-4 p-4 rounded-lg bg-muted/50 border border-border space-y-3">
					<Select bind:value={newRule.pattern_type} label="Pattern Type">
					<option value="title_contains">Title Contains</option>
					<option value="title_regex">Title Regex</option>
					<option value="release_group">Release Group</option>
					<option value="indexer">Indexer</option>
				</Select>
					<Input bind:value={newRule.pattern} label="Pattern" placeholder={newRule.pattern_type === 'title_regex' ? '.*YIFY.*' : newRule.pattern_type === 'release_group' ? 'YTS' : 'pattern'} />
					<Input bind:value={newRule.reason} label="Reason (optional)" placeholder="Why this pattern is blocked" />
					<div class="flex justify-end">
						<Button size="sm" onclick={createRule} loading={savingRule}>Add Rule</Button>
					</div>
				</div>
			{/if}

			{#if !rulesLoaded}
				<Skeleton rows={3} height="h-10" />
			{:else if rules.length === 0}
				<p class="text-sm text-muted-foreground py-4 text-center">No custom rules</p>
			{:else}
				<DataTable data={rules} columns={ruleColumns} searchable pageSize={50} noun="rules">
					{#snippet row(rule)}
						<T.Row>
							<T.Cell class="text-foreground font-mono text-xs max-w-xs truncate">{rule.pattern}</T.Cell>
							<T.Cell><Badge variant="default">{rule.pattern_type.replace('_', ' ')}</Badge></T.Cell>
							<T.Cell class="text-muted-foreground text-xs max-w-xs truncate">{rule.reason || '—'}</T.Cell>
							<T.Cell class="text-muted-foreground text-xs">
								{#if rule.source_id}
									{@const srcName = sources.find(s => s.id === rule.source_id)?.name}
									<Badge variant="info">{srcName ?? 'synced'}</Badge>
								{:else}
									<Badge variant="warning">manual</Badge>
								{/if}
							</T.Cell>
							<T.Cell class="text-right">
								{#if !rule.source_id}
									{#if confirmDeleteRule === rule.id}
										<span class="flex items-center justify-end gap-1">
											<button onclick={() => { deleteRule(rule.id); confirmDeleteRule = null; }} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
											<button onclick={() => confirmDeleteRule = null} class="rounded px-1.5 py-0.5 bg-secondary text-muted-foreground text-[10px] hover:bg-muted">No</button>
										</span>
									{:else}
										<button onclick={() => confirmDeleteRule = rule.id} class="text-muted-foreground hover:text-red-400 transition-colors" title="Delete">
											<Trash2 class="w-4 h-4" />
										</button>
									{/if}
								{:else}
									<span class="text-muted-foreground/50 text-xs">synced</span>
								{/if}
							</T.Cell>
						</T.Row>
					{/snippet}
				</DataTable>
			{/if}
		</Card>
	</div>
	{/if}
</div>
