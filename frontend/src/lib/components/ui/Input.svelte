<script lang="ts">
	import { cn } from '$lib/lib/utils';

	interface Props {
		value?: string | number;
		type?: string;
		placeholder?: string;
		label?: string;
		hint?: string;
		error?: string;
		disabled?: boolean;
		class?: string;
		oninput?: (e: Event) => void;
	}

	let {
		value = $bindable(''),
		type = 'text',
		placeholder = '',
		label = '',
		hint = '',
		error = '',
		disabled = false,
		class: className = '',
		oninput
	}: Props = $props();
</script>

<label class={cn('block space-y-1.5', className)}>
	{#if label}
		<span class="text-sm font-medium text-foreground">{label}</span>
	{/if}
	<input
		{type}
		{placeholder}
		{disabled}
		{oninput}
		bind:value
		class={cn(
			'flex h-9 w-full rounded-md border bg-transparent px-3 py-1 text-sm shadow-sm transition-colors',
			'file:border-0 file:bg-transparent file:text-sm file:font-medium file:text-foreground',
			'placeholder:text-muted-foreground',
			'focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring',
			'disabled:cursor-not-allowed disabled:opacity-50',
			error ? 'border-destructive focus-visible:ring-destructive' : 'border-input'
		)}
	/>
	{#if error}
		<p class="text-xs text-destructive">{error}</p>
	{:else if hint}
		<p class="text-xs text-muted-foreground">{hint}</p>
	{/if}
</label>
