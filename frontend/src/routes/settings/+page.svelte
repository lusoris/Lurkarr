<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const toasts = getToasts();

	interface GeneralSettings {
		secret_key: string;
		proxy_auth_bypass: boolean;
		ssl_verify: boolean;
		api_timeout: number;
		stateful_reset_hours: number;
		command_wait_delay: number;
		command_wait_attempts: number;
		min_download_queue_size: number;
	}

	interface OIDCSettings {
		enabled: boolean;
		issuer_url: string;
		client_id: string;
		client_secret: string;
		redirect_url: string;
		scopes: string;
		auto_create: boolean;
		admin_group: string;
	}

	let general = $state<GeneralSettings | null>(null);
	let oidc = $state<OIDCSettings | null>(null);
	let saving = $state(false);
	let savingOIDC = $state(false);

	async function load() {
		api.get<GeneralSettings>('/settings/general').then(r => general = r).catch(() => {});
		api.get<OIDCSettings>('/oidc/settings').then(r => oidc = r).catch(() => {});
	}

	async function saveGeneral() {
		if (!general) return;
		saving = true;
		try {
			await api.put('/settings/general', general);
			toasts.success('General settings saved');
		} catch {
			toasts.error('Failed to save general settings');
		}
		saving = false;
	}

	async function saveOIDC() {
		if (!oidc) return;
		savingOIDC = true;
		try {
			const result = await api.put<OIDCSettings>('/oidc/settings', oidc);
			oidc = result;
			toasts.success('OIDC settings saved');
		} catch {
			toasts.error('Failed to save OIDC settings');
		}
		savingOIDC = false;
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Settings - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-foreground">Settings</h1>

	{#if general}
	<!-- ── Lurking Behaviour ─────────────────────────────── -->
	<Card>
		<h2 class="text-lg font-semibold text-foreground mb-1">Lurking Behaviour</h2>
		<p class="text-xs text-muted-foreground mb-4">Controls how Lurkarr searches and manages your media libraries.</p>
		<div class="space-y-4">
			<Input bind:value={general.stateful_reset_hours} type="number" label="State Reset (hours)" hint="Hours after which lurk progress resets and starts fresh" />
			<Input bind:value={general.min_download_queue_size} type="number" label="Min Download Queue Size (-1 = disabled)" hint="Pause lurking if the download queue has fewer items. -1 disables" />
		</div>
	</Card>

	<!-- ── API & Command Execution ───────────────────────── -->
	<Card>
		<h2 class="text-lg font-semibold text-foreground mb-1">API &amp; Command Execution</h2>
		<p class="text-xs text-muted-foreground mb-4">Tune how Lurkarr communicates with your Arr apps.</p>
		<div class="space-y-4">
			<Input bind:value={general.api_timeout} type="number" label="API Timeout (seconds)" hint="How long to wait for arr API responses before timing out" />
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<Input bind:value={general.command_wait_delay} type="number" label="Command Wait Delay (ms)" hint="Delay between command-completion checks" />
				<Input bind:value={general.command_wait_attempts} type="number" label="Command Wait Attempts" hint="Max retries for command completion" />
			</div>
		</div>
	</Card>

	<!-- ── Security ──────────────────────────────────────── -->
	<Card>
		<h2 class="text-lg font-semibold text-foreground mb-1">Security</h2>
		<p class="text-xs text-muted-foreground mb-4">Connection security and authentication settings.</p>
		<div class="space-y-4">
			<Toggle bind:checked={general.ssl_verify} label="SSL Verification" hint="Verify TLS certificates when connecting to arr apps" />
			<Toggle bind:checked={general.proxy_auth_bypass} label="Proxy Auth Bypass" hint="Trust X-Forwarded headers from a reverse proxy for authentication" />
		</div>
	</Card>

	<div class="flex justify-end">
		<Button onclick={saveGeneral} loading={saving}>Save Settings</Button>
	</div>
	{:else}
	<Card>
		<div class="space-y-4">
			{#each Array(3) as _}
				<div class="h-20 rounded-xl bg-muted/50 animate-pulse"></div>
			{/each}
		</div>
	</Card>
	{/if}

	<!-- ── OIDC / SSO ───────────────────────────────────────────── -->
	<h2 class="text-xl font-bold text-foreground mt-2">Single Sign-On (OIDC)</h2>

	{#if oidc}
	<Card>
		<h2 class="text-lg font-semibold text-foreground mb-1">OpenID Connect Provider</h2>
		<p class="text-xs text-muted-foreground mb-4">Configure an OIDC provider (Authentik, Keycloak, Authelia, etc.) for SSO login.</p>
		<div class="space-y-4">
			<Toggle bind:checked={oidc.enabled} label="Enable OIDC" hint="Allow users to sign in via the configured OIDC provider" />

			{#if oidc.enabled}
			<Input bind:value={oidc.issuer_url} type="text" label="Issuer URL" hint="The OIDC provider's issuer URL (e.g. https://auth.example.com/application/o/lurkarr/)" />
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<Input bind:value={oidc.client_id} type="text" label="Client ID" hint="OAuth2 client ID from your OIDC provider" />
				<Input bind:value={oidc.client_secret} type="password" label="Client Secret" hint="OAuth2 client secret" />
			</div>
			<Input bind:value={oidc.redirect_url} type="text" label="Redirect URL" hint="Callback URL — usually https://your-domain/api/auth/oidc/callback" />
			<Input bind:value={oidc.scopes} type="text" label="Scopes" hint="Comma-separated scopes (default: openid,profile,email)" />
			{/if}
		</div>
	</Card>

	{#if oidc.enabled}
	<Card>
		<h2 class="text-lg font-semibold text-foreground mb-1">User Management</h2>
		<p class="text-xs text-muted-foreground mb-4">Control how OIDC users are provisioned.</p>
		<div class="space-y-4">
			<Toggle bind:checked={oidc.auto_create} label="Auto-Create Users" hint="Automatically create local accounts for new OIDC users on first login" />
			<Input bind:value={oidc.admin_group} type="text" label="Admin Group" hint="OIDC group claim value that grants admin privileges (leave empty to disable)" />
		</div>
	</Card>
	{/if}

	<div class="flex justify-end">
		<Button onclick={saveOIDC} loading={savingOIDC}>Save OIDC Settings</Button>
	</div>
	{:else}
	<Card>
		<div class="h-20 rounded-xl bg-muted/50 animate-pulse"></div>
	</Card>
	{/if}
</div>
