<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appTabLabel, appLogo } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const toasts = getToasts();

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

	let appSettings = $state<Record<string, AppSettings>>({});
	let selectedApp = $state<string>('sonarr');
	let saving = $state(false);

	async function loadAppSettings(app: string) {
		try {
			appSettings[app] = await api.get<AppSettings>(`/settings/${app}`);
		} catch { /* handled */ }
	}

	async function saveAppSettings() {
		const settings = appSettings[selectedApp];
		if (!settings) return;
		saving = true;
		try {
			await api.put(`/settings/${selectedApp}`, settings);
			toasts.success(`${appDisplayName(selectedApp)} settings saved`);
		} catch {
			toasts.error(`Failed to save ${appDisplayName(selectedApp)} settings`);
		}
		saving = false;
	}

	$effect(() => { loadAppSettings(selectedApp); });
</script>

<svelte:head><title>Lurk Settings - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Lurk Settings</h1>
	<p class="text-sm text-surface-400">Configure lurking behavior per app — how many items to search, modes, rate limits, and more.</p>

	<Card>
		<!-- App selector -->
		<div class="flex gap-1 mb-5 rounded-lg bg-surface-800/50 p-1 overflow-x-auto">
			{#each appTypes as app}
				{@const logo = appLogo(app)}
				<button
					onclick={() => { selectedApp = app; loadAppSettings(app); }}
					class="shrink-0 flex items-center gap-1.5 rounded-md px-2 py-1.5 text-xs font-medium transition-colors
						{selectedApp === app ? 'bg-lurk-600 text-white' : 'text-surface-400 hover:text-surface-200 hover:bg-surface-700'}"
				>
					{#if logo}<img src={logo} alt="" class="w-4 h-4 rounded-sm" />{/if}
					{appTabLabel(app)}
				</button>
			{/each}
		</div>

		{#if appSettings[selectedApp]}
			{@const settings = appSettings[selectedApp]}
			<div class="space-y-4">
				<div class="grid grid-cols-2 gap-4">
					<Input bind:value={settings.lurk_missing_count} type="number" label="Lurk Missing Count" hint="Number of missing items to search per lurk cycle" />
					<Input bind:value={settings.lurk_upgrade_count} type="number" label="Lurk Upgrade Count" hint="Number of cutoff-unmet items to search per cycle" />
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
				<Input bind:value={settings.sleep_duration} type="number" label="Sleep Duration (ms)" hint="Delay between individual API commands to avoid rate-limiting" />
				<Input bind:value={settings.hourly_cap} type="number" label="Hourly API Cap (0 = unlimited)" hint="Max API search commands per hour. 0 disables the limit" />
				<Toggle bind:checked={settings.monitored_only} label="Monitored Only" hint="Only search for items marked as monitored in the arr app" />
				<Toggle bind:checked={settings.skip_future} label="Skip Future Releases" hint="Skip items with a release date in the future" />
				<Toggle bind:checked={settings.random_selection} label="Random Selection" hint="Randomize which items are picked within the selected mode" />
				<Toggle bind:checked={settings.debug_mode} label="Debug Mode" hint="Log detailed information about each lurk cycle for troubleshooting" />
				<Button onclick={saveAppSettings} loading={saving}>Save {appDisplayName(selectedApp)} Settings</Button>
			</div>
		{:else}
			<p class="text-sm text-surface-500">Loading settings...</p>
		{/if}
	</Card>
</div>
