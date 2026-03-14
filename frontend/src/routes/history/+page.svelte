<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { appDisplayName, appColor, visibleAppTypes } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import DataTable from '$lib/components/ui/DataTable.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import { History, Trash2, ChevronLeft, ChevronRight } from 'lucide-svelte';

	const toasts = getToasts();
	const PAGE_SIZE = 50;

	type ActiveTab = 'lurking' | 'cleaner' | 'imports';
	let activeTab = $state<ActiveTab>('lurking');

	// --- Lurk History ---
	interface HistoryItem {
		id: number;
		app_type: string;
		instance_name: string;
		media_title: string;
		operation: string;
		created_at: string;
	}

	let items = $state<HistoryItem[]>([]);
	let total = $state(0);
	let search = $state('');
	let filterApp = $state('');
	let page = $state(1);
	let loading = $state(true);
	let confirmDelete = $state<string | null>(null);

	// --- Blocklist Log (Queue Cleaner) ---
	interface BlocklistEntry {
		id: number;
		app_type: string;
		instance_id: string;
		download_id: string;
		title: string;
		reason: string;
		blocklisted_at: string;
	}

	let blocklistItems = $state<BlocklistEntry[]>([]);
	let blocklistApp = $state('');
	let blocklistLoading = $state(false);

	// --- Auto-Import Log ---
	interface ImportEntry {
		id: number;
		app_type: string;
		instance_id: string;
		media_id: number;
		media_title: string;
		queue_item_id: number;
		action: string;
		reason: string;
		created_at: string;
	}

	let importItems = $state<ImportEntry[]>([]);
	let importApp = $state('');
	let importLoading = $state(false);

	// --- Lurk History ---
	async function load() {
		loading = true;
		try {
			const params = new URLSearchParams();
			if (search) params.set('search', search);
			if (filterApp) params.set('app', filterApp);
			params.set('limit', String(PAGE_SIZE));
			params.set('offset', String((page - 1) * PAGE_SIZE));
			const q = params.toString() ? `?${params}` : '';
			const res = await api.get<{ items: HistoryItem[]; total: number }>(`/history${q}`);
			items = res.items ?? [];
			total = res.total ?? 0;
		} catch {
			items = [];
			total = 0;
		}
		loading = false;
	}

	async function deleteHistory(appType: string) {
		try {
			await api.del(`/history/${appType}`);
			toasts.success(`${appDisplayName(appType)} history cleared`);
			confirmDelete = null;
			await load();
		} catch {
			toasts.error(`Failed to clear ${appDisplayName(appType)} history`);
		}
	}

	// --- Blocklist Log ---
	async function loadBlocklist() {
		if (!blocklistApp) { blocklistItems = []; return; }
		blocklistLoading = true;
		try {
			blocklistItems = await api.get<BlocklistEntry[]>(`/queue/blocklist/${blocklistApp}`);
		} catch {
			blocklistItems = [];
		}
		blocklistLoading = false;
	}

	// --- Auto-Import Log ---
	async function loadImports() {
		if (!importApp) { importItems = []; return; }
		importLoading = true;
		try {
			importItems = await api.get<ImportEntry[]>(`/queue/imports/${importApp}`);
		} catch {
			importItems = [];
		}
		importLoading = false;
	}

	let debounceTimer: ReturnType<typeof setTimeout>;

	onMount(() => {
		load();
		return () => clearTimeout(debounceTimer);
	});

	function onSearch() {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(() => { page = 1; load(); }, 300);
	}

	function onFilterChange() {
		page = 1;
		load();
	}

	function onTabChange(tab: string) {
		activeTab = tab as ActiveTab;
		if (tab === 'cleaner' && blocklistItems.length === 0 && blocklistApp) loadBlocklist();
		if (tab === 'imports' && importItems.length === 0 && importApp) loadImports();
	}

	const totalPages = $derived(Math.max(1, Math.ceil(total / PAGE_SIZE)));

	function prevPage() {
		if (page > 1) { page--; load(); }
	}

	function nextPage() {
		if (page < totalPages) { page++; load(); }
	}

	// Unique app types present in current results
	const presentApps = $derived.by(() => {
		const s = new Set(items.map(i => i.app_type));
		return [...s];
	});

	function reasonLabel(reason: string): string {
		return reason
			.replace(/_/g, ' ')
			.replace(/^blocklist /, '')
			.replace(/max strikes$/, '(max strikes)');
	}

	function reasonVariant(reason: string): 'default' | 'warning' | 'info' | 'error' {
		if (reason.includes('max_strikes')) return 'error';
		if (reason.includes('duplicate')) return 'warning';
		if (reason.includes('blocklist')) return 'info';
		return 'default';
	}
