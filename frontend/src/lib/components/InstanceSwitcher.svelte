<script lang="ts">
	import { appTypes, visibleAppTypes, appTabLabel, appLogo, appBgColor, appDisplayName } from '$lib';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import Select from '$lib/components/ui/Select.svelte';

	interface Props {
		/** Use all 6 backend app types (default) or only the 5 visible UI types. */
		allApps?: boolean;
		/** Show instance dropdown (default: true). Set false for app-only filtering. */
		showInstances?: boolean;
		/** Called when app or instance changes. */
		onchange?: () => void;
		class?: string;
	}

	let {
		allApps = true,
		showInstances = true,
		onchange,
		class: className = ''
	}: Props = $props();

	const store = getInstances();

	const types = $derived(allApps ? appTypes : visibleAppTypes);
	const appTabs = $derived(
		types.map((app) => ({
			value: app,
			label: appTabLabel(app),
			icon: appLogo(app),
			activeClass: appBgColor(app) + ' text-white shadow-sm'
		}))
	);

	function onAppChange(v: string) {
		store.selectedApp = v;
		onchange?.();
	}

	function onInstanceChange() {
		onchange?.();
	}
</script>

<div class="space-y-2 {className}">
	<Tabs tabs={appTabs} value={store.selectedApp} onchange={onAppChange} />

	{#if showInstances && store.currentInstances.length > 1}
		<Select
			value={store.selectedInstance}
			onchange={(e) => {
				store.selectedInstance = (e.target as HTMLSelectElement).value;
				onInstanceChange();
			}}
			class="max-w-xs"
		>
			<option value="">All {appDisplayName(store.selectedApp)} instances</option>
			{#each store.currentInstances as inst}
				<option value={inst.id}>{inst.name}</option>
			{/each}
		</Select>
	{/if}
</div>
