<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { Button as ShadcnButton, buttonVariants, type ButtonVariant, type ButtonSize } from './button';
	import { Loader2 } from 'lucide-svelte';

	// Map our app-level variant names to shadcn variant names.
	const variantMap: Record<string, ButtonVariant> = {
		primary: 'default',
		secondary: 'secondary',
		danger: 'destructive',
		ghost: 'ghost',
		outline: 'outline',
		link: 'link'
	};

	// Map our size names to shadcn size names.
	const sizeMap: Record<string, ButtonSize> = {
		sm: 'sm',
		md: 'default',
		lg: 'lg',
		icon: 'icon'
	};

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
	class={cn(buttonVariants({ variant: variantMap[variant], size: sizeMap[size] }), className)}
>
	{#if loading}
		<Loader2 class="h-4 w-4 animate-spin" />
	{/if}
	{@render children()}
</button>
