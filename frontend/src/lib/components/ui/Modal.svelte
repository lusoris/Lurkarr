<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { Dialog } from 'bits-ui';
	import { X } from 'lucide-svelte';

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
	<Dialog.Portal>
		<Dialog.Overlay class="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm data-[state=open]:animate-in data-[state=closed]:animate-out data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0" />
		<Dialog.Content
			class={cn(
				'fixed left-1/2 top-1/2 z-50 w-full max-w-lg -translate-x-1/2 -translate-y-1/2',
				'rounded-xl border border-border bg-card p-6 shadow-2xl duration-200',
				'data-[state=open]:animate-in data-[state=closed]:animate-out',
				'data-[state=closed]:fade-out-0 data-[state=open]:fade-in-0',
				'data-[state=closed]:zoom-out-95 data-[state=open]:zoom-in-95',
				className
			)}
		>
			{#if title}
				<div class="flex items-center justify-between mb-4">
					<Dialog.Title class="text-lg font-semibold text-foreground">{title}</Dialog.Title>
					<Dialog.Close class="rounded-sm opacity-70 ring-offset-background transition-opacity hover:opacity-100 focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2">
						<X class="h-4 w-4" />
						<span class="sr-only">Close</span>
					</Dialog.Close>
				</div>
			{/if}
			{@render children()}
		</Dialog.Content>
	</Dialog.Portal>
</Dialog.Root>
