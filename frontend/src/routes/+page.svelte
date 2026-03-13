<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appColor } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { RotateCcw } from 'lucide-svelte';

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
	let seerrCount = $state<number | null>(null);
	let loading = $state(true);
	let resettingInstance = $state<string | null>(null);
	let confirmResetStats = $state(false);
	let confirmResetInstance = $state<string | null>(null);

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

		// Load Seerr request count (non-blocking)
		try {
			const res = await api.get<{ count: number }>('/seerr/requests/count');
			seerrCount = res.count;
		} catch {
			seerrCount = null;
		}

		loading = false;
	}

	async function resetStats() {
		try {
			await api.post('/stats/reset');
			toasts.success('Stats reset');
			confirmResetStats = false;
			await load();
		} catch {
			toasts.error('Failed to reset stats');
		}
	}

	async function resetInstance(appType: string, instanceId: string) {
		const key = `${appType}:${instanceId}`;
		resettingInstance = key;
		try {
			await api.post(`/state/reset?app=${encodeURIComponent(appType)}&instance_id=${encodeURIComponent(instanceId)}`);
			toasts.success(`State reset for ${instanceName(appType, instanceId)}`);
			confirmResetInstance = null;
			await load();
		} catch {
			toasts.error('Failed to reset instance state');
		}
		resettingInstance = null;
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
		<h1 class="text-2xl font-bold text-foreground">Dashboard</h1>
		<div class="flex gap-2">
			<Badge variant="success">Running</Badge>
			{#if confirmResetStats}
				<span class="flex items-center gap-1 text-xs">
				<span class="text-muted-foreground">Reset all stats?</span>
					<button onclick={resetStats} class="rounded px-2 py-1 bg-destructive text-destructive-foreground text-xs hover:bg-destructive/90">Yes</button>
					<button onclick={() => confirmResetStats = false} class="rounded px-2 py-1 bg-secondary text-secondary-foreground text-xs hover:bg-secondary/80">No</button>
				</span>
			{:else}
				<Button size="sm" variant="ghost" onclick={() => confirmResetStats = true}>Reset Stats</Button>
			{/if}
		</div>
	</div>

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Array(6) as _}
				<div class="h-32 rounded-xl bg-muted/50 animate-pulse"></div>
			{/each}
		</div>
	{:else}
		<!-- App Summary Cards -->
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Object.entries(appTotals()) as [appType, totals]}
				<Card>
					<div class="flex items-center justify-between mb-3">
						<span class="text-sm font-medium {appColor(appType)}">
							{appDisplayName(appType)}
						</span>
						<Badge variant="default">{(groupedStats()[appType] ?? []).length} instance{(groupedStats()[appType] ?? []).length !== 1 ? 's' : ''}</Badge>
					</div>
					<div class="grid grid-cols-2 gap-4">
						<div>
							<p class="text-2xl font-bold text-foreground">{totals.lurked.toLocaleString()}</p>
							<p class="text-xs text-muted-foreground mt-0.5">Lurked</p>
						</div>
						<div>
							<p class="text-2xl font-bold text-foreground">{totals.upgraded.toLocaleString()}</p>
							<p class="text-xs text-muted-foreground mt-0.5">Upgraded</p>
						</div>
					</div>
					<!-- Per-instance breakdown -->
					{#if (groupedStats()[appType] ?? []).length > 0}
						<div class="mt-3 pt-3 border-t border-border space-y-1.5">
							{#each groupedStats()[appType] as s}
								<div class="flex items-center justify-between text-xs gap-2">
									<span class="text-muted-foreground truncate flex-1">{instanceName(appType, s.instance_id)}</span>
									<span class="text-foreground/80 font-mono shrink-0">{s.lurked} / {s.upgraded}</span>
									{#if confirmResetInstance === `${appType}:${s.instance_id}`}
										<span class="flex items-center gap-1 shrink-0">
											<button onclick={() => resetInstance(appType, s.instance_id)} class="rounded px-1.5 py-0.5 bg-destructive text-destructive-foreground text-[10px] hover:bg-destructive/90">Yes</button>
											<button onclick={() => confirmResetInstance = null} class="rounded px-1.5 py-0.5 bg-secondary text-secondary-foreground text-[10px] hover:bg-secondary/80">No</button>
										</span>
									{:else}
										<button
											onclick={() => confirmResetInstance = `${appType}:${s.instance_id}`}
											disabled={resettingInstance === `${appType}:${s.instance_id}`}
											class="shrink-0 text-muted-foreground hover:text-foreground transition-colors disabled:opacity-50"
											title="Reset state for this instance"
										>
											<RotateCcw class="h-3.5 w-3.5" />
										</button>
									{/if}
								</div>
							{/each}
						</div>
					{/if}
				</Card>
			{/each}
		</div>

		{#if Object.keys(appTotals()).length === 0}
			<Card>
			<p class="text-sm text-muted-foreground text-center py-8">No stats yet — configure app instances to get started</p>
			</Card>
		{/if}

		<!-- Hourly Caps -->
		{#if caps.length > 0}
			<h2 class="text-lg font-semibold text-foreground/80 mt-8">Hourly API Usage</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				{#each caps as cap}
					<Card>
						<div class="flex items-center justify-between">
							<div>
						<span class="text-sm {appColor(cap.app_type)}">
								{appDisplayName(cap.app_type)}
								</span>
								<span class="text-xs text-muted-foreground ml-2">{instanceName(cap.app_type, cap.instance_id)}</span>
							</div>
							<span class="text-lg font-mono font-bold text-foreground">{cap.api_hits}</span>
						</div>
					</Card>
				{/each}
			</div>
		{/if}

		<!-- Seerr -->
		{#if seerrCount !== null}
			<h2 class="text-lg font-semibold text-foreground/80 mt-8">Services</h2>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
				<Card>
					<div class="flex items-center justify-between mb-3">
						<span class="text-sm font-medium text-purple-400">Seerr</span>
						<Badge variant="info">Connected</Badge>
					</div>
					<div>
						<p class="text-2xl font-bold text-foreground">{seerrCount.toLocaleString()}</p>
						<p class="text-xs text-muted-foreground mt-0.5">Pending Requests</p>
					</div>
				</Card>
			</div>
		{/if}
	{/if}
</div>
