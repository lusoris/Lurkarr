<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appTabLabel, appLogo, appAccentBorder, appBgColor, appButtonClass } from '$lib';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import InstanceSwitcher from '$lib/components/InstanceSwitcher.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import Separator from '$lib/components/ui/Separator.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import { RotateCcw } from 'lucide-svelte';
	import type { AppSettings, StateEntry } from '$lib/types';

	const toasts = getToasts();
	const store = getInstances();

	let appSettings = $state<Record<string, AppSettings>>({});
	let selectedApp = $derived(store.selectedApp);
	let saving = $state(false);

	// State management
	let stateEntries = $state<StateEntry[]>([]);
	let loadingState = $state(false);
	let resettingInstance = $state<string | null>(null);
	let confirmResetId = $state<string | null>(null);

	let currentAppStates = $derived(
		stateEntries.filter(s => s.app_type === selectedApp ||
			(selectedApp === 'whisparr' && (s.app_type === 'whisparr' || s.app_type === 'eros')))
	);



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
	$effect(() => { loadState(); });

	async function loadState() {
		loadingState = true;
		try {
			stateEntries = await api.get<StateEntry[]>('/state');
		} catch { stateEntries = []; }
		loadingState = false;
	}

	async function resetInstanceState(appType: string, instanceId: string, name: string) {
		resettingInstance = instanceId;
		try {
			await api.post(`/state/reset?app=${appType}&instance_id=${instanceId}`, {});
			toasts.success(`State reset for ${name}`);
			await loadState();
		} catch {
			toasts.error(`Failed to reset state for ${name}`);
		}
		resettingInstance = null;
		confirmResetId = null;
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString();
	}
</script>

<svelte:head><title>Lurk Settings - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Lurk Settings" description="Configure lurking behavior per app — how many items to search, modes, rate limits, and more.">
		{#snippet actions()}
			<HelpDrawer page="lurk" />
		{/snippet}
	</PageHeader>

	<InstanceSwitcher showInstances={false} />

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
						<Input type="number" bind:value={settings.max_search_failures} label="Max Search Failures" hint="Stop retrying items after this many consecutive search failures (0 = no limit)" min={0} />
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

	<!-- Instance State -->
	<Card>
		<h3 class="text-sm font-semibold text-foreground mb-3">Lurk State</h3>
		<p class="text-xs text-muted-foreground mb-3">Per-instance lurk progress tracking. Resetting clears cached state so the next lurk cycle starts fresh.</p>
		{#if loadingState}
			<Skeleton rows={2} height="h-8" />
		{:else if currentAppStates.length === 0}
			<p class="text-sm text-muted-foreground">No state tracked for this app yet.</p>
		{:else}
			<div class="space-y-2">
				{#each currentAppStates as entry}
					<div class="flex items-center justify-between p-3 rounded-lg bg-muted/30">
						<div class="flex-1 min-w-0">
							<p class="text-sm font-medium text-foreground truncate">{entry.name}</p>
							<p class="text-xs text-muted-foreground">
								{#if entry.last_reset}
									Last reset: {formatDate(entry.last_reset)}
								{:else}
									Never reset
								{/if}
							</p>
						</div>
						<ConfirmAction
							active={confirmResetId === entry.instance_id}
							message="Reset state?"
							onconfirm={() => resetInstanceState(entry.app_type, entry.instance_id, entry.name)}
							oncancel={() => confirmResetId = null}
						>
							<Button
								size="sm"
								variant="ghost"
								onclick={() => confirmResetId = entry.instance_id}
								loading={resettingInstance === entry.instance_id}
							>
								<RotateCcw class="h-3.5 w-3.5" />
								Reset
							</Button>
						</ConfirmAction>
					</div>
				{/each}
			</div>
		{/if}
	</Card>
</div>

<ScrollToTop />
