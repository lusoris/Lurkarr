<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import { Input as ShadcnInput } from './input';
	import { Label } from './label';

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
		min?: number | string;
		max?: number | string;
		step?: number | string;
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
		oninput,
		min,
		max,
		step
	}: Props = $props();
</script>

<div class={cn('block space-y-1.5', className)}>
	{#if label}
		<Label>{label}</Label>
	{/if}
	<ShadcnInput
		{type}
		{placeholder}
		{disabled}
		{oninput}
		{min}
		{max}
		{step}
		bind:value
		aria-invalid={error ? true : undefined}
	/>
	{#if error}
		<p class="text-xs text-destructive">{error}</p>
	{:else if hint}
		<p class="text-xs text-muted-foreground">{hint}</p>
	{/if}
</div>
