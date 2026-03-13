<script lang="ts">
	import { goto } from '$app/navigation';
	import { getAuth } from '$lib/stores/auth.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const auth = getAuth();

	let username = $state('');
	let password = $state('');
	let totp = $state('');
	let recoveryCode = $state('');
	let error = $state('');
	let loading = $state(false);
	let showTotp = $state(false);
	let useRecovery = $state(false);
	let oidcEnabled = $state(false);
	let needsSetup = $state(false);
	let checkingSetup = $state(true);

	async function checkSetup() {
		try {
			const [setupRes, oidcRes] = await Promise.all([
				fetch('/api/auth/setup'),
				fetch('/api/auth/oidc/info')
			]);
			if (setupRes.ok) {
				const data = await setupRes.json();
				needsSetup = data.needs_setup === true;
			}
			if (oidcRes.ok) {
				const data = await oidcRes.json();
				oidcEnabled = data.enabled === true;
			}
		} catch {
			// Silently ignore
		} finally {
			checkingSetup = false;
		}
	}

	checkSetup();

	async function submit() {
		error = '';
		loading = true;
		try {
			if (needsSetup) {
				const res = await fetch('/api/auth/setup', {
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({ username, password })
				});
				if (!res.ok) {
					const data = await res.json().catch(() => ({ error: 'Setup failed' }));
					throw new Error(data.error || 'Setup failed');
				}
				await auth.check();
				goto('/');
			} else {
				await auth.login(username, password, showTotp ? totp : undefined, useRecovery ? recoveryCode : undefined);
				goto('/');
			}
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

<svelte:head><title>{needsSetup ? 'Setup' : 'Login'} - Lurkarr</title></svelte:head>

<div class="min-h-screen flex items-center justify-center bg-surface-950">
	<div class="w-full max-w-sm space-y-6">
		<div class="text-center">
			<img src="/banner.png" alt="Lurkarr" class="max-w-[16rem] w-full h-auto mx-auto object-contain" />
		</div>

		{#if checkingSetup}
			<div class="flex justify-center">
				<div class="h-6 w-6 animate-spin rounded-full border-2 border-lurk-500 border-t-transparent"></div>
			</div>
		{:else}
			{#if needsSetup}
				<p class="text-center text-sm text-surface-400">Create your admin account to get started.</p>
			{/if}

			<form onsubmit={(e: Event) => { e.preventDefault(); submit(); }} class="space-y-4 bg-surface-900 border border-surface-800 rounded-xl p-6">
				{#if error}
					<div class="rounded-lg bg-red-950/50 border border-red-800 px-4 py-3 text-sm text-red-300">
						{error}
					</div>
				{/if}

				<Input bind:value={username} label="Username" placeholder="admin" />
				<Input bind:value={password} type="password" label="Password" />

				{#if showTotp && !needsSetup}
					{#if useRecovery}
						<Input bind:value={recoveryCode} label="Recovery Code" placeholder="xxxx-xxxx" />
						<button type="button" onclick={() => useRecovery = false} class="text-xs text-surface-400 hover:text-surface-200 transition-colors">Use authenticator code instead</button>
					{:else}
						<Input bind:value={totp} label="2FA Code" placeholder="000000" />
						<button type="button" onclick={() => useRecovery = true} class="text-xs text-surface-400 hover:text-surface-200 transition-colors">Lost your authenticator? Use a recovery code</button>
					{/if}
				{/if}

				<Button type="submit" {loading} class="w-full">{needsSetup ? 'Create Account' : 'Sign In'}</Button>
			</form>

			{#if oidcEnabled && !needsSetup}
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
		{/if}
	</div>
</div>
