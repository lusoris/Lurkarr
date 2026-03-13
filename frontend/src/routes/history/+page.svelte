<script lang="ts">
	import { api } from '$lib/api';
	import { appDisplayName, appColor, visibleAppTypes } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const toasts = getToasts();

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
	let loading = $state(true);
	let confirmDelete = $state<string | null>(null);

	async function load() {
		loading = true;
		try {
			const params = new URLSearchParams();
			if (search) params.set('search', search);
			if (filterApp) params.set('app', filterApp);
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

	$effect(() => { load(); });

	let debounceTimer: ReturnType<typeof setTimeout>;
	function onSearch() {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(load, 300);
	}

	// Unique app types present in current results
	const presentApps = $derived(() => {
		const s = new Set(items.map(i => i.app_type));
		return [...s];
	});
</script>

<svelte:head><title>History - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-foreground">Lurk History</h1>
		{#if total > 0}
			<span class="text-sm text-muted-foreground">{total.toLocaleString()} total</span>
		{/if}
	</div>

	<div class="flex flex-col sm:flex-row gap-3">
		<div class="flex-1">
			<Input bind:value={search} placeholder="Search media titles..." oninput={onSearch} />
		</div>
		<select
			bind:value={filterApp}
			onchange={load}
			class="rounded-lg border border-input bg-card text-foreground px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-ring focus:ring-ring"
		>
			<option value="">All Apps</option>
			{#each visibleAppTypes as app}
				<option value={app}>{appDisplayName(app)}</option>
			{/each}
		</select>
	</div>

	<!-- Delete by app type -->
	{#if presentApps().length > 0}
		<div class="flex flex-wrap gap-2">
			{#each presentApps() as app}
				{#if confirmDelete === app}
					<div class="flex items-center gap-2 rounded-lg bg-red-900/30 border border-red-800 px-3 py-1.5">
						<span class="text-xs text-red-300">Delete all {appDisplayName(app)} history?</span>
						<button onclick={() => deleteHistory(app)} class="text-xs font-medium text-red-400 hover:text-red-300">Yes</button>
						<button onclick={() => { confirmDelete = null; }} class="text-xs text-muted-foreground hover:text-foreground">No</button>
					</div>
				{:else}
					<button
						onclick={() => { confirmDelete = app; }}
						class="flex items-center gap-1.5 rounded-lg border border-border px-2.5 py-1.5 text-xs text-muted-foreground hover:text-foreground hover:border-border transition-colors"
					>
						<svg class="w-3 h-3" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"/></svg>
						<span class="{appColor(app)}">{appDisplayName(app)}</span>
					</button>
				{/if}
			{/each}
		</div>
	{/if}

	{#if loading}
		<div class="space-y-2">
			{#each Array(5) as _}
				<div class="h-14 rounded-lg bg-muted/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if items.length === 0}
		<Card>
			<p class="text-sm text-muted-foreground text-center py-4">No history entries</p>
		</Card>
	{:else}
		<div class="rounded-xl border border-border overflow-x-auto">
			<table class="w-full text-sm min-w-[600px]">
				<thead class="bg-card text-muted-foreground text-xs uppercase">
					<tr>
						<th class="px-4 py-3 text-left">Media</th>
						<th class="px-4 py-3 text-left">App</th>
						<th class="px-4 py-3 text-left">Instance</th>
						<th class="px-4 py-3 text-left">Operation</th>
						<th class="px-4 py-3 text-left">Date</th>
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
			</table>
		</div>
	{/if}
</div>
