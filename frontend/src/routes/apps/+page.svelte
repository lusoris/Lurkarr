<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, visibleAppTypes, appDisplayName, appLogo, appWebsite, appTabLabel, appColor, appAccentBorder, appBgColor, appPlaceholderUrl, type AppType } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import { Plus, Cable as CableIcon } from 'lucide-svelte';

	const toasts = getToasts();

	// ── Types ───────────────────────────────────────────────
	interface AppInstance {
		id: string;
		app_type: string;
		name: string;
		api_url: string;
		api_key: string;
		enabled: boolean;
	}

	interface ProwlarrSettings {
		url: string;
		api_key: string;
		enabled: boolean;
		sync_indexers: boolean;
		timeout: number;
	}

	interface SeerrSettings {
		id: string;
		url: string;
		api_key: string;
		enabled: boolean;
		sync_interval_minutes: number;
		auto_approve: boolean;
	}

	interface DownloadClientInstance {
		id: string;
		name: string;
		client_type: string;
		url: string;
		api_key: string;
		username: string;
		password: string;
		category: string;
		enabled: boolean;
		timeout: number;
	}

	const clientTypes = ['qbittorrent', 'transmission', 'deluge', 'sabnzbd', 'nzbget'] as const;

	const clientDefaults: Record<string, { url: string; port: number }> = {
		qbittorrent: { url: 'http://qbittorrent', port: 8080 },
		transmission: { url: 'http://transmission', port: 9091 },
		deluge: { url: 'http://deluge', port: 8112 },
		sabnzbd: { url: 'http://sabnzbd', port: 8080 },
		nzbget: { url: 'http://nzbget', port: 6789 }
	};

	// ── Arr instances state ─────────────────────────────────
	let instances = $state<Record<string, AppInstance[]>>({});
	let healthStatus = $state<Record<string, { status: string; version: string }>>({});
	let loading = $state(true);
	let showArrModal = $state(false);
	let editingInstance = $state<AppInstance | null>(null);
	let modalApp = $state<AppType>('sonarr');
	let whisparrVersion = $state<'whisparr' | 'eros'>('eros');
	let arrForm = $state({ name: '', api_url: '', api_key: '', enabled: true });
	let saving = $state(false);
	let testing = $state(false);
	let deleteConfirm = $state<string | null>(null);
	const effectiveAppType = $derived(modalApp === 'whisparr' ? whisparrVersion : modalApp);

	// ── Service connection state ────────────────────────────
	let prowlarr = $state<ProwlarrSettings | null>(null);
	let seerr = $state<SeerrSettings | null>(null);
	let prowlarrHealth = $state<{ status: string; version?: string } | null>(null);
	let seerrHealth = $state<{ status: string; version?: string } | null>(null);
	let showProwlarrModal = $state(false);
	let showSeerrModal = $state(false);

	// ── Download client instances state ─────────────────────
	let dlClients = $state<DownloadClientInstance[]>([]);
	let dlHealthStatus = $state<Record<string, { status: string; version?: string }>>({});
	let showDlModal = $state(false);
	let editingDl = $state<DownloadClientInstance | null>(null);
	let dlForm = $state({ name: '', client_type: 'sabnzbd' as string, url: '', api_key: '', username: '', password: '', category: '', timeout: 30, enabled: true });

	// ── Add dropdown state ──────────────────────────────────
	let showAddDropdown = $state(false);

	// ── Derived: which app types actually have instances ────
	const populatedArrApps = $derived(
		visibleAppTypes.filter(app => {
			if (app === 'whisparr') {
				return (instances['whisparr']?.length ?? 0) + (instances['eros']?.length ?? 0) > 0;
			}
			return (instances[app]?.length ?? 0) > 0;
		})
	);

	const hasAnyArrApps = $derived(populatedArrApps.length > 0);
	const hasAnyDlClients = $derived(dlClients.length > 0);
	const hasAnything = $derived(hasAnyArrApps || hasAnyDlClients);

	// ── Loaders ─────────────────────────────────────────────
	async function loadAll() {
		loading = true;
		const results: Record<string, AppInstance[]> = {};
		await Promise.all(
			appTypes.map(async (app) => {
				try {
					results[app] = await api.get<AppInstance[]>(`/instances/${app}`);
				} catch {
					results[app] = [];
				}
			})
		);
		instances = results;
		loading = false;
		checkAllHealth(results);
	}

	async function checkAllHealth(allInstances: Record<string, AppInstance[]>) {
		const allInsts = Object.values(allInstances).flat();
		await Promise.allSettled(
			allInsts.map(async (inst) => {
				try {
					const res = await api.get<{ status: string; version: string }>(`/instances/${inst.id}/health`);
					healthStatus = { ...healthStatus, [inst.id]: { status: res.status, version: res.version } };
				} catch {
					healthStatus = { ...healthStatus, [inst.id]: { status: 'offline', version: '' } };
				}
			})
		);
	}

	async function loadServices() {
		api.get<ProwlarrSettings>('/prowlarr/settings').then(r => {
			prowlarr = r;
			if (r?.enabled && r.url && r.api_key) {
				api.post<{ status: string; version: string }>('/prowlarr/test', { url: r.url, api_key: r.api_key })
					.then(h => prowlarrHealth = h)
					.catch(() => prowlarrHealth = { status: 'offline' });
			}
		}).catch(() => { prowlarr = { url: '', api_key: '', enabled: false, sync_indexers: false, timeout: 30 }; });
		api.get<SeerrSettings>('/seerr/settings').then(r => {
			seerr = r;
			if (r?.enabled && r.url) {
				api.post<{ status: string; version: string }>('/seerr/test')
					.then(h => seerrHealth = h)
					.catch(() => seerrHealth = { status: 'offline' });
			}
		}).catch(() => { seerr = { id: '', url: '', api_key: '', enabled: false, sync_interval_minutes: 60, auto_approve: false }; });
	}

	async function loadDlClients() {
		try {
			dlClients = await api.get<DownloadClientInstance[]>('/download-clients');
		} catch {
			dlClients = [];
		}
		checkDlHealth();
	}

	async function checkDlHealth() {
		await Promise.allSettled(
			dlClients.filter(d => d.enabled).map(async (dl) => {
				try {
					const res = await api.get<{ status: string; version?: string }>(`/download-clients/${dl.id}/health`);
					dlHealthStatus = { ...dlHealthStatus, [dl.id]: res };
				} catch {
					dlHealthStatus = { ...dlHealthStatus, [dl.id]: { status: 'offline' } };
				}
			})
		);
	}

	// ── Arr instance actions ────────────────────────────────
	function openAddArr(app: (typeof visibleAppTypes)[number]) {
		showAddDropdown = false;
		editingInstance = null;
		modalApp = app;
		if (app === 'whisparr') whisparrVersion = 'eros';
		arrForm = { name: '', api_url: '', api_key: '', enabled: true };
		showArrModal = true;
	}

	function openEditArr(inst: AppInstance) {
		editingInstance = inst;
		deleteConfirm = null;
		const t = inst.app_type as AppType;
		modalApp = t === 'eros' ? 'whisparr' : t;
		if (t === 'whisparr' || t === 'eros') whisparrVersion = t;
		arrForm = { name: inst.name, api_url: inst.api_url, api_key: '', enabled: inst.enabled };
		showArrModal = true;
	}

	async function saveArrInstance() {
		saving = true;
		try {
			if (editingInstance) {
				await api.put(`/instances/${editingInstance.id}`, {
					name: arrForm.name,
					api_url: arrForm.api_url,
					api_key: arrForm.api_key,
					enabled: arrForm.enabled
				});
				toasts.success('Instance updated');
			} else {
				await api.post(`/instances/${effectiveAppType}`, {
					name: arrForm.name,
					api_url: arrForm.api_url,
					api_key: arrForm.api_key
				});
				toasts.success('Instance added');
			}
			showArrModal = false;
			await loadAll();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
		saving = false;
	}

	async function testArrConnection() {
		testing = true;
		try {
			const res = await api.post<{ status: string; app: string; version: string }>('/instances/test', {
				id: editingInstance?.id ?? '',
				api_url: arrForm.api_url,
				api_key: arrForm.api_key,
				app_type: effectiveAppType
			});
			toasts.success(`Connected — ${res.app} v${res.version}`);
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Connection failed');
		}
		testing = false;
	}

	async function deleteArrInstance(id: string) {
		try {
			await api.del(`/instances/${id}`);
			toasts.success('Instance deleted');
			deleteConfirm = null;
			showArrModal = false;
			await loadAll();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to delete');
		}
	}

	// ── Service save/test actions ───────────────────────────
	async function saveProwlarr() {
		if (!prowlarr) return;
		saving = true;
		try {
			await api.put('/prowlarr/settings', prowlarr);
			toasts.success('Prowlarr settings saved');
		} catch { toasts.error('Failed to save Prowlarr settings'); }
		saving = false;
	}

	async function testProwlarr() {
		if (!prowlarr) return;
		try {
			await api.post('/prowlarr/test', { url: prowlarr.url, api_key: prowlarr.api_key });
			toasts.success('Prowlarr connection successful');
		} catch { toasts.error('Prowlarr connection failed'); }
	}

	async function saveSeerr() {
		if (!seerr) return;
		saving = true;
		try {
			await api.put('/seerr/settings', seerr);
			toasts.success('Seerr settings saved');
		} catch { toasts.error('Failed to save Seerr settings'); }
		saving = false;
	}

	async function testSeerr() {
		try {
			const res = await api.post<{ status: string; version: string }>('/seerr/test');
			toasts.success(`Seerr connected — v${res.version}`);
		} catch { toasts.error('Seerr connection failed'); }
	}

	// ── Download client actions ─────────────────────────────
	function openAddDl(clientType?: string) {
		showAddDropdown = false;
		editingDl = null;
		dlForm = { name: '', client_type: clientType ?? 'sabnzbd', url: '', api_key: '', username: '', password: '', category: '', timeout: 30, enabled: true };
		showDlModal = true;
	}

	function openEditDl(dl: DownloadClientInstance) {
		editingDl = dl;
		deleteConfirm = null;
		dlForm = { name: dl.name, client_type: dl.client_type, url: dl.url, api_key: '', username: dl.username, password: '', category: dl.category, timeout: dl.timeout, enabled: dl.enabled };
		showDlModal = true;
	}

	async function saveDlClient() {
		saving = true;
		try {
			if (editingDl) {
				await api.put(`/download-clients/${editingDl.id}`, {
					name: dlForm.name,
					client_type: dlForm.client_type,
					url: dlForm.url,
					api_key: dlForm.api_key,
					username: dlForm.username,
					password: dlForm.password,
					category: dlForm.category,
					timeout: dlForm.timeout,
					enabled: dlForm.enabled
				});
				toasts.success('Download client updated');
			} else {
				await api.post('/download-clients', {
					name: dlForm.name,
					client_type: dlForm.client_type,
					url: dlForm.url,
					api_key: dlForm.api_key,
					username: dlForm.username,
					password: dlForm.password,
					category: dlForm.category,
					timeout: dlForm.timeout
				});
				toasts.success('Download client added');
			}
			showDlModal = false;
			await loadDlClients();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
		saving = false;
	}

	async function testDlConnection() {
		testing = true;
		try {
			const res = await api.post<{ status: string; version: string }>('/download-clients/test', {
				id: editingDl?.id ?? '',
				client_type: dlForm.client_type,
				url: dlForm.url,
				api_key: dlForm.api_key,
				username: dlForm.username,
				password: dlForm.password
			});
			toasts.success(`Connected — ${dlForm.client_type} v${res.version}`);
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Connection failed');
		}
		testing = false;
	}

	async function deleteDlClient(id: string) {
		try {
			await api.del(`/download-clients/${id}`);
			toasts.success('Download client deleted');
			deleteConfirm = null;
			showDlModal = false;
			await loadDlClients();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to delete');
		}
	}

	// ── Close dropdown on outside click ─────────────────────
	function handleWindowClick() {
		if (showAddDropdown) showAddDropdown = false;
	}

	// ── Effects ─────────────────────────────────────────────
	$effect(() => { loadAll(); loadServices(); loadDlClients(); });
</script>

<svelte:head><title>Connections - Lurkarr</title></svelte:head>
<svelte:window onclick={handleWindowClick} />

<div class="space-y-8">
	<!-- ── Header with Add Dropdown ──────────────────────── -->
	<PageHeader title="Connections" description="Manage your Arr apps, download clients, and services.">
		{#snippet actions()}
			<div class="relative">
				<Button size="sm" onclick={(e) => { e.stopPropagation(); showAddDropdown = !showAddDropdown; }}>
					<Plus class="h-4 w-4" />
					Add Connection
				</Button>
			{#if showAddDropdown}
				<!-- svelte-ignore a11y_no_static_element_interactions -->
				<div
					class="absolute right-0 top-full mt-2 w-56 rounded-xl border border-border bg-muted shadow-xl shadow-black/30 z-50 overflow-hidden"
					role="menu"
					tabindex="-1"
					onclick={(e) => e.stopPropagation()}
					onkeydown={(e) => { if (e.key === 'Escape') showAddDropdown = false; }}
				>
					<!-- Arr Apps -->
					<div class="px-3 pt-3 pb-1">
						<span class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Arr Apps</span>
					</div>
					{#each visibleAppTypes as app}
						{@const logo = appLogo(app)}
						<button
							class="flex items-center gap-2.5 w-full px-3 py-2 text-sm text-foreground/80 hover:text-foreground hover:bg-primary/10 transition-colors"
							onclick={() => openAddArr(app)}
							role="menuitem"
						>
							{#if logo}
								<img src={logo} alt="" class="w-4 h-4 rounded shrink-0" />
							{/if}
							<span>{appDisplayName(app)}</span>
						</button>
					{/each}

					<div class="mx-3 border-t border-border"></div>

					<!-- Download Clients -->
					<div class="px-3 pt-3 pb-1">
						<span class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Download Clients</span>
					</div>
					{#each clientTypes as ct}
						{@const logo = appLogo(ct)}
						<button
							class="flex items-center gap-2.5 w-full px-3 py-2 text-sm text-foreground/80 hover:text-foreground hover:bg-primary/10 transition-colors"
							onclick={() => openAddDl(ct)}
							role="menuitem"
						>
							{#if logo}
								<img src={logo} alt="" class="w-4 h-4 rounded shrink-0" />
							{/if}
							<span>{appDisplayName(ct)}</span>
						</button>
					{/each}

					<div class="mx-3 border-t border-border"></div>

					<!-- Services -->
					<div class="px-3 pt-3 pb-1">
						<span class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Services</span>
					</div>
					<button
						class="flex items-center gap-2.5 w-full px-3 py-2 text-sm text-foreground/80 hover:text-foreground hover:bg-primary/10 transition-colors"
						onclick={() => { showAddDropdown = false; showProwlarrModal = true; }}
						role="menuitem"
					>
						<img src={appLogo('prowlarr')} alt="" class="w-4 h-4 rounded shrink-0" />
						<span>Prowlarr</span>
					</button>
					<button
						class="flex items-center gap-2.5 w-full px-3 py-2 pb-3 text-sm text-foreground/80 hover:text-foreground hover:bg-primary/10 transition-colors"
						onclick={() => { showAddDropdown = false; showSeerrModal = true; }}
						role="menuitem"
					>
						<img src={appLogo('seerr')} alt="" class="w-4 h-4 rounded shrink-0" />
						<span>Seerr</span>
					</button>
				</div>
			{/if}
			</div>
		{/snippet}
	</PageHeader>

	{#if loading}
		<!-- ── Skeleton loader ──────────────────────────────── -->
		<Skeleton rows={6} height="h-28" />
	{:else if !hasAnything}
		<!-- ── Empty state ──────────────────────────────────── -->
		<EmptyState icon={CableIcon} title="No connections yet" description="Add your first Arr app, download client, or service to get started with Lurkarr.">
			{#snippet actions()}
				<Button onclick={(e) => { e.stopPropagation(); showAddDropdown = !showAddDropdown; }}>
					<Plus class="h-4 w-4" />
					Add Connection
				</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<!-- ── Arr Apps (only populated) ────────────────────── -->
		{#if hasAnyArrApps}
			<section>
				<h2 class="text-lg font-semibold text-foreground mb-4">Arr Apps</h2>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
					{#each populatedArrApps as app}
						{@const appInstances = app === 'whisparr'
							? [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])]
							: (instances[app] ?? [])}
						{#each appInstances as inst}
							{@const instLogo = appLogo(inst.app_type)}
							<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder(inst.app_type)}" onclick={() => openEditArr(inst)}>
								<div class="flex items-start justify-between gap-2 mb-2">
									<div class="flex items-center gap-2 min-w-0">
										{#if instLogo}
										<img src={instLogo} alt="" class="w-5 h-5 rounded shrink-0" />
										{/if}
										<span class="font-medium text-sm text-foreground truncate">{inst.name}</span>
										{#if app === 'whisparr'}
											<span class="text-[10px] {appColor(inst.app_type)} shrink-0">({inst.app_type === 'eros' ? 'v3' : 'v2'})</span>
										{/if}
									</div>
									{#if healthStatus[inst.id]}
										{#if healthStatus[inst.id].status === 'ok'}
											<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/15 text-emerald-400 shrink-0">
												<span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
												{healthStatus[inst.id].version ? (healthStatus[inst.id].version.startsWith('v') ? healthStatus[inst.id].version : `v${healthStatus[inst.id].version}`) : 'online'}
											</span>
										{:else}
											<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/15 text-red-400 shrink-0">
												<span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
												offline
											</span>
										{/if}
									{:else}
										<span class="w-3 h-3 rounded-full border-2 border-muted-foreground/50 border-t-muted-foreground animate-spin shrink-0"></span>
									{/if}
								</div>
								<p class="text-xs text-muted-foreground truncate mb-1">{inst.api_url}</p>
								<Badge variant={inst.enabled ? 'success' : 'error'}>
									{inst.enabled ? 'Enabled' : 'Disabled'}
								</Badge>
							</Card>
						{/each}
					{/each}
				</div>
			</section>
		{/if}

		<!-- ── Download Clients (only if populated) ─────────── -->
		{#if hasAnyDlClients}
			<section>
				<h2 class="text-lg font-semibold text-foreground mb-4">Download Clients</h2>
				<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
					{#each dlClients as dl}
						{@const logo = appLogo(dl.client_type)}
						<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder(dl.client_type)}" onclick={() => openEditDl(dl)}>
							<div class="flex items-start justify-between gap-2 mb-2">
								<div class="flex items-center gap-2 min-w-0">
									{#if logo}
									<img src={logo} alt="" class="w-5 h-5 rounded shrink-0" />
									{/if}
										<span class="font-medium text-sm text-foreground truncate">{dl.name}</span>
								</div>
								{#if dl.enabled && dlHealthStatus[dl.id]}
									{#if dlHealthStatus[dl.id].status === 'ok'}
										<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/15 text-emerald-400 shrink-0">
											<span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
											{dlHealthStatus[dl.id]?.version ? (dlHealthStatus[dl.id].version!.startsWith('v') ? dlHealthStatus[dl.id].version : `v${dlHealthStatus[dl.id].version}`) : 'online'}
										</span>
									{:else}
										<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/15 text-red-400 shrink-0">
											<span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
											offline
										</span>
									{/if}
								{:else if dl.enabled}
										<span class="w-3 h-3 rounded-full border-2 border-muted-foreground/50 border-t-muted-foreground animate-spin shrink-0"></span>
								{/if}
							</div>
							<div class="flex items-center gap-2 mb-1">
								<Badge variant="info">{appDisplayName(dl.client_type)}</Badge>
								{#if dl.category}
										<span class="text-[10px] text-muted-foreground">cat: {dl.category}</span>
								{/if}
							</div>
							<p class="text-xs text-muted-foreground truncate mb-1">{dl.url}</p>
							<Badge variant={dl.enabled ? 'success' : 'error'}>
								{dl.enabled ? 'Enabled' : 'Disabled'}
							</Badge>
						</Card>
					{/each}
				</div>
			</section>
		{/if}
	{/if}

	<!-- ── Services (always shown) ───────────────────────── -->
	<section>
		<h2 class="text-lg font-semibold text-foreground mb-4">Services</h2>
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
			<!-- Prowlarr card -->
			<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder('prowlarr')}" onclick={() => showProwlarrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('prowlarr')} alt="Prowlarr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-foreground">Prowlarr</span>
					</div>
					{#if prowlarr?.enabled && prowlarrHealth}
						{#if prowlarrHealth.status === 'ok'}
							<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/15 text-emerald-400 shrink-0">
								<span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
								{prowlarrHealth.version ? (prowlarrHealth.version.startsWith('v') ? prowlarrHealth.version : `v${prowlarrHealth.version}`) : 'online'}
							</span>
						{:else}
							<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/15 text-red-400 shrink-0">
								<span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
								offline
							</span>
						{/if}
					{:else if prowlarr?.enabled}
						<span class="w-3 h-3 rounded-full border-2 border-muted-foreground/50 border-t-muted-foreground animate-spin shrink-0"></span>
					{:else}
						<Badge variant={prowlarr?.enabled ? 'success' : 'error'}>
							{prowlarr?.enabled ? 'Enabled' : 'Disabled'}
						</Badge>
					{/if}
				</div>
				{#if prowlarr?.url}
					<p class="text-xs text-muted-foreground truncate mb-1">{prowlarr.url}</p>
				{:else}
					<p class="text-xs text-muted-foreground/50 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-muted-foreground/50">Indexer manager</p>
			</Card>

			<!-- Seerr card -->
			<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder('seerr')}" onclick={() => showSeerrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('seerr')} alt="Seerr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-foreground">Seerr</span>
					</div>
					{#if seerr?.enabled && seerrHealth}
						{#if seerrHealth.status === 'ok'}
							<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/15 text-emerald-400 shrink-0">
								<span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
								{seerrHealth.version ? (seerrHealth.version.startsWith('v') ? seerrHealth.version : `v${seerrHealth.version}`) : 'online'}
							</span>
						{:else}
							<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/15 text-red-400 shrink-0">
								<span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
								offline
							</span>
						{/if}
					{:else if seerr?.enabled}
						<span class="w-3 h-3 rounded-full border-2 border-muted-foreground/50 border-t-muted-foreground animate-spin shrink-0"></span>
					{:else}
						<Badge variant={seerr?.enabled ? 'success' : 'error'}>
							{seerr?.enabled ? 'Enabled' : 'Disabled'}
						</Badge>
					{/if}
				</div>
				{#if seerr?.url}
					<p class="text-xs text-muted-foreground truncate mb-1">{seerr.url}</p>
				{:else}
					<p class="text-xs text-muted-foreground/50 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-muted-foreground/50">Request management</p>
			</Card>
		</div>
	</section>
</div>

<!-- ── Add/Edit Arr Instance Modal ───────────────────────── -->
<Modal bind:open={showArrModal} title={editingInstance ? `Edit ${appDisplayName(modalApp)} Instance` : `Add ${appDisplayName(modalApp)} Instance`} onclose={() => showArrModal = false}>
	<form onsubmit={(e: Event) => { e.preventDefault(); saveArrInstance(); }} class="space-y-4">
		{#if modalApp === 'whisparr'}
			<Select bind:value={whisparrVersion} label="Version" disabled={!!editingInstance}>
				<option value="whisparr">v2</option>
				<option value="eros">v3</option>
			</Select>
		{/if}
		<Input bind:value={arrForm.name} label="Name" placeholder="My {appDisplayName(modalApp)}" />
		<Input bind:value={arrForm.api_url} label="URL" placeholder={appPlaceholderUrl(effectiveAppType)} />
		<Input bind:value={arrForm.api_key} type="password" label={editingInstance ? 'API Key (leave empty to keep current)' : 'API Key'} />
		{#if editingInstance}
			<Toggle bind:checked={arrForm.enabled} label="Enabled" color={appBgColor(effectiveAppType)} />
		{/if}
		<div class="flex items-center gap-2 pt-2">
			<Button type="submit" loading={saving}>{editingInstance ? 'Update' : 'Add'}</Button>
			<Button variant="secondary" loading={testing} onclick={testArrConnection}>Test Connection</Button>
			{#if editingInstance}
				<div class="ml-auto">
					{#if deleteConfirm === editingInstance.id}
						<Button size="sm" variant="danger" onclick={() => { if (editingInstance) deleteArrInstance(editingInstance.id); }}>Confirm Delete</Button>
						<Button size="sm" variant="ghost" onclick={() => deleteConfirm = null}>Cancel</Button>
					{:else}
						<Button size="sm" variant="danger" onclick={() => { if (editingInstance) deleteConfirm = editingInstance.id; }}>Delete</Button>
					{/if}
				</div>
			{/if}
		</div>
	</form>
</Modal>

<!-- ── Add/Edit Download Client Modal ────────────────────── -->
<Modal bind:open={showDlModal} title={editingDl ? 'Edit Download Client' : 'Add Download Client'} onclose={() => showDlModal = false}>
	<form onsubmit={(e: Event) => { e.preventDefault(); saveDlClient(); }} class="space-y-4">
		<Input bind:value={dlForm.name} label="Name" placeholder="My {appDisplayName(dlForm.client_type)}" />
		<Select bind:value={dlForm.client_type} label="Client Type" disabled={!!editingDl}>
			{#each clientTypes as ct}
				<option value={ct}>{appDisplayName(ct)}</option>
			{/each}
		</Select>
		<Input bind:value={dlForm.url} label="URL" placeholder="{clientDefaults[dlForm.client_type]?.url ?? 'http://localhost'}:{clientDefaults[dlForm.client_type]?.port ?? 8080}" hint="Including port number" />
		{#if dlForm.client_type === 'sabnzbd' || dlForm.client_type === 'nzbget'}
			<Input bind:value={dlForm.api_key} type="password" label={editingDl ? 'API Key (leave empty to keep current)' : 'API Key'} hint="Found in {dlForm.client_type === 'sabnzbd' ? 'SABnzbd → Config → General' : 'NZBGet → Settings → Security'}" />
		{/if}
		{#if dlForm.client_type !== 'sabnzbd'}
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-4">
				<Input bind:value={dlForm.username} label="Username" />
				<Input bind:value={dlForm.password} type="password" label={editingDl ? 'Password (leave empty to keep current)' : 'Password'} />
			</div>
		{/if}
		{#if dlForm.client_type === 'sabnzbd'}
			<Input bind:value={dlForm.category} label="Category" placeholder="Optional" hint="SABnzbd category to assign to downloads" />
		{/if}
		<Input bind:value={dlForm.timeout} type="number" label="Timeout (seconds)" hint="Connection timeout" />
		{#if editingDl}
			<Toggle bind:checked={dlForm.enabled} label="Enabled" color={appBgColor(dlForm.client_type)} />
		{/if}
		<div class="flex items-center gap-2 pt-2">
			<Button type="submit" loading={saving}>{editingDl ? 'Update' : 'Add'}</Button>
			<Button variant="secondary" loading={testing} onclick={testDlConnection}>Test Connection</Button>
			{#if editingDl}
				<div class="ml-auto">
					{#if deleteConfirm === editingDl.id}
						<Button size="sm" variant="danger" onclick={() => { if (editingDl) deleteDlClient(editingDl.id); }}>Confirm Delete</Button>
						<Button size="sm" variant="ghost" onclick={() => deleteConfirm = null}>Cancel</Button>
					{:else}
						<Button size="sm" variant="danger" onclick={() => { if (editingDl) deleteConfirm = editingDl.id; }}>Delete</Button>
					{/if}
				</div>
			{/if}
		</div>
	</form>
</Modal>

<!-- ── Prowlarr Settings Modal ───────────────────────────── -->
<Modal bind:open={showProwlarrModal} title="Prowlarr Settings" onclose={() => showProwlarrModal = false}>
	{#if prowlarr}
		<div class="space-y-4">
			<Toggle bind:checked={prowlarr.enabled} label="Enabled" color={appBgColor('prowlarr')} />
			<Input bind:value={prowlarr.url} label="URL" placeholder="http://prowlarr:9696" hint="Base URL of your Prowlarr instance" />
			<Input bind:value={prowlarr.api_key} label="API Key" type="password" hint="Found in Prowlarr → Settings → General" />
			<Toggle bind:checked={prowlarr.sync_indexers} label="Sync Indexers" hint="Automatically sync Prowlarr indexers to connected arr apps" color={appBgColor('prowlarr')} />
			<Input bind:value={prowlarr.timeout} type="number" label="Timeout (seconds)" hint="Connection timeout for Prowlarr API requests" />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveProwlarr} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testProwlarr}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground text-center py-4">Loading Prowlarr settings...</p>
	{/if}
</Modal>

<!-- ── Seerr Settings Modal ──────────────────────────────── -->
<Modal bind:open={showSeerrModal} title="Seerr Settings" onclose={() => showSeerrModal = false}>
	{#if seerr}
		<div class="space-y-4">
			<Toggle bind:checked={seerr.enabled} label="Enabled" color={appBgColor('seerr')} />
			<Input bind:value={seerr.url} label="URL" placeholder="http://seerr:5055" hint="Base URL of your Seerr instance" />
			<Input bind:value={seerr.api_key} label="API Key" type="password" hint="Found in Seerr → Settings → General" />
			<Input bind:value={seerr.sync_interval_minutes} type="number" label="Sync Interval (minutes)" hint="How often to poll Seerr for new requests" />
			<Toggle bind:checked={seerr.auto_approve} label="Auto-Approve Requests" hint="Automatically approve incoming media requests" color={appBgColor('seerr')} />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveSeerr} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testSeerr}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground text-center py-4">Loading Seerr settings...</p>
	{/if}
</Modal>
