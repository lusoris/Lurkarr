<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';

	interface Props {
		children: Snippet;
		class?: string;
		onclick?: () => void;
	}

	let { children, class: className = '', onclick }: Props = $props();
</script>

{#if onclick}
	<div
		class={cn(
			'rounded-xl border border-border bg-card text-card-foreground shadow-sm p-5',
			'cursor-pointer hover:bg-accent/50 transition-colors',
			className
		)}
		role="button"
		tabindex="0"
		onclick={onclick}
		onkeydown={(e) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onclick(); } }}
	>
		{@render children()}
	</div>
{:else}
	<div
		class={cn(
			'rounded-xl border border-border bg-card text-card-foreground shadow-sm p-5',
			className
		)}
	>
		{@render children()}
	</div>
{/if}
