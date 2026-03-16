<script lang="ts">
	import { ArrowUp } from 'lucide-svelte';
	import { Button } from '$lib/components/ui/button';

	let visible = $state(false);

	function onScroll() {
		visible = window.scrollY > 400;
	}

	function scrollToTop() {
		window.scrollTo({ top: 0, behavior: 'smooth' });
	}
</script>

<svelte:window onscroll={onScroll} />

{#if visible}
	<div class="fixed bottom-6 right-6 z-50">
		<Button variant="outline" size="icon" class="rounded-full shadow-lg" onclick={scrollToTop}>
			<ArrowUp class="h-4 w-4" />
		</Button>
	</div>
{/if}
