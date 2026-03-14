<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { badgeVariants } from './badge';
	import { tv } from 'tailwind-variants';

	const appBadgeVariants = tv({
		extend: badgeVariants,
		variants: {
			variant: {
				default: 'bg-secondary text-secondary-foreground border-transparent',
				success: 'bg-emerald-500/15 text-emerald-400 border-transparent',
				warning: 'bg-amber-500/15 text-amber-400 border-transparent',
				error: 'bg-destructive/15 text-destructive border-transparent',
				info: 'bg-primary/15 text-primary border-transparent'
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

<span class={cn(appBadgeVariants({ variant }), className)}>
	{@render children()}
</span>
