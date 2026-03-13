<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		children: Snippet;
		class?: string;
		onclick?: () => void;
	}

	let { children, class: className = '', onclick }: Props = $props();
</script>

<div class="rounded-xl border border-surface-800 bg-surface-900/50 backdrop-blur-sm p-5 {className}" role={onclick ? 'button' : undefined} tabindex={onclick ? 0 : undefined} onclick={onclick} onkeydown={onclick ? (e: KeyboardEvent) => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); onclick(); } } : undefined}>
	{@render children()}
</div>
