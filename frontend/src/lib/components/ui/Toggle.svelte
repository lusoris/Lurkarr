<script lang="ts">
	interface Props {
		checked?: boolean;
		label?: string;
		disabled?: boolean;
		onchange?: (checked: boolean) => void;
	}

	let {
		checked = $bindable(false),
		label = '',
		disabled = false,
		onchange
	}: Props = $props();

	function toggle() {
		if (disabled) return;
		checked = !checked;
		onchange?.(checked);
	}
</script>

<button
	type="button"
	role="switch"
	aria-checked={checked}
	{disabled}
	onclick={toggle}
	class="group inline-flex items-center gap-3 disabled:opacity-50 disabled:cursor-not-allowed"
>
	<span
		class="relative inline-flex h-6 w-11 shrink-0 rounded-full transition-colors
			{checked ? 'bg-lurk-600' : 'bg-surface-700'}"
	>
		<span
			class="inline-block h-5 w-5 rounded-full bg-white shadow-sm transition-transform mt-0.5
				{checked ? 'translate-x-5.5 ml-0' : 'translate-x-0.5'}"
		/>
	</span>
	{#if label}
		<span class="text-sm text-surface-300">{label}</span>
	{/if}
</button>
