<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { Label } from './label';

	interface Props {
		value?: string | number;
		label?: string;
		hint?: string;
		disabled?: boolean;
		class?: string;
		onchange?: (e: Event) => void;
		children: Snippet;
	}

	let {
		value = $bindable(''),
		label = '',
		hint = '',
		disabled = false,
		class: className = '',
		onchange,
		children
	}: Props = $props();
</script>

<div class={cn('block space-y-1.5', className)}>
	{#if label}
		<Label>{label}</Label>
	{/if}
	<select
		bind:value
		{disabled}
		{onchange}
		class={cn(
			'border-input bg-background selection:bg-primary selection:text-primary-foreground ring-offset-background placeholder:text-muted-foreground flex h-9 w-full items-center rounded-md border px-3 py-1 text-sm shadow-xs transition-[color,box-shadow] outline-none',
			'focus-visible:border-ring focus-visible:ring-ring/50 focus-visible:ring-[3px]',
			'disabled:cursor-not-allowed disabled:opacity-50',
			'[&>option]:bg-popover [&>option]:text-popover-foreground'
		)}
	>
		{@render children()}
	</select>
	{#if hint}
		<p class="text-xs text-muted-foreground">{hint}</p>
	{/if}
</div>
