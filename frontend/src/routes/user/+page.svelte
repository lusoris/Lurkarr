<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';

	const toasts = getToasts();

	interface User {
		id: string;
		username: string;
		has_2fa: boolean;
		is_admin: boolean;
		auth_provider: string;
		created_at: string;
	}

	interface Session {
		id: string;
		created_at: string;
		expires_at: string;
		ip_address: string;
		user_agent: string;
		current: boolean;
	}

	let user = $state<User | null>(null);
	let sessions = $state<Session[]>([]);
	let newUsername = $state('');
	let currentPassword = $state('');
	let newPassword = $state('');
	let saving = $state(false);

	// TOTP state
	let totpEnabling = $state(false);
	let totpSecret = $state('');
	let totpQR = $state('');
	let totpCode = $state('');
	let recoveryCodes = $state<string[]>([]);
	let showRecoveryCodes = $state(false);
	let showTOTPSetup = $state(false);
	let confirmDisable2FA = $state(false);
	let confirmRevokeSession = $state<string | null>(null);
	let confirmRevokeAll = $state(false);

	// Passkey state
	interface Passkey {
		id: string;
		name: string;
		created_at: string;
	}
	let passkeys = $state<Passkey[]>([]);
	let passkeyRegistering = $state(false);
	let passkeyEnabled = $state(false);
	let confirmDeletePasskey = $state<string | null>(null);
	let renamingPasskey = $state<string | null>(null);
	let renameValue = $state('');

	async function checkPasskeySupport() {
		try {
			const res = await fetch('/api/auth/passkey/info');
			if (res.ok) {
				const data = await res.json();
				passkeyEnabled = data.enabled === true;
			}
		} catch { /* ignore */ }
	}

	async function loadPasskeys() {
		if (!passkeyEnabled) return;
		try {
			passkeys = await api.get<Passkey[]>('/passkeys');
		} catch { /* handled */ }
	}

	function bufferToBase64url(buffer: ArrayBuffer): string {
		const bytes = new Uint8Array(buffer);
		let str = '';
		for (const b of bytes) str += String.fromCharCode(b);
		return btoa(str).replace(/\+/g, '-').replace(/\//g, '_').replace(/=/g, '');
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

	async function registerPasskey() {
		passkeyRegistering = true;
		try {
			const options = await api.post<any>('/passkeys/register/begin');

			// Convert base64url fields for the browser API.
			options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
			options.publicKey.user.id = base64urlToBuffer(options.publicKey.user.id);
			if (options.publicKey.excludeCredentials) {
				for (const c of options.publicKey.excludeCredentials) {
					c.id = base64urlToBuffer(c.id);
				}
			}

			const credential = await navigator.credentials.create({ publicKey: options.publicKey }) as PublicKeyCredential;
			if (!credential) throw new Error('Registration cancelled');

			const response = credential.response as AuthenticatorAttestationResponse;

			const name = prompt('Name this passkey (e.g. "MacBook Touch ID")') || 'Passkey';

			const body = {
				id: bufferToBase64url(credential.rawId),
				rawId: bufferToBase64url(credential.rawId),
				type: credential.type,
				response: {
					attestationObject: bufferToBase64url(response.attestationObject),
					clientDataJSON: bufferToBase64url(response.clientDataJSON)
				}
			};

			// Send as raw fetch since the server reads the body as an http.Request
			const res = await fetch(`/api/passkeys/register/finish?name=${encodeURIComponent(name)}`, {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json',
					'X-CSRF-Token': api.getCsrfToken()
				},
				credentials: 'same-origin',
				body: JSON.stringify(body)
			});

			if (!res.ok) {
				const data = await res.json().catch(() => ({ error: 'Registration failed' }));
				throw new Error(data.error || 'Registration failed');
			}

			toasts.success('Passkey registered!');
			await loadPasskeys();
		} catch (e) {
			const msg = e instanceof Error ? e.message : 'Failed to register passkey';
			if (!msg.includes('cancelled') && !msg.includes('abort')) {
				toasts.error(msg);
			}
		}
		passkeyRegistering = false;
	}

	async function deletePasskey(id: string) {
		try {
			await api.del(`/passkeys/${id}`);
			toasts.success('Passkey deleted');
			confirmDeletePasskey = null;
			await loadPasskeys();
		} catch {
			toasts.error('Failed to delete passkey');
		}
	}

	async function renamePasskey(id: string) {
		if (!renameValue.trim()) return;
		try {
			await api.post(`/passkeys/${id}/rename`, { name: renameValue.trim() });
			toasts.success('Passkey renamed');
			renamingPasskey = null;
			renameValue = '';
			await loadPasskeys();
		} catch {
			toasts.error('Failed to rename passkey');
		}
	}

	async function load() {
		try {
			user = await api.get<User>('/user');
			newUsername = user?.username ?? '';
		} catch { /* handled */ }
	}

	async function loadSessions() {
		try {
			sessions = await api.get<Session[]>('/sessions');
		} catch { /* handled */ }
	}

	async function updateUsername() {
		saving = true;
		try {
			await api.post('/user/username', { username: newUsername });
			toasts.success('Username updated');
			await load();
		} catch {
			toasts.error('Failed to update username');
		}
		saving = false;
	}

	async function updatePassword() {
		if (!currentPassword || !newPassword) {
			toasts.error('Both fields required');
			return;
		}
		saving = true;
		try {
			await api.post('/user/password', { current_password: currentPassword, new_password: newPassword });
			toasts.success('Password updated');
			currentPassword = '';
			newPassword = '';
		} catch {
			toasts.error('Failed to update password');
		}
		saving = false;
	}

	async function enable2FA() {
		totpEnabling = true;
		try {
			const res = await api.post<{ secret: string; qr_base64: string; recovery_codes: string[] }>('/auth/2fa/enable');
			totpSecret = res.secret;
			totpQR = res.qr_base64;
			recoveryCodes = res.recovery_codes;
			showTOTPSetup = true;
		} catch {
			toasts.error('Failed to enable 2FA');
		}
		totpEnabling = false;
	}

	async function verify2FA() {
		if (!totpCode) {
			toasts.error('Enter the code from your authenticator');
			return;
		}
		saving = true;
		try {
			await api.post('/auth/2fa/verify', { code: totpCode });
			toasts.success('2FA enabled successfully');
			showTOTPSetup = false;
			showRecoveryCodes = true;
			totpCode = '';
			await load();
		} catch {
			toasts.error('Invalid code — try again');
		}
		saving = false;
	}

	async function disable2FA() {
		saving = true;
		try {
			await api.post('/auth/2fa/disable');
			toasts.success('2FA disabled');
			confirmDisable2FA = false;
			await load();
		} catch {
			toasts.error('Failed to disable 2FA');
		}
		saving = false;
	}

	async function regenerateCodes() {
		saving = true;
		try {
			const res = await api.post<{ recovery_codes: string[] }>('/auth/2fa/recovery-codes');
			recoveryCodes = res.recovery_codes;
			showRecoveryCodes = true;
			toasts.success('New recovery codes generated');
		} catch {
			toasts.error('Failed to regenerate codes');
		}
		saving = false;
	}

	async function revokeSession(id: string) {
		try {
			await api.del(`/sessions/${id}`);
			toasts.success('Session revoked');
			await loadSessions();
		} catch {
			toasts.error('Failed to revoke session');
		}
	}

	async function revokeAllSessions() {
		try {
			await api.del('/sessions');
			toasts.success('All other sessions revoked');
			await loadSessions();
		} catch {
			toasts.error('Failed to revoke sessions');
		}
	}

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString();
	}

	function parseUA(ua: string) {
		if (!ua) return 'Unknown';
		if (ua.includes('Firefox')) return 'Firefox';
		if (ua.includes('Edg/')) return 'Edge';
		if (ua.includes('Chrome')) return 'Chrome';
		if (ua.includes('Safari')) return 'Safari';
		return ua.slice(0, 40);
	}

	$effect(() => { load(); loadSessions(); checkPasskeySupport().then(() => loadPasskeys()); });
