<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import { Checkbox as ShadcnCheckbox } from './checkbox';

	interface Props {
		checked?: boolean;
		label?: string;
		hint?: string;
		disabled?: boolean;
		class?: string;
		onchange?: (checked: boolean) => void;
	}

	let {
		checked = $bindable(false),
		label = '',
		hint = '',
		disabled = false,
		class: className = '',
		onchange
	}: Props = $props();
</script>

<div class={cn('flex items-center gap-2', className)}>
	<ShadcnCheckbox
		bind:checked
		{disabled}
		onCheckedChange={(v) => { if (typeof v === 'boolean') onchange?.(v); }}
	/>
	{#if label}
		<span class="text-sm text-foreground select-none">{label}</span>
	{/if}
	{#if hint}
		<span class="text-xs text-muted-foreground">{hint}</span>
	{/if}
</div>
