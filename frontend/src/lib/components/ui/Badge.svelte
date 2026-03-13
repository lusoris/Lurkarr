<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { tv } from 'tailwind-variants';

	const badgeVariants = tv({
		base: 'inline-flex items-center rounded-md border px-2 py-0.5 text-xs font-medium transition-colors',
		variants: {
			variant: {
				default: 'border-transparent bg-secondary text-secondary-foreground',
				success: 'border-transparent bg-emerald-500/15 text-emerald-400',
				warning: 'border-transparent bg-amber-500/15 text-amber-400',
				error: 'border-transparent bg-destructive/15 text-destructive',
				info: 'border-transparent bg-primary/15 text-primary'
			}
		},
		defaultVariants: {
			variant: 'default'
		}
	});

	interface Props {
		variant?: 'default' | 'success' | 'warning' | 'error' | 'info';
		class?: string;
		children: Snippet;
	}

	let { variant = 'default', class: className = '', children }: Props = $props();
</script>

<span class={cn(badgeVariants({ variant }), className)}>
	{@render children()}
</span>
