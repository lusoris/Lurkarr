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

	interface SeerrSettings {
		id: string;
		url: string;
		api_key: string;
		enabled: boolean;
		sync_interval_minutes: number;
		auto_approve: boolean;
	}

	interface DownloadClientSettings {
		app_type: string;
		client_type: string;
		url: string;
		username: string;
		password: string;
		enabled: boolean;
		timeout: number;
	}

	interface AppSettings {
		app_type: string;
		lurk_missing_count: number;
		lurk_upgrade_count: number;
		lurk_missing_mode: string;
		upgrade_mode: string;
		sleep_duration: number;
		monitored_only: boolean;
		skip_future: boolean;
		hourly_cap: number;
		random_selection: boolean;
		debug_mode: boolean;
	}

	const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;
	const clientTypes = ['qbittorrent', 'transmission', 'deluge', 'sabnzbd', 'nzbget'] as const;
	type Tab = 'general' | 'apps' | 'prowlarr' | 'sabnzbd' | 'seerr' | 'download-clients';

	let activeTab = $state<Tab>('general');
	let general = $state<GeneralSettings | null>(null);
	let prowlarr = $state<ProwlarrSettings | null>(null);
	let sabnzbd = $state<SABnzbdSettings | null>(null);
	let seerr = $state<SeerrSettings | null>(null);
	let dlClient = $state<DownloadClientSettings | null>(null);
	let appSettings = $state<Record<string, AppSettings>>({});
	let selectedApp = $state<string>('sonarr');
	let selectedDlApp = $state<string>('sonarr');
	let saving = $state(false);

	async function load() {
		try {
			[general, prowlarr, sabnzbd, seerr] = await Promise.all([
				api.get<GeneralSettings>('/settings/general'),
				api.get<ProwlarrSettings>('/prowlarr/settings'),
				api.get<SABnzbdSettings>('/sabnzbd/settings'),
				api.get<SeerrSettings>('/seerr/settings')
			]);
		} catch { /* handled */ }
	}

	async function loadAppSettings(app: string) {
		try {
			appSettings[app] = await api.get<AppSettings>(`/settings/${app}`);
		} catch { /* handled */ }
	}

	async function loadDlClient(app: string) {
		try {
			dlClient = await api.get<DownloadClientSettings>(`/queue/download-client/${app}`);
		} catch {
			dlClient = { app_type: app, client_type: 'qbittorrent', url: '', username: '', password: '', enabled: false, timeout: 30 };
		}
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

	async function saveAppSettings() {
		const settings = appSettings[selectedApp];
		if (!settings) return;
		saving = true;
		try {
			await api.put(`/settings/${selectedApp}`, settings);
			toasts.success(`${selectedApp} settings saved`);
		} catch {
			toasts.error(`Failed to save ${selectedApp} settings`);
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

	async function saveSeerr() {
		if (!seerr) return;
		saving = true;
		try {
			await api.put('/seerr/settings', seerr);
			toasts.success('Seerr settings saved');
		} catch {
			toasts.error('Failed to save Seerr settings');
		}
		saving = false;
	}

	async function testSeerr() {
		try {
			const res = await api.post<{ status: string; version: string }>('/seerr/test');
			toasts.success(`Seerr connected — v${res.version}`);
		} catch {
			toasts.error('Seerr connection failed');
		}
	}

	async function saveDlClient() {
		if (!dlClient) return;
		saving = true;
		try {
			await api.put(`/queue/download-client/${selectedDlApp}`, dlClient);
			toasts.success('Download client settings saved');
		} catch {
			toasts.error('Failed to save download client settings');
		}
		saving = false;
	}

	$effect(() => { load(); });
	$effect(() => { if (activeTab === 'apps') loadAppSettings(selectedApp); });
	$effect(() => { if (activeTab === 'download-clients') loadDlClient(selectedDlApp); });

	const tabs: { id: Tab; label: string }[] = [
		{ id: 'general', label: 'General' },
		{ id: 'apps', label: 'Lurk Settings' },
		{ id: 'prowlarr', label: 'Prowlarr' },
		{ id: 'sabnzbd', label: 'SABnzbd' },
		{ id: 'seerr', label: 'Seerr' },
		{ id: 'download-clients', label: 'Download Clients' }
	];
</script>

<svelte:head><title>Settings - Lurkarr</title></svelte:head>

<div class="space-y-6 max-w-2xl">
	<h1 class="text-2xl font-bold text-surface-50">Settings</h1>

	<!-- Tab navigation -->
	<div class="flex gap-1 rounded-lg bg-surface-900 border border-surface-800 p-1 overflow-x-auto">
		{#each tabs as tab}
			<button
				onclick={() => activeTab = tab.id}
				class="shrink-0 rounded-md px-3 py-2 text-sm font-medium transition-colors
					{activeTab === tab.id ? 'bg-lurk-600 text-white' : 'text-surface-400 hover:text-surface-200 hover:bg-surface-800'}"
			>{tab.label}</button>
		{/each}
	</div>

	<!-- General Settings -->
	{#if activeTab === 'general' && general}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">General</h2>
			<div class="space-y-4">
				<Input bind:value={general.api_timeout} type="number" label="API Timeout (seconds)" />
				<Input bind:value={general.stateful_reset_hours} type="number" label="State Reset (hours)" />
				<Input bind:value={general.command_wait_delay} type="number" label="Command Wait Delay (ms)" />
				<Input bind:value={general.command_wait_attempts} type="number" label="Command Wait Attempts" />
				<Input bind:value={general.min_download_queue_size} type="number" label="Min Download Queue Size (-1 = disabled)" />
				<Toggle bind:checked={general.ssl_verify} label="SSL Verification" />
				<Toggle bind:checked={general.proxy_auth_bypass} label="Proxy Auth Bypass" />
				<Button onclick={saveGeneral} loading={saving}>Save General</Button>
			</div>
		</Card>
	{/if}

	<!-- Per-App Lurk Settings -->
	{#if activeTab === 'apps'}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Lurk Settings</h2>
			<div class="flex gap-1 mb-4 rounded-lg bg-surface-800/50 p-1">
				{#each appTypes as app}
					<button
						onclick={() => { selectedApp = app; loadAppSettings(app); }}
						class="flex-1 rounded-md px-2 py-1.5 text-xs font-medium capitalize transition-colors
							{selectedApp === app ? 'bg-lurk-600 text-white' : 'text-surface-400 hover:text-surface-200 hover:bg-surface-700'}"
					>{app}</button>
				{/each}
			</div>

			{@const settings = appSettings[selectedApp]}
			{#if settings}
				<div class="space-y-4">
					<div class="grid grid-cols-2 gap-4">
						<Input bind:value={settings.lurk_missing_count} type="number" label="Lurk Missing Count" />
						<Input bind:value={settings.lurk_upgrade_count} type="number" label="Lurk Upgrade Count" />
					</div>
					<div class="grid grid-cols-2 gap-4">
						<label class="block">
							<span class="block text-sm font-medium text-surface-300 mb-1.5">Missing Mode</span>
							<select bind:value={settings.lurk_missing_mode} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
								<option value="oldest">Oldest First</option>
								<option value="newest">Newest First</option>
								<option value="random">Random</option>
							</select>
						</label>
						<label class="block">
							<span class="block text-sm font-medium text-surface-300 mb-1.5">Upgrade Mode</span>
							<select bind:value={settings.upgrade_mode} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
								<option value="oldest">Oldest First</option>
								<option value="newest">Newest First</option>
								<option value="random">Random</option>
							</select>
						</label>
					</div>
					<Input bind:value={settings.sleep_duration} type="number" label="Sleep Duration (ms)" />
					<Input bind:value={settings.hourly_cap} type="number" label="Hourly API Cap (0 = unlimited)" />
					<Toggle bind:checked={settings.monitored_only} label="Monitored Only" />
					<Toggle bind:checked={settings.skip_future} label="Skip Future Releases" />
					<Toggle bind:checked={settings.random_selection} label="Random Selection" />
					<Toggle bind:checked={settings.debug_mode} label="Debug Mode" />
					<Button onclick={saveAppSettings} loading={saving}>Save {selectedApp} Settings</Button>
				</div>
			{:else}
				<p class="text-sm text-surface-500">Loading settings...</p>
			{/if}
		</Card>
	{/if}

	<!-- Prowlarr Settings -->
	{#if activeTab === 'prowlarr' && prowlarr}
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
	{#if activeTab === 'sabnzbd' && sabnzbd}
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

	<!-- Seerr Settings -->
	{#if activeTab === 'seerr' && seerr}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Seerr / Overseerr / Jellyseerr</h2>
			<div class="space-y-4">
				<Toggle bind:checked={seerr.enabled} label="Enabled" />
				<Input bind:value={seerr.url} label="URL" placeholder="http://overseerr:5055" />
				<Input bind:value={seerr.api_key} label="API Key" type="password" />
				<Input bind:value={seerr.sync_interval_minutes} type="number" label="Sync Interval (minutes)" />
				<Toggle bind:checked={seerr.auto_approve} label="Auto-Approve Requests" />
				<div class="flex gap-2">
					<Button onclick={saveSeerr} loading={saving}>Save</Button>
					<Button variant="secondary" onclick={testSeerr}>Test Connection</Button>
				</div>
			</div>
		</Card>
	{/if}

	<!-- Download Client Settings -->
	{#if activeTab === 'download-clients'}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Download Clients</h2>
			<div class="flex gap-1 mb-4 rounded-lg bg-surface-800/50 p-1 overflow-x-auto">
				{#each appTypes as app}
					<button
						onclick={() => { selectedDlApp = app; loadDlClient(app); }}
						class="shrink-0 rounded-md px-2 py-1.5 text-xs font-medium capitalize transition-colors
							{selectedDlApp === app ? 'bg-lurk-600 text-white' : 'text-surface-400 hover:text-surface-200 hover:bg-surface-700'}"
					>{app}</button>
				{/each}
			</div>

			{#if dlClient}
				<div class="space-y-4">
					<Toggle bind:checked={dlClient.enabled} label="Enabled" />
					<label class="block">
						<span class="block text-sm font-medium text-surface-300 mb-1.5">Client Type</span>
						<select bind:value={dlClient.client_type} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
							{#each clientTypes as ct}
								<option value={ct}>{ct}</option>
							{/each}
						</select>
					</label>
					<Input bind:value={dlClient.url} label="URL" placeholder="http://qbittorrent:8080" />
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input bind:value={dlClient.username} label="Username" />
						<Input bind:value={dlClient.password} label="Password" type="password" />
					</div>
					<Input bind:value={dlClient.timeout} type="number" label="Timeout (seconds)" />
					<Button onclick={saveDlClient} loading={saving}>Save</Button>
				</div>
			{:else}
				<p class="text-sm text-surface-500">Loading settings...</p>
			{/if}
		</Card>
	{/if}
</div>
