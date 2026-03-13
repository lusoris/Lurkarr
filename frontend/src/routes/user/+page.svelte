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

	$effect(() => { load(); loadSessions(); });
</script>

<svelte:head><title>Profile - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Profile</h1>

	{#if user}
		<!-- Username -->
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Username</h2>
			<div class="space-y-3">
				<Input bind:value={newUsername} label="Username" />
				<Button onclick={updateUsername} loading={saving}>Update Username</Button>
			</div>
		</Card>

		<!-- Password -->
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Change Password</h2>
			<div class="space-y-3">
				<Input bind:value={currentPassword} type="password" label="Current Password" />
				<Input bind:value={newPassword} type="password" label="New Password" />
				<Button onclick={updatePassword} loading={saving}>Update Password</Button>
			</div>
		</Card>

		<!-- Two-Factor Authentication -->
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Two-Factor Authentication</h2>
			{#if user.has_2fa}
				<div class="space-y-3">
					<div class="flex items-center gap-2">
						<span class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-green-500/20 text-green-400">
							<svg class="w-3.5 h-3.5" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/></svg>
							Enabled
						</span>
					</div>
					<p class="text-sm text-surface-400">Your account is protected with TOTP two-factor authentication.</p>
					<div class="flex gap-2">
						<Button variant="secondary" onclick={regenerateCodes} loading={saving}>Regenerate Recovery Codes</Button>
						{#if confirmDisable2FA}
							<span class="flex items-center gap-2 text-xs">
								<span class="text-surface-400">Disable 2FA?</span>
								<button onclick={disable2FA} class="rounded px-2 py-1 bg-red-600 text-white text-xs hover:bg-red-500">Yes</button>
								<button onclick={() => confirmDisable2FA = false} class="rounded px-2 py-1 bg-surface-700 text-surface-300 text-xs hover:bg-surface-600">No</button>
							</span>
						{:else}
							<Button variant="danger" onclick={() => confirmDisable2FA = true} loading={saving}>Disable 2FA</Button>
						{/if}
					</div>
				</div>
			{:else}
				<div class="space-y-3">
					<p class="text-sm text-surface-400">Add an extra layer of security with a TOTP authenticator app.</p>
					<Button onclick={enable2FA} loading={totpEnabling}>Enable 2FA</Button>
				</div>
			{/if}
		</Card>

		<!-- Active Sessions -->
		<Card>
			<div class="flex items-center justify-between mb-4">
				<h2 class="text-lg font-semibold text-surface-200">Active Sessions</h2>
				{#if sessions.length > 1}
					{#if confirmRevokeAll}
						<span class="flex items-center gap-2 text-xs">
							<span class="text-surface-400">Revoke all?</span>
							<button onclick={() => { revokeAllSessions(); confirmRevokeAll = false; }} class="rounded px-2 py-1 bg-red-600 text-white text-xs hover:bg-red-500">Yes</button>
							<button onclick={() => confirmRevokeAll = false} class="rounded px-2 py-1 bg-surface-700 text-surface-300 text-xs hover:bg-surface-600">No</button>
						</span>
					{:else}
						<Button variant="danger" size="sm" onclick={() => confirmRevokeAll = true}>Revoke All Others</Button>
					{/if}
				{/if}
			</div>
			{#if sessions.length > 0}
				<div class="space-y-2">
					{#each sessions as s}
						<div class="flex items-center justify-between p-3 rounded-lg bg-surface-800/50 border border-surface-700/50">
							<div class="min-w-0">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium text-surface-200">{parseUA(s.user_agent)}</span>
									{#if s.current}
										<span class="px-1.5 py-0.5 text-[10px] rounded bg-lurk-600/30 text-lurk-400 font-medium">Current</span>
									{/if}
								</div>
								<p class="text-xs text-surface-500 mt-0.5">
									{s.ip_address || 'Unknown IP'} &middot; Created {formatDate(s.created_at)} &middot; Expires {formatDate(s.expires_at)}
								</p>
							</div>
							{#if !s.current}
								{#if confirmRevokeSession === s.id}
									<span class="flex items-center gap-1 shrink-0">
										<button onclick={() => { revokeSession(s.id); confirmRevokeSession = null; }} class="rounded px-1.5 py-0.5 bg-red-600 text-white text-[10px] hover:bg-red-500">Yes</button>
										<button onclick={() => confirmRevokeSession = null} class="rounded px-1.5 py-0.5 bg-surface-700 text-surface-300 text-[10px] hover:bg-surface-600">No</button>
									</span>
								{:else}
									<Button variant="ghost" size="sm" onclick={() => confirmRevokeSession = s.id}>Revoke</Button>
								{/if}
							{/if}
						</div>
					{/each}
				</div>
			{:else}
				<p class="text-sm text-surface-500 text-center py-2">No active sessions</p>
			{/if}
		</Card>

		<!-- Account info -->
		<Card>
			<p class="text-xs text-surface-500">
				Account created: {new Date(user.created_at).toLocaleDateString()}
				&middot; Provider: {user.auth_provider}
				{#if user.is_admin}&middot; <span class="text-lurk-400">Admin</span>{/if}
			</p>
		</Card>
	{:else}
		<div class="space-y-6">
			{#each Array(4) as _}
				<div class="h-32 rounded-xl bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{/if}
</div>

<!-- TOTP Setup Modal -->
<Modal open={showTOTPSetup} title="Set Up Two-Factor Authentication" onclose={() => showTOTPSetup = false}>
	<div class="space-y-4">
		<p class="text-sm text-surface-300">Scan this QR code with your authenticator app (Google Authenticator, Authy, etc):</p>
		{#if totpQR}
			<div class="flex justify-center">
				<img src="data:image/png;base64,{totpQR}" alt="TOTP QR Code" class="w-48 h-48 rounded-lg bg-white p-2" />
			</div>
		{/if}
		<details class="text-xs">
			<summary class="text-surface-400 cursor-pointer hover:text-surface-200">Can't scan? Enter manually</summary>
			<code class="block mt-2 p-2 rounded bg-surface-800 text-surface-200 break-all select-all">{totpSecret}</code>
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
				<code class="block p-2 rounded bg-surface-800 text-surface-200 text-center text-sm font-mono select-all">{code}</code>
			{/each}
		</div>
		<Button onclick={() => showRecoveryCodes = false}>I've Saved These Codes</Button>
	</div>
</Modal>
