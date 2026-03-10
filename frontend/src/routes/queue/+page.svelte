<script lang="ts">
	import { api } from '$lib/api';
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

	const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;
	type Tab = 'cleaner' | 'scoring' | 'blocklist' | 'imports';

	let activeTab = $state<Tab>('cleaner');
	let selectedApp = $state<string>('sonarr');
	let cleanerSettings = $state<Record<string, QueueCleanerSettings>>({});
	let scoringProfiles = $state<Record<string, ScoringProfile>>({});
	let blocklist = $state<BlocklistLog[]>([]);
	let imports = $state<AutoImportLog[]>([]);
	let saving = $state(false);
	let loading = $state(false);

	async function loadCleaner(app: string) {
		try {
			cleanerSettings[app] = await api.get<QueueCleanerSettings>(`/queue/settings/${app}`);
		} catch { /* first load may 404 */ }
	}

	async function loadScoring(app: string) {
		try {
			scoringProfiles[app] = await api.get<ScoringProfile>(`/queue/scoring/${app}`);
		} catch { /* handled */ }
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

<div class="space-y-6 max-w-3xl">
	<h1 class="text-2xl font-bold text-surface-50">Queue Management</h1>

	<!-- App selector -->
	<div class="flex gap-1 rounded-lg bg-surface-900 border border-surface-800 p-1">
		{#each appTypes as app}
			<button
				onclick={() => { selectedApp = app; loadTabData(); }}
				class="flex-1 rounded-md px-2 py-1.5 text-xs font-medium capitalize transition-colors
					{selectedApp === app ? 'bg-lurk-600 text-white' : 'text-surface-400 hover:text-surface-200 hover:bg-surface-800'}"
			>{app}</button>
		{/each}
	</div>

	<!-- Tab navigation -->
	<div class="flex gap-1 rounded-lg bg-surface-800/50 p-1">
		{#each tabs as tab}
			<button
				onclick={() => { activeTab = tab.id; loadTabData(); }}
				class="flex-1 rounded-md px-3 py-2 text-sm font-medium transition-colors
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
					<Toggle bind:checked={settings.enabled} label="Enable Queue Cleaner" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Stall Detection</h3>
					<div class="grid grid-cols-2 gap-4">
						<Input bind:value={settings.stalled_threshold_minutes} type="number" label="Stalled Threshold (min)" />
						<Input bind:value={settings.slow_threshold_bytes_per_sec} type="number" label="Slow Threshold (bytes/s)" />
					</div>
					<Input bind:value={settings.slow_ignore_above_bytes} type="number" label="Ignore Slow Above (bytes, 0 = disabled)" />
					<Input bind:value={settings.metadata_stuck_minutes} type="number" label="Metadata Stuck (min, 0 = disabled)" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Strike System</h3>
					<div class="grid grid-cols-2 gap-4">
						<Input bind:value={settings.max_strikes} type="number" label="Max Strikes" />
						<Input bind:value={settings.strike_window_hours} type="number" label="Strike Window (hours)" />
					</div>
					<Toggle bind:checked={settings.strike_public} label="Strike Public Trackers" />
					<Toggle bind:checked={settings.strike_private} label="Strike Private Trackers" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Actions</h3>
					<Input bind:value={settings.check_interval_seconds} type="number" label="Check Interval (seconds)" />
					<Toggle bind:checked={settings.remove_from_client} label="Remove from Download Client" />
					<Toggle bind:checked={settings.blocklist_on_remove} label="Blocklist on Remove" />

					<h3 class="text-sm font-semibold text-surface-300 pt-2">Failed Imports</h3>
					<Toggle bind:checked={settings.failed_import_remove} label="Remove Failed Imports" />
					<Toggle bind:checked={settings.failed_import_blocklist} label="Blocklist Failed Imports" />

					<Button onclick={saveCleaner} loading={saving}>Save Cleaner Settings</Button>
				</div>
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
		{:else}
			<Card>
				<p class="text-sm text-surface-500 text-center py-4">Loading scoring profile...</p>
			</Card>
		{/if}
	{/if}

	<!-- Blocklist Log -->
	{#if activeTab === 'blocklist'}
		{#if loading}
			<Card><p class="text-sm text-surface-500 text-center py-4">Loading...</p></Card>
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
			<Card><p class="text-sm text-surface-500 text-center py-4">Loading...</p></Card>
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
</div>
