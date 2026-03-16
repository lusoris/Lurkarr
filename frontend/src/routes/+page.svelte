<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { appTypes, appDisplayName, appColor, appLogo, appAccentBorder } from '$lib';
	import { fmtVersion, timeAgo } from '$lib/format';
	import type { Stats, HourlyCap, AppInstance, AppSettings, HealthInfo, ServiceSettings, ActivityEvent } from '$lib/types';
	import { getToasts } from '$lib/stores/toast.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import ConnectionCard from '$lib/components/ui/ConnectionCard.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Progress } from '$lib/components/ui/progress';
	import { RotateCcw, LayoutDashboard, Plug, Moon, AlertTriangle, Flame, ListOrdered, CalendarDays, Search, Shield, Zap, ArrowRightLeft, Clock } from '@lucide/svelte';

	const toasts = getToasts();
	const instanceStore = getInstances();

	interface DlClient {
		id: string;
		name: string;
		client_type: string;
		enabled: boolean;
		url: string;
	}

	let stats = $state<Stats[]>([]);
	let caps = $state<HourlyCap[]>([]);
	const instances = $derived(instanceStore.cache);
	let dlClients = $state<DlClient[]>([]);
	let healthStatus = $state<Record<string, HealthInfo>>({});
	let dlHealthStatus = $state<Record<string, HealthInfo>>({});
	let prowlarrSettings = $state<ServiceSettings | null>(null);
	let prowlarrHealth = $state<HealthInfo | null>(null);
	let seerrSettings = $state<ServiceSettings | null>(null);
	let seerrHealth = $state<HealthInfo | null>(null);
	let seerrCount = $state<number | null>(null);
	let recentActivity = $state<ActivityEvent[]>([]);
	let appSettingsMap = $state<Record<string, AppSettings>>({});
	let loading = $state(true);
	let resettingInstance = $state<string | null>(null);
	let confirmResetStats = $state(false);
	let confirmResetInstance = $state<string | null>(null);

	async function load() {
		loading = true;
		try {
			const [s, c, dls] = await Promise.all([
				api.get<Stats[]>('/stats'),
				api.get<HourlyCap[]>('/stats/hourly-caps'),
				api.get<DlClient[]>('/download-clients').catch(() => [] as DlClient[]),
			]);
			stats = s;
			caps = c;
			dlClients = dls ?? [];
		} catch (e) {
			console.error('Failed to load dashboard data', e);
			toasts.error('Failed to load dashboard data');
		}

		// Non-blocking: health checks, services, seerr count, activity, app settings
		checkAllHealth();
		checkDlHealth();
		loadServiceHealth();
		loadSeerrCount();
		loadRecentActivity();
		loadAppSettings();

		loading = false;
	}

	async function checkAllHealth() {
		const allInsts = Object.values(instances).flat();
		await Promise.allSettled(
			allInsts.map(async (inst) => {
				try {
					const res = await api.get<HealthInfo>(`/instances/${inst.id}/health`);
					healthStatus = { ...healthStatus, [inst.id]: res };
				} catch {
					healthStatus = { ...healthStatus, [inst.id]: { status: 'offline' } };
				}
			})
		);
	}

	async function checkDlHealth() {
		await Promise.allSettled(
			dlClients.filter(d => d.enabled).map(async (dl) => {
				try {
					const res = await api.get<HealthInfo>(`/download-clients/${dl.id}/health`);
					dlHealthStatus = { ...dlHealthStatus, [dl.id]: res };
				} catch {
					dlHealthStatus = { ...dlHealthStatus, [dl.id]: { status: 'offline' } };
				}
			})
		);
	}

	async function loadServiceHealth() {
		// Prowlarr
		try {
			const r = await api.get<ServiceSettings>('/prowlarr/settings');
			prowlarrSettings = r;
			if (r?.enabled && r.url && r.api_key) {
				try {
					prowlarrHealth = await api.post<HealthInfo>('/prowlarr/test', { url: r.url, api_key: r.api_key });
				} catch {
					prowlarrHealth = { status: 'offline' };
				}
			}
		} catch { /* service not configured — expected */ }

		// Seerr
		try {
			const r = await api.get<ServiceSettings>('/seerr/settings');
			seerrSettings = r;
			if (r?.enabled && r.url) {
				try {
					seerrHealth = await api.post<HealthInfo>('/seerr/test');
				} catch {
					seerrHealth = { status: 'offline' };
				}
			}
		} catch { /* service not configured — expected */ }
	}

	async function loadSeerrCount() {
		try {
			const res = await api.get<{ total: number }>('/seerr/requests/count');
			seerrCount = res.total;
		} catch {
			seerrCount = null;
		}
	}

	async function loadRecentActivity() {
		try {
			recentActivity = await api.get<ActivityEvent[]>('/activity?limit=5');
		} catch {
			recentActivity = [];
		}
	}

	async function loadAppSettings() {
		const uniqueApps = new Set<string>();
		for (const [app] of Object.entries(instances)) {
			uniqueApps.add(app);
		}
		await Promise.allSettled(
			[...uniqueApps].map(async (app) => {
				try {
					const settings = await api.get<AppSettings>(`/settings/${app}`);
					appSettingsMap = { ...appSettingsMap, [app]: settings };
				} catch { /* app not configured — expected */ }
			})
		);
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

	onMount(() => {
		load();
		const interval = setInterval(() => {
			// Silently refresh health + stats in the background
			checkAllHealth();
			checkDlHealth();
			loadServiceHealth();
			loadRecentActivity();
			api.get<Stats[]>('/stats').then(s => stats = s).catch(() => {});
			api.get<HourlyCap[]>('/stats/hourly-caps').then(c => caps = c).catch(() => {});
		}, 30000);
		return () => clearInterval(interval);
	});

	function instanceName(appType: string, instanceId: string): string {
		// Search both whisparr and eros buckets for merged cards
		const buckets = appType === 'whisparr' || appType === 'eros'
			? [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])]
			: (instances[appType] ?? []);
		const inst = buckets.find(i => i.id === instanceId);
		return inst?.name ?? instanceId.slice(0, 8);
	}

	// All arr instances flat
	const allInstances = $derived(Object.values(instances).flat());
	const hasAnyInstances = $derived(allInstances.length > 0);
	const enabledDlClients = $derived(dlClients.filter(d => d.enabled));
	const hasProwlarr = $derived(prowlarrSettings?.enabled ?? false);
	const hasSeerr = $derived(seerrSettings?.enabled ?? false);
	const hasAnything = $derived(hasAnyInstances || enabledDlClients.length > 0 || hasProwlarr || hasSeerr);

	// Health summary counts
	const healthCounts = $derived.by(() => {
		let online = 0, offline = 0, checking = 0;
		for (const inst of allInstances) {
			const h = healthStatus[inst.id];
			if (!h) checking++;
			else if (h.status === 'ok') online++;
			else offline++;
		}
		for (const dl of enabledDlClients) {
			const h = dlHealthStatus[dl.id];
			if (!h) checking++;
			else if (h.status === 'ok') online++;
			else offline++;
		}
		if (hasProwlarr) {
			if (!prowlarrHealth) checking++;
			else if (prowlarrHealth.status === 'ok') online++;
			else offline++;
		}
		if (hasSeerr) {
			if (!seerrHealth) checking++;
			else if (seerrHealth.status === 'ok') online++;
			else offline++;
		}
		return { online, offline, checking };
	});

	// Group stats by app_type
	// Merge eros stats under whisparr so they share one dashboard card.
	function mergeKey(appType: string): string {
		return appType === 'eros' ? 'whisparr' : appType;
	}

	const groupedStats = $derived.by(() => {
		const groups: Record<string, Stats[]> = {};
		for (const s of stats) {
			const key = mergeKey(s.app_type);
			if (!groups[key]) groups[key] = [];
			groups[key].push(s);
		}
		return groups;
	});

	// Aggregate per-app totals
	const appTotals = $derived.by(() => {
		const totals: Record<string, { lurked: number; upgraded: number }> = {};
		for (const s of stats) {
			const key = mergeKey(s.app_type);
			if (!totals[key]) totals[key] = { lurked: 0, upgraded: 0 };
			totals[key].lurked += s.lurked;
			totals[key].upgraded += s.upgraded;
		}
		return totals;
	});

	// App types that have instances but no stats yet
	const idleAppTypes = $derived.by(() => {
		const seen = new Set<string>();
		const types: string[] = [];
		for (const [app, insts] of Object.entries(instances)) {
			const key = mergeKey(app);
			if (seen.has(key)) continue;
			// For whisparr, count eros instances too
			const merged = key === 'whisparr'
				? [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])]
				: insts;
			if (merged.length > 0 && !appTotals[key]) {
				types.push(key);
			}
			seen.add(key);
		}
		return types;
	});
	// Offline instance names for alert banner
	const offlineNames = $derived.by(() => {
		const names: string[] = [];
		for (const inst of allInstances) {
			const h = healthStatus[inst.id];
			if (h && h.status !== 'ok') names.push(inst.name);
		}
		for (const dl of enabledDlClients) {
			const h = dlHealthStatus[dl.id];
			if (h && h.status !== 'ok') names.push(dl.name);
		}
		if (prowlarrHealth && prowlarrHealth.status !== 'ok') names.push('Prowlarr');
		if (seerrHealth && seerrHealth.status !== 'ok') names.push('Seerr');
		return names;
	});

	// Hourly cap with progress (merge eros → whisparr for lookup)
	function capLimit(appType: string): number {
		const s = appSettingsMap[appType] ?? appSettingsMap[appType === 'eros' ? 'whisparr' : appType];
		return s?.hourly_cap ?? 0;
	}

	// Activity source icon
	function sourceIcon(source: string) {
		switch (source) {
			case 'lurk': return Search;
			case 'blocklist': return Shield;
			case 'auto_import': return Zap;
			case 'cross_instance': return ArrowRightLeft;
			case 'schedule': return Clock;
			default: return Search;
		}
	}

	function sourceBadgeVariant(source: string): 'default' | 'success' | 'warning' | 'error' | 'info' {
		switch (source) {
			case 'lurk': return 'info';
			case 'blocklist': return 'error';
			case 'auto_import': return 'success';
			case 'cross_instance': return 'warning';
			case 'schedule': return 'default';
			default: return 'default';
		}
	}
