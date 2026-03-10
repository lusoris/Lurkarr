<script lang="ts">
	import { getLogStream } from '$lib/stores/websocket.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';

	const logs = getLogStream();
	let filter = $state('');
	let autoScroll = $state(true);
	let logContainer: HTMLDivElement;

	$effect(() => { logs.connect(); return () => logs.disconnect(); });

	$effect(() => {
		if (autoScroll && logs.messages.length > 0 && logContainer) {
			logContainer.scrollTop = logContainer.scrollHeight;
		}
	});

	const filtered = $derived(
		filter
			? logs.messages.filter((m) =>
					m.message.toLowerCase().includes(filter.toLowerCase()) ||
					m.app_type.toLowerCase().includes(filter.toLowerCase())
				)
			: logs.messages
	);

	const levelColors: Record<string, string> = {
		error: 'text-red-400',
		warn: 'text-amber-400',
		info: 'text-lurk-400',
		debug: 'text-surface-500'
	};
</script>

<svelte:head><title>Logs - Lurkarr</title></svelte:head>

<div class="space-y-4 h-[calc(100vh-6rem)] flex flex-col">
	<div class="flex items-center justify-between gap-4">
		<h1 class="text-2xl font-bold text-surface-50">Live Logs</h1>
		<div class="flex items-center gap-3">
			<Badge variant={logs.connected ? 'success' : 'error'}>
				{logs.connected ? 'Connected' : 'Disconnected'}
			</Badge>
			<Button variant="ghost" size="sm" onclick={() => logs.clear()}>Clear</Button>
		</div>
	</div>

	<Input bind:value={filter} placeholder="Filter logs..." />

	<div
		bind:this={logContainer}
		class="flex-1 overflow-y-auto rounded-xl border border-surface-800 bg-surface-950 p-4 font-mono text-xs leading-relaxed"
	>
		{#each filtered as msg (msg.id)}
			<div class="flex gap-3 py-0.5 hover:bg-surface-900/50">
				<span class="text-surface-600 shrink-0 w-20">{new Date(msg.created_at).toLocaleTimeString()}</span>
				<span class="shrink-0 w-12 uppercase {levelColors[msg.level] ?? 'text-surface-400'}">{msg.level}</span>
				<span class="shrink-0 w-16 text-surface-500">{msg.app_type}</span>
				<span class="text-surface-200">{msg.message}</span>
			</div>
		{/each}
		{#if filtered.length === 0}
			<p class="text-surface-600 text-center py-8">No log entries{filter ? ' matching filter' : ''}</p>
		{/if}
	</div>
</div>
