<script lang="ts">
	import { api } from '$lib/api';
	import { appDisplayName, appLogo } from '$lib';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Progress from '$lib/components/ui/progress/progress.svelte';
	import { ArrowDown, ArrowUp, Cable, Pause, Play } from '@lucide/svelte';
	import type { DownloadClientInstance, ClientStatus, DownloadItem, SABnzbdSettings, SABnzbdQueueSlot, SABnzbdQueue, SABnzbdStats } from '$lib/types';
	import { formatBytes, formatSpeed, formatETA } from '$lib/format';

	let clients = $state<DownloadClientInstance[]>([]);
	let statuses = $state<Record<string, ClientStatus | null>>({});
	let items = $state<Record<string, DownloadItem[]>>({});
	let errors = $state<Record<string, string>>({});
	let loading = $state(true);
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	// SABnzbd state
	let sabEnabled = $state(false);
	let sabQueue = $state<SABnzbdQueue | null>(null);
	let sabStats = $state<SABnzbdStats | null>(null);
	let sabError = $state('');
	let sabPausing = $state(false);

	async function loadClients() {
		try {
			clients = await api.get<DownloadClientInstance[]>('/download-clients');
		} catch {
			clients = [];
		}
	}

	async function loadSABnzbdSettings() {
		try {
			const settings = await api.get<SABnzbdSettings>('/sabnzbd/settings');
			sabEnabled = settings.enabled && !!settings.url;
		} catch {
			sabEnabled = false;
		}
	}

	async function loadSABnzbdData() {
		if (!sabEnabled) return;
		try {
			const [queue, stats] = await Promise.all([
				api.get<SABnzbdQueue>('/sabnzbd/queue'),
				api.get<SABnzbdStats>('/sabnzbd/stats')
			]);
			sabQueue = queue;
			sabStats = stats;
			sabError = '';
		} catch {
			sabError = 'offline';
			sabQueue = null;
			sabStats = null;
		}
	}

	async function toggleSABnzbdPause() {
		if (!sabQueue) return;
		sabPausing = true;
		try {
			if (sabQueue.paused) {
				await api.post('/sabnzbd/resume');
			} else {
				await api.post('/sabnzbd/pause');
			}
			await loadSABnzbdData();
		} catch {
			console.warn('Failed to toggle SABnzbd pause');
		}
		sabPausing = false;
	}

	async function loadData() {
		const enabled = clients.filter(c => c.enabled);
		const promises: Promise<void>[] = [];
		if (enabled.length > 0) {
			promises.push(...enabled.map(async (cl) => {
				try {
					const [status, dlItems] = await Promise.all([
						api.get<ClientStatus>(`/download-clients/${cl.id}/status`),
						api.get<DownloadItem[]>(`/download-clients/${cl.id}/items`)
					]);
					statuses = { ...statuses, [cl.id]: status };
					items = { ...items, [cl.id]: dlItems ?? [] };
					errors = { ...errors, [cl.id]: '' };
				} catch {
					errors = { ...errors, [cl.id]: 'offline' };
					statuses = { ...statuses, [cl.id]: null };
					items = { ...items, [cl.id]: [] };
				}
			}));
		}
		promises.push(loadSABnzbdData());
		await Promise.allSettled(promises);
		loading = false;
	}

	async function init() {
		loading = true;
		await Promise.all([loadClients(), loadSABnzbdSettings()]);
		await loadData();
		pollTimer = setInterval(loadData, 15000);
	}

	$effect(() => {
		init();
		return () => { if (pollTimer) clearInterval(pollTimer); };
	});

	function statusColor(status: string): 'success' | 'warning' | 'error' | 'info' | 'default' {
		switch (status.toLowerCase()) {
			case 'downloading': return 'info';
			case 'seeding': case 'uploading': return 'success';
			case 'paused': case 'stalledDL': return 'warning';
			case 'error': case 'missingFiles': return 'error';
			default: return 'default';
		}
	}

	function sabSlotStatusColor(status: string): 'success' | 'warning' | 'error' | 'info' | 'default' {
		switch (status.toLowerCase()) {
			case 'downloading': return 'info';
			case 'completed': return 'success';
			case 'paused': return 'warning';
			case 'failed': return 'error';
			case 'queued': return 'default';
			default: return 'default';
		}
	}

	const hasAnyClients = $derived(clients.filter(c => c.enabled).length > 0 || sabEnabled);

	const totalDown = $derived.by(() => {
		let speed = 0;
		for (const st of Object.values(statuses)) {
			if (st) speed += st.download_speed;
		}
		return speed;
	});

	const totalUp = $derived.by(() => {
		let speed = 0;
		for (const st of Object.values(statuses)) {
			if (st) speed += st.upload_speed;
		}
		return speed;
	});

	const totalItems = $derived.by(() => {
		let count = 0;
		for (const list of Object.values(items)) {
			count += list.length;
		}
		if (sabQueue) count += sabQueue.slots?.length ?? 0;
		return count;
	});
