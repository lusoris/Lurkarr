<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import { Switch as ShadcnSwitch } from './switch';

	interface Props {
		checked?: boolean;
		label?: string;
		hint?: string;
		disabled?: boolean;
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
</script>

<div class="flex flex-col gap-1">
	<div class="flex items-center gap-3">
		<ShadcnSwitch
			bind:checked
			{disabled}
			onCheckedChange={(v) => onchange?.(v)}
			class={cn(color && checked ? color : '')}
		/>
		{#if label}
			<span class="text-sm text-foreground">{label}</span>
		{/if}
	</div>
	{#if hint}
		<p class="text-xs text-muted-foreground ml-12">{hint}</p>
	{/if}
</div>
