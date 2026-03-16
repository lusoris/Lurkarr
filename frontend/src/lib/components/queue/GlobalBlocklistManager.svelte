<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import { SquarePen, Trash2 } from '@lucide/svelte';
	import type { BlocklistSource, BlocklistRule } from '$lib/types';

	interface Props {
		sources: BlocklistSource[];
		rules: BlocklistRule[];
		sourcesLoaded: boolean;
		rulesLoaded: boolean;
		showAddSource: boolean;
		showAddRule: boolean;
		editingSource: BlocklistSource | null;
		newSource: Omit<BlocklistSource, 'id' | 'created_at' | 'updated_at' | 'last_synced_at'>;
		newRule: Omit<BlocklistRule, 'id' | 'created_at' | 'enabled'>;
		regexTestInput: string;
		regexTestResult: boolean | 'invalid' | null;
		savingSource: boolean;
		savingRule: boolean;
		confirmDeleteSource: string | null;
		confirmDeleteRule: string | null;
		onCreateSource: () => Promise<void>;
		onUpdateSource: (src: BlocklistSource) => Promise<void>;
		onDeleteSource: (id: string) => Promise<void>;
		onCreateRule: () => Promise<void>;
		onDeleteRule: (id: string) => Promise<void>;
		onShowAddSourceChange: (show: boolean) => void;
		onShowAddRuleChange: (show: boolean) => void;
		onEditingSourceChange: (src: BlocklistSource | null) => void;
		onNewSourceChange: (partial: Partial<Omit<BlocklistSource, 'id' | 'created_at' | 'updated_at' | 'last_synced_at'>>) => void;
		onNewRuleChange: (partial: Partial<Omit<BlocklistRule, 'id' | 'created_at' | 'enabled'>>) => void;
		onRegexTestInputChange: (input: string) => void;
		onConfirmDeleteSourceChange: (id: string | null) => void;
		onConfirmDeleteRuleChange: (id: string | null) => void;
	}

	let {
		sources = $bindable(),
		rules = $bindable(),
		sourcesLoaded,
		rulesLoaded,
		showAddSource = $bindable(),
		showAddRule = $bindable(),
		editingSource = $bindable(),
		newSource = $bindable(),
		newRule = $bindable(),
		regexTestInput = $bindable(),
		regexTestResult,
		savingSource,
		savingRule,
		confirmDeleteSource = $bindable(),
		confirmDeleteRule = $bindable(),
		onCreateSource,
		onUpdateSource,
		onDeleteSource,
		onCreateRule,
		onDeleteRule,
		onShowAddSourceChange,
		onShowAddRuleChange,
		onEditingSourceChange,
		onNewSourceChange,
		onNewRuleChange,
		onRegexTestInputChange,
		onConfirmDeleteSourceChange,
		onConfirmDeleteRuleChange
	}: Props = $props();

	const toasts = getToasts();
</script>

