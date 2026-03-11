<script lang="ts">
	import { goto } from '$app/navigation';
	import { getAuth } from '$lib/stores/auth.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const auth = getAuth();

	let username = $state('');
	let password = $state('');
	let totp = $state('');
	let error = $state('');
	let loading = $state(false);
	let showTotp = $state(false);
	let oidcEnabled = $state(false);

	async function checkOIDC() {
		try {
			const res = await fetch('/api/auth/oidc/info');
			if (res.ok) {
				const data = await res.json();
				oidcEnabled = data.enabled === true;
			}
		} catch {
			// Silently ignore — OIDC button just won't show.
		}
	}

	checkOIDC();

	async function submit() {
		error = '';
		loading = true;
		try {
			await auth.login(username, password, showTotp ? totp : undefined);
			goto('/');
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Login failed';
			if (msg.includes('2fa') || msg.includes('totp')) {
				showTotp = true;
			} else {
				error = msg;
			}
		}
		loading = false;
	}

	function loginOIDC() {
		window.location.href = '/api/auth/oidc/login';
	}
</script>

<svelte:head><title>Login - Lurkarr</title></svelte:head>

<div class="min-h-screen flex items-center justify-center bg-surface-950">
	<div class="w-full max-w-sm space-y-6">
		<div class="text-center">
			<span class="text-5xl">&#x1F438;</span>
			<h1 class="text-2xl font-bold text-surface-50 mt-3">Lurkarr</h1>
			<p class="text-sm text-surface-500 mt-1">Sign in to continue</p>
		</div>

		<form onsubmit={submit} class="space-y-4 bg-surface-900 border border-surface-800 rounded-xl p-6">
			{#if error}
				<div class="rounded-lg bg-red-950/50 border border-red-800 px-4 py-3 text-sm text-red-300">
					{error}
				</div>
			{/if}

			<Input bind:value={username} label="Username" placeholder="admin" />
			<Input bind:value={password} type="password" label="Password" />

			{#if showTotp}
				<Input bind:value={totp} label="2FA Code" placeholder="000000" />
			{/if}

			<Button type="submit" {loading} class="w-full">Sign In</Button>
		</form>

		{#if oidcEnabled}
			<div class="relative">
				<div class="absolute inset-0 flex items-center">
					<div class="w-full border-t border-surface-700"></div>
				</div>
				<div class="relative flex justify-center text-xs">
					<span class="bg-surface-950 px-2 text-surface-500">or</span>
				</div>
			</div>

			<button
				onclick={loginOIDC}
				class="w-full rounded-lg bg-surface-800 border border-surface-700 px-4 py-3 text-sm font-medium text-surface-200 hover:bg-surface-700 transition-colors"
			>
				Sign in with SSO
			</button>
		{/if}
	</div>
</div>
