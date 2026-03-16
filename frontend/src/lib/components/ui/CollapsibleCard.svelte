<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import Card from './Card.svelte';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { ChevronRight } from '@lucide/svelte';

	interface Props {
		title: string;
		children: Snippet;
		open?: boolean;
		class?: string;
	}

	let { title, children, open = $bindable(true), class: className = '' }: Props = $props();
</script>

<Collapsible.Root bind:open>
	<Card class={className}>
		<Collapsible.Trigger class="flex items-center gap-2 w-full text-left cursor-pointer select-none group">
			<ChevronRight class="h-4 w-4 text-muted-foreground transition-transform duration-200 {open ? 'rotate-90' : ''}" />
			<h3 class="text-sm font-semibold text-foreground">{title}</h3>
		</Collapsible.Trigger>
		<Collapsible.Content>
			<div class="pt-3">
				{@render children()}
			</div>
		</Collapsible.Content>
	</Card>
</Collapsible.Root>
