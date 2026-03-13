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
	let passkeyEnabled = $state(false);
	let needsSetup = $state(false);
	let checkingSetup = $state(true);

	async function checkSetup() {
		try {
			const [setupRes, oidcRes, passkeyRes] = await Promise.all([
				fetch('/api/auth/setup'),
				fetch('/api/auth/oidc/info'),
				fetch('/api/auth/passkey/info')
			]);
			if (setupRes.ok) {
				const data = await setupRes.json();
				needsSetup = data.needs_setup === true;
			}
			if (oidcRes.ok) {
				const data = await oidcRes.json();
				oidcEnabled = data.enabled === true;
			}
			if (passkeyRes.ok) {
				const data = await passkeyRes.json();
				passkeyEnabled = data.enabled === true;
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

	function base64urlToBuffer(base64url: string): ArrayBuffer {
		const base64 = base64url.replace(/-/g, '+').replace(/_/g, '/');
		const pad = base64.length % 4;
		const padded = pad ? base64 + '='.repeat(4 - pad) : base64;
		const binary = atob(padded);
		const bytes = new Uint8Array(binary.length);
		for (let i = 0; i < binary.length; i++) bytes[i] = binary.charCodeAt(i);
		return bytes.buffer;
	}

	function bufferToBase64url(buffer: ArrayBuffer): string {
		const bytes = new Uint8Array(buffer);
		let str = '';
		for (const b of bytes) str += String.fromCharCode(b);
		return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
	}

	async function loginPasskey() {
		error = '';
		loading = true;
		try {
			// Begin discoverable login
			const beginRes = await fetch('/api/auth/passkey/login/begin', { method: 'POST' });
			if (!beginRes.ok) throw new Error('Failed to start passkey login');
			const options = await beginRes.json();

			// Convert for browser
			options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
			if (options.publicKey.allowCredentials) {
				for (const c of options.publicKey.allowCredentials) {
					c.id = base64urlToBuffer(c.id);
				}
			}

			const assertion = await navigator.credentials.get({ publicKey: options.publicKey }) as PublicKeyCredential;
			if (!assertion) throw new Error('Passkey login cancelled');

			const response = assertion.response as AuthenticatorAssertionResponse;

			const body = {
				id: bufferToBase64url(assertion.rawId),
				rawId: bufferToBase64url(assertion.rawId),
				type: assertion.type,
				response: {
					authenticatorData: bufferToBase64url(response.authenticatorData),
					clientDataJSON: bufferToBase64url(response.clientDataJSON),
					signature: bufferToBase64url(response.signature),
					userHandle: response.userHandle ? bufferToBase64url(response.userHandle) : ''
				}
			};

			const finishRes = await fetch('/api/auth/passkey/login/finish', {
				method: 'POST',
				headers: { 'Content-Type': 'application/json' },
				credentials: 'same-origin',
				body: JSON.stringify(body)
			});
			if (!finishRes.ok) {
				const data = await finishRes.json().catch(() => ({ error: 'Passkey login failed' }));
				throw new Error(data.error || 'Passkey login failed');
			}

			await auth.check();
			goto('/');
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Passkey login failed';
			if (!msg.includes('cancelled') && !msg.includes('abort')) {
				error = msg;
			}
		}
		loading = false;
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

			{#if (oidcEnabled || passkeyEnabled) && !needsSetup}
				<div class="relative">
					<div class="absolute inset-0 flex items-center">
						<div class="w-full border-t border-surface-700"></div>
					</div>
					<div class="relative flex justify-center text-xs">
						<span class="bg-surface-950 px-2 text-surface-500">or</span>
					</div>
				</div>

				{#if passkeyEnabled}
					<button
						onclick={loginPasskey}
						disabled={loading}
						class="w-full rounded-lg bg-surface-800 border border-surface-700 px-4 py-3 text-sm font-medium text-surface-200 hover:bg-surface-700 transition-colors flex items-center justify-center gap-2 disabled:opacity-50"
					>
						<svg class="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.864 4.243A7.5 7.5 0 0119.5 10.5c0 2.92-.556 5.709-1.568 8.268M5.742 6.364A7.465 7.465 0 004.5 10.5a48.667 48.667 0 00-1.298 8.568M5.742 6.364L3 4.5M5.742 6.364l2.121 2.121m0 0A7.465 7.465 0 0110.5 7.5c1.56 0 3.03.476 4.243 1.293M7.864 8.485l2.121 2.121m0 0a7.465 7.465 0 014.53-1.606c.896 0 1.76.157 2.56.442M10 10.5l2.121 2.121M12.121 12.621A48.578 48.578 0 0120.25 18.4M12.121 12.621L10.5 14.242"/></svg>
						Sign in with Passkey
					</button>
				{/if}

				{#if oidcEnabled}
					<button
						onclick={loginOIDC}
						class="w-full rounded-lg bg-surface-800 border border-surface-700 px-4 py-3 text-sm font-medium text-surface-200 hover:bg-surface-700 transition-colors"
					>
						Sign in with SSO
					</button>
				{/if}
			{/if}
		{/if}
	</div>
</div>
