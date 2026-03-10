<script lang="ts">
	import { api } from '$lib/api';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';

	interface AppInstance {
		id: string;
		app_type: string;
		name: string;
		api_url: string;
		api_key: string;
		enabled: boolean;
	}

	const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;
	let instances = $state<Record<string, AppInstance[]>>({});
	let loading = $state(true);

	async function loadAll() {
		loading = true;
		const results: Record<string, AppInstance[]> = {};
		await Promise.all(
			appTypes.map(async (app) => {
				try {
					results[app] = await api.get<AppInstance[]>(`/instances/${app}`);
				} catch {
					results[app] = [];
				}
			})
		);
		instances = results;
		loading = false;
	}

	$effect(() => { loadAll(); });
</script>

<svelte:head><title>Apps - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">App Instances</h1>

	{#if loading}
		<div class="space-y-4">
			{#each Array(3) as _}
				<div class="h-24 rounded-xl bg-surface-800/50 animate-pulse" />
			{/each}
		</div>
	{:else}
		{#each appTypes as app}
			{@const appInstances = instances[app] ?? []}
			<div>
				<h2 class="text-lg font-semibold text-surface-200 mb-3 capitalize">{app}</h2>
				{#if appInstances.length === 0}
					<Card>
						<p class="text-sm text-surface-500">No instances configured</p>
					</Card>
				{:else}
					<div class="space-y-2">
						{#each appInstances as inst}
							<Card class="flex items-center justify-between">
								<div>
									<span class="font-medium text-surface-100">{inst.name}</span>
									<span class="text-xs text-surface-500 ml-2">{inst.api_url}</span>
								</div>
								<Badge variant={inst.enabled ? 'success' : 'error'}>
									{inst.enabled ? 'Enabled' : 'Disabled'}
								</Badge>
							</Card>
						{/each}
					</div>
				{/if}
			</div>
		{/each}
	{/if}
</div>
