<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import CollapsibleCard from '$lib/components/ui/CollapsibleCard.svelte';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';

	const toasts = getToasts();
	import type { GeneralSettings, OIDCSettings } from '$lib/types';

	type SettingsTab = 'general' | 'sso';
	let activeTab = $state<SettingsTab>('general');

	let general = $state<GeneralSettings | null>(null);
	let oidc = $state<OIDCSettings | null>(null);
	let saving = $state(false);
	let savingOIDC = $state(false);
	let testingOIDC = $state(false);

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

	async function testOIDC() {
		if (!oidc?.issuer_url) {
			toasts.error('Issuer URL is required to test');
			return;
		}
		testingOIDC = true;
		try {
			await api.post('/oidc/test', { issuer_url: oidc.issuer_url, client_id: oidc.client_id });
			toasts.success('OIDC provider is reachable');
		} catch {
			toasts.error('Could not reach OIDC provider — check the Issuer URL');
		}
		testingOIDC = false;
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Settings - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Settings" description="General behaviour, security and single sign-on configuration.">
		{#snippet actions()}
			<HelpDrawer page="settings" />
		{/snippet}
	</PageHeader>

	<Tabs
		tabs={[
			{ value: 'general', label: 'General' },
			{ value: 'sso', label: 'Single Sign-On' }
		]}
		bind:value={activeTab}
	/>

	{#if activeTab === 'general'}
		{#if general}
			<!-- ── Lurking Behaviour ─────────────────────────────── -->
			<CollapsibleCard title="Lurking Behaviour">
				<p class="text-xs text-muted-foreground mb-4">Controls how Lurkarr searches and manages your media libraries.</p>
				<div class="space-y-4">
					<Input bind:value={general.stateful_reset_hours} type="number" label="State Reset (hours)" hint="Hours after which lurk progress resets and starts fresh" />
					<Input bind:value={general.max_download_queue_size} type="number" label="Max Download Queue Size (0 = disabled)" hint="Pause lurking when the download queue has this many items or more. 0 disables" />
				</div>
			</CollapsibleCard>

			<!-- ── API & Command Execution ───────────────────────────── -->
			<CollapsibleCard title="API & Command Execution">
				<p class="text-xs text-muted-foreground mb-4">Tune how Lurkarr communicates with your Arr apps.</p>
				<div class="space-y-4">
					<Input bind:value={general.api_timeout} type="number" label="API Timeout (seconds)" hint="How long to wait for arr API responses before timing out" />
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input bind:value={general.command_wait_delay} type="number" label="Command Wait Delay (ms)" hint="Delay between command-completion checks" />
						<Input bind:value={general.command_wait_attempts} type="number" label="Command Wait Attempts" hint="Max retries for command completion" />
					</div>
				</div>
			</CollapsibleCard>

			<!-- ── Security ──────────────────────────────────────── -->
			<CollapsibleCard title="Security">
				<p class="text-xs text-muted-foreground mb-4">Connection security and authentication settings.</p>
				<div class="space-y-4">
					<Toggle bind:checked={general.ssl_verify} label="SSL Verification" hint="Verify TLS certificates when connecting to arr apps" />
					<Toggle bind:checked={general.proxy_auth_bypass} label="Proxy Auth Bypass" hint="Trust X-Forwarded headers from a reverse proxy for authentication" />
				</div>
			</CollapsibleCard>

			<div class="flex justify-end">
				<Button onclick={saveGeneral} loading={saving}>Save Settings</Button>
			</div>
		{:else}
			<Skeleton rows={3} height="h-20" />
		{/if}
	{:else if activeTab === 'sso'}
		{#if oidc}
			<CollapsibleCard title="OpenID Connect Provider">
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
			</CollapsibleCard>

			{#if oidc.enabled}
				<CollapsibleCard title="User Management">
					<p class="text-xs text-muted-foreground mb-4">Control how OIDC users are provisioned.</p>
					<div class="space-y-4">
						<Toggle bind:checked={oidc.auto_create} label="Auto-Create Users" hint="Automatically create local accounts for new OIDC users on first login" />
						<Input bind:value={oidc.admin_group} type="text" label="Admin Group" hint="OIDC group claim value that grants admin privileges (leave empty to disable)" />
					</div>
				</CollapsibleCard>
			{/if}

			<div class="flex justify-end gap-2">
				{#if oidc.enabled}
					<Button variant="outline" onclick={testOIDC} loading={testingOIDC}>Test Connection</Button>
				{/if}
				<Button onclick={saveOIDC} loading={savingOIDC}>Save OIDC Settings</Button>
			</div>
		{:else}
			<Skeleton rows={1} height="h-20" />
		{/if}
	{/if}
</div>

<ScrollToTop />
