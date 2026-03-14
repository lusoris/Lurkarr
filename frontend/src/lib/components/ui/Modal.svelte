<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import * as Dialog from './dialog';

	interface Props {
		open: boolean;
		title?: string;
		onclose: () => void;
		children: Snippet;
		class?: string;
	}

	let { open = $bindable(false), title = '', onclose, children, class: className = '' }: Props = $props();
</script>

<Dialog.Root bind:open onOpenChange={(v) => { if (!v) onclose(); }}>
	<Dialog.Content class={cn('max-h-[90vh] overflow-y-auto', className)}>
		{#if title}
			<Dialog.Header>
				<Dialog.Title>{title}</Dialog.Title>
			</Dialog.Header>
		{/if}
		{@render children()}
	</Dialog.Content>
</Dialog.Root>
