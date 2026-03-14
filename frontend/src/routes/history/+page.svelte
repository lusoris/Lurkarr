<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { appDisplayName, appColor, visibleAppTypes } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import { History, Trash2 } from 'lucide-svelte';

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
	let mounted = false;

	onMount(() => {
		load();
		mounted = true;
		return () => clearTimeout(debounceTimer);
	});

	// React to server-side search/filter/page changes (debounced for search, immediate for page/filter)
	let prevSearch = '';
	let prevFilterApp = '';
	let prevPage = 1;

	$effect(() => {
		// Track reactive deps
		const s = search;
		const f = filterApp;
		const p = page;
		if (!mounted) return;

		if (s !== prevSearch) {
			// Search changed — debounce and reset page
			prevSearch = s;
			clearTimeout(debounceTimer);
			debounceTimer = setTimeout(() => { page = 1; prevPage = 1; load(); }, 300);
		} else if (f !== prevFilterApp) {
			// Filter changed — immediate reload, reset page
			prevFilterApp = f;
			page = 1;
			prevPage = 1;
			load();
		} else if (p !== prevPage) {
			// Page changed — immediate reload
			prevPage = p;
			load();
		}
	});

	function onTabChange(tab: string) {
		activeTab = tab as ActiveTab;
		if (tab === 'cleaner' && blocklistItems.length === 0 && blocklistApp) loadBlocklist();
		if (tab === 'imports' && importItems.length === 0 && importApp) loadImports();
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

	// --- Column definitions ---
	const historyColumns: Column<HistoryItem>[] = [
		{ key: 'media_title', header: 'Media', sortable: true },
		{ key: 'app_type', header: 'App', accessor: (r) => appDisplayName(r.app_type) },
		{ key: 'instance_name', header: 'Instance' },
		{ key: 'operation', header: 'Operation', sortable: true },
		{ key: 'created_at', header: 'Date', sortable: true }
	];

	const blocklistColumns: Column<BlocklistEntry>[] = [
		{ key: 'title', header: 'Title', sortable: true },
		{ key: 'reason', header: 'Reason', sortable: true, accessor: (r) => reasonLabel(r.reason) },
		{ key: 'blocklisted_at', header: 'Date', sortable: true }
	];

	const importColumns: Column<ImportEntry>[] = [
		{ key: 'media_title', header: 'Media', sortable: true },
		{ key: 'action', header: 'Action', sortable: true },
		{ key: 'reason', header: 'Reason' },
		{ key: 'created_at', header: 'Date', sortable: true }
	];
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
			<DataTable
				data={items}
				columns={historyColumns}
				searchable
				searchPlaceholder="Search media titles..."
				bind:search
				pageSize={PAGE_SIZE}
				bind:page
				totalItems={total}
				noun="entries"
			>
				{#snippet toolbar()}
					<Select bind:value={filterApp} class="sm:w-48">
						<option value="">All Apps</option>
						{#each visibleAppTypes as app}
							<option value={app}>{appDisplayName(app)}</option>
						{/each}
					</Select>
				{/snippet}
				{#snippet row(item)}
					<T.Row>
						<T.Cell class="text-foreground">{item.media_title}</T.Cell>
						<T.Cell><Badge>{appDisplayName(item.app_type)}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground">{item.instance_name}</T.Cell>
						<T.Cell><Badge variant={item.operation === 'missing' ? 'warning' : 'info'}>{item.operation}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground text-xs">{new Date(item.created_at).toLocaleString()}</T.Cell>
					</T.Row>
				{/snippet}
			</DataTable>
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
			<DataTable data={blocklistItems} columns={blocklistColumns} searchable pageSize={50} noun="actions">
				{#snippet row(entry)}
					<T.Row>
						<T.Cell class="text-foreground max-w-xs truncate" title={entry.title}>{entry.title}</T.Cell>
						<T.Cell><Badge variant={reasonVariant(entry.reason)}>{reasonLabel(entry.reason)}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground text-xs">{new Date(entry.blocklisted_at).toLocaleString()}</T.Cell>
					</T.Row>
				{/snippet}
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
			<DataTable data={importItems} columns={importColumns} searchable pageSize={50} noun="imports">
				{#snippet row(entry)}
					<T.Row>
						<T.Cell class="text-foreground">{entry.media_title}</T.Cell>
						<T.Cell><Badge>{entry.action}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground">{entry.reason}</T.Cell>
						<T.Cell class="text-muted-foreground text-xs">{new Date(entry.created_at).toLocaleString()}</T.Cell>
					</T.Row>
				{/snippet}
			</DataTable>
		{/if}
	{/if}
</div>
