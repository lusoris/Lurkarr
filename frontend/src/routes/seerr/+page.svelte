<script lang="ts">
	import { api } from '$lib/api';
	import { onMount } from 'svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import { Clapperboard, Tv, ScanSearch, Film, RefreshCw } from '@lucide/svelte';

	const toasts = getToasts();
	import type { SeerrUser, SeerrMedia, MediaRequest, RequestCount, DuplicateFlag, DupScanResult } from '$lib/types';

	interface PageInfo {
		pages: number;
		pageSize: number;
		results: number;
		page: number;
	}

	interface RequestsResponse {
		pageInfo: PageInfo;
		results: MediaRequest[];
	}

	type FilterTab = 'all' | 'pending' | 'approved' | 'declined' | 'processing' | 'available';

	let activeTab = $state<FilterTab>('all');
	let requests = $state<MediaRequest[]>([]);
	let counts = $state<RequestCount | null>(null);
	let loading = $state(true);
	let countsLoading = $state(true);
	let scanning = $state(false);
	let duplicates = $state<DuplicateFlag[]>([]);

	const filterMap: Record<FilterTab, string> = {
		all: '',
		pending: 'pending',
		approved: 'approved',
		declined: 'declined',
		processing: 'processing',
		available: 'available'
	};

	async function loadRequests() {
		loading = true;
		try {
			const filter = filterMap[activeTab];
			const q = filter ? `?filter=${filter}` : '';
			const resp = await api.get<RequestsResponse>(`/seerr/requests${q}`);
			requests = resp.results ?? [];
		} catch {
			requests = [];
		}
		loading = false;
	}

	async function loadCounts() {
		countsLoading = true;
		try {
			counts = await api.get<RequestCount>('/seerr/requests/count');
		} catch {
			counts = null;
		}
		countsLoading = false;
	}

	async function scanDuplicates() {
		scanning = true;
		try {
			const result = await api.post<DupScanResult>('/seerr/scan-duplicates');
			duplicates = result.duplicates ?? [];
			if (duplicates.length === 0) {
				toasts.success('No duplicates found');
			} else {
				toasts.success(`Found ${duplicates.length} potential duplicate(s)`);
			}
		} catch {
			toasts.error('Failed to scan for duplicates');
		}
		scanning = false;
	}

	function onTabChange(tab: string) {
		activeTab = tab as FilterTab;
		loadRequests();
	}

	onMount(() => {
		loadRequests();
		loadCounts();
	});

	function statusLabel(status: number): string {
		switch (status) {
			case 1: return 'Pending';
			case 2: return 'Approved';
			case 3: return 'Declined';
			case 4: return 'Failed';
			case 5: return 'Completed';
			default: return 'Unknown';
		}
	}

	function statusVariant(status: number): 'default' | 'warning' | 'success' | 'error' | 'info' {
		switch (status) {
			case 1: return 'warning';
			case 2: return 'success';
			case 3: return 'error';
			case 4: return 'error';
			case 5: return 'info';
			default: return 'default';
		}
	}

	function mediaStatusLabel(status: number): string {
		switch (status) {
			case 1: return 'Unknown';
			case 2: return 'Pending';
			case 3: return 'Processing';
			case 4: return 'Partially Available';
			case 5: return 'Available';
			case 6: return 'Deleted';
			default: return 'Unknown';
		}
	}

	function mediaStatusVariant(status: number): 'default' | 'warning' | 'success' | 'error' | 'info' {
		switch (status) {
			case 2: return 'warning';
			case 3: return 'info';
			case 4: return 'warning';
			case 5: return 'success';
			case 6: return 'error';
			default: return 'default';
		}
	}

	const tabList = $derived([
		{ value: 'all', label: `All${counts ? ` (${counts.total})` : ''}` },
		{ value: 'pending', label: `Pending${counts ? ` (${counts.pending})` : ''}` },
		{ value: 'approved', label: `Approved${counts ? ` (${counts.approved})` : ''}` },
		{ value: 'declined', label: `Declined${counts ? ` (${counts.declined})` : ''}` },
		{ value: 'processing', label: `Processing${counts ? ` (${counts.processing})` : ''}` },
		{ value: 'available', label: `Available${counts ? ` (${counts.available})` : ''}` }
	]);
</script>

