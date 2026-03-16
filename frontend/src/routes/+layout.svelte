<script lang="ts">
	import '../app.css';
	import { page } from '$app/state';
	import { getAuth } from '$lib/stores/auth.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Sidebar from '$lib/components/Sidebar.svelte';
	import CommandPalette from '$lib/components/CommandPalette.svelte';
	import Breadcrumbs from '$lib/components/Breadcrumbs.svelte';
	import Toast from '$lib/components/Toast.svelte';
	import * as S from '$lib/components/ui/sidebar';
	import { Loader2 } from 'lucide-svelte';
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	const auth = getAuth();
	const instances = getInstances();
	const isLogin = $derived(page.url.pathname.startsWith('/login'));

	$effect(() => {
		if (!isLogin) auth.check();
	});

	// Pre-fetch instances once authenticated.
	$effect(() => {
		if (auth.user) instances.fetch();
	});
</script>

{#if isLogin}
	{@render children()}
{:else if auth.loading}
	<div class="flex items-center justify-center h-screen bg-background">
		<Loader2 class="h-8 w-8 animate-spin text-primary" />
	</div>
{:else if auth.user}
	<S.Provider>
		<Sidebar />
		<CommandPalette />
		<S.Inset>
			<header class="flex items-center gap-2 border-b border-border px-4 h-12 md:hidden">
				<S.Trigger />
				<img src="/banner.png" alt="Lurkarr" class="h-7 w-auto object-contain" />
			</header>
			<main class="flex-1 overflow-y-auto bg-background">
				<div class="max-w-7xl mx-auto px-4 sm:px-6 py-4 sm:py-6">
					<Breadcrumbs />
					{@render children()}
				</div>
			</main>
		</S.Inset>
	</S.Provider>
{:else}
	<div class="flex flex-col items-center justify-center h-screen bg-background gap-4">
		<p class="text-sm text-muted-foreground">Unable to authenticate. Please try again.</p>
		<a href="/login" class="text-sm text-primary hover:underline">Go to Login</a>
	</div>
{/if}

<Toast />