</script>

<svelte:head><title>Profile - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-foreground">Profile</h1>

	{#if user}
		<!-- Username -->
		<Card>
			<h2 class="text-lg font-semibold text-foreground mb-4">Username</h2>
			<div class="space-y-3">
				<Input bind:value={newUsername} label="Username" />
				<Button onclick={updateUsername} loading={saving}>Update Username</Button>
			</div>
		</Card>

		<!-- Password -->
		<Card>
			<h2 class="text-lg font-semibold text-foreground mb-4">Change Password</h2>
			<div class="space-y-3">
				<Input bind:value={currentPassword} type="password" label="Current Password" />
				<Input bind:value={newPassword} type="password" label="New Password" />
				<Button onclick={updatePassword} loading={saving}>Update Password</Button>
			</div>
		</Card>

		<!-- Two-Factor Authentication -->
		<Card>
			<h2 class="text-lg font-semibold text-foreground mb-4">Two-Factor Authentication</h2>
			{#if user.has_2fa}
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/></svg>
							Enabled
						</span>
					</div>
					<p class="text-sm text-muted-foreground">Your account is protected with TOTP two-factor authentication.</p>
					<div class="flex gap-2">
						<Button variant="secondary" onclick={regenerateCodes} loading={saving}>Regenerate Recovery Codes</Button>
						{#if confirmDisable2FA}
							<span class="flex items-center gap-2 text-xs">
								<span class="text-muted-foreground">Disable 2FA?</span>
								<button onclick={disable2FA} class="rounded px-2 py-1 bg-red-600 text-white text-xs hover:bg-red-500">Yes</button>
								<button onclick={() => confirmDisable2FA = false} class="rounded px-2 py-1 bg-secondary text-muted-foreground text-xs hover:bg-muted">No</button>
							</span>
						{:else}
							<Button variant="danger" onclick={() => confirmDisable2FA = true} loading={saving}>Disable 2FA</Button>
						{/if}
					</div>
				</div>
			{:else}
				<div class="space-y-3">
					<p class="text-sm text-muted-foreground">Add an extra layer of security with a TOTP authenticator app.</p>
					<Button onclick={enable2FA} loading={totpEnabling}>Enable 2FA</Button>
				</div>
			{/if}
		</Card>

		<!-- Passkeys -->
		{#if passkeyEnabled}
			<Card>
				<div class="flex items-center justify-between mb-4">
					<h2 class="text-lg font-semibold text-foreground">Passkeys</h2>
					<Button size="sm" onclick={registerPasskey} loading={passkeyRegistering}>
						<svg class="w-4 h-4 mr-1 inline" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/></svg>
						Add Passkey
					</Button>
				</div>
				<p class="text-sm text-muted-foreground mb-4">Sign in without a password using your device's biometrics, security key, or platform authenticator.</p>
				{#if passkeys.length > 0}
					<div class="space-y-2">
						{#each passkeys as pk}
							<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50 border border-border/50">
								<div class="min-w-0 flex items-center gap-3">
									<svg class="w-5 h-5 text-muted-foreground shrink-0" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M7.864 4.243A7.5 7.5 0 0119.5 10.5c0 2.92-.556 5.709-1.568 8.268M5.742 6.364A7.465 7.465 0 004.5 10.5a48.667 48.667 0 00-1.298 8.568M5.742 6.364L3 4.5M5.742 6.364l2.121 2.121m0 0A7.465 7.465 0 0110.5 7.5c1.56 0 3.03.476 4.243 1.293M7.864 8.485l2.121 2.121m0 0a7.465 7.465 0 014.53-1.606c.896 0 1.76.157 2.56.442M10 10.5l2.121 2.121M12.121 12.621A48.578 48.578 0 0120.25 18.4M12.121 12.621L10.5 14.242"/></svg>
									<div>
										{#if renamingPasskey === pk.id}
											<form onsubmit={(e: Event) => { e.preventDefault(); renamePasskey(pk.id); }} class="flex items-center gap-2">
<input bind:value={renameValue} class="bg-secondary border border-border rounded px-2 py-1 text-sm text-foreground w-40" />
													<button type="submit" class="text-xs text-primary hover:text-primary/80">Save</button>
													<button type="button" onclick={() => renamingPasskey = null} class="text-xs text-muted-foreground hover:text-foreground">Cancel</button>
											</form>
										{:else}
											<span class="text-sm font-medium text-foreground">{pk.name}</span>
											<button onclick={() => { renamingPasskey = pk.id; renameValue = pk.name; }} class="ml-2 text-xs text-muted-foreground hover:text-foreground" title="Rename">
												<svg class="w-3 h-3 inline" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M16.862 4.487l1.687-1.688a1.875 1.875 0 112.652 2.652L6.832 19.82a4.5 4.5 0 01-1.897 1.13l-2.685.8.8-2.685a4.5 4.5 0 011.13-1.897L16.863 4.487zm0 0L19.5 7.125"/></svg>
											</button>
										{/if}
										<p class="text-xs text-muted-foreground mt-0.5">Added {formatDate(pk.created_at)}</p>
									</div>
								</div>
								<div class="shrink-0">
									{#if confirmDeletePasskey === pk.id}
										<span class="flex items-center gap-1">
											<button onclick={() => deletePasskey(pk.id)} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
											<button onclick={() => confirmDeletePasskey = null} class="rounded px-1.5 py-0.5 bg-secondary text-muted-foreground text-[10px] hover:bg-muted">No</button>
										</span>
									{:else}
										<Button variant="ghost" size="sm" onclick={() => confirmDeletePasskey = pk.id}>Delete</Button>
									{/if}
								</div>
							</div>
						{/each}
					</div>
				{:else}
					<p class="text-sm text-muted-foreground text-center py-2">No passkeys registered yet</p>
				{/if}
			</Card>
		{/if}

		<!-- Active Sessions -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-foreground">Active Sessions</h2>
				{#if sessions.length > 1}
					{#if confirmRevokeAll}
						<span class="flex items-center gap-2 text-xs">
						<span class="text-muted-foreground">Revoke all?</span>
						<button onclick={() => { revokeAllSessions(); confirmRevokeAll = false; }} class="rounded px-2 py-1 bg-red-600 text-white text-xs hover:bg-red-500">Yes</button>
						<button onclick={() => confirmRevokeAll = false} class="rounded px-2 py-1 bg-secondary text-muted-foreground text-xs hover:bg-muted">No</button>
						</span>
					{:else}
						<Button variant="danger" size="sm" onclick={() => confirmRevokeAll = true}>Revoke All Others</Button>
					{/if}
				{/if}
			</div>
			{#if sessions.length > 0}
				<div class="space-y-2">
					{#each sessions as s}
					<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50 border border-border/50">
						<div class="min-w-0">
							<div class="flex items-center gap-2">
								<span class="text-sm font-medium text-foreground">{parseUA(s.user_agent)}</span>
								{#if s.current}
									<span class="px-1.5 py-0.5 text-[10px] rounded bg-primary/30 text-primary font-medium">Current</span>
									{/if}
								</div>
								<p class="text-xs text-muted-foreground mt-0.5">
									{s.ip_address || 'Unknown IP'} &middot; Created {formatDate(s.created_at)} &middot; Expires {formatDate(s.expires_at)}
								</p>
							</div>
							{#if !s.current}
								{#if confirmRevokeSession === s.id}
									<span class="flex items-center gap-1 shrink-0">
										<button onclick={() => { revokeSession(s.id); confirmRevokeSession = null; }} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
										<button onclick={() => confirmRevokeSession = null} class="rounded px-1.5 py-0.5 bg-secondary text-muted-foreground text-[10px] hover:bg-muted">No</button>
									</span>
								{:else}
									<Button variant="ghost" size="sm" onclick={() => confirmRevokeSession = s.id}>Revoke</Button>
								{/if}
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<p class="text-sm text-muted-foreground text-center py-2">No active sessions</p>
			{/if}
		</Card>

		<!-- Account info -->
		<Card>
			<p class="text-xs text-muted-foreground">
				Account created: {new Date(user.created_at).toLocaleDateString()}
				&middot; Provider: {user.auth_provider}
				{#if user.is_admin}&middot; <span class="text-primary">Admin</span>{/if}
			</p>
		</Card>
	{:else}
		<div class="space-y-6">
			{#each Array(4) as _}
				<div class="h-32 rounded-xl bg-muted/50 animate-pulse"></div>
			{/each}
		</div>
	{/if}
</div>

<!-- TOTP Setup Modal -->
<Modal open={showTOTPSetup} title="Set Up Two-Factor Authentication" onclose={() => showTOTPSetup = false}>
	<div class="space-y-4">
		<p class="text-sm text-muted-foreground">Scan this QR code with your authenticator app (Google Authenticator, Authy, etc):</p>
		{#if totpQR}
			<div class="flex justify-center">
				<img src="data:image/png;base64,{totpQR}" alt="TOTP QR Code" class="w-48 h-48 rounded-lg bg-white p-2" />
			</div>
		{/if}
		<details class="text-xs">
			<summary class="text-muted-foreground cursor-pointer hover:text-foreground">Can't scan? Enter manually</summary>
			<code class="block mt-2 p-2 rounded bg-muted text-foreground break-all select-all">{totpSecret}</code>
		</details>
		<Input bind:value={totpCode} label="Verification Code" placeholder="Enter 6-digit code" />
		<div class="flex gap-2">
			<Button onclick={verify2FA} loading={saving}>Verify &amp; Enable</Button>
			<Button variant="ghost" onclick={() => showTOTPSetup = false}>Cancel</Button>
		</div>
	</div>
</Modal>

<!-- Recovery Codes Modal -->
<Modal open={showRecoveryCodes} title="Recovery Codes" onclose={() => showRecoveryCodes = false}>
	<div class="space-y-4">
		<div class="p-3 rounded-lg bg-yellow-500/10 border border-yellow-500/20">
			<p class="text-sm text-yellow-300">Save these codes somewhere safe. Each code can only be used once. If you lose your authenticator, these are the only way to access your account.</p>
		</div>
		<div class="grid grid-cols-2 gap-2">
			{#each recoveryCodes as code}
				<code class="block p-2 rounded bg-muted text-foreground text-center text-sm font-mono select-all">{code}</code>
			{/each}
		</div>
		<Button onclick={() => showRecoveryCodes = false}>I've Saved These Codes</Button>
	</div>
</Modal>
