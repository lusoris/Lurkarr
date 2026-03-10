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

<div class="space-y-4">
	<div class="flex items-center justify-between">
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
				<div class="h-16 rounded-lg bg-surface-800/50 animate-pulse" />
			{/each}
		</div>
	{:else if error}
		<Card>
			<p class="text-sm text-red-400 text-center py-4">{error}</p>
			<p class="text-xs text-surface-500 text-center">Make sure SABnzbd is configured in Settings</p>
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
						/>
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
