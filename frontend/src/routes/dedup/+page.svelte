<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import { appTypes, appDisplayName, appTabLabel, appLogo, appBgColor, appAccentBorder, appColor } from '$lib';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import InstanceSwitcher from '$lib/components/InstanceSwitcher.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Separator from '$lib/components/ui/Separator.svelte';
	import { Layers, RefreshCw, Search, CheckCircle, XCircle, Minus, AlertTriangle } from 'lucide-svelte';
	import type { InstanceGroupMember, InstanceGroup, CrossInstancePresence, CrossInstanceMedia, CrossInstanceAction, DuplicateFlag, DupScanResult } from '$lib/types';

	const toasts = getToasts();
	const store = getInstances();

	// --- State ---

	let loading = $state(true);
	let selectedApp = $derived(store.selectedApp);
	let groups = $state<InstanceGroup[]>([]);
	let selectedGroupId = $state<string | null>(null);
	let overlaps = $state<CrossInstanceMedia[]>([]);
	let actions = $state<CrossInstanceAction[]>([]);
	let loadingOverlaps = $state(false);
	let scanning = $state(false);
	let scanResult = $state<DupScanResult | null>(null);

	const selectedGroup = $derived(groups.find(g => g.id === selectedGroupId) ?? null);
	const memberInstances = $derived(selectedGroup?.members?.sort((a, b) => a.quality_rank - b.quality_rank) ?? []);
	const hasGroups = $derived(groups.length > 0);

	const modeDescriptions: Record<string, string> = {
		quality_hierarchy: 'Rank-1 instance keeps the file; lower-ranked duplicates are removed or unmonitored.',
		overlap_detect: 'Flags media present in multiple instances without automatic removal.',
		split_season: 'Splits seasons across instances using configured rules (e.g. odd/even, range).'
	};



	// --- Loaders ---

	async function loadGroups() {
		try {
			groups = await api.get<InstanceGroup[]>(`/instance-groups/${selectedApp}`);
			if (groups.length > 0 && !selectedGroupId) {
				selectedGroupId = groups[0].id;
			} else if (groups.length > 0 && !groups.find(g => g.id === selectedGroupId)) {
				selectedGroupId = groups[0].id;
			} else if (groups.length === 0) {
				selectedGroupId = null;
				overlaps = [];
			}
		} catch {
			toasts.error('Failed to load instance groups');
			groups = [];
		}
	}

	async function loadOverlaps() {
		if (!selectedGroupId) {
			overlaps = [];
			return;
		}
		loadingOverlaps = true;
		try {
			overlaps = await api.get<CrossInstanceMedia[]>(`/instance-groups/by-id/${selectedGroupId}/overlaps`);
		} catch {
			toasts.error('Failed to load overlaps');
			overlaps = [];
		} finally {
			loadingOverlaps = false;
		}
	}

	async function loadActions() {
		try {
			actions = await api.get<CrossInstanceAction[]>('/instance-groups/actions?limit=50');
		} catch {
			actions = [];
		}
	}

	async function scanDuplicates() {
		scanning = true;
		scanResult = null;
		try {
			scanResult = await api.post<DupScanResult>('/seerr/scan-duplicates');
			if (scanResult.duplicates?.length === 0) {
				toasts.success('No duplicate requests found');
			} else {
				toasts.success(`Found ${scanResult.duplicates.length} duplicate request(s)`);
			}
		} catch {
			toasts.error('Failed to scan for duplicates');
		} finally {
			scanning = false;
		}
	}

	async function refreshAll() {
		loading = true;
		try {
			await Promise.all([loadGroups(), loadActions()]);
		} finally {
			loading = false;
		}
	}

	// --- Effects ---

	$effect(() => {
		selectedApp;
		selectedGroupId = null;
		overlaps = [];
		refreshAll();
	});

	$effect(() => {
		if (selectedGroupId) {
			loadOverlaps();
		}
	});

	// --- Helpers ---

	function presenceCell(media: CrossInstanceMedia, instanceId: string): CrossInstancePresence | null {
		return media.presence?.find(p => p.instance_id === instanceId) ?? null;
	}

	function cellColor(presence: CrossInstancePresence | null, member: InstanceGroupMember): string {
		if (!presence) return 'bg-muted/30 text-muted-foreground';
		if (presence.has_file && member.quality_rank === 1) return 'bg-emerald-500/15 text-emerald-400';
		if (presence.has_file) return 'bg-destructive/15 text-destructive';
		if (presence.monitored) return 'bg-amber-500/15 text-amber-400';
		return 'bg-muted/30 text-muted-foreground';
	}

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
	}

	// --- Column definitions ---

	const overlapColumns = $derived<Column<CrossInstanceMedia>[]>([
		{ key: 'title', header: 'Media', sortable: true },
		{ key: 'external_id', header: 'External ID', headerClass: 'text-xs' },
		...memberInstances.map((m): Column<CrossInstanceMedia> => ({
			key: `inst_${m.instance_id}`,
			header: m.instance_name ?? m.instance_id.slice(0, 8),
			headerClass: 'text-center text-xs whitespace-nowrap'
		}))
	]);

	const dupColumns: Column<DuplicateFlag>[] = [
		{ key: 'media_title', header: 'Title', sortable: true },
		{ key: 'request_type', header: 'Type', headerClass: 'text-xs', sortable: true },
		{ key: 'requested_by', header: 'Requested By', headerClass: 'text-xs' },
		{ key: 'reason', header: 'Reason', headerClass: 'text-xs' }
	];

	const actionColumns: Column<CrossInstanceAction>[] = [
		{ key: 'title', header: 'Title', sortable: true },
		{ key: 'action', header: 'Action', headerClass: 'text-xs', sortable: true },
		{ key: 'reason', header: 'Reason', headerClass: 'text-xs' },
		{ key: 'executed_at', header: 'When', headerClass: 'text-xs', sortable: true }
	];
