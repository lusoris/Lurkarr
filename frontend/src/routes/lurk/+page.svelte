<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appTabLabel, appLogo, appAccentBorder, appBgColor, appButtonClass } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import Separator from '$lib/components/ui/Separator.svelte';

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
		selection_mode: string;
		debug_mode: boolean;
	}

	let appSettings = $state<Record<string, AppSettings>>({});
	let selectedApp = $state<string>('sonarr');
	let saving = $state(false);

	const tabs = appTypes.map(app => ({
		value: app,
		label: appTabLabel(app),
		icon: appLogo(app),
		activeClass: appBgColor(app) + ' text-white shadow-sm'
	}));

	async function loadAppSettings(app: string) {
		try {
			appSettings[app] = await api.get<AppSettings>(`/settings/${app}`);
		} catch { /* handled */ }
	}

	async function saveAppSettings() {
		const app = selectedApp;
		const settings = appSettings[app];
		if (!settings) return;
		saving = true;
		try {
			await api.put(`/settings/${app}`, settings);
			toasts.success(`${appDisplayName(app)} settings saved`);
		} catch {
			toasts.error(`Failed to save ${appDisplayName(app)} settings`);
		}
		saving = false;
	}

	$effect(() => { loadAppSettings(selectedApp); });
</script>

<svelte:head><title>Lurk Settings - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Lurk Settings" description="Configure lurking behavior per app — how many items to search, modes, rate limits, and more." />

	<Tabs {tabs} bind:value={selectedApp} />

	<Card class="border-l-2 {appAccentBorder(selectedApp)}">
		{#if appSettings[selectedApp]}
			{@const settings = appSettings[selectedApp]}
			<div class="space-y-6">
				<!-- Search Counts -->
				<div>
					<h3 class="text-sm font-semibold text-foreground mb-3">Search Counts</h3>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input bind:value={settings.lurk_missing_count} type="number" label="Missing Count" hint="Number of missing items to search per lurk cycle" />
						<Input bind:value={settings.lurk_upgrade_count} type="number" label="Upgrade Count" hint="Number of cutoff-unmet items to search per cycle" />
					</div>
				</div>

				<Separator />

				<!-- Search Mode -->
				<div>
					<h3 class="text-sm font-semibold text-foreground mb-3">Search Mode</h3>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Select bind:value={settings.lurk_missing_mode} label="Missing Mode" hint="How missing items are selected for search">
							<option value="oldest">Oldest First</option>
							<option value="newest">Newest First</option>
							<option value="random">Random</option>
						</Select>
						<Select bind:value={settings.upgrade_mode} label="Upgrade Mode" hint="How upgrade candidates are selected for search">
							<option value="oldest">Oldest First</option>
							<option value="newest">Newest First</option>
							<option value="random">Random</option>
						</Select>
					</div>
				</div>

				<Separator />

				<!-- Rate Limiting -->
				<div>
					<h3 class="text-sm font-semibold text-foreground mb-3">Rate Limiting</h3>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
						<Input bind:value={settings.sleep_duration} type="number" label="Sleep Duration (ms)" hint="Delay between individual API commands" />
						<Input bind:value={settings.hourly_cap} type="number" label="Hourly API Cap" hint="Max API search commands per hour (0 = unlimited)" />
					</div>
				</div>

				<Separator />

				<!-- Behaviour -->
				<div>
					<h3 class="text-sm font-semibold text-foreground mb-3">Behaviour</h3>
					<div class="space-y-4">
						<Toggle bind:checked={settings.monitored_only} label="Monitored Only" hint="Only search for items marked as monitored in the arr app" />
						<Toggle bind:checked={settings.skip_future} label="Skip Future Releases" hint="Skip items with a release date in the future" />
						<Select bind:value={settings.selection_mode} label="Selection Mode" hint="How items are chosen from the candidate pool">
							<option value="random">Random</option>
							<option value="newest">Newest First</option>
							<option value="oldest">Oldest First</option>
							<option value="least_recent">Least Recently Searched</option>
						</Select>
						<Toggle bind:checked={settings.debug_mode} label="Debug Mode" hint="Log detailed information about each lurk cycle for troubleshooting" />
					</div>
				</div>

				<Separator />

				<div class="flex justify-end">
					<Button onclick={saveAppSettings} loading={saving} class={appButtonClass(selectedApp)}>Save {appDisplayName(selectedApp)} Settings</Button>
				</div>
			</div>
		{:else}
			<Skeleton rows={6} height="h-10" />
		{/if}
	</Card>
</div>
