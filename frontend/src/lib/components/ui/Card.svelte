<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { Card as ShadcnCard } from './card';

	interface Props {
		children: Snippet;
		class?: string;
		onclick?: () => void;
	}

	let { children, class: className = '', onclick }: Props = $props();
</script>

{#if onclick}
	<ShadcnCard
		class={cn('p-5 gap-0 cursor-pointer hover:bg-accent/50 transition-colors', className)}
		role="button"
		tabindex={0}
		onclick={onclick}
		onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onclick(); } }}
	>
		{@render children()}
	</ShadcnCard>
{:else}
	<ShadcnCard class={cn('p-5 gap-0', className)}>
		{@render children()}
	</ShadcnCard>
{/if}
