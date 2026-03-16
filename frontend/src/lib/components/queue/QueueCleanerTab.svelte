<script lang="ts">
	import { api } from '$lib/api';
	import { appDisplayName, appButtonClass } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import CollapsibleCard from '$lib/components/ui/CollapsibleCard.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import * as Alert from '$lib/components/ui/alert';
	import * as Collapsible from '$lib/components/ui/collapsible';
	import { SquarePen, Trash2 } from '@lucide/svelte';
	import type { QueueCleanerSettings, SeedingRuleGroup } from '$lib/types';

	interface Props {
		app: string;
		settings: QueueCleanerSettings | undefined;
		loaded: boolean;
		seedingGroups: SeedingRuleGroup[];
		showAddGroup: boolean;
		editingGroup: SeedingRuleGroup | null;
		newGroup: Omit<SeedingRuleGroup, 'id'>;
		savingGroup: boolean;
		confirmDeleteGroup: number | null;
		onSave: () => Promise<void>;
		onCreateGroup: () => Promise<void>;
		onUpdateGroup: (g: SeedingRuleGroup) => Promise<void>;
		onDeleteGroup: (id: number) => Promise<void>;
		onShowAddGroupChange: (show: boolean) => void;
		onEditingGroupChange: (group: SeedingRuleGroup | null) => void;
		onNewGroupChange: (partial: Partial<Omit<SeedingRuleGroup, 'id'>>) => void;
		onConfirmDeleteGroupChange: (id: number | null) => void;
		onSettingsChange?: (partial: Partial<QueueCleanerSettings>) => void;
	}

	let {
		app,
		settings,
		loaded,
		seedingGroups,
		showAddGroup = $bindable(),
		editingGroup = $bindable(),
		newGroup = $bindable(),
		savingGroup,
		confirmDeleteGroup = $bindable(),
		onSave,
		onCreateGroup,
		onUpdateGroup,
		onDeleteGroup,
		onShowAddGroupChange,
		onEditingGroupChange,
		onNewGroupChange,
		onConfirmDeleteGroupChange,
		onSettingsChange
	}: Props = $props();

	const toasts = getToasts();
	let saving = $state(false);

	async function save() {
		saving = true;
		try {
			await onSave();
		} finally {
			saving = false;
		}
	}

	function updateSetting<K extends keyof QueueCleanerSettings>(key: K, value: QueueCleanerSettings[K]) {
		if (settings) {
			settings[key] = value;
			onSettingsChange?.({ [key]: value } as Partial<QueueCleanerSettings>);
		}
	}
</script>

