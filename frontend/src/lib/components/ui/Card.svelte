<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		class?: string;
	}

	let { children, class: className = '' }: Props = $props();
</script>

<div class="rounded-xl border border-surface-800 bg-surface-900/50 backdrop-blur-sm p-5 {className}">
	{@render children()}
</div>