</script>

<svelte:head><title>Downloads - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Downloads" description="Monitor active downloads across all clients">
		{#snippet actions()}
			{#if !loading && hasAnyClients}
				<div class="flex items-center gap-3 text-sm">
					<span class="flex items-center gap-1.5 text-muted-foreground">
						<ArrowDown class="h-3.5 w-3.5 text-info" />
						<span class="font-mono">{formatSpeed(totalDown)}</span>
					</span>
					<span class="flex items-center gap-1.5 text-muted-foreground">
						<ArrowUp class="h-3.5 w-3.5 text-success" />
						<span class="font-mono">{formatSpeed(totalUp)}</span>
					</span>
					<Badge>{totalItems} item{totalItems !== 1 ? 's' : ''}</Badge>
				</div>
			{/if}
			<HelpDrawer page="downloads" />
		{/snippet}
	</PageHeader>

	{#if loading}
		<Skeleton rows={4} height="h-20" />
	{:else if !hasAnyClients}
		<EmptyState
			icon={Cable}
			title="No download clients configured"
			description="Add a download client on the Connections page to monitor active downloads."
		>
			{#snippet actions()}
				<Button href="/apps" variant="link" class="gap-1.5">
					Go to Connections →
				</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<div class="space-y-6">
			<!-- Torrent clients -->
			{#each clients.filter(c => c.enabled) as cl}
				{@const logo = appLogo(cl.client_type)}
				{@const status = statuses[cl.id]}
				{@const clientItems = items[cl.id] ?? []}
				{@const hasError = !!errors[cl.id]}

				<div class="space-y-3">
					<!-- Client Header -->
					<Card class="!p-4">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-3">
								{#if logo}
									<img src={logo} alt="" class="w-6 h-6 rounded" />
								{/if}
								<div>
									<div class="flex items-center gap-2">
										<span class="font-medium text-foreground">{cl.name}</span>
										{#if hasError}
											<Badge variant="error">offline</Badge>
										{:else}
											<Badge variant="success">online</Badge>
										{/if}
									</div>
									{#if status}
										<p class="text-xs text-muted-foreground mt-0.5">
											{appDisplayName(cl.client_type)} {status.version}{status.paused ? ' · Paused' : ''}
										</p>
									{/if}
								</div>
							</div>
							{#if status && !hasError}
								<div class="flex items-center gap-4 text-xs text-muted-foreground">
									<span class="flex items-center gap-1">
										<ArrowDown class="h-3 w-3 text-info" />
										<span class="font-mono">{formatSpeed(status.download_speed)}</span>
									</span>
									<span class="flex items-center gap-1">
										<ArrowUp class="h-3 w-3 text-success" />
										<span class="font-mono">{formatSpeed(status.upload_speed)}</span>
									</span>
									<span class="font-mono">{status.item_count} item{status.item_count !== 1 ? 's' : ''}</span>
								</div>
							{/if}
						</div>
					</Card>

					<!-- Download Items -->
					{#if !hasError && clientItems.length > 0}
						<div class="space-y-1.5 pl-2">
							{#each clientItems as item}
								{@const pct = Math.round(item.progress * 100)}
								<div class="rounded-lg border border-border bg-card/50 p-3">
									<div class="flex items-start justify-between gap-3 mb-2">
										<div class="min-w-0 flex-1">
											<p class="text-sm font-medium text-foreground truncate" title={item.name}>{item.name}</p>
											<div class="flex items-center gap-2 mt-0.5">
												<Badge variant={statusColor(item.status)}>{item.status}</Badge>
												{#if item.category}
													<span class="text-xs text-muted-foreground">{item.category}</span>
												{/if}
											</div>
										</div>
										<div class="text-right text-xs text-muted-foreground shrink-0">
											<p class="font-mono">{formatBytes(item.total_size - item.remaining_size)} / {formatBytes(item.total_size)}</p>
											{#if item.download_speed > 0}
												<p class="font-mono text-info">{formatSpeed(item.download_speed)}</p>
											{/if}
											{#if item.eta > 0}
												<p>ETA {formatETA(item.eta)}</p>
											{/if}
										</div>
									</div>
									<!-- Progress bar -->
								<Progress value={Math.min(pct, 100)} max={100} class="h-1.5 bg-muted {pct >= 100 ? '[&>[data-slot=progress-indicator]]:bg-success' : ''}" />
									<p class="text-[10px] text-muted-foreground mt-1 text-right font-mono">{pct}%</p>
								</div>
							{/each}
						</div>
					{:else if !hasError}
						<div class="pl-2 py-3">
							<p class="text-sm text-muted-foreground">No active downloads</p>
						</div>
					{/if}
				</div>
			{/each}

			<!-- SABnzbd Section -->
			{#if sabEnabled}
				<div class="space-y-3">
					<Card class="!p-4">
						<div class="flex items-center justify-between">
							<div class="flex items-center gap-3">
								{#if appLogo('sabnzbd')}
									<img src={appLogo('sabnzbd')} alt="SABnzbd logo" class="w-6 h-6 rounded" />
								{/if}
								<div>
									<div class="flex items-center gap-2">
										<span class="font-medium text-foreground">SABnzbd</span>
										{#if sabError}
											<Badge variant="error">offline</Badge>
										{:else}
											<Badge variant="success">online</Badge>
										{/if}
										<Badge variant="default">Usenet</Badge>
									</div>
									{#if sabQueue && !sabError}
										<p class="text-xs text-muted-foreground mt-0.5">
											{sabQueue.status}{sabQueue.paused ? ' · Paused' : ''}{sabQueue.speed ? ` · ${sabQueue.speed}` : ''}
										</p>
									{/if}
								</div>
							</div>
							<div class="flex items-center gap-3">
								{#if sabStats && !sabError}
									<div class="text-xs text-muted-foreground text-right">
										<span class="font-mono">Today: {sabStats.day}</span>
										<span class="mx-1">·</span>
										<span class="font-mono">Total: {sabStats.total}</span>
									</div>
								{/if}
								{#if sabQueue && !sabError}
									<Button
										size="sm"
										variant="outline"
										onclick={toggleSABnzbdPause}
										disabled={sabPausing}
									>
										{#if sabQueue.paused}
											<Play class="h-3.5 w-3.5" />
										{:else}
											<Pause class="h-3.5 w-3.5" />
										{/if}
									</Button>
								{/if}
							</div>
						</div>
					</Card>

					{#if !sabError && sabQueue && sabQueue.slots?.length > 0}
						<div class="space-y-1.5 pl-2">
							{#each sabQueue.slots as slot}
								{@const pct = parseInt(slot.percentage) || 0}
								<div class="rounded-lg border border-border bg-card/50 p-3">
									<div class="flex items-start justify-between gap-3 mb-2">
										<div class="min-w-0 flex-1">
											<p class="text-sm font-medium text-foreground truncate" title={slot.filename}>{slot.filename}</p>
											<div class="flex items-center gap-2 mt-0.5">
												<Badge variant={sabSlotStatusColor(slot.status)}>{slot.status}</Badge>
												{#if slot.cat && slot.cat !== '*'}
													<span class="text-xs text-muted-foreground">{slot.cat}</span>
												{/if}
											</div>
										</div>
										<div class="text-right text-xs text-muted-foreground shrink-0">
											<p class="font-mono">{slot.mb} MB ({slot.mbleft} MB left)</p>
											{#if slot.timeleft && slot.timeleft !== '0:00:00'}
												<p>ETA {slot.timeleft}</p>
											{/if}
										</div>
									</div>
									<Progress value={Math.min(pct, 100)} max={100} class="h-1.5 bg-muted {pct >= 100 ? '[&>[data-slot=progress-indicator]]:bg-success' : ''}" />
									<p class="text-[10px] text-muted-foreground mt-1 text-right font-mono">{pct}%</p>
								</div>
							{/each}
						</div>
					{:else if !sabError}
						<div class="pl-2 py-3">
							<p class="text-sm text-muted-foreground">No active Usenet downloads</p>
						</div>
					{/if}
				</div>
			{/if}
		</div>
	{/if}
</div>

<ScrollToTop />