{#if settings}
	<div class="space-y-3">
		<Card>
			<Toggle bind:checked={settings.enabled} label="Enable Queue Cleaner" hint="Automatically manage stalled, slow, and failed downloads" />
			{#if settings.enabled}
				<div class="mt-2">
					<Toggle bind:checked={settings.dry_run} label="Dry-Run Mode" hint="Preview what would be removed without actually deleting anything — check your logs" />
				</div>
				{#if settings.dry_run}
					<div class="mt-2">
						<Alert.Root variant="warning">
							<Alert.Description>Dry-run mode is active — no items will actually be removed. Check your logs to see what would happen.</Alert.Description>
						</Alert.Root>
					</div>
				{/if}
				<div class="mt-2">
					<Input bind:value={settings.protected_tags} type="text" label="Protected Tags" hint="Comma-separated tag names — items with these tags are never removed" />
				</div>
				<div class="mt-2">
					<Input bind:value={settings.ignored_indexers} type="text" label="Ignored Indexers" hint="Comma-separated indexer names — items from these indexers skip all cleanup" />
				</div>
				<div class="mt-2">
					<Input bind:value={settings.ignored_download_clients} type="text" label="Ignored Download Clients" hint="Comma-separated download client names — items from these clients skip all cleanup" />
				</div>
			{/if}
		</Card>

		<CollapsibleCard title="Stall Detection">
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
				<Input bind:value={settings.stalled_threshold_minutes} type="number" label="Stalled (min)" hint="No progress threshold" />
				<Input bind:value={settings.slow_threshold_bytes_per_sec} type="number" label="Slow (bytes/s)" hint="Below this = slow" />
				<Input bind:value={settings.slow_ignore_above_bytes} type="number" label="Ignore Slow Above" hint="0 = disabled" />
				<Input bind:value={settings.metadata_stuck_minutes} type="number" label="Metadata Stuck (min)" hint="0 = disabled" />
				<Input bind:value={settings.bandwidth_limit_bytes_per_sec} type="number" label="Bandwidth Limit (bytes/s)" hint="Skip slow detection when >80% saturated (0 = disabled)" />
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Strike System">
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-3">
				<Input bind:value={settings.max_strikes} type="number" label="Max Strikes (global)" hint="Default for all reasons" />
				<Input bind:value={settings.strike_window_hours} type="number" label="Window (hours)" hint="Expiry time" />
			</div>
			<p class="text-xs text-muted-foreground mb-2">Per-reason overrides (0 = use global)</p>
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mb-3">
				<Input bind:value={settings.max_strikes_stalled} type="number" label="Stalled" hint="Stalled torrents" />
				<Input bind:value={settings.max_strikes_slow} type="number" label="Slow" hint="Below speed threshold" />
				<Input bind:value={settings.max_strikes_metadata} type="number" label="Metadata Stuck" hint="No size info" />
				<Input bind:value={settings.max_strikes_paused} type="number" label="Paused" hint="Paused in SABnzbd" />
				<Input bind:value={settings.max_strikes_queued} type="number" label="Queued" hint="Stuck in queue" />
			</div>
			<Input bind:value={settings.ignore_above_bytes} type="number" label="Ignore Above (bytes)" hint="Skip stalled/slow/metadata for items above this size (0 = disabled)" class="mb-3" />
			<div class="space-y-2">
				<Toggle bind:checked={settings.strike_public} label="Strike Public Trackers" />
				<Toggle bind:checked={settings.strike_private} label="Strike Private Trackers" />
				<Toggle bind:checked={settings.strike_queued} label="Strike Queued Items" hint="Flag items stuck in queued state" />
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Actions">
			<div class="space-y-2">
				<Input bind:value={settings.check_interval_seconds} type="number" label="Check Interval (seconds)" />
				<Toggle bind:checked={settings.remove_from_client} label="Remove from Download Client" />
				<Toggle bind:checked={settings.keep_archives} label="Keep Archives" hint="Preserve downloaded files for unpackerr — overrides 'Remove from Download Client' to keep files on disk" />
				<Toggle bind:checked={settings.blocklist_on_remove} label="Blocklist on Remove (global default)" hint="Fallback for reasons without a specific toggle below" />
				<Collapsible.Root class="mt-1 pl-1 border-l-2 border-border">
					<Collapsible.Trigger class="text-xs text-muted-foreground cursor-pointer select-none py-1 hover:text-foreground transition-colors flex items-center gap-1">Per-reason blocklist overrides</Collapsible.Trigger>
					<Collapsible.Content>
						<div class="space-y-2 pt-2 pl-2">
							<Toggle bind:checked={settings.blocklist_stalled} label="Blocklist Stalled" />
							<Toggle bind:checked={settings.blocklist_slow} label="Blocklist Slow" />
							<Toggle bind:checked={settings.blocklist_metadata} label="Blocklist Metadata Stuck" />
							<Toggle bind:checked={settings.blocklist_duplicate} label="Blocklist Duplicates" />
							<Toggle bind:checked={settings.blocklist_unregistered} label="Blocklist Unregistered" />
							<Toggle bind:checked={settings.blocklist_mismatch} label="Blocklist Mismatch" />
						</div>
					</Collapsible.Content>
				</Collapsible.Root>
				<Toggle bind:checked={settings.search_on_remove} label="Re-search on Remove" hint="Trigger a new search when an item is removed (blocklist, stalled, failed import)" />
				{#if settings.search_on_remove}
					<Input bind:value={settings.search_cooldown_hours} type="number" label="Search Cooldown (hours)" hint="Min hours between re-searches for the same media (0 = no cooldown)" />
					<Input bind:value={settings.max_searches_per_run} type="number" label="Max Searches per Run" hint="Limit re-searches per cleanup cycle per instance (0 = unlimited)" />
					<Input bind:value={settings.max_search_failures} type="number" label="Max Search Failures" hint="Stop retrying items after this many consecutive failures (0 = no limit)" />
				{/if}
				<Toggle bind:checked={settings.tag_instead_of_delete} label="Tag Media on Removal" hint="Apply an obsolete tag to the media item when removing from queue" />
				{#if settings.tag_instead_of_delete}
					<Input bind:value={settings.obsolete_tag_label} label="Obsolete Tag Label" hint="Tag name applied in the *arr app (e.g. lurkarr-obsolete)" />
				{/if}
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Failed Imports">
			<div class="space-y-2">
				<Toggle bind:checked={settings.failed_import_remove} label="Remove Failed Imports" />
				<Toggle bind:checked={settings.failed_import_blocklist} label="Blocklist Failed Imports" />
				<Input bind:value={settings.failed_import_patterns} label="Message Patterns" hint="Comma-separated substrings to match (empty = built-in defaults: import failed, no files found, etc.)" />
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Metadata Mismatch">
			<div class="space-y-2">
				<Toggle bind:checked={settings.mismatch_enabled} label="Detect Metadata Mismatches" hint="Strike downloads whose metadata doesn't match the expected media (wrong series/movie/episode)" />
				{#if settings.mismatch_enabled}
					<Input bind:value={settings.max_strikes_mismatch} type="number" label="Max Strikes (mismatch)" hint="0 = use global max strikes" />
					<Input bind:value={settings.custom_mismatch_keywords} label="Extra Mismatch Keywords" hint="Comma-separated extra phrases to detect (added to built-in defaults)" />
				{/if}
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Unregistered Torrents">
			<div class="space-y-2">
				<Toggle bind:checked={settings.unregistered_enabled} label="Detect Unregistered Torrents" hint="Strike torrents that have been removed or unregistered from their tracker" />
				{#if settings.unregistered_enabled}
					<Input bind:value={settings.max_strikes_unregistered} type="number" label="Max Strikes (unregistered)" hint="0 = use global max strikes" />
					<Input bind:value={settings.custom_unregistered_keywords} label="Extra Unregistered Keywords" hint="Comma-separated extra phrases to detect (added to built-in defaults)" />
				{/if}
			</div>
		</CollapsibleCard>

		<CollapsibleCard title="Seeding Rules">
			<Toggle bind:checked={settings.seeding_enabled} label="Enable Seeding Enforcement" />
			{#if settings.seeding_enabled}
				<div class="grid grid-cols-1 sm:grid-cols-2 gap-3 mt-3">
					<Input bind:value={settings.seeding_max_ratio} type="number" label="Max Ratio" hint="0 = disabled" />
					<Input bind:value={settings.seeding_max_hours} type="number" label="Max Hours" hint="0 = disabled" />
				</div>
				<Select bind:value={settings.seeding_mode} label="Mode" class="mt-2">
					<option value="or">Either condition (OR)</option>
					<option value="and">Both conditions (AND)</option>
				</Select>
				<div class="space-y-2 mt-2">
					<Toggle bind:checked={settings.seeding_delete_files} label="Delete Files on Removal" />
					<Toggle bind:checked={settings.seeding_skip_private} label="Skip Private Trackers" />
				</div>
			{/if}
		</CollapsibleCard>

		{#if settings.seeding_enabled}
			<Card>
				<div class="flex items-center justify-between mb-3">
					<h3 class="text-sm font-semibold text-foreground">Seeding Rule Groups</h3>
					<Button size="sm" onclick={() => onShowAddGroupChange(true)}>Add Group</Button>
				</div>
				<p class="text-xs text-muted-foreground mb-3">Override seeding limits per tracker, category, or tag. First match wins (highest priority). Items not matching any group use the global settings above.</p>

				{#if showAddGroup}
					<div class="border border-border rounded-lg p-3 space-y-2 mb-3">
						<Input bind:value={newGroup.name} label="Name" />
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							<Select bind:value={newGroup.match_type} label="Match Type">
								<option value="tracker">Tracker (domain contains)</option>
								<option value="category">Category (exact)</option>
								<option value="tag">Tag (exact)</option>
							</Select>
							<Input bind:value={newGroup.match_pattern} label="Pattern" />
						</div>
						<Input bind:value={newGroup.priority} type="number" label="Priority" hint="Higher = checked first" />
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							<Input bind:value={newGroup.max_ratio} type="number" label="Max Ratio" hint="0 = disabled" />
							<Input bind:value={newGroup.max_hours} type="number" label="Max Hours" hint="0 = disabled" />
						</div>
						<Select bind:value={newGroup.seeding_mode} label="Mode">
							<option value="or">Either (OR)</option>
							<option value="and">Both (AND)</option>
						</Select>
						<div class="space-y-2">
							<Toggle bind:checked={newGroup.skip_removal} label="Skip Removal" hint="Never remove torrents matching this group" />
							<Toggle bind:checked={newGroup.delete_files} label="Delete Files" />
						</div>
						<div class="flex gap-2">
							<Button size="sm" onclick={onCreateGroup} disabled={savingGroup}>Create</Button>
							<Button size="sm" variant="ghost" onclick={() => onShowAddGroupChange(false)}>Cancel</Button>
						</div>
					</div>
				{/if}

				{#each seedingGroups as group (group.id)}
					{#if editingGroup?.id === group.id}
						<div class="border border-border rounded-lg p-3 space-y-2 mb-2">
							<Input bind:value={editingGroup.name} label="Name" />
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
								<Select bind:value={editingGroup.match_type} label="Match Type">
									<option value="tracker">Tracker (domain contains)</option>
									<option value="category">Category (exact)</option>
									<option value="tag">Tag (exact)</option>
								</Select>
								<Input bind:value={editingGroup.match_pattern} label="Pattern" />
							</div>
							<Input bind:value={editingGroup.priority} type="number" label="Priority" />
							<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
								<Input bind:value={editingGroup.max_ratio} type="number" label="Max Ratio" />
								<Input bind:value={editingGroup.max_hours} type="number" label="Max Hours" />
							</div>
							<Select bind:value={editingGroup.seeding_mode} label="Mode">
								<option value="or">Either (OR)</option>
								<option value="and">Both (AND)</option>
							</Select>
							<div class="space-y-2">
								<Toggle bind:checked={editingGroup.skip_removal} label="Skip Removal" />
								<Toggle bind:checked={editingGroup.delete_files} label="Delete Files" />
							</div>
							<div class="flex gap-2">
								<Button size="sm" onclick={() => onUpdateGroup(editingGroup!)}>Save</Button>
								<Button size="sm" variant="ghost" onclick={() => onEditingGroupChange(null)}>Cancel</Button>
							</div>
						</div>
					{:else}
						<div class="flex items-center justify-between border border-border rounded-lg p-2 mb-2">
							<div class="flex-1 min-w-0">
								<div class="flex items-center gap-2">
									<span class="font-medium text-sm">{group.name}</span>
									<Badge>{group.match_type}: {group.match_pattern}</Badge>
									<Badge variant="info">P{group.priority}</Badge>
									{#if group.skip_removal}<Badge variant="warning">skip</Badge>{/if}
								</div>
								<div class="text-xs text-muted-foreground mt-0.5">
									{group.max_ratio > 0 ? `Ratio ≥${group.max_ratio}` : ''}
									{group.max_ratio > 0 && group.max_hours > 0 ? ` ${group.seeding_mode.toUpperCase()} ` : ''}
									{group.max_hours > 0 ? `${group.max_hours}h` : ''}
									{group.max_ratio <= 0 && group.max_hours <= 0 && !group.skip_removal ? 'No limits set' : ''}
								</div>
							</div>
							<div class="flex gap-1">
								<Button size="icon" variant="ghost" class="h-auto w-auto p-1" onclick={() => onEditingGroupChange({...group})} aria-label="Edit seeding group">
									<SquarePen class="w-3.5 h-3.5" />
								</Button>
								<ConfirmAction active={confirmDeleteGroup === group.id} onconfirm={() => { onDeleteGroup(group.id); onConfirmDeleteGroupChange(null); }} oncancel={() => onConfirmDeleteGroupChange(null)}>
									<Button size="icon" variant="ghost" class="h-auto w-auto p-1 text-muted-foreground hover:text-destructive" onclick={() => onConfirmDeleteGroupChange(group.id)} aria-label="Delete seeding group">
										<Trash2 class="w-3.5 h-3.5" />
									</Button>
								</ConfirmAction>
							</div>
						</div>
					{/if}
				{:else}
					<p class="text-xs text-muted-foreground italic">No seeding groups defined — all torrents use global settings.</p>
				{/each}
			</Card>
		{/if}

		<CollapsibleCard title="Orphan Cleanup">
			<Toggle bind:checked={settings.orphan_enabled} label="Enable Orphan Detection" />
			{#if settings.orphan_enabled}
				<div class="space-y-2 mt-3">
					<Input bind:value={settings.orphan_grace_minutes} type="number" label="Grace Period (minutes)" />
					<Toggle bind:checked={settings.orphan_delete_files} label="Delete Orphan Files" />
					<Input bind:value={settings.orphan_excluded_categories} label="Excluded Categories" hint="Comma-separated" />
				</div>
			{/if}
		</CollapsibleCard>

		<CollapsibleCard title="Advanced">
			<div class="space-y-2">
				<Toggle bind:checked={settings.hardlink_protection} label="Hardlink Protection" />
				<Toggle bind:checked={settings.skip_cross_seeds} label="Skip Cross-Seeded Torrents" />
				<Toggle bind:checked={settings.cross_arr_sync} label="Cross-Arr Blocklist Sync" />
			</div>
		</CollapsibleCard>

		<Button onclick={save} loading={saving} class="w-full">Save Cleaner Settings</Button>
	</div>
{:else if loaded}
	<Card>
		<p class="text-sm text-muted-foreground text-center py-4">No cleaner settings configured for {appDisplayName(app)}.</p>
	</Card>
{:else}
	<Card>
		<p class="text-sm text-muted-foreground text-center py-4">Loading cleaner settings...</p>
	</Card>
{/if}
