<script lang="ts">
	import { appTypes, visibleAppTypes, appTabLabel, appLogo, appBgColor, appDisplayName } from '$lib';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Button from '$lib/components/ui/Button.svelte';

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

	function onAppChange(app: string) {
		store.selectedApp = app;
		onchange?.();
	}
</script>

<div class="space-y-2 {className}">
	<div class="inline-flex items-center gap-1 rounded-lg bg-muted p-1 overflow-x-auto">
		{#each types as app}
			{@const active = store.selectedApp === app}
			{@const logo = appLogo(app)}
			<Button
				variant={active ? 'primary' : 'ghost'}
				size="sm"
				class="inline-flex items-center gap-1.5 whitespace-nowrap {active ? appBgColor(app) + ' text-white shadow-sm hover:opacity-90' : ''}"
				onclick={() => onAppChange(app)}
			>
				{#if logo}
					<img src={logo} alt="{appTabLabel(app)} logo" class="w-4 h-4 rounded-sm" />
				{/if}
				{appTabLabel(app)}
			</Button>
		{/each}
	</div>

	{#if showInstances && store.currentInstances.length > 1}
		<Select
			value={store.selectedInstance}
			onchange={(e) => {
				store.selectedInstance = (e.target as HTMLSelectElement).value;
				onchange?.();
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
