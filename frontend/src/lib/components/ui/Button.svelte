<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { tv } from 'tailwind-variants';
	import { Loader2 } from 'lucide-svelte';

	const buttonVariants = tv({
		base: 'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-all focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 focus-visible:ring-offset-background disabled:pointer-events-none disabled:opacity-50',
		variants: {
			variant: {
				primary: 'bg-primary text-primary-foreground shadow-sm hover:bg-primary/90',
				secondary: 'bg-secondary text-secondary-foreground shadow-sm hover:bg-secondary/80',
				danger: 'bg-destructive text-destructive-foreground shadow-sm hover:bg-destructive/90',
				ghost: 'hover:bg-accent hover:text-accent-foreground',
				outline: 'border border-input bg-background shadow-sm hover:bg-accent hover:text-accent-foreground',
				link: 'text-primary underline-offset-4 hover:underline'
			},
			size: {
				sm: 'h-8 rounded-md px-3 text-xs',
				md: 'h-9 px-4 py-2',
				lg: 'h-10 rounded-md px-6',
				icon: 'h-9 w-9'
			}
		},
		defaultVariants: {
			variant: 'primary',
			size: 'md'
		}
	});

	interface Props {
		variant?: 'primary' | 'secondary' | 'danger' | 'ghost' | 'outline' | 'link';
		size?: 'sm' | 'md' | 'lg' | 'icon';
		disabled?: boolean;
		loading?: boolean;
		type?: 'button' | 'submit';
		class?: string;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
	}

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		loading = false,
		type = 'button',
		class: className = '',
		onclick,
		children
	}: Props = $props();
</script>

<button
	{type}
	disabled={disabled || loading}
	{onclick}
	class={cn(buttonVariants({ variant, size }), className)}
>
	{#if loading}
		<Loader2 class="h-4 w-4 animate-spin" />
	{/if}
	{@render children()}
</button>