<div class="space-y-4">
	<h3 class="text-sm font-semibold text-foreground">Global Blocklist Management</h3>
	<p class="text-xs text-muted-foreground">Manage community blocklist sources and custom rules that apply across all apps.</p>

	<!-- Sources -->
	<Card>
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-sm font-semibold text-muted-foreground">Blocklist Sources</h3>
			<Button size="sm" onclick={() => onShowAddSourceChange(!showAddSource)}>
				{showAddSource ? 'Cancel' : '+ Add Source'}
			</Button>
		</div>

		{#if showAddSource}
			<div class="mb-4 p-4 rounded-lg bg-muted/50 border border-border space-y-3">
				<Input bind:value={newSource.name} label="Name" placeholder="e.g. Trash Guides blocklist" />
				<Input bind:value={newSource.url} label="URL" placeholder="https://example.com/blocklist.txt" />
				<Input bind:value={newSource.sync_interval_hours} type="number" label="Sync Interval (hours)" />
				<div class="flex justify-end">
					<Button size="sm" onclick={onCreateSource} loading={savingSource}>Add Source</Button>
				</div>
			</div>
		{/if}

		{#if !sourcesLoaded}
			<Skeleton rows={3} height="h-16" />
		{:else if sources.length === 0}
			<p class="text-sm text-muted-foreground py-4 text-center">No blocklist sources configured</p>
		{:else}
			<div class="space-y-2">
				{#each sources as src}
					{#if editingSource?.id === src.id}
						<div class="p-3 rounded-lg bg-muted/50 border border-border space-y-3">
							<Input bind:value={editingSource.name} label="Name" />
							<Input bind:value={editingSource.url} label="URL" />
							<Input bind:value={editingSource.sync_interval_hours} type="number" label="Sync Interval (hours)" />
							<Toggle bind:checked={editingSource.enabled} label="Enabled" />
							<div class="flex justify-end gap-2">
								<Button size="sm" variant="ghost" onclick={() => onEditingSourceChange(null)}>Cancel</Button>
								<Button size="sm" onclick={() => onUpdateSource(editingSource!)}>Save</Button>
							</div>
						</div>
					{:else}
						<div class="flex items-center justify-between p-3 rounded-lg bg-card border border-border">
							<div class="min-w-0 flex-1">
								<div class="flex items-center gap-2">
									<span class="text-sm font-medium text-foreground truncate">{src.name}</span>
									{#if !src.enabled}
										<Badge variant="default">Disabled</Badge>
									{/if}
								</div>
								<p class="text-xs text-muted-foreground mt-0.5 truncate">{src.url}</p>
								{#if src.last_synced_at}
									<p class="text-xs text-muted-foreground/50 mt-0.5">Last synced: {new Date(src.last_synced_at).toLocaleString()}</p>
								{/if}
							</div>
							<div class="flex items-center gap-1.5 ml-3 shrink-0">
								<Button size="icon" variant="ghost" class="h-auto w-auto p-0" onclick={() => onEditingSourceChange({...src})} aria-label="Edit source">
									<SquarePen class="w-4 h-4" />
								</Button>
								<ConfirmAction active={confirmDeleteSource === src.id} onconfirm={() => { onDeleteSource(src.id); onConfirmDeleteSourceChange(null); }} oncancel={() => onConfirmDeleteSourceChange(null)}>
									<Button size="icon" variant="ghost" class="h-auto w-auto p-0 text-muted-foreground hover:text-destructive" onclick={() => onConfirmDeleteSourceChange(src.id)} aria-label="Delete source">
										<Trash2 class="w-4 h-4" />
									</Button>
								</ConfirmAction>
							</div>
						</div>
					{/if}
				{/each}
			</div>
		{/if}
	</Card>

	<!-- Rules -->
	<Card>
		<div class="flex items-center justify-between mb-4">
			<h3 class="text-sm font-semibold text-muted-foreground">Custom Rules</h3>
			<Button size="sm" onclick={() => onShowAddRuleChange(!showAddRule)}>
				{showAddRule ? 'Cancel' : '+ Add Rule'}
			</Button>
		</div>

		{#if showAddRule}
			<div class="mb-4 p-4 rounded-lg bg-muted/50 border border-border space-y-3">
				<Select bind:value={newRule.pattern_type} label="Pattern Type">
					<option value="title_contains">Title Contains</option>
					<option value="title_regex">Title Regex</option>
					<option value="release_group">Release Group</option>
					<option value="indexer">Indexer</option>
				</Select>
				<Input 
					bind:value={newRule.pattern} 
					label="Pattern" 
					placeholder={newRule.pattern_type === 'title_regex' ? '.*YIFY.*' : newRule.pattern_type === 'release_group' ? 'YTS' : 'pattern'} 
				/>
				{#if newRule.pattern_type === 'title_regex' && newRule.pattern}
					<div class="space-y-2">
						<Input bind:value={regexTestInput} label="Test Regex" placeholder="Enter text to test the pattern" />
						{#if regexTestResult !== null}
							{#if regexTestResult === 'invalid'}
								<p class="text-xs text-destructive">Invalid regex pattern</p>
							{:else if regexTestResult}
								<p class="text-xs text-green-600">Pattern matches ✓</p>
							{:else}
								<p class="text-xs text-muted-foreground">Pattern does not match</p>
							{/if}
						{/if}
					</div>
				{/if}
				<Input bind:value={newRule.reason} label="Reason" placeholder="Why this rule exists" />
				<div class="flex justify-end">
					<Button size="sm" onclick={onCreateRule} loading={savingRule}>Add Rule</Button>
				</div>
			</div>
		{/if}

		{#if !rulesLoaded}
			<Skeleton rows={3} height="h-12" />
		{:else if rules.length === 0}
			<p class="text-sm text-muted-foreground py-4 text-center">No custom rules configured</p>
		{:else}
			<div class="space-y-2">
				{#each rules as rule}
					<div class="flex items-center justify-between p-3 rounded-lg bg-card border border-border">
						<div class="min-w-0 flex-1">
							<div class="flex items-center gap-2 flex-wrap">
								<span class="text-sm font-medium text-foreground">{rule.pattern}</span>
								<Badge variant="info">{rule.pattern_type}</Badge>
								{#if rule.reason}
									<span class="text-xs text-muted-foreground">{rule.reason}</span>
								{/if}
							</div>
						</div>
						<div class="flex items-center gap-1 ml-3 shrink-0">
							<ConfirmAction active={confirmDeleteRule === rule.id} onconfirm={() => { onDeleteRule(rule.id); onConfirmDeleteRuleChange(null); }} oncancel={() => onConfirmDeleteRuleChange(null)}>
								<Button size="icon" variant="ghost" class="h-auto w-auto p-0 text-muted-foreground hover:text-destructive" onclick={() => onConfirmDeleteRuleChange(rule.id)} aria-label="Delete rule">
									<Trash2 class="w-4 h-4" />
								</Button>
							</ConfirmAction>
						</div>
					</div>
				{/each}
			</div>
		{/if}
	</Card>
</div>
