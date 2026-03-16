<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import { Switch as ShadcnSwitch } from './switch';

	interface Props {
		checked?: boolean;
		label?: string;
		hint?: string;
		disabled?: boolean;
		/** Pass a bg-* class (e.g. from appBgColor) to color the switch when checked. */
		color?: string;
		onchange?: (checked: boolean) => void;
	}

	let {
		checked = $bindable(false),
		label = '',
		hint = '',
		disabled = false,
		color = '',
		onchange
	}: Props = $props();

	// Convert e.g. "bg-sky-500" → "data-[state=checked]:bg-sky-500" so it
	// properly overrides the switch primitive's default data-[state=checked]:bg-primary.
	const checkedColorClass = $derived(
		color ? `data-[state=checked]:${color}` : ''
	);
</script>

<div class="flex flex-col gap-1">
	<div class="flex items-center gap-3">
		<ShadcnSwitch
			bind:checked
			{disabled}
			onCheckedChange={(v) => onchange?.(v)}
			class={cn(checkedColorClass)}
		/>
		{#if label}
			<span class="text-sm text-foreground">{label}</span>
		{/if}
	</div>
	{#if hint}
		<p class="text-xs text-muted-foreground ml-12">{hint}</p>
	{/if}
</div>
