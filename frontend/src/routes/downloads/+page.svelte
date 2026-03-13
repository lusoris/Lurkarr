<script lang="ts">
	import { api } from '$lib/api';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	interface QueueSlot {
		nzo_id: string;
		filename: string;
		status: string;
		mb: string;
		mbleft: string;
		percentage: string;
		timeleft: string;
		cat: string;
	}

	interface Queue {
		status: string;
		speed: string;
		sizeleft: string;
		paused: boolean;
		slots: QueueSlot[];
	}

	let queue = $state<Queue | null>(null);
	let loading = $state(true);
	let error = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			queue = await api.get<Queue>('/sabnzbd/queue');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load';
			queue = null;
		}
		loading = false;
	}

	async function togglePause() {
		if (!queue) return;
		try {
			if (queue.paused) {
				await api.post('/sabnzbd/resume');
			} else {
				await api.post('/sabnzbd/pause');
			}
			await load();
		} catch { /* handled */ }
	}

	$effect(() => { load(); const interval = setInterval(load, 5000); return () => clearInterval(interval); });
</script>

<svelte:head><title>Downloads - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
		<h1 class="text-2xl font-bold text-surface-50">Downloads</h1>
		{#if queue}
			<div class="flex items-center gap-3">
				<span class="text-sm text-surface-400">
					{queue.speed}/s &middot; {queue.sizeleft} remaining
				</span>
				<Button size="sm" variant="secondary" onclick={togglePause}>
					{queue.paused ? 'Resume' : 'Pause'}
				</Button>
			</div>
		{/if}
	</div>

	{#if loading && !queue}
		<div class="space-y-2">
			{#each Array(3) as _}
				<div class="h-16 rounded-lg bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if error}
		<Card>
			<div class="text-center py-6">
				<p class="text-sm text-red-400 mb-2">Could not connect to SABnzbd</p>
				<p class="text-xs text-surface-500 mb-4">Make sure SABnzbd is configured with a valid URL and API key.</p>
				<a href="/settings" class="inline-flex items-center gap-1.5 text-sm font-medium text-lurk-400 hover:text-lurk-300 transition-colors">
					Go to Settings
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
					</svg>
				</a>
			</div>
		</Card>
	{:else if queue && queue.slots.length === 0}
		<Card>
			<p class="text-sm text-surface-500 text-center py-8">Queue is empty</p>
		</Card>
	{:else if queue}
		<div class="space-y-2">
			{#each queue.slots as slot}
				<Card>
					<div class="flex items-center justify-between mb-2">
						<span class="text-sm font-medium text-surface-100 truncate flex-1">{slot.filename}</span>
						<Badge>{slot.status}</Badge>
					</div>
					<div class="w-full bg-surface-800 rounded-full h-1.5">
						<div
							class="bg-lurk-500 h-1.5 rounded-full transition-all"
							style="width: {slot.percentage}%"
						></div>
					</div>
					<div class="flex justify-between mt-1.5 text-xs text-surface-500">
						<span>{slot.percentage}%</span>
						<span>{slot.timeleft}</span>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