</script>

<svelte:head><title>History - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="History" />

	<Tabs
		tabs={[
			{ value: 'lurking', label: 'Lurking' },
			{ value: 'cleaner', label: 'Queue Cleaner' },
			{ value: 'imports', label: 'Auto-Import' }
		]}
		bind:value={activeTab}
		onchange={onTabChange}
	/>

	<!-- ==================== Lurk History Tab ==================== -->
	{#if activeTab === 'lurking'}
		{#if total > 0}
			<span class="text-sm text-muted-foreground">{total.toLocaleString()} total</span>
		{/if}

		<!-- Filters -->
		<div class="flex flex-col sm:flex-row gap-3">
			<div class="flex-1">
				<Input bind:value={search} placeholder="Search media titles..." oninput={onSearch} />
			</div>
			<Select bind:value={filterApp} onchange={onFilterChange} class="sm:w-48">
				<option value="">All Apps</option>
				{#each visibleAppTypes as app}
					<option value={app}>{appDisplayName(app)}</option>
				{/each}
			</Select>
		</div>

		<!-- Delete by app type -->
		{#if presentApps.length > 0}
			<div class="flex flex-wrap gap-2">
				{#each presentApps as app}
					{#if confirmDelete === app}
						<div class="flex items-center gap-2 rounded-lg bg-destructive/10 border border-destructive/30 px-3 py-1.5">
							<span class="text-xs text-destructive">Delete all {appDisplayName(app)} history?</span>
							<button onclick={() => deleteHistory(app)} class="text-xs font-medium text-destructive hover:text-destructive/80">Yes</button>
							<button onclick={() => { confirmDelete = null; }} class="text-xs text-muted-foreground hover:text-foreground">No</button>
						</div>
					{:else}
						<button
							onclick={() => { confirmDelete = app; }}
							class="flex items-center gap-1.5 rounded-lg border border-border px-2.5 py-1.5 text-xs text-muted-foreground hover:text-foreground hover:border-destructive/50 transition-colors"
						>
							<Trash2 class="w-3 h-3" />
							<span class="{appColor(app)}">{appDisplayName(app)}</span>
						</button>
					{/if}
				{/each}
			</div>
		{/if}

		{#if loading}
			<Skeleton rows={5} height="h-14" />
		{:else if items.length === 0}
			<EmptyState icon={History} title="No history entries" description="Lurk history will appear here once Lurkarr starts searching." />
		{:else}
			<DataTable>
				<thead class="bg-muted/50 text-muted-foreground text-xs uppercase">
					<tr>
						<th class="px-4 py-3 text-left font-medium">Media</th>
						<th class="px-4 py-3 text-left font-medium">App</th>
						<th class="px-4 py-3 text-left font-medium">Instance</th>
						<th class="px-4 py-3 text-left font-medium">Operation</th>
						<th class="px-4 py-3 text-left font-medium">Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each items as item}
						<tr class="hover:bg-muted/30 transition-colors">
							<td class="px-4 py-3 text-foreground">{item.media_title}</td>
							<td class="px-4 py-3"><Badge>{appDisplayName(item.app_type)}</Badge></td>
							<td class="px-4 py-3 text-muted-foreground">{item.instance_name}</td>
							<td class="px-4 py-3">
								<Badge variant={item.operation === 'missing' ? 'warning' : 'info'}>{item.operation}</Badge>
							</td>
							<td class="px-4 py-3 text-muted-foreground text-xs">{new Date(item.created_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</DataTable>

			<!-- Pagination -->
			{#if totalPages > 1}
				<div class="flex items-center justify-between">
					<p class="text-sm text-muted-foreground">
						Page {page} of {totalPages}
					</p>
					<div class="flex items-center gap-1">
						<Button size="sm" variant="ghost" disabled={page <= 1} onclick={prevPage}>
							<ChevronLeft class="h-4 w-4" />
						</Button>
						<Button size="sm" variant="ghost" disabled={page >= totalPages} onclick={nextPage}>
							<ChevronRight class="h-4 w-4" />
						</Button>
					</div>
				</div>
			{/if}
		{/if}

	<!-- ==================== Queue Cleaner Log Tab ==================== -->
	{:else if activeTab === 'cleaner'}
		<div class="flex flex-col sm:flex-row gap-3">
			<Select bind:value={blocklistApp} onchange={loadBlocklist} class="sm:w-48">
				<option value="">Select App</option>
				{#each visibleAppTypes as app}
					<option value={app}>{appDisplayName(app)}</option>
				{/each}
			</Select>
		</div>

		{#if blocklistLoading}
			<Skeleton rows={5} height="h-14" />
		{:else if !blocklistApp}
			<EmptyState icon={History} title="Select an app" description="Choose an app type above to view its queue cleaner action log." />
		{:else if blocklistItems.length === 0}
			<EmptyState icon={History} title="No actions recorded" description="Queue cleaner actions for {appDisplayName(blocklistApp)} will appear here." />
		{:else}
			<p class="text-sm text-muted-foreground">{blocklistItems.length} recent actions</p>
			<DataTable>
				<thead class="bg-muted/50 text-muted-foreground text-xs uppercase">
					<tr>
						<th class="px-4 py-3 text-left font-medium">Title</th>
						<th class="px-4 py-3 text-left font-medium">Reason</th>
						<th class="px-4 py-3 text-left font-medium">Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each blocklistItems as entry}
						<tr class="hover:bg-muted/30 transition-colors">
							<td class="px-4 py-3 text-foreground text-sm truncate max-w-xs" title={entry.title}>{entry.title}</td>
							<td class="px-4 py-3">
								<Badge variant={reasonVariant(entry.reason)}>{reasonLabel(entry.reason)}</Badge>
							</td>
							<td class="px-4 py-3 text-muted-foreground text-xs">{new Date(entry.blocklisted_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</DataTable>
		{/if}

	<!-- ==================== Auto-Import Log Tab ==================== -->
	{:else if activeTab === 'imports'}
		<div class="flex flex-col sm:flex-row gap-3">
			<Select bind:value={importApp} onchange={loadImports} class="sm:w-48">
				<option value="">Select App</option>
				{#each visibleAppTypes as app}
					<option value={app}>{appDisplayName(app)}</option>
				{/each}
			</Select>
		</div>

		{#if importLoading}
			<Skeleton rows={5} height="h-14" />
		{:else if !importApp}
			<EmptyState icon={History} title="Select an app" description="Choose an app type above to view its auto-import log." />
		{:else if importItems.length === 0}
			<EmptyState icon={History} title="No imports recorded" description="Auto-import actions for {appDisplayName(importApp)} will appear here." />
		{:else}
			<p class="text-sm text-muted-foreground">{importItems.length} recent actions</p>
			<DataTable>
				<thead class="bg-muted/50 text-muted-foreground text-xs uppercase">
					<tr>
						<th class="px-4 py-3 text-left font-medium">Media</th>
						<th class="px-4 py-3 text-left font-medium">Action</th>
						<th class="px-4 py-3 text-left font-medium">Reason</th>
						<th class="px-4 py-3 text-left font-medium">Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-border">
					{#each importItems as entry}
						<tr class="hover:bg-muted/30 transition-colors">
							<td class="px-4 py-3 text-foreground">{entry.media_title}</td>
							<td class="px-4 py-3"><Badge>{entry.action}</Badge></td>
							<td class="px-4 py-3 text-muted-foreground text-sm">{entry.reason}</td>
							<td class="px-4 py-3 text-muted-foreground text-xs">{new Date(entry.created_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</DataTable>
		{/if}
	{/if}
</div>