</script>

<svelte:head><title>Dedup Dashboard - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Cross-Instance Dedup" description="Matrix view of media across instances — identify and manage duplicates.">
		{#snippet actions()}
			<div class="flex items-center gap-2">
				<Button size="sm" variant="outline" onclick={scanDuplicates} loading={scanning}>
					{#snippet children()}<Search class="h-4 w-4 mr-1.5" />Scan Seerr{/snippet}
				</Button>
				<Button size="sm" variant="outline" onclick={refreshAll} loading={loading}>
					{#snippet children()}<RefreshCw class="h-4 w-4 mr-1.5" />Refresh{/snippet}
				</Button>
			</div>
			<HelpDrawer page="dedup" />
		{/snippet}
	</PageHeader>

	<InstanceSwitcher showInstances={false} onchange={refreshAll} />

	{#if loading}
		<Skeleton rows={6} height="h-12" />
	{:else if !hasGroups}
		<EmptyState icon={Layers} title="No instance groups" description="Create instance groups on the Connections page to group multiple instances of the same app and start tracking cross-instance media.">
			{#snippet actions()}
				<Button size="sm" variant="outline" onclick={() => window.location.href = '/apps'}>
					{#snippet children()}Go to Connections{/snippet}
				</Button>
			{/snippet}
		</EmptyState>
		<Card class="!p-4">
			<p class="text-sm text-muted-foreground">
				<strong class="text-foreground">How it works:</strong> On the <a href="/apps" class="underline hover:text-foreground">Connections</a> page, click <em>Add Connection → Instance Group</em> to group e.g. your Radarr and Radarr 4K instances. Then come back here to scan for overlaps.
			</p>
		</Card>
	{:else}
		<!-- Group selector (if multiple groups for this app) -->
		{#if groups.length > 1}
			<div class="flex gap-2 flex-wrap">
				{#each groups as group}
					<Button
						size="sm"
						variant={group.id === selectedGroupId ? 'primary' : 'outline'}
						class={group.id === selectedGroupId ? appBgColor(selectedApp) + ' text-white border-transparent' : ''}
						onclick={() => { selectedGroupId = group.id; }}
					>
						{group.name}
					</Button>
				{/each}
			</div>
		{/if}

		{#if selectedGroup}
			<Card class="border-l-2 {appAccentBorder(selectedApp)}">
				<div class="space-y-5">
					<!-- Group info bar -->
					<div class="flex flex-wrap items-center gap-3">
						<h3 class="text-sm font-semibold text-foreground">{selectedGroup.name}</h3>
						<Badge variant={selectedGroup.mode === 'quality_hierarchy' ? 'info' : 'default'}>
							{#snippet children()}{selectedGroup.mode.replace('_', ' ')}{/snippet}
						</Badge>
						<span class="text-xs text-muted-foreground">{memberInstances.length} instance{memberInstances.length !== 1 ? 's' : ''}</span>
						{#if modeDescriptions[selectedGroup.mode]}
							<span class="text-xs text-muted-foreground italic">{modeDescriptions[selectedGroup.mode]}</span>
						{/if}
					</div>

					<!-- Member legend -->
					<div class="flex flex-wrap gap-4 text-xs">
						{#each memberInstances as member}
							<div class="flex items-center gap-1.5">
								<span class="inline-block w-2 h-2 rounded-full {member.quality_rank === 1 ? 'bg-emerald-400' : 'bg-muted-foreground'}"></span>
								<span class="text-foreground font-medium">{member.instance_name ?? member.instance_id.slice(0, 8)}</span>
								<span class="text-muted-foreground">rank {member.quality_rank}</span>
								{#if member.is_independent}
									<Badge variant="warning">{#snippet children()}indie{/snippet}</Badge>
								{/if}
							</div>
						{/each}
					</div>

					<!-- Color legend -->
					<div class="flex flex-wrap gap-4 text-xs text-muted-foreground">
						<span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded bg-emerald-500/15 border border-emerald-500/30"></span> Winner (rank 1 + file)</span>
						<span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded bg-destructive/15 border border-destructive/30"></span> Duplicate (lower rank + file)</span>
						<span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded bg-amber-500/15 border border-amber-500/30"></span> Monitored (no file)</span>
						<span class="flex items-center gap-1"><span class="inline-block w-3 h-3 rounded bg-muted/30 border border-border"></span> Not present</span>
					</div>

					<Separator />

					<!-- Overlap matrix -->
					{#if loadingOverlaps}
						<Skeleton rows={4} height="h-10" />
					{:else if overlaps.length === 0}
						<div class="text-center py-8">
							<CheckCircle class="h-8 w-8 text-emerald-400 mx-auto mb-2" />
							<p class="text-sm text-foreground font-medium">No overlapping media</p>
							<p class="text-xs text-muted-foreground mt-1">All media in this group is unique to each instance.</p>
						</div>
					{:else}
						<DataTable data={overlaps} columns={overlapColumns} searchable searchPlaceholder="Search media..." pageSize={50} noun="items">
							{#snippet row(media)}
								<T.Row>
									<T.Cell class="font-medium text-foreground max-w-[200px] truncate" title={media.title}>
										{media.title}
									</T.Cell>
									<T.Cell class="text-xs text-muted-foreground font-mono">{media.external_id}</T.Cell>
									{#each memberInstances as member}
										{@const p = presenceCell(media, member.instance_id)}
										<T.Cell class="text-center">
											<span class="inline-flex items-center justify-center w-full rounded px-2 py-0.5 text-xs font-medium {cellColor(p, member)}">
												{#if !p}
													<Minus class="h-3 w-3" />
												{:else if p.has_file}
													<CheckCircle class="h-3 w-3 mr-1" />file
												{:else if p.monitored}
													<AlertTriangle class="h-3 w-3 mr-1" />mon
												{:else}
													<XCircle class="h-3 w-3" />
												{/if}
											</span>
										</T.Cell>
									{/each}
								</T.Row>
							{/snippet}
						</DataTable>

						<p class="text-xs text-muted-foreground text-right">{overlaps.length} overlapping media item{overlaps.length !== 1 ? 's' : ''}</p>
					{/if}
				</div>
			</Card>
		{/if}

		<!-- Seerr Duplicate Scan Results -->
		{#if scanResult}
			<Card>
				<div class="space-y-4">
					<div class="flex items-center justify-between">
						<h3 class="text-sm font-semibold text-foreground">Seerr Duplicate Scan</h3>
						<span class="text-xs text-muted-foreground">{scanResult.total_scanned} requests scanned</span>
					</div>

					{#if scanResult.duplicates.length === 0}
						<div class="text-center py-4">
							<CheckCircle class="h-6 w-6 text-emerald-400 mx-auto mb-1" />
							<p class="text-sm text-foreground">No duplicates found</p>
						</div>
					{:else}
						<DataTable data={scanResult?.duplicates ?? []} columns={dupColumns} searchable pageSize={25} noun="duplicates">
						{#snippet row(dup)}
							<T.Row>
								<T.Cell class="font-medium text-foreground">{dup.media_title}</T.Cell>
								<T.Cell>
									<Badge variant={dup.is4k ? 'info' : 'default'}>
										{#snippet children()}{dup.request_type}{dup.is4k ? ' 4K' : ''}{/snippet}
									</Badge>
								</T.Cell>
								<T.Cell class="text-sm text-muted-foreground">{dup.requested_by}</T.Cell>
								<T.Cell class="text-xs text-muted-foreground max-w-[250px] truncate" title={dup.reason}>{dup.reason}</T.Cell>
							</T.Row>
						{/snippet}
					</DataTable>
					{/if}
				</div>
			</Card>
		{/if}

		<!-- Recent Actions / Audit Log -->
		{#if actions.length > 0}
			<Card>
				<div class="space-y-4">
					<h3 class="text-sm font-semibold text-foreground">Recent Routing Actions</h3>
					<DataTable data={actions} columns={actionColumns} searchable pageSize={25} noun="actions">
					{#snippet row(action)}
						<T.Row>
							<T.Cell class="font-medium text-foreground">{action.title}</T.Cell>
							<T.Cell>
								<Badge variant={action.action === 'decline' ? 'error' : action.action === 'approve' ? 'success' : 'default'}>
									{#snippet children()}{action.action}{/snippet}
								</Badge>
							</T.Cell>
							<T.Cell class="text-xs text-muted-foreground max-w-[250px] truncate" title={action.reason}>{action.reason}</T.Cell>
							<T.Cell class="text-xs text-muted-foreground whitespace-nowrap">{formatDate(action.executed_at)}</T.Cell>
						</T.Row>
					{/snippet}
				</DataTable>
				</div>
			</Card>
		{/if}
	{/if}
</div>
