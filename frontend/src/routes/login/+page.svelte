<script lang="ts">
	import { goto } from '$app/navigation';
	import { getAuth } from '$lib/stores/auth.svelte';
	import { base64urlToBuffer, bufferToBase64url } from '$lib/webauthn';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Separator from '$lib/components/ui/Separator.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import { Loader2, Fingerprint, CircleAlert } from 'lucide-svelte';

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

<div class="min-h-screen flex items-center justify-center bg-background">
	<div class="w-full max-w-sm space-y-6">
		<div class="text-center">
			<img src="/banner.png" alt="Lurkarr" class="max-w-[16rem] w-full h-auto mx-auto object-contain" />
		</div>

		{#if checkingSetup}
			<div class="flex justify-center">
				<Loader2 class="h-6 w-6 animate-spin text-primary" />
			</div>
		{:else}
			{#if needsSetup}
				<p class="text-center text-sm text-muted-foreground">Create your admin account to get started.</p>
			{/if}

			<form onsubmit={(e: Event) => { e.preventDefault(); submit(); }} class="space-y-4 border border-border bg-card rounded-xl p-6 shadow-lg">
				{#if error}
					<Alert.Root variant="destructive">
						<CircleAlert class="h-4 w-4" />
						<Alert.Description>{error}</Alert.Description>
					</Alert.Root>
				{/if}

				<Input bind:value={username} label="Username" placeholder="admin" />
				<Input bind:value={password} type="password" label="Password" />

				{#if showTotp && !needsSetup}
					{#if useRecovery}
						<Input bind:value={recoveryCode} label="Recovery Code" placeholder="xxxx-xxxx" />
						<Button type="button" variant="link" class="h-auto p-0 text-xs text-muted-foreground" onclick={() => useRecovery = false}>Use authenticator code instead</Button>
					{:else}
						<Input bind:value={totp} label="2FA Code" placeholder="000000" />
						<Button type="button" variant="link" class="h-auto p-0 text-xs text-muted-foreground" onclick={() => useRecovery = true}>Lost your authenticator? Use a recovery code</Button>
					{/if}
				{/if}

				<Button type="submit" {loading} class="w-full">{needsSetup ? 'Create Account' : 'Sign In'}</Button>
			</form>

			{#if (oidcEnabled || passkeyEnabled) && !needsSetup}
				<div class="relative">
					<div class="absolute inset-0 flex items-center">
						<Separator />
					</div>
					<div class="relative flex justify-center text-xs">
						<span class="bg-background px-2 text-muted-foreground">or</span>
					</div>
				</div>

				{#if passkeyEnabled}
					<Button variant="outline" onclick={loginPasskey} disabled={loading} class="w-full h-11">
						<Fingerprint class="h-5 w-5" />
						Sign in with Passkey
					</Button>
				{/if}

				{#if oidcEnabled}
					<Button variant="outline" onclick={loginOIDC} class="w-full h-11">
						Sign in with SSO
					</Button>
				{/if}
			{/if}
		{/if}
	</div>
</div>
