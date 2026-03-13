<script lang="ts">
	import { api } from '$lib/api';
	import { appDisplayName } from '$lib';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	interface HistoryItem {
		id: number;
		app_type: string;
		instance_name: string;
		media_title: string;
		operation: string;
		created_at: string;
	}

	let items = $state<HistoryItem[]>([]);
	let search = $state('');
	let loading = $state(true);

	async function load() {
		loading = true;
		try {
			const q = search ? `?search=${encodeURIComponent(search)}` : '';
			items = await api.get<HistoryItem[]>(`/history${q}`);
		} catch {
			items = [];
		}
		loading = false;
	}

	$effect(() => { load(); });

	let debounceTimer: ReturnType<typeof setTimeout>;
	function onSearch(e: Event) {
		clearTimeout(debounceTimer);
		debounceTimer = setTimeout(load, 300);
	}
</script>

<svelte:head><title>History - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Lurk History</h1>

	<Input bind:value={search} placeholder="Search media titles..." oninput={onSearch} />

	{#if loading}
		<div class="space-y-2">
			{#each Array(5) as _}
				<div class="h-14 rounded-lg bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if items.length === 0}
		<Card>
			<p class="text-sm text-surface-500 text-center py-4">No history entries</p>
		</Card>
	{:else}
		<div class="rounded-xl border border-surface-800 overflow-x-auto">
			<table class="w-full text-sm min-w-[600px]">
				<thead class="bg-surface-900 text-surface-400 text-xs uppercase">
					<tr>
						<th class="px-4 py-3 text-left">Media</th>
						<th class="px-4 py-3 text-left">App</th>
						<th class="px-4 py-3 text-left">Instance</th>
						<th class="px-4 py-3 text-left">Operation</th>
						<th class="px-4 py-3 text-left">Date</th>
					</tr>
				</thead>
				<tbody class="divide-y divide-surface-800">
					{#each items as item}
						<tr class="hover:bg-surface-800/30 transition-colors">
							<td class="px-4 py-3 text-surface-100">{item.media_title}</td>
							<td class="px-4 py-3"><Badge>{appDisplayName(item.app_type)}</Badge></td>
							<td class="px-4 py-3 text-surface-400">{item.instance_name}</td>
							<td class="px-4 py-3">
								<Badge variant={item.operation === 'missing' ? 'warning' : 'info'}>{item.operation}</Badge>
							</td>
							<td class="px-4 py-3 text-surface-500 text-xs">{new Date(item.created_at).toLocaleString()}</td>
						</tr>
					{/each}
				</tbody>
			</table>
		</div>
	{/if}
</div>
