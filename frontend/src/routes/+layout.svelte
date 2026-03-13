<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import { getAuth } from '$lib/stores/auth.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import { Loader2 } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	const auth = getAuth();
	const isLogin = $derived(page.url.pathname.startsWith('/login'));

	$effect(() => {
		if (!isLogin) auth.check();
	});
</script>

{#if isLogin}
	{@render children()}
{:else if auth.loading}
	<div class="flex items-center justify-center h-screen bg-background">
		<Loader2 class="h-8 w-8 animate-spin text-primary" />
	</div>
{:else if auth.user}
	<div class="flex flex-col md:flex-row h-screen overflow-hidden">
		<Sidebar />

		<main class="flex-1 overflow-y-auto bg-background">
			<div class="max-w-7xl mx-auto px-4 sm:px-6 py-4 sm:py-6">
				{@render children()}
			</div>
		</main>
	</div>
{/if}

<Toast />
