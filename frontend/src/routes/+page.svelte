<script lang="ts">
	import { api } from '$lib/api';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	interface Stats {
		app_type: string;
		hunted: number;
		upgraded: number;
	}

	interface HourlyCap {
		app_type: string;
		api_hits: number;
	}

	let stats = $state<Stats[]>([]);
	let caps = $state<HourlyCap[]>([]);
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

	async function load() {
		loading = true;
		try {
			[stats, caps] = await Promise.all([
				api.get<Stats[]>('/stats'),
				api.get<HourlyCap[]>('/stats/hourly-caps')
			]);
		} catch { /* handled by error boundary */ }
		loading = false;
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Dashboard - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-surface-50">Dashboard</h1>
		<Badge variant="success">Running</Badge>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Array(6) as _}
				<div class="h-32 rounded-xl bg-surface-800/50 animate-pulse" />
			{/each}
		</div>
	{:else}
		<!-- Stats Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each stats as stat}
				<Card>
					<div class="flex items-center justify-between mb-3">
						<span class="text-sm font-medium {appColors[stat.app_type] ?? 'text-surface-300'}">
							{appLabels[stat.app_type] ?? stat.app_type}
						</span>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-2xl font-bold text-surface-50">{stat.hunted.toLocaleString()}</p>
							<p class="text-xs text-surface-500 mt-0.5">Hunted</p>
						</div>
						<div>
							<p class="text-2xl font-bold text-surface-50">{stat.upgraded.toLocaleString()}</p>
							<p class="text-xs text-surface-500 mt-0.5">Upgraded</p>
						</div>
					</div>
				</Card>
			{/each}
		</div>

		<!-- Hourly Caps -->
		{#if caps.length > 0}
			<h2 class="text-lg font-semibold text-surface-200 mt-8">Hourly API Usage</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each caps as cap}
					<Card>
						<div class="flex items-center justify-between">
							<span class="text-sm {appColors[cap.app_type] ?? 'text-surface-300'}">
								{appLabels[cap.app_type] ?? cap.app_type}
							</span>
							<span class="text-lg font-mono font-bold text-surface-100">{cap.api_hits}</span>
						</div>
					</Card>
				{/each}
			</div>
		{/if}
	{/if}
</div>
