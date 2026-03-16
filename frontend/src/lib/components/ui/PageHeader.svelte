<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';

	interface Props {
		title: string;
		description?: string;
		class?: string;
		actions?: Snippet;
	}

	let { title, description = '', class: className = '', actions }: Props = $props();
</script>

<div class={cn('flex flex-col sm:flex-row sm:items-center justify-between gap-2', className)}>
	<div>
		<h1 class="text-2xl font-bold tracking-tight text-foreground">{title}</h1>
		{#if description}
			<p class="text-sm text-muted-foreground mt-1">{description}</p>
		{/if}
	</div>
	{#if actions}
		<div class="flex items-center gap-2 shrink-0">
			{@render actions()}
		</div>
	{/if}
</div>
