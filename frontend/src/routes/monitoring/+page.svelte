<script lang="ts">
	import { api } from '$lib/api';
	import { appTabLabel } from '$lib';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import * as Table from '$lib/components/ui/table';
	import { Activity, ExternalLink, RefreshCw, Heart, Terminal, BarChart3, Trash2 } from '@lucide/svelte';
	import type { HealthInfo, Stats, HourlyCap } from '$lib/types';

	const store = getInstances();
	const toasts = getToasts();
	let instances = $derived(store.cache);

	let liveness = $state<HealthInfo | null>(null);
	let readiness = $state<HealthInfo | null>(null);
	let livenessError = $state(false);
	let readinessError = $state(false);
	let checking = $state(false);

	let stats = $state<Stats[]>([]);
	let caps = $state<HourlyCap[]>([]);
	let loadingStats = $state(true);
	let resettingStats = $state(false);
	let confirmReset = $state(false);

	async function checkHealth() {
		checking = true;
		try {
			const res = await fetch('/healthz');
			liveness = await res.json();
			livenessError = !res.ok;
		} catch {
			livenessError = true;
		}
		try {
			const res = await fetch('/readyz');
			readiness = await res.json();
			readinessError = !res.ok;
		} catch {
			readinessError = true;
		}
		checking = false;
	}

	async function loadStats() {
		loadingStats = true;
		try {
			const [s, c] = await Promise.all([
				api.get<Stats[]>('/stats'),
				api.get<HourlyCap[]>('/stats/hourly-caps')
			]);
			stats = s;
			caps = c;
		} catch (e) {
			console.error('Failed to load monitoring stats', e);
		}
		loadingStats = false;
	}

	async function resetStats() {
		resettingStats = true;
		try {
			await api.post('/stats/reset', {});
			toasts.success('All statistics reset');
			stats = [];
			caps = [];
		} catch {
			toasts.error('Failed to reset statistics');
		}
		resettingStats = false;
		confirmReset = false;
	}

	function instanceName(appType: string, instanceId: string): string {
		const buckets = appType === 'whisparr' || appType === 'eros'
			? [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])]
			: (instances[appType] ?? []);
		const inst = buckets.find(i => i.id === instanceId);
		return inst?.name ?? instanceId.slice(0, 8);
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString();
	}

	$effect(() => {
		store.fetch();
		checkHealth();
		loadStats();
	});

	const endpoints = [
		{ name: 'Prometheus Metrics', desc: 'Scrape target for Prometheus', path: '/metrics', icon: BarChart3 },
		{ name: 'API Documentation', desc: 'Interactive Scalar API reference', path: '/api/docs', icon: ExternalLink },
		{ name: 'OpenAPI Spec', desc: 'Raw OpenAPI 3.1 YAML', path: '/api/spec', icon: Terminal },
	];
	const dashboards = [
		{ name: 'Lurkarr Overview', desc: 'Search rates, trends, errors, durations' },
		{ name: 'Lurkarr System', desc: 'Goroutines, heap, GC, CPU, threads' },
		{ name: 'Lurkarr Logs', desc: 'Loki log exploration, volume, errors' }
	];
</script>

