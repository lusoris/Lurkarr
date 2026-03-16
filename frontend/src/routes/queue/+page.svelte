<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import InstanceSwitcher from '$lib/components/InstanceSwitcher.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import QueueCleanerTab from '$lib/components/queue/QueueCleanerTab.svelte';
	import QueueScoringTab from '$lib/components/queue/QueueScoringTab.svelte';
	import QueueBlocklistTab from '$lib/components/queue/QueueBlocklistTab.svelte';
	import QueueImportsTab from '$lib/components/queue/QueueImportsTab.svelte';
	import GlobalBlocklistManager from '$lib/components/queue/GlobalBlocklistManager.svelte';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import type { QueueCleanerSettings, ScoringProfile, BlocklistEntry, AutoImportLog, SeedingRuleGroup, BlocklistSource, BlocklistRule } from '$lib/types';

	const toasts = getToasts();
	const store = getInstances();

	type Tab = 'cleaner' | 'scoring' | 'blocklist' | 'imports';

	let activeTab = $state<Tab>('cleaner');
	let selectedApp = $derived(store.selectedApp);

	// --- Cleaner Tab ---
	let cleanerSettings = $state<Record<string, QueueCleanerSettings>>({});
	let loadedCleaners = $state<Set<string>>(new Set());
	let seedingGroups = $state<SeedingRuleGroup[]>([]);
	let seedingGroupsLoaded = $state(false);
	let showAddGroup = $state(false);
	let editingGroup = $state<SeedingRuleGroup | null>(null);
	let savingGroup = $state(false);
	let confirmDeleteGroup = $state<number | null>(null);
	let newGroup = $state<Omit<SeedingRuleGroup, 'id'>>({
		name: '', priority: 0, match_type: 'tracker', match_pattern: '',
		max_ratio: 0, max_hours: 0, seeding_mode: 'or', skip_removal: false, delete_files: false
	});

	// --- Scoring Tab ---
	let scoringProfiles = $state<Record<string, ScoringProfile>>({});
	let loadedScoring = $state<Set<string>>(new Set());

	// --- Blocklist Tab ---
	let blocklist = $state<BlocklistEntry[]>([]);
	let loadingBlocklist = $state(false);

	// --- Imports Tab ---
	let imports = $state<AutoImportLog[]>([]);
	let loadingImports = $state(false);

	// --- Global Blocklist Management ---
	let sources = $state<BlocklistSource[]>([]);
	let rules = $state<BlocklistRule[]>([]);
	let sourcesLoaded = $state(false);
	let rulesLoaded = $state(false);
	let showAddSource = $state(false);
	let showAddRule = $state(false);
	let editingSource = $state<BlocklistSource | null>(null);
	let newSource = $state({ name: '', url: '', enabled: true, sync_interval_hours: 24 });
	let newRule = $state({ pattern: '', pattern_type: 'title_contains' as const, reason: '' });
	let regexTestInput = $state('');
	let regexTestResult = $derived.by(() => {
		if (newRule.pattern_type !== 'title_regex' || !newRule.pattern || !regexTestInput) return null;
		try {
			const re = new RegExp(newRule.pattern, 'i');
			return re.test(regexTestInput);
		} catch {
			return 'invalid';
		}
	});
	let savingSource = $state(false);
	let savingRule = $state(false);
	let confirmDeleteSource = $state<string | null>(null);
	let confirmDeleteRule = $state<string | null>(null);

	// --- Seeding Groups ---
	async function loadSeedingGroups() {
		try {
			seedingGroups = await api.get<SeedingRuleGroup[]>('/queue/seeding-groups');
		} catch { seedingGroups = []; }
		seedingGroupsLoaded = true;
	}

	async function createSeedingGroup() {
		savingGroup = true;
		try {
			await api.post('/queue/seeding-groups', newGroup);
			toasts.success('Seeding group created');
			newGroup = { name: '', priority: 0, match_type: 'tracker', match_pattern: '', max_ratio: 0, max_hours: 0, seeding_mode: 'or', skip_removal: false, delete_files: false };
			showAddGroup = false;
			await loadSeedingGroups();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to create group');
		}
		savingGroup = false;
	}

	async function updateSeedingGroup(g: SeedingRuleGroup) {
		try {
			await api.put(`/queue/seeding-groups/${g.id}`, g);
			toasts.success('Seeding group updated');
			editingGroup = null;
			await loadSeedingGroups();
		} catch {
			toasts.error('Failed to update group');
		}
	}

	async function deleteSeedingGroup(id: number) {
		try {
			await api.del(`/queue/seeding-groups/${id}`);
			toasts.success('Seeding group deleted');
			await loadSeedingGroups();
		} catch {
			toasts.error('Failed to delete group');
		}
	}

	// --- Cleaner Settings ---
	async function loadCleaner(app: string) {
		try {
			cleanerSettings[app] = await api.get<QueueCleanerSettings>(`/queue/settings/${app}`);
		} catch { /* 404 expected on first load */ }
		loadedCleaners = new Set([...loadedCleaners, app]);
	}

	async function saveCleaner() {
		const settings = cleanerSettings[selectedApp];
		if (!settings) return;
		try {
			await api.put(`/queue/settings/${selectedApp}`, settings);
			toasts.success('Queue cleaner settings saved');
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
	}

	// --- Scoring Settings ---
	async function loadScoring(app: string) {
		try {
			scoringProfiles[app] = await api.get<ScoringProfile>(`/queue/scoring/${app}`);
		} catch { /* not yet configured */ }
		loadedScoring = new Set([...loadedScoring, app]);
	}

	async function saveScoring() {
		const profile = scoringProfiles[selectedApp];
		if (!profile) return;
		try {
			await api.put(`/queue/scoring/${selectedApp}`, profile);
			toasts.success('Scoring profile saved');
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
	}

	// --- Blocklist Logs ---
	async function loadBlocklist(app: string) {
		loadingBlocklist = true;
		try {
			blocklist = await api.get<BlocklistEntry[]>(`/queue/blocklist/${app}`);
		} catch {
			blocklist = [];
		}
		loadingBlocklist = false;
	}

	// --- Import Logs ---
	async function loadImports(app: string) {
		loadingImports = true;
		try {
			imports = await api.get<AutoImportLog[]>(`/queue/imports/${app}`);
		} catch {
			imports = [];
		}
		loadingImports = false;
	}

	// --- Global Blocklist Sources ---
	async function loadSources() {
		try {
			sources = await api.get<BlocklistSource[]>('/blocklist/sources');
		} catch { sources = []; }
		sourcesLoaded = true;
	}

	async function createSource() {
		savingSource = true;
		try {
			await api.post('/blocklist/sources', newSource);
			toasts.success('Source added');
			newSource = { name: '', url: '', enabled: true, sync_interval_hours: 24 };
			showAddSource = false;
			await loadSources();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to create source');
		}
		savingSource = false;
	}

	async function updateSource(src: BlocklistSource) {
		try {
			await api.put(`/blocklist/sources/${src.id}`, src);
			toasts.success('Source updated');
			editingSource = null;
			await loadSources();
		} catch {
			toasts.error('Failed to update source');
		}
	}

	async function deleteSource(id: string) {
		try {
			await api.del(`/blocklist/sources/${id}`);
			toasts.success('Source deleted');
			await Promise.all([loadSources(), loadRules()]);
		} catch {
			toasts.error('Failed to delete source');
		}
	}

	// --- Global Blocklist Rules ---
	async function loadRules() {
		try {
			rules = await api.get<BlocklistRule[]>('/blocklist/rules');
		} catch { rules = []; }
		rulesLoaded = true;
	}

	async function createRule() {
		savingRule = true;
		try {
			await api.post('/blocklist/rules', { ...newRule, enabled: true });
			toasts.success('Rule added');
			newRule = { pattern: '', pattern_type: 'title_contains', reason: '' };
			showAddRule = false;
			await loadRules();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to create rule');
		}
		savingRule = false;
	}

	async function deleteRule(id: string) {
		try {
			await api.del(`/blocklist/rules/${id}`);
			toasts.success('Rule deleted');
			await loadRules();
		} catch {
			toasts.error('Failed to delete rule');
		}
	}

	// --- Tab Loading ---
	function loadTabData() {
		if (activeTab === 'cleaner') loadCleaner(selectedApp);
		else if (activeTab === 'scoring') loadScoring(selectedApp);
		else if (activeTab === 'blocklist') loadBlocklist(selectedApp);
		else if (activeTab === 'imports') loadImports(selectedApp);
	}

	$effect(() => { loadTabData(); });

	$effect(() => {
		if (!sourcesLoaded) loadSources();
		if (!rulesLoaded) loadRules();
		if (!seedingGroupsLoaded) loadSeedingGroups();
	});
</script>

<svelte:head><title>Queue Management - Lurkarr</title></svelte:head>

<div class="space-y-4">
	<PageHeader title="Queue Management" description="Queue cleaner, scoring profiles, blocklist, and import history.">
		{#snippet actions()}
			<HelpDrawer page="queue" />
		{/snippet}
	</PageHeader>

	<!-- App selector -->
	<InstanceSwitcher showInstances={false} onchange={loadTabData} />

	<!-- Tab navigation -->
	<Tabs
		tabs={[
			{ value: 'cleaner', label: 'Queue Cleaner' },
			{ value: 'scoring', label: 'Scoring' },
			{ value: 'blocklist', label: 'Blocklist' },
			{ value: 'imports', label: 'Import Log' }
		]}
		bind:value={activeTab}
	/>

	<!-- Queue Cleaner Settings -->
	{#if activeTab === 'cleaner'}
		<QueueCleanerTab
			app={selectedApp}
			settings={cleanerSettings[selectedApp]}
			loaded={loadedCleaners.has(selectedApp)}
			seedingGroups={seedingGroups}
			bind:showAddGroup={showAddGroup}
			bind:editingGroup={editingGroup}
			bind:newGroup={newGroup}
			savingGroup={savingGroup}
			bind:confirmDeleteGroup={confirmDeleteGroup}
			onSave={saveCleaner}
			onCreateGroup={createSeedingGroup}
			onUpdateGroup={updateSeedingGroup}
			onDeleteGroup={deleteSeedingGroup}
			onShowAddGroupChange={(show) => (showAddGroup = show)}
			onEditingGroupChange={(group) => (editingGroup = group)}
			onNewGroupChange={(partial) => (newGroup = { ...newGroup, ...partial })}
			onConfirmDeleteGroupChange={(id) => (confirmDeleteGroup = id)}
		/>
	{:else if activeTab === 'scoring'}
		<QueueScoringTab
			app={selectedApp}
			profile={scoringProfiles[selectedApp]}
			loaded={loadedScoring.has(selectedApp)}
			saving={false}
			onSave={saveScoring}
		/>
	{:else if activeTab === 'blocklist'}
		<QueueBlocklistTab app={selectedApp} blocklist={blocklist} loading={loadingBlocklist} />
	{:else if activeTab === 'imports'}
		<QueueImportsTab app={selectedApp} imports={imports} loading={loadingImports} />
	{/if}

	<!-- Global Blocklist Management -->
	{#if activeTab === 'blocklist'}
		<div class="border-t border-border pt-6 mt-6">
			<GlobalBlocklistManager
				bind:sources={sources}
				bind:rules={rules}
				sourcesLoaded={sourcesLoaded}
				rulesLoaded={rulesLoaded}
				bind:showAddSource={showAddSource}
				bind:showAddRule={showAddRule}
				bind:editingSource={editingSource}
				bind:newSource={newSource}
				bind:newRule={newRule}
				bind:regexTestInput={regexTestInput}
				regexTestResult={regexTestResult}
				savingSource={savingSource}
				savingRule={savingRule}
				bind:confirmDeleteSource={confirmDeleteSource}
				bind:confirmDeleteRule={confirmDeleteRule}
				onCreateSource={createSource}
				onUpdateSource={updateSource}
				onDeleteSource={deleteSource}
				onCreateRule={createRule}
				onDeleteRule={deleteRule}
				onShowAddSourceChange={(show) => (showAddSource = show)}
				onShowAddRuleChange={(show) => (showAddRule = show)}
				onEditingSourceChange={(src) => (editingSource = src)}
				onNewSourceChange={(partial) => (newSource = { ...newSource, ...partial })}
				onNewRuleChange={(partial) => (newRule = { ...newRule, ...partial })}
				onRegexTestInputChange={(input) => (regexTestInput = input)}
				onConfirmDeleteSourceChange={(id) => (confirmDeleteSource = id)}
				onConfirmDeleteRuleChange={(id) => (confirmDeleteRule = id)}
			/>
		</div>
	{/if}

	<ScrollToTop />
</div>
