<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost';
		size?: 'sm' | 'md' | 'lg';
		disabled?: boolean;
		loading?: boolean;
		type?: 'button' | 'submit';
		class?: string;
		onclick?: () => void;
		children: Snippet;
	}

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		loading = false,
		type = 'button',
		class: className = '',
		onclick,
		children
	}: Props = $props();

	const variants: Record<string, string> = {
		primary: 'bg-lurk-600 hover:bg-lurk-500 text-white',
		secondary: 'bg-surface-700 hover:bg-surface-600 text-surface-100',
		danger: 'bg-red-600 hover:bg-red-500 text-white',
		ghost: 'bg-transparent hover:bg-surface-800 text-surface-300'
	};

	const sizes: Record<string, string> = {
		sm: 'px-3 py-1.5 text-xs',
		md: 'px-4 py-2 text-sm',
		lg: 'px-6 py-3 text-base'
	};
</script>

<button
	{type}
	{disabled}
	{onclick}
	class="inline-flex items-center justify-center gap-2 rounded-lg font-medium transition-colors
		focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-lurk-500
		disabled:opacity-50 disabled:cursor-not-allowed
		{variants[variant]} {sizes[size]} {className}"
>
	{#if loading}
		<svg class="animate-spin h-4 w-4" viewBox="0 0 24 24" fill="none">
			<circle cx="12" cy="12" r="10" stroke="currentColor" stroke-width="4" class="opacity-25" />
			<path fill="currentColor" d="M4 12a8 8 0 018-8V0C5.4 0 0 5.4 0 12h4z" class="opacity-75" />
		</svg>
	{/if}
	{@render children()}
</button>
