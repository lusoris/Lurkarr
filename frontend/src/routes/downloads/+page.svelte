<script lang="ts">
	import { api } from '$lib/api';
	import { appDisplayName, appLogo } from '$lib';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	interface DownloadClientInstance {
		id: string;
		name: string;
		client_type: string;
		url: string;
		enabled: boolean;
	}

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

	let clients = $state<DownloadClientInstance[]>([]);
	let queues = $state<Record<string, Queue | null>>({});
	let errors = $state<Record<string, string>>({});
	let loading = $state(true);
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	async function loadClients() {
		try {
			clients = await api.get<DownloadClientInstance[]>('/download-clients');
		} catch {
			clients = [];
		}
	}

	async function loadQueues() {
		const enabled = clients.filter(c => c.enabled);
		if (enabled.length === 0) {
			loading = false;
			return;
		}
		await Promise.allSettled(
			enabled.map(async (cl) => {
				try {
					const q = await api.get<Queue>(`/download-clients/${cl.id}/health`);
					queues = { ...queues, [cl.id]: q };
					errors = { ...errors, [cl.id]: '' };
				} catch {
					errors = { ...errors, [cl.id]: 'offline' };
				}
			})
		);
		loading = false;
	}

	async function init() {
		loading = true;
		await loadClients();
		await loadQueues();
	}

	$effect(() => {
		init();
		return () => { if (pollTimer) clearInterval(pollTimer); };
	});
</script>

<svelte:head><title>Downloads - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Downloads</h1>

	{#if loading}
		<div class="space-y-2">
			{#each Array(3) as _}
				<div class="h-16 rounded-lg bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if clients.filter(c => c.enabled).length === 0}
		<Card>
			<div class="text-center py-6">
				<p class="text-sm text-surface-400 mb-2">No download clients configured</p>
				<p class="text-xs text-surface-500 mb-4">Add a download client on the Connections page to see active downloads.</p>
				<a href="/apps" class="inline-flex items-center gap-1.5 text-sm font-medium text-lurk-400 hover:text-lurk-300 transition-colors">
					Go to Connections
					<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
						<path stroke-linecap="round" stroke-linejoin="round" d="M13.5 4.5L21 12m0 0l-7.5 7.5M21 12H3" />
					</svg>
				</a>
			</div>
		</Card>
	{:else}
		{#each clients.filter(c => c.enabled) as cl}
			{@const logo = appLogo(cl.client_type)}
			<div class="space-y-2">
				<div class="flex items-center gap-2">
					{#if logo}
						<img src={logo} alt="" class="w-5 h-5 rounded" />
					{/if}
					<span class="text-sm font-semibold text-surface-300">{cl.name}</span>
					{#if errors[cl.id]}
						<Badge variant="error">offline</Badge>
					{:else}
						<Badge variant="success">connected</Badge>
					{/if}
				</div>
			</div>
		{/each}
	{/if}
</div>
