<script lang="ts">
	import { api } from '$lib/api';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	interface Schedule {
		id: string;
		app_type: string;
		action: string;
		days: string[];
		hour: number;
		minute: number;
		enabled: boolean;
	}

	let schedules = $state<Schedule[]>([]);
	let loading = $state(true);

	async function load() {
		loading = true;
		try {
			schedules = await api.get<Schedule[]>('/schedules');
		} catch {
			schedules = [];
		}
		loading = false;
	}

	$effect(() => { load(); });

	function formatTime(h: number, m: number): string {
		return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
	}
</script>

<svelte:head><title>Scheduling - Lurkarr</title></svelte:head>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-surface-50">Schedules</h1>
		<Button size="sm">Add Schedule</Button>
	</div>

	{#if loading}
		<div class="space-y-2">
			{#each Array(3) as _}
				<div class="h-16 rounded-lg bg-surface-800/50 animate-pulse" />
			{/each}
		</div>
	{:else if schedules.length === 0}
		<Card>
			<p class="text-sm text-surface-500 text-center py-8">No schedules configured</p>
		</Card>
	{:else}
		<div class="space-y-2">
			{#each schedules as sched}
				<Card class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<Badge variant={sched.enabled ? 'success' : 'error'}>
							{sched.enabled ? 'Active' : 'Inactive'}
						</Badge>
						<div>
							<span class="font-medium text-surface-100 capitalize">{sched.app_type}</span>
							<span class="text-surface-500 mx-2">&middot;</span>
							<span class="text-surface-300">{sched.action}</span>
						</div>
					</div>
					<div class="text-right">
						<span class="font-mono text-surface-200">{formatTime(sched.hour, sched.minute)}</span>
						<p class="text-xs text-surface-500">{sched.days.join(', ') || 'Every day'}</p>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
