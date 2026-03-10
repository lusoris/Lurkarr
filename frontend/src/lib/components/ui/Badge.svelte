<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		variant?: 'default' | 'success' | 'warning' | 'error' | 'info';
		children: Snippet;
	}

	let { variant = 'default', children }: Props = $props();

	const variants: Record<string, string> = {
		default: 'bg-surface-700 text-surface-300',
		success: 'bg-lurk-900/50 text-lurk-400 border border-lurk-700/30',
		warning: 'bg-amber-900/50 text-amber-400 border border-amber-700/30',
		error: 'bg-red-900/50 text-red-400 border border-red-700/30',
		info: 'bg-blue-900/50 text-blue-400 border border-blue-700/30'
	};
</script>

<span class="inline-flex items-center rounded-md px-2 py-0.5 text-xs font-medium {variants[variant]}">
	{@render children()}
</span>
