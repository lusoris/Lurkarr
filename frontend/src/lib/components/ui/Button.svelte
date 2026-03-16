<script lang="ts">
	import type { Snippet } from 'svelte';
	import { cn } from '$lib/lib/utils';
	import { Button as ShadcnButton, buttonVariants, type ButtonVariant, type ButtonSize } from './button';
	import { Loader2 } from '@lucide/svelte';

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
		href?: string;
		class?: string;
		onclick?: (e: MouseEvent) => void;
		children: Snippet;
		[key: string]: unknown;
	}

	let {
		variant = 'primary',
		size = 'md',
		disabled = false,
		loading = false,
		type = 'button',
		href,
		class: className = '',
		onclick,
		children,
		...restProps
	}: Props = $props();

	const classes = $derived(cn(buttonVariants({ variant: variantMap[variant], size: sizeMap[size] }), className));
</script>

{#if href}
<a
	{href}
	class={classes}
	{...restProps}
>
	{#if loading}
		<Loader2 class="h-4 w-4 animate-spin" />
	{/if}
	{@render children()}
</a>
{:else}
<button
	{type}
	disabled={disabled || loading}
	{onclick}
	class={classes}
	{...restProps}
>
	{#if loading}
		<Loader2 class="h-4 w-4 animate-spin" />
	{/if}
	{@render children()}
</button>
{/if}
