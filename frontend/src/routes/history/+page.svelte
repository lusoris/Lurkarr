<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import { appDisplayName, appColor, visibleAppTypes } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import { History, Trash2 } from 'lucide-svelte';
	import type { HistoryItem, BlocklistEntry, ImportEntry, StrikeEntry } from '$lib/types';

	const toasts = getToasts();
	const PAGE_SIZE = 50;

	type ActiveTab = 'lurking' | 'cleaner' | 'imports' | 'strikes';
	let activeTab = $state<ActiveTab>('lurking');

	// --- Lurk History ---
	let items = $state<HistoryItem[]>([]);
	let total = $state(0);
	let search = $state('');
	let filterApp = $state('');
	let page = $state(1);
	let loading = $state(true);
	let confirmDelete = $state<string | null>(null);

	// --- Blocklist Log (Queue Cleaner) ---
	let blocklistItems = $state<BlocklistEntry[]>([]);
	let blocklistApp = $state('');
	let blocklistLoading = $state(false);

	// --- Auto-Import Log ---
	let importItems = $state<ImportEntry[]>([]);
	let importApp = $state('');
	let importLoading = $state(false);

	// --- Strike Log ---
	let strikeItems = $state<StrikeEntry[]>([]);
	let strikeApp = $state('');
	let strikeLoading = $state(false);

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

	// --- Strike Log ---
	async function loadStrikes() {
		if (!strikeApp) { strikeItems = []; return; }
		strikeLoading = true;
		try {
			strikeItems = await api.get<StrikeEntry[]>(`/queue/strikes/${strikeApp}`);
		} catch {
			strikeItems = [];
		}
		strikeLoading = false;
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
		if (tab === 'strikes' && strikeItems.length === 0 && strikeApp) loadStrikes();
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

	function strikeReasonVariant(reason: string): 'default' | 'warning' | 'info' | 'error' {
		if (reason === 'unregistered') return 'error';
		if (reason === 'stalled') return 'warning';
		if (reason === 'slow') return 'info';
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

	const strikeColumns: Column<StrikeEntry>[] = [
		{ key: 'title', header: 'Title', sortable: true },
		{ key: 'reason', header: 'Reason', sortable: true },
		{ key: 'struck_at', header: 'Date', sortable: true }
	];
</script>

<svelte:head><title>History - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="History">
		{#snippet actions()}
			<HelpDrawer page="history" />
		{/snippet}
	</PageHeader>

	<Tabs
		tabs={[
			{ value: 'lurking', label: 'Lurking' },
			{ value: 'cleaner', label: 'Queue Cleaner' },
			{ value: 'imports', label: 'Auto-Import' },
			{ value: 'strikes', label: 'Strikes' }
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
					<ConfirmAction active={confirmDelete === app} message="Delete all {appDisplayName(app)} history?" onconfirm={() => deleteHistory(app)} oncancel={() => { confirmDelete = null; }}>
						<Button size="sm" variant="outline" class="h-auto px-2.5 py-1.5 text-xs" onclick={() => { confirmDelete = app; }}>
							<Trash2 class="w-3 h-3 mr-1.5" />
							<span class="{appColor(app)}">{appDisplayName(app)}</span>
						</Button>
					</ConfirmAction>
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

	<!-- ==================== Strike Log Tab ==================== -->
	{:else if activeTab === 'strikes'}
		<div class="flex flex-col sm:flex-row gap-3">
			<Select bind:value={strikeApp} onchange={loadStrikes} class="sm:w-48">
				<option value="">Select App</option>
				{#each visibleAppTypes as app}
					<option value={app}>{appDisplayName(app)}</option>
				{/each}
			</Select>
		</div>

		{#if strikeLoading}
			<Skeleton rows={5} height="h-14" />
		{:else if !strikeApp}
			<EmptyState icon={History} title="Select an app" description="Choose an app type above to view its strike log." />
		{:else if strikeItems.length === 0}
			<EmptyState icon={History} title="No strikes recorded" description="Strikes for {appDisplayName(strikeApp)} will appear here when the queue cleaner issues them." />
		{:else}
			<DataTable data={strikeItems} columns={strikeColumns} searchable pageSize={50} noun="strikes">
				{#snippet row(entry)}
					<T.Row>
						<T.Cell class="text-foreground max-w-xs truncate" title={entry.title}>{entry.title}</T.Cell>
						<T.Cell><Badge variant={strikeReasonVariant(entry.reason)}>{reasonLabel(entry.reason)}</Badge></T.Cell>
						<T.Cell class="text-muted-foreground text-xs">{new Date(entry.struck_at).toLocaleString()}</T.Cell>
					</T.Row>
				{/snippet}
			</DataTable>
		{/if}
	{/if}
</div>

<ScrollToTop />
