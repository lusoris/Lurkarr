<script lang="ts">
	import type { Snippet } from 'svelte';
	import type { ComponentType } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import Card from './Card.svelte';

	interface Props {
		icon?: ComponentType;
		title: string;
		description?: string;
		class?: string;
		actions?: Snippet;
	}

	let { icon, title, description = '', class: className = '', actions }: Props = $props();
</script>

<Card class={cn('text-center py-10', className)}>
	<div role="status">
	{#if icon}
		{@const Icon = icon}
		<div class="mx-auto mb-3 flex h-12 w-12 items-center justify-center rounded-full bg-muted">
			<Icon class="h-6 w-6 text-muted-foreground" />
		</div>
	{/if}
	<p class="text-sm font-medium text-foreground">{title}</p>
	{#if description}
		<p class="text-sm text-muted-foreground mt-1 max-w-sm mx-auto">{description}</p>
	{/if}
	{#if actions}
		<div class="mt-4">
			{@render actions()}
		</div>
	{/if}
	</div>
</Card>
