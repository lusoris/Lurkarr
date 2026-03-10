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

	interface ProwlarrSettings {
		url: string;
		api_key: string;
		enabled: boolean;
		sync_indexers: boolean;
		timeout: number;
	}

	interface SABnzbdSettings {
		url: string;
		api_key: string;
		enabled: boolean;
		timeout: number;
		category: string;
	}

	let general = $state<GeneralSettings | null>(null);
	let prowlarr = $state<ProwlarrSettings | null>(null);
	let sabnzbd = $state<SABnzbdSettings | null>(null);
	let saving = $state(false);

	async function load() {
		try {
			[general, prowlarr, sabnzbd] = await Promise.all([
				api.get<GeneralSettings>('/settings/general'),
				api.get<ProwlarrSettings>('/prowlarr/settings'),
				api.get<SABnzbdSettings>('/sabnzbd/settings')
			]);
		} catch { /* handled */ }
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

	async function saveProwlarr() {
		if (!prowlarr) return;
		saving = true;
		try {
			await api.put('/prowlarr/settings', prowlarr);
			toasts.success('Prowlarr settings saved');
		} catch {
			toasts.error('Failed to save Prowlarr settings');
		}
		saving = false;
	}

	async function saveSabnzbd() {
		if (!sabnzbd) return;
		saving = true;
		try {
			await api.put('/sabnzbd/settings', sabnzbd);
			toasts.success('SABnzbd settings saved');
		} catch {
			toasts.error('Failed to save SABnzbd settings');
		}
		saving = false;
	}

	async function testProwlarr() {
		if (!prowlarr) return;
		try {
			await api.post('/prowlarr/test', { url: prowlarr.url, api_key: prowlarr.api_key });
			toasts.success('Prowlarr connection successful');
		} catch {
			toasts.error('Prowlarr connection failed');
		}
	}

	async function testSabnzbd() {
		if (!sabnzbd) return;
		try {
			await api.post('/sabnzbd/test', { url: sabnzbd.url, api_key: sabnzbd.api_key });
			toasts.success('SABnzbd connection successful');
		} catch {
			toasts.error('SABnzbd connection failed');
		}
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Settings - Lurkarr</title></svelte:head>

<div class="space-y-8 max-w-2xl">
	<h1 class="text-2xl font-bold text-surface-50">Settings</h1>

	<!-- General Settings -->
	{#if general}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">General</h2>
			<div class="space-y-4">
				<Input bind:value={general.api_timeout} type="number" label="API Timeout (seconds)" />
				<Input bind:value={general.stateful_reset_hours} type="number" label="State Reset (hours)" />
				<Input bind:value={general.min_download_queue_size} type="number" label="Min Download Queue Size" />
				<Toggle bind:checked={general.ssl_verify} label="SSL Verification" />
				<Toggle bind:checked={general.proxy_auth_bypass} label="Proxy Auth Bypass" />
				<Button onclick={saveGeneral} loading={saving}>Save General</Button>
			</div>
		</Card>
	{/if}

	<!-- Prowlarr Settings -->
	{#if prowlarr}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Prowlarr</h2>
			<div class="space-y-4">
				<Toggle bind:checked={prowlarr.enabled} label="Enabled" />
				<Input bind:value={prowlarr.url} label="URL" placeholder="http://prowlarr:9696" />
				<Input bind:value={prowlarr.api_key} label="API Key" type="password" />
				<Toggle bind:checked={prowlarr.sync_indexers} label="Sync Indexers" />
				<Input bind:value={prowlarr.timeout} type="number" label="Timeout (seconds)" />
				<div class="flex gap-2">
					<Button onclick={saveProwlarr} loading={saving}>Save</Button>
					<Button variant="secondary" onclick={testProwlarr}>Test Connection</Button>
				</div>
			</div>
		</Card>
	{/if}

	<!-- SABnzbd Settings -->
	{#if sabnzbd}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">SABnzbd</h2>
			<div class="space-y-4">
				<Toggle bind:checked={sabnzbd.enabled} label="Enabled" />
				<Input bind:value={sabnzbd.url} label="URL" placeholder="http://sabnzbd:8080" />
				<Input bind:value={sabnzbd.api_key} label="API Key" type="password" />
				<Input bind:value={sabnzbd.category} label="Category" placeholder="Optional" />
				<Input bind:value={sabnzbd.timeout} type="number" label="Timeout (seconds)" />
				<div class="flex gap-2">
					<Button onclick={saveSabnzbd} loading={saving}>Save</Button>
					<Button variant="secondary" onclick={testSabnzbd}>Test Connection</Button>
				</div>
			</div>
		</Card>
	{/if}
</div>
