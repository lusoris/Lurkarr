<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const toasts = getToasts();

	interface Stats {
		app_type: string;
		instance_id: string;
		lurked: number;
		upgraded: number;
	}

	interface HourlyCap {
		app_type: string;
		instance_id: string;
		api_hits: number;
	}

	interface AppInstance {
		id: string;
		app_type: string;
		name: string;
		enabled: boolean;
	}

	let stats = $state<Stats[]>([]);
	let caps = $state<HourlyCap[]>([]);
	let instances = $state<Record<string, AppInstance[]>>({});
	let loading = $state(true);

	const appLabels: Record<string, string> = {
		sonarr: 'Sonarr',
		radarr: 'Radarr',
		lidarr: 'Lidarr',
		readarr: 'Readarr',
		whisparr: 'Whisparr',
		eros: 'Eros',
		prowlarr: 'Prowlarr'
	};

	const appColors: Record<string, string> = {
		sonarr: 'text-sky-400',
		radarr: 'text-amber-400',
		lidarr: 'text-emerald-400',
		readarr: 'text-rose-400',
		whisparr: 'text-pink-400',
		eros: 'text-purple-400',
		prowlarr: 'text-orange-400'
	};

	const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;

	async function load() {
		loading = true;
		try {
			const instResults: Record<string, AppInstance[]> = {};
			const [s, c] = await Promise.all([
				api.get<Stats[]>('/stats'),
				api.get<HourlyCap[]>('/stats/hourly-caps'),
				...appTypes.map(async (app) => {
					try {
						instResults[app] = await api.get<AppInstance[]>(`/instances/${app}`);
					} catch {
						instResults[app] = [];
					}
				})
			]);
			stats = s as Stats[];
			caps = c as HourlyCap[];
			instances = instResults;
		} catch { /* handled by error boundary */ }
		loading = false;
	}

	async function resetStats() {
		try {
			await api.post('/stats/reset');
			toasts.success('Stats reset');
			await load();
		} catch {
			toasts.error('Failed to reset stats');
		}
	}

	$effect(() => { load(); });

	function instanceName(appType: string, instanceId: string): string {
		const list = instances[appType] ?? [];
		const inst = list.find(i => i.id === instanceId);
		return inst?.name ?? instanceId.slice(0, 8);
	}

	// Group stats by app_type
	const groupedStats = $derived(() => {
		const groups: Record<string, Stats[]> = {};
		for (const s of stats) {
			if (!groups[s.app_type]) groups[s.app_type] = [];
			groups[s.app_type].push(s);
		}
		return groups;
	});

	// Aggregate per-app totals
	const appTotals = $derived(() => {
		const totals: Record<string, { lurked: number; upgraded: number }> = {};
		for (const s of stats) {
			if (!totals[s.app_type]) totals[s.app_type] = { lurked: 0, upgraded: 0 };
			totals[s.app_type].lurked += s.lurked;
			totals[s.app_type].upgraded += s.upgraded;
		}
		return totals;
	});
</script>

<svelte:head><title>Dashboard - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-surface-50">Dashboard</h1>
		<div class="flex gap-2">
			<Badge variant="success">Running</Badge>
			<Button size="sm" variant="ghost" onclick={resetStats}>Reset Stats</Button>
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Array(6) as _}
				<div class="h-32 rounded-xl bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else}
		<!-- App Summary Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Object.entries(appTotals()) as [appType, totals]}
				<Card>
					<div class="flex items-center justify-between mb-3">
						<span class="text-sm font-medium {appColors[appType] ?? 'text-surface-300'}">
							{appLabels[appType] ?? appType}
						</span>
						<Badge variant="default">{(groupedStats()[appType] ?? []).length} instance{(groupedStats()[appType] ?? []).length !== 1 ? 's' : ''}</Badge>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-2xl font-bold text-surface-50">{totals.lurked.toLocaleString()}</p>
							<p class="text-xs text-surface-500 mt-0.5">Lurked</p>
						</div>
						<div>
							<p class="text-2xl font-bold text-surface-50">{totals.upgraded.toLocaleString()}</p>
							<p class="text-xs text-surface-500 mt-0.5">Upgraded</p>
						</div>
					</div>
					<!-- Per-instance breakdown -->
					{#if (groupedStats()[appType] ?? []).length > 1}
						<div class="mt-3 pt-3 border-t border-surface-800 space-y-1.5">
							{#each groupedStats()[appType] as s}
								<div class="flex items-center justify-between text-xs">
									<span class="text-surface-400 truncate">{instanceName(appType, s.instance_id)}</span>
									<span class="text-surface-300 font-mono">{s.lurked} / {s.upgraded}</span>
								</div>
							{/each}
						</div>
					{/if}
				</Card>
			{/each}
		</div>

		{#if Object.keys(appTotals()).length === 0}
			<Card>
				<p class="text-sm text-surface-500 text-center py-8">No stats yet — configure app instances to get started</p>
			</Card>
		{/if}

		<!-- Hourly Caps -->
		{#if caps.length > 0}
			<h2 class="text-lg font-semibold text-surface-200 mt-8">Hourly API Usage</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each caps as cap}
					<Card>
						<div class="flex items-center justify-between">
							<div>
								<span class="text-sm {appColors[cap.app_type] ?? 'text-surface-300'}">
									{appLabels[cap.app_type] ?? cap.app_type}
								</span>
								<span class="text-xs text-surface-500 ml-2">{instanceName(cap.app_type, cap.instance_id)}</span>
							</div>
							<span class="text-lg font-mono font-bold text-surface-100">{cap.api_hits}</span>
						</div>
					</Card>
				{/each}
			</div>
		{/if}
	{/if}
</div>
