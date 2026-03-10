<script lang="ts">
	interface Props {
		value?: string | number;
		type?: string;
		placeholder?: string;
		label?: string;
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
		error = '',
		disabled = false,
		class: className = '',
		oninput
	}: Props = $props();
</script>

<label class="block space-y-1.5 {className}">
	{#if label}
		<span class="text-sm font-medium text-surface-300">{label}</span>
	{/if}
	<input
		{type}
		{placeholder}
		{disabled}
		{oninput}
		bind:value
		class="w-full rounded-lg border px-3 py-2 text-sm bg-surface-900 text-surface-100
			placeholder:text-surface-500 transition-colors
			{error ? 'border-red-500 focus:ring-red-500' : 'border-surface-700 focus:border-lurk-500 focus:ring-lurk-500'}
			focus:outline-none focus:ring-1 disabled:opacity-50"
	/>
	{#if error}
		<p class="text-xs text-red-400">{error}</p>
	{/if}
</label>
