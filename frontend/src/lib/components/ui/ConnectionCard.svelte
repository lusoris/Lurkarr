<script lang="ts">
	import type { HealthInfo } from '$lib/types';
	import { appAccentBorder, appDisplayName, appLogo } from '$lib';
	import HealthBadge from './HealthBadge.svelte';

	interface Props {
		name: string;
		subtitle?: string;
		appType: string;
		health: HealthInfo | null | undefined;
	}

	let { name, subtitle, appType, health }: Props = $props();

	const logo = $derived(appLogo(appType));
	const ok = $derived(health?.status === 'ok');
</script>

<div class="flex items-center gap-3 rounded-lg border px-3 py-2.5 border-l-2 {appAccentBorder(appType)} {health && !ok ? 'bg-destructive/5' : ''}">
	{#if logo}
		<img src={logo} alt="" class="w-5 h-5 rounded shrink-0" />
	{/if}
	<div class="min-w-0 flex-1">
		<p class="text-sm font-medium text-foreground truncate">{name}</p>
		<p class="text-[10px] text-muted-foreground truncate">{subtitle ?? appDisplayName(appType)}</p>
	</div>
	<HealthBadge {health} />
</div>
