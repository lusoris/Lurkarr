<script lang="ts">
	import '../app.css';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();
</script>

<div class="flex flex-col md:flex-row h-screen overflow-hidden">
	<Sidebar />

	<main class="flex-1 overflow-y-auto">
		<div class="max-w-7xl mx-auto px-4 sm:px-6 py-4 sm:py-6">
			{@render children()}
		</div>
	</main>
</div>

<Toast />
