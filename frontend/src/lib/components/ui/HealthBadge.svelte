<script lang="ts">
	import type { HealthInfo } from '$lib/types';
	import { fmtVersion } from '$lib/format';
	import Badge from './Badge.svelte';
	import * as Popover from './popover';

	interface Props {
		health: HealthInfo | null | undefined;
		name?: string;
	}

	let { health, name }: Props = $props();

	const ok = $derived(health?.status === 'ok');
</script>

{#if !health}
	<span class="w-3 h-3 rounded-full border-2 border-muted-foreground/50 border-t-muted-foreground animate-spin shrink-0"></span>
{:else}
	<Popover.Root>
		<Popover.Trigger class="cursor-default">
			{#if ok}
				<Badge variant="success" class="text-[10px] px-1.5 py-0.5 shrink-0">
					<span class="w-1.5 h-1.5 rounded-full bg-emerald-400 mr-1"></span>
					{fmtVersion(health.version) || 'online'}
				</Badge>
			{:else}
				<Badge variant="error" class="text-[10px] px-1.5 py-0.5 shrink-0">
					<span class="w-1.5 h-1.5 rounded-full bg-destructive mr-1"></span>
					{health.version ? fmtVersion(health.version) : 'offline'}
				</Badge>
			{/if}
		</Popover.Trigger>
		<Popover.Content class="w-48 p-3 text-xs space-y-1.5" side="bottom" align="end">
			{#if name}
				<p class="font-medium text-foreground">{name}</p>
			{/if}
			<div class="flex justify-between">
				<span class="text-muted-foreground">Status</span>
				<span class={ok ? 'text-emerald-400' : 'text-destructive'}>{ok ? 'Online' : 'Offline'}</span>
			</div>
			{#if health.version}
				<div class="flex justify-between">
					<span class="text-muted-foreground">Version</span>
					<span class="text-foreground font-mono">{fmtVersion(health.version)}</span>
				</div>
			{/if}
		</Popover.Content>
	</Popover.Root>
{/if}
