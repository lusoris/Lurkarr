<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { appDisplayName, appColor } from '$lib';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { ScrollText, RefreshCw, Search, Shield, ArrowRightLeft, Clock, Zap } from '@lucide/svelte';
	import type { ActivityEvent } from '$lib/types';
	import { timeAgo } from '$lib/format';

	let events = $state<ActivityEvent[]>([]);
	let loading = $state(true);
	let refreshing = $state(false);
	let filterSource = $state('');

	const sourceLabels: Record<string, string> = {
		lurk: 'Lurk',
		blocklist: 'Blocklist',
		auto_import: 'Auto Import',
		cross_instance: 'Cross Instance',
		schedule: 'Schedule'
	};

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

	function sourceIcon(source: string) {
		switch (source) {
			case 'lurk': return Search;
			case 'blocklist': return Shield;
			case 'auto_import': return Zap;
			case 'cross_instance': return ArrowRightLeft;
			case 'schedule': return Clock;
			default: return ScrollText;
		}
	}

	let filtered = $derived(
		filterSource ? events.filter(e => e.source === filterSource) : events
	);

	async function loadActivity() {
		try {
			const res = await api.get('/activity?limit=100') as { items?: ActivityEvent[] };
			events = res.items ?? [];
		} catch {
			events = [];
		} finally {
			loading = false;
			refreshing = false;
		}
	}

	async function refresh() {
		refreshing = true;
		await loadActivity();
	}

	onMount(() => { loadActivity(); });
</script>

<div class="space-y-6">
	<PageHeader title="Activity" description="Unified timeline of all Lurkarr actions — searches, blocklist hits, imports, and more.">
		{#snippet actions()}
			<Button variant="secondary" size="sm" loading={refreshing} onclick={refresh}>
				<RefreshCw class="h-4 w-4 mr-1.5" />
				Refresh
			</Button>
			<HelpDrawer page="history" />
		{/snippet}
	</PageHeader>

	<div class="flex items-center gap-3">
		<Select label="" bind:value={filterSource} class="w-48">
			<option value="">All Sources</option>
			{#each Object.entries(sourceLabels) as [key, label]}
				<option value={key}>{label}</option>
			{/each}
		</Select>
		{#if filterSource}
			<Button size="sm" variant="link" class="h-auto p-0 text-xs text-muted-foreground" onclick={() => filterSource = ''}>
				Clear filter
			</Button>
		{/if}
		<span class="text-xs text-muted-foreground ml-auto">{filtered.length} events</span>
	</div>

	{#if loading}
		<Skeleton rows={8} height="h-16" />
	{:else if filtered.length === 0}
		<EmptyState icon={ScrollText} title="No activity yet" description="Actions will appear here as Lurkarr operates." />
	{:else}
		<div class="space-y-2">
			{#each filtered as event (event.id)}
				{@const Icon = sourceIcon(event.source)}
				<Card class="!p-3">
					<div class="flex items-start gap-3">
						<div class="mt-0.5 flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted">
							<Icon class="h-4 w-4 text-muted-foreground" />
						</div>
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2 flex-wrap">
								<span class="text-sm font-medium text-foreground truncate">{event.title}</span>
								<Badge variant={sourceBadgeVariant(event.source)}>
									{sourceLabels[event.source] ?? event.source}
								</Badge>
								{#if event.app_type}
									<span class="text-xs {appColor(event.app_type)}">{appDisplayName(event.app_type)}</span>
								{/if}
							</div>
							{#if event.detail}
								<p class="text-xs text-muted-foreground mt-0.5 truncate">{event.action} — {event.detail}</p>
							{:else}
								<p class="text-xs text-muted-foreground mt-0.5">{event.action}</p>
							{/if}
						</div>
						<span class="shrink-0 text-xs text-muted-foreground whitespace-nowrap" title={new Date(event.timestamp).toLocaleString()}>
							{timeAgo(event.timestamp)}
						</span>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
