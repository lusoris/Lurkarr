<script lang="ts">
	import { getToasts } from '$lib/stores/toast.svelte';
	import { fly } from 'svelte/transition';
	import { cubicOut } from 'svelte/easing';

	const toasts = getToasts();

	const typeStyles: Record<string, string> = {
		success: 'border-lurk-500 bg-lurk-950/80',
		error: 'border-red-500 bg-red-950/80',
		info: 'border-blue-500 bg-blue-950/80',
		warning: 'border-amber-500 bg-amber-950/80'
	};

	const typeIcons: Record<string, string> = {
		success: '\u2713',
		error: '\u2717',
		info: '\u2139',
		warning: '\u26A0'
	};
</script>

<div class="fixed bottom-4 right-4 z-[100] flex flex-col gap-2 max-w-sm">
	{#each toasts.items as toast (toast.id)}
		<div
			transition:fly={{ x: 50, duration: 250, easing: cubicOut }}
			class="flex items-center gap-3 rounded-lg border px-4 py-3 text-sm text-surface-100 shadow-xl backdrop-blur-sm
				{typeStyles[toast.type]}"
		>
			<span class="text-lg">{typeIcons[toast.type]}</span>
			<span class="flex-1">{toast.message}</span>
			<button
				onclick={() => toasts.remove(toast.id)}
				class="text-surface-400 hover:text-surface-100 shrink-0"
			>&times;</button>
		</div>
	{/each}
</div>
