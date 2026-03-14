<script lang="ts">
	import { cn } from '$lib/lib/utils';
	import { Skeleton as ShadcnSkeleton } from './skeleton';

	interface Props {
		class?: string;
		rows?: number;
		height?: string;
	}

	let { class: className = '', rows = 1, height = 'h-10' }: Props = $props();
</script>

<div class={cn('space-y-2', className)}>
	{#each Array(rows) as _}
		<ShadcnSkeleton class={height} />
	{/each}
</div>
