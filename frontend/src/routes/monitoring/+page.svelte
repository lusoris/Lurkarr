<script lang="ts">
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { api } from '$lib/api';

	interface HealthStatus {
		status: string;
	}

	let health = $state<HealthStatus | null>(null);
	let healthError = $state(false);

	async function checkHealth() {
		try {
			const res = await fetch('/api/health');
			health = await res.json();
			healthError = !res.ok;
		} catch {
			healthError = true;
		}
	}

	$effect(() => { checkHealth(); });
</script>

<svelte:head><title>Monitoring - Lurkarr</title></svelte:head>

<div class="space-y-6 max-w-3xl">
	<h1 class="text-2xl font-bold text-surface-50">Monitoring</h1>

	<!-- Health Status -->
	<Card>
		<h2 class="text-sm font-semibold text-surface-300 mb-3">Health</h2>
		{#if health}
			<div class="flex items-center gap-2">
				<Badge variant={healthError ? 'error' : 'success'}>{health.status}</Badge>
				<span class="text-sm text-surface-400">Application health check</span>
			</div>
		{:else}
			<p class="text-sm text-surface-500">Checking...</p>
		{/if}
	</Card>

	<!-- Endpoints -->
	<Card>
		<h2 class="text-sm font-semibold text-surface-300 mb-3">Endpoints</h2>
		<div class="space-y-3">
			<div class="flex items-center justify-between p-3 rounded-lg bg-surface-800/50">
				<div>
					<p class="text-sm font-medium text-surface-100">Prometheus Metrics</p>
					<p class="text-xs text-surface-500">Scrape target for Prometheus</p>
				</div>
				<a href="/metrics" target="_blank" rel="noopener" class="text-sm text-lurk-400 hover:text-lurk-300 font-mono">/metrics</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-surface-800/50">
				<div>
					<p class="text-sm font-medium text-surface-100">API Documentation</p>
					<p class="text-xs text-surface-500">Interactive Scalar API reference</p>
				</div>
				<a href="/api/docs" target="_blank" rel="noopener" class="text-sm text-lurk-400 hover:text-lurk-300 font-mono">/api/docs</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-surface-800/50">
				<div>
					<p class="text-sm font-medium text-surface-100">OpenAPI Spec</p>
					<p class="text-xs text-surface-500">Raw OpenAPI 3.1 YAML</p>
				</div>
				<a href="/api/spec" target="_blank" rel="noopener" class="text-sm text-lurk-400 hover:text-lurk-300 font-mono">/api/spec</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-surface-800/50">
				<div>
					<p class="text-sm font-medium text-surface-100">Health Check</p>
					<p class="text-xs text-surface-500">Load balancer probe endpoint</p>
				</div>
				<a href="/api/health" target="_blank" rel="noopener" class="text-sm text-lurk-400 hover:text-lurk-300 font-mono">/api/health</a>
			</div>
		</div>
	</Card>

	<!-- Grafana Dashboards -->
	<Card>
		<h2 class="text-sm font-semibold text-surface-300 mb-3">Grafana Dashboards</h2>
		<p class="text-sm text-surface-400 mb-4">Pre-built dashboards are included in the <code class="text-lurk-400">deploy/grafana/dashboards/</code> directory.</p>
		<div class="space-y-2">
			<div class="p-3 rounded-lg bg-surface-800/50">
				<p class="text-sm font-medium text-surface-100">Lurkarr Overview</p>
				<p class="text-xs text-surface-500">Search rates, missing/upgrade trends, error rates, durations</p>
			</div>
			<div class="p-3 rounded-lg bg-surface-800/50">
				<p class="text-sm font-medium text-surface-100">Lurkarr System</p>
				<p class="text-xs text-surface-500">Go runtime: goroutines, heap, GC, CPU, threads</p>
			</div>
			<div class="p-3 rounded-lg bg-surface-800/50">
				<p class="text-sm font-medium text-surface-100">Lurkarr Logs</p>
				<p class="text-xs text-surface-500">Loki log exploration, volume by level, error aggregation</p>
			</div>
		</div>
	</Card>

	<!-- Setup Info -->
	<Card>
		<h2 class="text-sm font-semibold text-surface-300 mb-3">Monitoring Stack Setup</h2>
		<p class="text-sm text-surface-400">
			A complete monitoring stack (Prometheus + Loki + Grafana) is included in
			<code class="text-lurk-400">deploy/docker-compose.monitoring.yml</code>. Add it alongside your main compose file:
		</p>
		<pre class="mt-3 rounded-lg bg-surface-900 border border-surface-800 p-3 text-xs text-surface-300 overflow-x-auto">docker compose -f docker-compose.yml -f deploy/docker-compose.monitoring.yml up -d</pre>
	</Card>
</div>
