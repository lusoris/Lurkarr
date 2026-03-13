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

	let general = $state<GeneralSettings | null>(null);
	let saving = $state(false);

	async function load() {
		api.get<GeneralSettings>('/settings/general').then(r => general = r).catch(() => {});
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

	$effect(() => { load(); });
</script>

<svelte:head><title>Settings - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Settings</h1>

	{#if general}
	<Card>
		<h2 class="text-lg font-semibold text-surface-200 mb-4">General</h2>
		<div class="space-y-4">
			<Input bind:value={general.api_timeout} type="number" label="API Timeout (seconds)" hint="How long to wait for arr API responses before timing out" />
			<Input bind:value={general.stateful_reset_hours} type="number" label="State Reset (hours)" hint="Hours after which lurk progress resets and starts fresh" />
			<Input bind:value={general.command_wait_delay} type="number" label="Command Wait Delay (ms)" hint="Delay between checking if an arr command has completed" />
			<Input bind:value={general.command_wait_attempts} type="number" label="Command Wait Attempts" hint="Max retries when waiting for an arr command to finish" />
			<Input bind:value={general.min_download_queue_size} type="number" label="Min Download Queue Size (-1 = disabled)" hint="Pause lurking if the download queue has fewer items. -1 disables" />
			<Toggle bind:checked={general.ssl_verify} label="SSL Verification" hint="Verify TLS certificates when connecting to arr apps" />
			<Toggle bind:checked={general.proxy_auth_bypass} label="Proxy Auth Bypass" hint="Trust X-Forwarded headers from a reverse proxy for authentication" />
			<Button onclick={saveGeneral} loading={saving}>Save General</Button>
		</div>
	</Card>
	{:else}
	<Card>
		<p class="text-sm text-surface-500 text-center py-4">Loading settings...</p>
	</Card>
	{/if}
</div>
