<script lang="ts">
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import { api } from '$lib/api';

	interface HealthStatus {
		status: string;
	}

	let liveness = $state<HealthStatus | null>(null);
	let readiness = $state<HealthStatus | null>(null);
	let livenessError = $state(false);
	let readinessError = $state(false);

	async function checkHealth() {
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
	}

	$effect(() => { checkHealth(); });
</script>

<svelte:head><title>Monitoring - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-foreground">Monitoring</h1>

	<!-- Health Status -->
	<Card>
		<h2 class="text-sm font-semibold text-muted-foreground mb-3">Health</h2>
		<div class="space-y-3">
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div class="flex items-center gap-2">
					{#if liveness}
						<Badge variant={livenessError ? 'error' : 'success'}>{liveness.status}</Badge>
					{:else}
						<Badge variant="secondary">...</Badge>
					{/if}
					<span class="text-sm text-foreground">Liveness</span>
				</div>
				<span class="text-xs font-mono text-muted-foreground">/healthz</span>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div class="flex items-center gap-2">
					{#if readiness}
						<Badge variant={readinessError ? 'error' : 'success'}>{readiness.status}</Badge>
					{:else}
						<Badge variant="secondary">...</Badge>
					{/if}
					<span class="text-sm text-foreground">Readiness</span>
				</div>
				<span class="text-xs font-mono text-muted-foreground">/readyz</span>
			</div>
		</div>
	</Card>

	<!-- Endpoints -->
	<Card>
		<h2 class="text-sm font-semibold text-muted-foreground mb-3">Endpoints</h2>
		<div class="space-y-3">
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div>
					<p class="text-sm font-medium text-foreground">Prometheus Metrics</p>
					<p class="text-xs text-muted-foreground">Scrape target for Prometheus</p>
				</div>
				<a href="/metrics" target="_blank" rel="noopener" class="text-sm text-primary hover:text-primary/80 font-mono">/metrics</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div>
					<p class="text-sm font-medium text-foreground">API Documentation</p>
					<p class="text-xs text-muted-foreground">Interactive Scalar API reference</p>
				</div>
				<a href="/api/docs" target="_blank" rel="noopener" class="text-sm text-primary hover:text-primary/80 font-mono">/api/docs</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div>
					<p class="text-sm font-medium text-foreground">OpenAPI Spec</p>
					<p class="text-xs text-muted-foreground">Raw OpenAPI 3.1 YAML</p>
				</div>
				<a href="/api/spec" target="_blank" rel="noopener" class="text-sm text-primary hover:text-primary/80 font-mono">/api/spec</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div>
					<p class="text-sm font-medium text-foreground">Liveness Probe</p>
					<p class="text-xs text-muted-foreground">Kubernetes liveness check</p>
				</div>
				<a href="/healthz" target="_blank" rel="noopener" class="text-sm text-primary hover:text-primary/80 font-mono">/healthz</a>
			</div>
			<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50">
				<div>
					<p class="text-sm font-medium text-foreground">Readiness Probe</p>
					<p class="text-xs text-muted-foreground">Kubernetes readiness check (includes DB)</p>
				</div>
				<a href="/readyz" target="_blank" rel="noopener" class="text-sm text-primary hover:text-primary/80 font-mono">/readyz</a>
			</div>
		</div>
	</Card>

	<!-- Grafana Dashboards -->
	<Card>
		<h2 class="text-sm font-semibold text-muted-foreground mb-3">Grafana Dashboards</h2>
		<p class="text-sm text-muted-foreground mb-4">Pre-built dashboards are included in the <code class="text-primary">deploy/grafana/dashboards/</code> directory.</p>
		<div class="space-y-2">
			<div class="p-3 rounded-lg bg-muted/50">
				<p class="text-sm font-medium text-foreground">Lurkarr Overview</p>
				<p class="text-xs text-muted-foreground">Search rates, missing/upgrade trends, error rates, durations</p>
			</div>
			<div class="p-3 rounded-lg bg-muted/50">
				<p class="text-sm font-medium text-foreground">Lurkarr System</p>
				<p class="text-xs text-muted-foreground">Go runtime: goroutines, heap, GC, CPU, threads</p>
			</div>
			<div class="p-3 rounded-lg bg-muted/50">
				<p class="text-sm font-medium text-foreground">Lurkarr Logs</p>
				<p class="text-xs text-muted-foreground">Loki log exploration, volume by level, error aggregation</p>
			</div>
		</div>
	</Card>

	<!-- Setup Info -->
	<Card>
		<h2 class="text-sm font-semibold text-muted-foreground mb-3">Monitoring Stack Setup</h2>
		<p class="text-sm text-muted-foreground">
			A complete monitoring stack (Prometheus + Loki + Grafana) is included in
			<code class="text-primary">deploy/docker-compose.monitoring.yml</code>. Add it alongside your main compose file:
		</p>
		<pre class="mt-3 rounded-lg bg-card border border-border p-3 text-xs text-muted-foreground overflow-x-auto">docker compose -f docker-compose.yml -f deploy/docker-compose.monitoring.yml up -d</pre>
	</Card>
</div>
