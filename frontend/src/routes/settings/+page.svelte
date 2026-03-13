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
	<!-- ── Lurking Behaviour ─────────────────────────────── -->
	<Card>
		<h2 class="text-lg font-semibold text-surface-200 mb-1">Lurking Behaviour</h2>
		<p class="text-xs text-surface-500 mb-4">Controls how Lurkarr searches and manages your media libraries.</p>
		<div class="space-y-4">
			<Input bind:value={general.stateful_reset_hours} type="number" label="State Reset (hours)" hint="Hours after which lurk progress resets and starts fresh" />
			<Input bind:value={general.min_download_queue_size} type="number" label="Min Download Queue Size (-1 = disabled)" hint="Pause lurking if the download queue has fewer items. -1 disables" />
		</div>
	</Card>

	<!-- ── API & Command Execution ───────────────────────── -->
	<Card>
		<h2 class="text-lg font-semibold text-surface-200 mb-1">API &amp; Command Execution</h2>
		<p class="text-xs text-surface-500 mb-4">Tune how Lurkarr communicates with your Arr apps.</p>
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
		<h2 class="text-lg font-semibold text-surface-200 mb-1">Security</h2>
		<p class="text-xs text-surface-500 mb-4">Connection security and authentication settings.</p>
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
				<div class="h-20 rounded-xl bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	</Card>
	{/if}
</div>