<svelte:head><title>Seerr - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Seerr" description="View and manage media requests from Seerr">
		{#snippet actions()}
			<div class="flex items-center gap-2">
				<Button size="sm" variant="outline" onclick={() => { loadRequests(); loadCounts(); }} disabled={loading}>
					<RefreshCw class="h-4 w-4 mr-1.5 {loading ? 'animate-spin' : ''}" />
					Refresh
				</Button>
				<Button size="sm" variant="outline" onclick={scanDuplicates} disabled={scanning}>
					<ScanSearch class="h-4 w-4 mr-1.5 {scanning ? 'animate-spin' : ''}" />
					{scanning ? 'Scanning...' : 'Scan Duplicates'}
				</Button>
			</div>
			<HelpDrawer page="seerr" />
		{/snippet}
	</PageHeader>

	<!-- Count Summary Cards -->
	{#if countsLoading}
		<Skeleton rows={1} height="h-16" />
	{:else if counts}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3">
			<Card class="!p-3 text-center">
				<p class="text-2xl font-bold text-foreground">{counts.total}</p>
				<p class="text-xs text-muted-foreground">Total Requests</p>
			</Card>
			<Card class="!p-3 text-center">
				<p class="text-2xl font-bold text-warning">{counts.pending}</p>
				<p class="text-xs text-muted-foreground">Pending</p>
			</Card>
			<Card class="!p-3 text-center">
				<p class="text-2xl font-bold text-info">{counts.movie}</p>
				<p class="text-xs text-muted-foreground">Movies</p>
			</Card>
			<Card class="!p-3 text-center">
				<p class="text-2xl font-bold text-info">{counts.tv}</p>
				<p class="text-xs text-muted-foreground">TV Shows</p>
			</Card>
		</div>
	{/if}

	<Tabs tabs={tabList} bind:value={activeTab} onchange={onTabChange} />

	<!-- Duplicate Scan Results -->
	{#if duplicates.length > 0}
		<Card class="border-warning/50">
			<div class="p-4 space-y-3">
				<h3 class="text-sm font-semibold text-warning mb-3">Potential Duplicates Found ({duplicates.length})</h3>
				{#each duplicates as dup}
					<div class="flex items-center justify-between rounded-lg bg-warning/5 border border-warning/20 px-3 py-2 text-sm">
						<span class="text-foreground">{dup.title ?? `Request #${dup.id ?? dup.request_id ?? '?'}`}</span>
						<Badge variant="warning">duplicate</Badge>
					</div>
				{/each}
			</div>
		</Card>
	{/if}

	<!-- Request List -->
	{#if loading}
		<Skeleton rows={6} height="h-16" />
	{:else if requests.length === 0}
		<EmptyState
			icon={Film}
			title="No requests found"
			description={activeTab === 'all' ? 'Seerr requests will appear here once configured.' : `No ${activeTab} requests found.`}
		/>
	{:else}
		<div class="space-y-2">
			{#each requests as req}
				{@const isMovie = req.type === 'movie'}
				<Card class="!p-0">
					<div class="flex items-center gap-4 p-4">
						<!-- Type Icon -->
						<div class="shrink-0 flex items-center justify-center w-10 h-10 rounded-lg bg-muted/50">
							{#if isMovie}
								<Clapperboard class="h-5 w-5 text-amber-400" />
							{:else}
								<Tv class="h-5 w-5 text-sky-400" />
							{/if}
						</div>

						<!-- Main Info -->
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 flex-wrap">
								<span class="font-medium text-foreground truncate">
									Request #{req.id}
								</span>
								<Badge variant={statusVariant(req.status)}>{statusLabel(req.status)}</Badge>
								<Badge variant={isMovie ? 'warning' : 'info'}>{isMovie ? 'Movie' : 'TV'}</Badge>
								{#if req.is4k}
									<Badge variant="default">4K</Badge>
								{/if}
								{#if req.isAutoRequest}
									<Badge variant="default">Auto</Badge>
								{/if}
							</div>
							<div class="flex items-center gap-3 mt-1 text-xs text-muted-foreground">
								<span>by {req.requestedBy?.displayName ?? 'Unknown'}</span>
								<span>·</span>
								<span>{new Date(req.createdAt).toLocaleDateString()}</span>
								{#if req.media}
									<span>·</span>
									<Badge variant={mediaStatusVariant(req.media.status)} class="text-[10px]">
										{mediaStatusLabel(req.media.status)}
									</Badge>
								{/if}
							</div>
						</div>

						<!-- External Links -->
						<div class="shrink-0 flex items-center gap-2">
							{#if req.media?.tmdbId}
								<a
									href="https://www.themoviedb.org/{isMovie ? 'movie' : 'tv'}/{req.media.tmdbId}"
									target="_blank"
									rel="noopener noreferrer"
									class="text-xs text-muted-foreground hover:text-foreground transition-colors"
								>
									TMDB ↗
								</a>
							{/if}
							{#if req.media?.imdbId}
								<a
									href="https://www.imdb.com/title/{req.media.imdbId}"
									target="_blank"
									rel="noopener noreferrer"
									class="text-xs text-muted-foreground hover:text-foreground transition-colors"
								>
									IMDb ↗
								</a>
							{/if}
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>
