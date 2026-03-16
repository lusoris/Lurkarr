<script lang="ts">
	import { api } from '$lib/api';
	import { base64urlToBuffer, bufferToBase64url } from '$lib/webauthn';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import { ShieldCheck, Plus, Fingerprint, Pencil } from '@lucide/svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import * as Collapsible from '$lib/components/ui/collapsible';

	const toasts = getToasts();
	import type { LurkarrUser, UserSession, Passkey } from '$lib/types';

	let user = $state<LurkarrUser | null>(null);
	let sessions = $state<UserSession[]>([]);
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
	let passkeys = $state<Passkey[]>([]);
	let passkeyRegistering = $state(false);
	let passkeyEnabled = $state(false);
	let confirmDeletePasskey = $state<string | null>(null);
	let renamingPasskey = $state<string | null>(null);
	let renameValue = $state('');
	let pendingPasskeyCredential = $state<any>(null);
	let passkeyNameInput = $state('');
	let showPasskeyNameModal = $state(false);

	async function checkPasskeySupport() {
		try {
			const res = await fetch('/api/auth/passkey/info');
			if (res.ok) {
				const data = await res.json();
				passkeyEnabled = data.enabled === true;
			}
		} catch { /* passkey info not available — feature disabled */ }
	}

	async function loadPasskeys() {
		if (!passkeyEnabled) return;
		try {
			passkeys = await api.get<Passkey[]>('/passkeys');
		} catch {
			console.warn('Failed to load passkeys');
		}
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

			pendingPasskeyCredential = credential;
			passkeyNameInput = '';
			showPasskeyNameModal = true;
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Passkey registration failed');
			passkeyRegistering = false;
		}
	}

	async function finishPasskeyRegistration() {
		const credential = pendingPasskeyCredential;
		if (!credential) return;
		const name = passkeyNameInput.trim() || 'Passkey';
		showPasskeyNameModal = false;
		try {
			const response = credential.response as AuthenticatorAttestationResponse;

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
			user = await api.get<LurkarrUser>('/user');
			newUsername = user?.username ?? '';
		} catch {
			console.warn('Failed to load user profile');
		}
	}

	async function loadSessions() {
		try {
			sessions = await api.get<UserSession[]>('/sessions');
		} catch {
			console.warn('Failed to load sessions');
		}
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
	<PageHeader title="Profile" description="Manage your account, security, and active sessions.">
		{#snippet actions()}
			<HelpDrawer page="user" />
		{/snippet}
	</PageHeader>

	{#if user}
		<!-- Username -->
		<Card>
			<h3 class="text-sm font-semibold text-foreground mb-3">Username</h3>
			<div class="space-y-3">
				<Input bind:value={newUsername} label="Username" />
				<Button onclick={updateUsername} loading={saving}>Update Username</Button>
			</div>
		</Card>

		<!-- Password -->
		<Card>
			<h3 class="text-sm font-semibold text-foreground mb-3">Change Password</h3>
			<div class="space-y-3">
				<Input bind:value={currentPassword} type="password" label="Current Password" />
				<Input bind:value={newPassword} type="password" label="New Password" />
				<Button onclick={updatePassword} loading={saving}>Update Password</Button>
			</div>
		</Card>

		<!-- Two-Factor Authentication -->
		<Card>
			<h3 class="text-sm font-semibold text-foreground mb-3">Two-Factor Authentication</h3>
			{#if user.has_2fa}
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<Badge variant="success" class="rounded-full px-2.5 py-1">
						<ShieldCheck class="w-3.5 h-3.5 mr-1" />
							Enabled
						</Badge>
					</div>
					<p class="text-sm text-muted-foreground">Your account is protected with TOTP two-factor authentication.</p>
					<div class="flex gap-2">
						<Button variant="secondary" onclick={regenerateCodes} loading={saving}>Regenerate Recovery Codes</Button>
						<ConfirmAction active={confirmDisable2FA} message="Disable 2FA?" onconfirm={disable2FA} oncancel={() => confirmDisable2FA = false}>
							<Button variant="danger" onclick={() => confirmDisable2FA = true} loading={saving}>Disable 2FA</Button>
						</ConfirmAction>
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
					<h3 class="text-sm font-semibold text-foreground">Passkeys</h3>
					<Button size="sm" onclick={registerPasskey} loading={passkeyRegistering}>
					<Plus class="w-4 h-4 mr-1 inline" />
						Add Passkey
					</Button>
				</div>
				<p class="text-sm text-muted-foreground mb-4">Sign in without a password using your device's biometrics, security key, or platform authenticator.</p>
				{#if passkeys.length > 0}
					<div class="space-y-2">
						{#each passkeys as pk}
							<div class="flex items-center justify-between p-3 rounded-lg bg-muted/50 border border-border/50">
								<div class="min-w-0 flex items-center gap-3">
								<Fingerprint class="w-5 h-5 text-muted-foreground shrink-0" />
									<div>
										{#if renamingPasskey === pk.id}
											<form onsubmit={(e: Event) => { e.preventDefault(); renamePasskey(pk.id); }} class="flex items-center gap-2">
														<Input bind:value={renameValue} class="w-40" />
														<Button type="submit" size="sm" variant="link" class="h-auto p-0 text-xs">Save</Button>
														<Button type="button" size="sm" variant="ghost" class="h-auto p-0 text-xs" onclick={() => renamingPasskey = null}>Cancel</Button>
											</form>
										{:else}
											<span class="text-sm font-medium text-foreground">{pk.name}</span>
												<Button size="sm" variant="ghost" class="ml-2 h-auto p-0" onclick={() => { renamingPasskey = pk.id; renameValue = pk.name; }}>
											<Pencil class="w-3 h-3 inline" />
											</Button>
										{/if}
										<p class="text-xs text-muted-foreground mt-0.5">Added {formatDate(pk.created_at)}</p>
									</div>
								</div>
								<div class="shrink-0">
								<ConfirmAction active={confirmDeletePasskey === pk.id} onconfirm={() => deletePasskey(pk.id)} oncancel={() => confirmDeletePasskey = null}>
									<Button variant="ghost" size="sm" onclick={() => confirmDeletePasskey = pk.id}>Delete</Button>
								</ConfirmAction>
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
				<h3 class="text-sm font-semibold text-foreground">Active Sessions</h3>
				{#if sessions.length > 1}
					<ConfirmAction active={confirmRevokeAll} message="Revoke all?" onconfirm={() => { revokeAllSessions(); confirmRevokeAll = false; }} oncancel={() => confirmRevokeAll = false}>
						<Button variant="danger" size="sm" onclick={() => confirmRevokeAll = true}>Revoke All Others</Button>
					</ConfirmAction>
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
								<ConfirmAction active={confirmRevokeSession === s.id} onconfirm={() => { revokeSession(s.id); confirmRevokeSession = null; }} oncancel={() => confirmRevokeSession = null}>
									<Button variant="ghost" size="sm" onclick={() => confirmRevokeSession = s.id}>Revoke</Button>
								</ConfirmAction>
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
		<Skeleton rows={4} height="h-32" />
	{/if}
</div>

<!-- Passkey Name Modal -->
<Modal open={showPasskeyNameModal} title="Name Your Passkey" onclose={() => { showPasskeyNameModal = false; passkeyRegistering = false; }}>
	<form onsubmit={(e: Event) => { e.preventDefault(); finishPasskeyRegistration(); }} class="space-y-4">
		<Input bind:value={passkeyNameInput} label="Passkey Name" placeholder="e.g. MacBook Touch ID" />
		<div class="flex justify-end gap-2">
			<Button variant="secondary" onclick={() => { showPasskeyNameModal = false; passkeyRegistering = false; }}>Cancel</Button>
			<Button type="submit">Save</Button>
		</div>
	</form>
</Modal>

<!-- TOTP Setup Modal -->
<Modal open={showTOTPSetup} title="Set Up Two-Factor Authentication" onclose={() => showTOTPSetup = false}>
	<div class="space-y-4">
		<p class="text-sm text-muted-foreground">Scan this QR code with your authenticator app (Google Authenticator, Authy, etc):</p>
		{#if totpQR}
			<div class="flex justify-center">
				<img src="data:image/jpeg;base64,{totpQR}" alt="TOTP QR Code" class="w-48 h-48 rounded-lg bg-white p-2" />
			</div>
		{/if}
		<Collapsible.Root class="text-xs">
			<Collapsible.Trigger class="text-muted-foreground cursor-pointer hover:text-foreground transition-colors">Can't scan? Enter manually</Collapsible.Trigger>
			<Collapsible.Content>
				<code class="block mt-2 p-2 rounded bg-muted text-foreground break-all select-all">{totpSecret}</code>
			</Collapsible.Content>
		</Collapsible.Root>
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
		<Alert.Root variant="warning">
			<Alert.Description>Save these codes somewhere safe. Each code can only be used once. If you lose your authenticator, these are the only way to access your account.</Alert.Description>
		</Alert.Root>
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
			{#each recoveryCodes as code}
				<code class="block p-2 rounded bg-muted text-foreground text-center text-sm font-mono select-all">{code}</code>
			{/each}
		</div>
		<Button onclick={() => showRecoveryCodes = false}>I've Saved These Codes</Button>
	</div>
</Modal>

<ScrollToTop />
