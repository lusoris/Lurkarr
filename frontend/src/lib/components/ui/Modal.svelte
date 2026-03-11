<script lang="ts">
	import type { Snippet } from 'svelte';

	interface Props {
		open: boolean;
		title?: string;
		onclose: () => void;
		children: Snippet;
	}

	let { open = $bindable(false), title = '', onclose, children }: Props = $props();
</script>

{#if open}
	<div class="fixed inset-0 z-50 flex items-center justify-center">
		<!-- Backdrop -->
		<button class="absolute inset-0 bg-black/60 backdrop-blur-sm" onclick={onclose} aria-label="Close"></button>

		<!-- Panel -->
		<div class="relative z-10 w-full max-w-lg rounded-xl border border-surface-700 bg-surface-900 p-6 shadow-2xl">
			{#if title}
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-surface-100">{title}</h2>
					<button onclick={onclose} class="text-surface-400 hover:text-surface-100 transition-colors" aria-label="Close">
						<svg class="w-5 h-5" viewBox="0 0 20 20" fill="currentColor">
							<path fill-rule="evenodd" d="M4.293 4.293a1 1 0 011.414 0L10 8.586l4.293-4.293a1 1 0 111.414 1.414L11.414 10l4.293 4.293a1 1 0 01-1.414 1.414L10 11.414l-4.293 4.293a1 1 0 01-1.414-1.414L8.586 10 4.293 5.707a1 1 0 010-1.414z" clip-rule="evenodd" />
						</svg>
					</button>
				</div>
			{/if}
			{@render children()}
		</div>
	</div>
{/if}