</script>

<svelte:head><title>Dashboard - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Dashboard">
		{#snippet actions()}
			{#if healthCounts.offline > 0}
				<Badge variant="error">{healthCounts.offline} offline</Badge>
			{/if}
			{#if healthCounts.online > 0}
				<Badge variant="success">{healthCounts.online} online</Badge>
			{/if}
			{#if stats.length > 0}
				<ConfirmAction active={confirmResetStats} message="Reset all stats?" onconfirm={resetStats} oncancel={() => confirmResetStats = false}>
					<span title="Zero out search/action counters — does not affect lurk progress or queue state">
						<Button size="sm" variant="ghost" onclick={() => confirmResetStats = true}>Reset Stats</Button>
					</span>
				</ConfirmAction>
			{/if}
			<HelpDrawer page="dashboard" />
		{/snippet}
	</PageHeader>

	<!-- ── Quick Actions ──────────────────────────────────── -->
	{#if hasAnything && !loading}
		<div class="flex flex-wrap gap-2">
			<Button href="/lurk" size="sm" variant="outline" class="gap-1.5"><Flame class="h-3.5 w-3.5" />Lurk Settings</Button>
			<Button href="/queue" size="sm" variant="outline" class="gap-1.5"><ListOrdered class="h-3.5 w-3.5" />Queue</Button>
			<Button href="/scheduling" size="sm" variant="outline" class="gap-1.5"><CalendarDays class="h-3.5 w-3.5" />Scheduling</Button>
		</div>
	{/if}

	{#if loading}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
			{#each Array(6) as _}
				<Skeleton rows={1} height="h-32" />
			{/each}
		</div>
	{:else if !hasAnything}
		<EmptyState icon={Plug} title="Nothing configured" description="Add your apps and download clients on the Connections page to get started." />
	{:else}
		{#if offlineNames.length > 0}
			<Alert.Root variant="destructive">
				<AlertTriangle class="h-4 w-4" />
				<Alert.Description>
					{offlineNames.length === 1 ? `${offlineNames[0]} is offline` : `${offlineNames.length} services offline: ${offlineNames.join(', ')}`}
					— schedules and queue cleaning may be affected.
				</Alert.Description>
			</Alert.Root>
		{/if}
		<!-- ── Connection Health ──────────────────────────────── -->
		<section>
			<h3 class="text-sm font-semibold text-foreground mb-3">Connections</h3>
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-3">
				{#each allInstances as inst}
					<ConnectionCard name={inst.name} appType={inst.app_type} health={healthStatus[inst.id]} />
				{/each}

				{#each enabledDlClients as dl}
					<ConnectionCard name={dl.name} appType={dl.client_type} health={dlHealthStatus[dl.id]} />
				{/each}

				{#if hasProwlarr}
					<ConnectionCard name="Prowlarr" appType="prowlarr" subtitle="Indexer Manager" health={prowlarrHealth} />
				{/if}

				{#if hasSeerr}
					<ConnectionCard name="Seerr" appType="seerr" subtitle="Request Management" health={seerrHealth} />
				{/if}
			</div>
		</section>

		<!-- ── Lurk Activity ─────────────────────────────────── -->
		<section>
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-sm font-semibold text-foreground mb-3">Lurk Activity</h3>
			</div>

			{#if Object.keys(appTotals).length > 0 || idleAppTypes.length > 0}
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4">
					<!-- Active apps with stats -->
					{#each Object.entries(appTotals) as [appType, totals]}
						{@const logo = appLogo(appType)}
						<Card class="border-l-2 {appAccentBorder(appType)}">
							<div class="flex items-center justify-between mb-3">
								<div class="flex items-center gap-2">
									{#if logo}
										<img src={logo} alt="{appDisplayName(appType)} logo" class="w-4 h-4 rounded shrink-0" />
									{/if}
									<span class="text-sm font-medium {appColor(appType)}">
										{appDisplayName(appType)}
									</span>
								</div>
								<Badge variant="default">{(groupedStats[appType] ?? []).length} instance{(groupedStats[appType] ?? []).length !== 1 ? 's' : ''}</Badge>
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
							{#if (groupedStats[appType] ?? []).length > 0}
								<div class="mt-3 pt-3 border-t border-border space-y-1.5">
									{#each groupedStats[appType] as s}
										<div class="flex items-center justify-between text-xs gap-2">
											<span class="text-muted-foreground truncate flex-1">{instanceName(s.app_type, s.instance_id)}</span>
											<span class="text-foreground/80 font-mono shrink-0">{s.lurked} / {s.upgraded}</span>
											<ConfirmAction active={confirmResetInstance === `${s.app_type}:${s.instance_id}`} onconfirm={() => resetInstance(s.app_type, s.instance_id)} oncancel={() => confirmResetInstance = null}>
												<Button
													size="icon"
													variant="ghost"
													class="h-auto w-auto p-0 shrink-0"
													onclick={() => confirmResetInstance = `${s.app_type}:${s.instance_id}`}
													disabled={resettingInstance === `${s.app_type}:${s.instance_id}`}
													aria-label="Reset instance stats"
												>
													<RotateCcw class="h-3.5 w-3.5" />
												</Button>
											</ConfirmAction>
										</div>
									{/each}
								</div>
							{/if}
						</Card>
					{/each}

					<!-- Idle apps — configured but no lurk activity yet -->
					{#each idleAppTypes as appType}
						{@const insts = appType === 'whisparr' ? [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])] : (instances[appType] ?? [])}
						{@const logo = appLogo(appType)}
						<Card class="border-l-2 {appAccentBorder(appType)}">
							<div class="flex items-center justify-between mb-3">
								<div class="flex items-center gap-2">
									{#if logo}
										<img src={logo} alt="{appDisplayName(appType)} logo" class="w-4 h-4 rounded shrink-0" />
									{/if}
									<span class="text-sm font-medium {appColor(appType)}">
										{appDisplayName(appType)}
									</span>
								</div>
								<Badge variant="default">{insts.length} instance{insts.length !== 1 ? 's' : ''}</Badge>
							</div>
							<div class="flex items-center gap-2 py-3 text-muted-foreground">
								<Moon class="h-4 w-4 shrink-0" />
								<span class="text-sm">Idle — waiting for first lurk cycle</span>
							</div>
						</Card>
					{/each}
				</div>
			{:else if hasAnyInstances}
				<!-- Instances exist but absolutely no stats and no idle apps (shouldn't normally happen) -->
				<Card class="text-center py-8">
					<Moon class="h-8 w-8 text-muted-foreground mx-auto mb-2" />
					<p class="text-sm text-muted-foreground">All apps idle — no lurk activity yet</p>
				</Card>
			{/if}
		</section>

		<!-- ── Hourly Caps ───────────────────────────────────── -->
		{#if caps.length > 0}
			<section>
				<h3 class="text-sm font-semibold text-foreground mb-3">Hourly API Usage</h3>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
					{#each caps as cap}
						{@const limit = capLimit(cap.app_type)}
						{@const pct = limit > 0 ? Math.min(100, Math.round((cap.api_hits / limit) * 100)) : 0}
						<Card>
							<div class="flex items-center justify-between mb-2">
								<div class="min-w-0">
									<span class="text-sm {appColor(cap.app_type)}">
										{appDisplayName(cap.app_type)}
									</span>
									<span class="text-xs text-muted-foreground ml-2 truncate">{instanceName(cap.app_type, cap.instance_id)}</span>
								</div>
								<span class="text-lg font-mono font-bold text-foreground shrink-0">
									{cap.api_hits}{#if limit > 0}<span class="text-xs text-muted-foreground font-normal">/{limit}</span>{/if}
								</span>
							</div>
							{#if limit > 0}
								<Progress value={pct} class="h-1.5 {pct >= 90 ? '[&>div]:bg-destructive' : pct >= 70 ? '[&>div]:bg-yellow-500' : ''}" />
							{/if}
						</Card>
					{/each}
				</div>
			</section>
		{/if}

		<!-- ── Seerr Requests ────────────────────────────────── -->
		{#if seerrCount !== null}
			<section>
				<h3 class="text-sm font-semibold text-foreground mb-3">Services</h3>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
					<Card class="border-l-2 border-l-purple-400">
						<div class="flex items-center justify-between mb-3">
							<div class="flex items-center gap-2">
								<img src={appLogo('seerr')} alt="Seerr logo" class="w-4 h-4 rounded shrink-0" />
								<span class="text-sm font-medium text-purple-400">Seerr</span>
							</div>
							<Badge variant="info">Connected</Badge>
						</div>
						<div>
							<p class="text-2xl font-bold text-foreground">{seerrCount.toLocaleString()}</p>
							<p class="text-xs text-muted-foreground mt-0.5">Pending Requests</p>
						</div>
					</Card>
				</div>
			</section>
		{/if}

		<!-- ── Recent Activity ───────────────────────────────── -->
		{#if recentActivity.length > 0}
			<section>
				<div class="flex items-center justify-between mb-3">
					<h3 class="text-sm font-semibold text-foreground">Recent Activity</h3>
					<Button href="/activity" size="sm" variant="ghost" class="text-xs">View All</Button>
				</div>
				<Card class="divide-y divide-border p-0">
					{#each recentActivity as event}
						{@const Icon = sourceIcon(event.source)}
						<div class="flex items-start gap-3 px-4 py-3">
							<div class="mt-0.5 shrink-0 flex h-6 w-6 items-center justify-center rounded-full bg-muted">
								<Icon class="h-3 w-3 text-muted-foreground" />
							</div>
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium text-foreground truncate">{event.title}</span>
									<Badge variant={sourceBadgeVariant(event.source)} class="text-[10px] shrink-0">{event.source.replace('_', ' ')}</Badge>
								</div>
								{#if event.action}
									<p class="text-xs text-muted-foreground mt-0.5 truncate">{event.action}{#if event.detail} — {event.detail}{/if}</p>
								{/if}
							</div>
							<span class="text-[10px] text-muted-foreground shrink-0 mt-0.5">{timeAgo(event.timestamp)}</span>
						</div>
					{/each}
				</Card>
			</section>
		{/if}
	{/if}
</div>
