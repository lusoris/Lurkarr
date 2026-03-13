<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import { Switch } from 'bits-ui';

	interface Props {
		checked?: boolean;
		label?: string;
		hint?: string;
		disabled?: boolean;
		onchange?: (checked: boolean) => void;
	}

	let {
		checked = $bindable(false),
		label = '',
		hint = '',
		disabled = false,
		onchange
	}: Props = $props();

	function handleChange(v: boolean) {
		checked = v;
		onchange?.(v);
	}
</script>

<div class="flex flex-col gap-1">
	<div class="flex items-center gap-3">
		<Switch.Root
			bind:checked
			{disabled}
			onCheckedChange={handleChange}
			class={cn(
				'peer inline-flex h-5 w-9 shrink-0 cursor-pointer items-center rounded-full border-2 border-transparent shadow-sm transition-colors',
				'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background',
				'disabled:cursor-not-allowed disabled:opacity-50',
				checked ? 'bg-primary' : 'bg-input'
			)}
		>
			<Switch.Thumb
				class={cn(
					'pointer-events-none block h-4 w-4 rounded-full bg-background shadow-lg ring-0 transition-transform',
					checked ? 'translate-x-4' : 'translate-x-0'
				)}
			/>
		</Switch.Root>
		{#if label}
			<span class="text-sm text-foreground">{label}</span>
		{/if}
	</div>
	{#if hint}
		<p class="text-xs text-muted-foreground ml-12">{hint}</p>
	{/if}
</div>