<svelte:head><title>Monitoring - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Monitoring" description="Application health, observability endpoints, and Grafana dashboards.">
		{#snippet actions()}
			<Button size="sm" variant="secondary" onclick={() => { checkHealth(); loadStats(); }} loading={checking}>
				<RefreshCw class="h-3.5 w-3.5" />
				Refresh
			</Button>
			<HelpDrawer page="monitoring" />
		{/snippet}
	</PageHeader>

	<!-- Health Probes -->
	<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
		<Card>
			<div class="flex items-center gap-3 mb-3">
				<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
					<Heart class="h-4 w-4 text-muted-foreground" />
				</div>
				<div class="flex-1">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-foreground">Liveness</span>
						{#if liveness}
							<Badge variant={livenessError ? 'error' : 'success'}>{liveness.status}</Badge>
						{:else}
							<Badge variant="default">checking...</Badge>
						{/if}
					</div>
					<p class="text-xs text-muted-foreground mt-0.5">Kubernetes liveness probe</p>
				</div>
			</div>
			<span class="text-xs font-mono text-muted-foreground">/healthz</span>
		</Card>
		<Card>
			<div class="flex items-center gap-3 mb-3">
				<div class="flex h-9 w-9 items-center justify-center rounded-lg bg-muted">
					<Activity class="h-4 w-4 text-muted-foreground" />
				</div>
				<div class="flex-1">
					<div class="flex items-center justify-between">
						<span class="text-sm font-medium text-foreground">Readiness</span>
						{#if readiness}
							<Badge variant={readinessError ? 'error' : 'success'}>{readiness.status}</Badge>
						{:else}
							<Badge variant="default">checking...</Badge>
						{/if}
					</div>
					<p class="text-xs text-muted-foreground mt-0.5">Includes database connectivity</p>
				</div>
			</div>
			<span class="text-xs font-mono text-muted-foreground">/readyz</span>
		</Card>
	</div>

	<!-- Endpoints -->
	<Card>
		<h3 class="text-sm font-semibold text-foreground mb-3">Observability Endpoints</h3>
		<div class="space-y-3">
			{#each endpoints as ep}
				{@const Icon = ep.icon}
				<a
					href={ep.path}
					target="_blank"
					rel="noopener noreferrer"
					class="flex items-center justify-between p-3 rounded-lg bg-muted/30 hover:bg-muted/60 transition-colors group"
				>
					<div class="flex items-center gap-3">
						<div class="flex h-8 w-8 items-center justify-center rounded-md bg-muted">
							<Icon class="h-4 w-4 text-muted-foreground" />
						</div>
						<div>
							<p class="text-sm font-medium text-foreground">{ep.name}</p>
							<p class="text-xs text-muted-foreground">{ep.desc}</p>
						</div>
					</div>
					<span class="text-sm text-primary font-mono group-hover:underline">{ep.path}</span>
				</a>
			{/each}
		</div>
	</Card>

	<!-- Grafana Dashboards -->
	<Card>
		<h3 class="text-sm font-semibold text-foreground mb-3">Grafana Dashboards</h3>
		<p class="text-sm text-muted-foreground mb-4">Pre-built dashboards included in <code class="text-primary text-xs">deploy/grafana/dashboards/</code></p>
		<div class="grid grid-cols-1 sm:grid-cols-3 gap-3">
			{#each dashboards as dash}
				<div class="p-3 rounded-lg bg-muted/30 border border-border">
					<p class="text-sm font-medium text-foreground">{dash.name}</p>
					<p class="text-xs text-muted-foreground mt-1">{dash.desc}</p>
				</div>
			{/each}
		</div>
	</Card>

	<!-- Lurk Stats -->
	<Card>
		<div class="flex items-center justify-between mb-3">
			<h3 class="text-sm font-semibold text-foreground">Lurk Statistics</h3>
			<ConfirmAction active={confirmReset} message="Reset all stats?" onconfirm={resetStats} oncancel={() => confirmReset = false}>
				<Button size="sm" variant="ghost" onclick={() => confirmReset = true} loading={resettingStats}>
					<Trash2 class="h-3.5 w-3.5" />
					Reset
				</Button>
			</ConfirmAction>
		</div>
		{#if loadingStats}
			<p class="text-sm text-muted-foreground">Loading statistics...</p>
		{:else if stats.length === 0}
			<p class="text-sm text-muted-foreground">No lurk statistics yet. Stats appear after the first lurk cycle.</p>
		{:else}
			<div class="overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>App</Table.Head>
							<Table.Head>Instance</Table.Head>
							<Table.Head class="text-right">Lurked</Table.Head>
							<Table.Head class="text-right">Upgraded</Table.Head>
							<Table.Head class="text-right">Last Updated</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each stats as s}
							<Table.Row>
								<Table.Cell>{appTabLabel(s.app_type)}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{instanceName(s.app_type, s.instance_id)}</Table.Cell>
								<Table.Cell class="text-right font-mono">{s.lurked.toLocaleString()}</Table.Cell>
								<Table.Cell class="text-right font-mono">{s.upgraded.toLocaleString()}</Table.Cell>
								<Table.Cell class="text-right text-xs text-muted-foreground">{s.updated_at ? formatDate(s.updated_at) : '—'}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		{/if}
	</Card>

	<!-- Hourly API Caps -->
	{#if caps.length > 0}
		<Card>
			<h3 class="text-sm font-semibold text-foreground mb-3">Hourly API Cap Usage</h3>
			<div class="overflow-x-auto">
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>App</Table.Head>
							<Table.Head>Instance</Table.Head>
							<Table.Head>Hour</Table.Head>
							<Table.Head class="text-right">API Hits</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each caps as cap}
							<Table.Row>
								<Table.Cell>{appTabLabel(cap.app_type)}</Table.Cell>
								<Table.Cell class="text-muted-foreground">{instanceName(cap.app_type, cap.instance_id)}</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">{cap.hour_bucket ? formatDate(cap.hour_bucket) : '—'}</Table.Cell>
								<Table.Cell class="text-right font-mono">{cap.api_hits}</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			</div>
		</Card>
	{/if}

	<!-- Setup -->
	<Card>
		<h3 class="text-sm font-semibold text-foreground mb-3">Monitoring Stack Setup</h3>
		<p class="text-sm text-muted-foreground mb-3">
			A complete stack (Prometheus + Loki + Grafana) is included in
			<code class="text-primary text-xs">deploy/docker-compose.monitoring.yml</code>.
		</p>
		<pre class="rounded-lg bg-background border border-border p-3 text-xs text-muted-foreground overflow-x-auto font-mono">docker compose -f docker-compose.yml -f deploy/docker-compose.monitoring.yml up -d</pre>
	</Card>
</div>

<ScrollToTop />
