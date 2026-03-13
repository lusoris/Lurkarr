<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, visibleAppTypes, appDisplayName, appLogo, appWebsite, appTabLabel, appColor, appPlaceholderUrl, type AppType } from '$lib';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';

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
	let showProwlarrModal = $state(false);
	let showSeerrModal = $state(false);

	// ── Download client instances state ─────────────────────
	let dlClients = $state<DownloadClientInstance[]>([]);
	let dlHealthStatus = $state<Record<string, { status: string; version?: string }>>({});
	let showDlModal = $state(false);
	let editingDl = $state<DownloadClientInstance | null>(null);
	let dlForm = $state({ name: '', client_type: 'sabnzbd' as string, url: '', api_key: '', username: '', password: '', category: '', timeout: 30, enabled: true });

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
		api.get<ProwlarrSettings>('/prowlarr/settings').then(r => prowlarr = r).catch(() => {});
		api.get<SeerrSettings>('/seerr/settings').then(r => seerr = r).catch(() => {});
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
		editingInstance = null;
		modalApp = app;
		if (app === 'whisparr') whisparrVersion = 'eros';
		arrForm = { name: '', api_url: '', api_key: '', enabled: true };
		showArrModal = true;
	}

	function openEditArr(inst: AppInstance) {
		editingInstance = inst;
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
				api_url: arrForm.api_url,
				api_key: arrForm.api_key || editingInstance?.api_key,
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
	function openAddDl() {
		editingDl = null;
		dlForm = { name: '', client_type: 'sabnzbd', url: '', api_key: '', username: '', password: '', category: '', timeout: 30, enabled: true };
		showDlModal = true;
	}

	function openEditDl(dl: DownloadClientInstance) {
		editingDl = dl;
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
				client_type: dlForm.client_type,
				url: dlForm.url,
				api_key: dlForm.api_key || (editingDl ? 'keep' : ''),
				username: dlForm.username,
				password: dlForm.password || (editingDl ? 'keep' : '')
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
			await loadDlClients();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to delete');
		}
	}

	// ── Effects ─────────────────────────────────────────────
	$effect(() => { loadAll(); loadServices(); loadDlClients(); });
</script>

<svelte:head><title>Connections - Lurkarr</title></svelte:head>

<div class="space-y-8">
	<h1 class="text-2xl font-bold text-surface-50">Connections</h1>

	<!-- ── Arr Apps ───────────────────────────────────────── -->
	<section>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-surface-200">Arr Apps</h2>
		</div>
		{#if loading}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
				{#each Array(6) as _}
					<div class="h-28 rounded-xl bg-surface-800/50 animate-pulse"></div>
				{/each}
			</div>
		{:else}
			{#each visibleAppTypes as app}
				{@const appInstances = app === 'whisparr'
					? [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])]
					: (instances[app] ?? [])}
				{@const logo = appLogo(app)}
				{@const website = appWebsite(app)}
				<div class="mb-5">
					<div class="flex items-center justify-between mb-2">
						<div class="flex items-center gap-2">
							{#if logo}
								<img src={logo} alt={appDisplayName(app)} class="w-5 h-5 rounded" />
							{/if}
							{#if website}
								<a href={website} target="_blank" rel="noopener noreferrer" class="text-sm font-semibold text-surface-300 hover:text-lurk-400 transition-colors">
									{appDisplayName(app)}
								</a>
							{:else}
								<span class="text-sm font-semibold text-surface-300">{appDisplayName(app)}</span>
							{/if}
						</div>
						<Button size="sm" onclick={() => openAddArr(app)}>+ Add</Button>
					</div>
					{#if appInstances.length === 0}
						<p class="text-xs text-surface-600 ml-7">No instances configured</p>
					{:else}
						<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
							{#each appInstances as inst}
								{@const instLogo = appLogo(inst.app_type)}
								<Card class="!p-4 cursor-pointer hover:border-surface-600 transition-colors" onclick={() => openEditArr(inst)}>
									<div class="flex items-start justify-between gap-2 mb-2">
										<div class="flex items-center gap-2 min-w-0">
											{#if instLogo && app === 'whisparr'}
												<img src={instLogo} alt="" class="w-5 h-5 rounded shrink-0" />
											{/if}
											<span class="font-medium text-sm text-surface-100 truncate">{inst.name}</span>
											{#if app === 'whisparr'}
												<span class="text-[10px] {appColor(inst.app_type)} shrink-0">({inst.app_type === 'eros' ? 'v3' : 'v2'})</span>
											{/if}
										</div>
										{#if healthStatus[inst.id]}
											{#if healthStatus[inst.id].status === 'ok'}
												<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/15 text-emerald-400 shrink-0">
													<span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
													v{healthStatus[inst.id].version}
												</span>
											{:else}
												<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/15 text-red-400 shrink-0">
													<span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
													offline
												</span>
											{/if}
										{:else}
											<span class="w-3 h-3 rounded-full border-2 border-surface-600 border-t-surface-400 animate-spin shrink-0"></span>
										{/if}
									</div>
									<p class="text-xs text-surface-500 truncate mb-1">{inst.api_url}</p>
									<Badge variant={inst.enabled ? 'success' : 'error'}>
										{inst.enabled ? 'Enabled' : 'Disabled'}
									</Badge>
								</Card>
							{/each}
						</div>
					{/if}
				</div>
			{/each}
		{/if}
	</section>

	<!-- ── Download Clients ──────────────────────────────── -->
	<section>
		<div class="flex items-center justify-between mb-4">
			<h2 class="text-lg font-semibold text-surface-200">Download Clients</h2>
			<Button size="sm" onclick={openAddDl}>+ Add</Button>
		</div>
		{#if dlClients.length === 0}
			<p class="text-sm text-surface-500">No download clients configured</p>
		{:else}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
				{#each dlClients as dl}
					{@const logo = appLogo(dl.client_type)}
					<Card class="!p-4 cursor-pointer hover:border-surface-600 transition-colors" onclick={() => openEditDl(dl)}>
						<div class="flex items-start justify-between gap-2 mb-2">
							<div class="flex items-center gap-2 min-w-0">
								{#if logo}
									<img src={logo} alt="" class="w-5 h-5 rounded shrink-0" />
								{/if}
								<span class="font-medium text-sm text-surface-100 truncate">{dl.name}</span>
							</div>
							{#if dl.enabled && dlHealthStatus[dl.id]}
								{#if dlHealthStatus[dl.id].status === 'ok'}
									<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-emerald-500/15 text-emerald-400 shrink-0">
										<span class="w-1.5 h-1.5 rounded-full bg-emerald-400"></span>
										{dlHealthStatus[dl.id].version ? `v${dlHealthStatus[dl.id].version}` : 'online'}
									</span>
								{:else}
									<span class="inline-flex items-center gap-1 px-1.5 py-0.5 rounded text-[10px] font-medium bg-red-500/15 text-red-400 shrink-0">
										<span class="w-1.5 h-1.5 rounded-full bg-red-400"></span>
										offline
									</span>
								{/if}
							{:else if dl.enabled}
								<span class="w-3 h-3 rounded-full border-2 border-surface-600 border-t-surface-400 animate-spin shrink-0"></span>
							{/if}
						</div>
						<div class="flex items-center gap-2 mb-1">
							<Badge variant="info">{appDisplayName(dl.client_type)}</Badge>
							{#if dl.category}
								<span class="text-[10px] text-surface-500">cat: {dl.category}</span>
							{/if}
						</div>
						<p class="text-xs text-surface-500 truncate mb-1">{dl.url}</p>
						<Badge variant={dl.enabled ? 'success' : 'error'}>
							{dl.enabled ? 'Enabled' : 'Disabled'}
						</Badge>
					</Card>
				{/each}
			</div>
		{/if}
	</section>

	<!-- ── Services (Prowlarr & Seerr) ───────────────────── -->
	<section>
		<h2 class="text-lg font-semibold text-surface-200 mb-4">Services</h2>
		<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
			<!-- Prowlarr card -->
			<Card class="!p-4 cursor-pointer hover:border-surface-600 transition-colors" onclick={() => showProwlarrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('prowlarr')} alt="Prowlarr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-surface-100">Prowlarr</span>
					</div>
					<Badge variant={prowlarr?.enabled ? 'success' : 'error'}>
						{prowlarr?.enabled ? 'Enabled' : 'Disabled'}
					</Badge>
				</div>
				{#if prowlarr?.url}
					<p class="text-xs text-surface-500 truncate mb-1">{prowlarr.url}</p>
				{:else}
					<p class="text-xs text-surface-600 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-surface-600">Indexer manager</p>
			</Card>

			<!-- Seerr card -->
			<Card class="!p-4 cursor-pointer hover:border-surface-600 transition-colors" onclick={() => showSeerrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('seerr')} alt="Seerr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-surface-100">Seerr</span>
					</div>
					<Badge variant={seerr?.enabled ? 'success' : 'error'}>
						{seerr?.enabled ? 'Enabled' : 'Disabled'}
					</Badge>
				</div>
				{#if seerr?.url}
					<p class="text-xs text-surface-500 truncate mb-1">{seerr.url}</p>
				{:else}
					<p class="text-xs text-surface-600 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-surface-600">Request management</p>
			</Card>
		</div>
	</section>
</div>

<!-- ── Add/Edit Arr Instance Modal ───────────────────────── -->
<Modal bind:open={showArrModal} title={editingInstance ? `Edit ${appDisplayName(modalApp)} Instance` : `Add ${appDisplayName(modalApp)} Instance`} onclose={() => showArrModal = false}>
	<form onsubmit={(e: Event) => { e.preventDefault(); saveArrInstance(); }} class="space-y-4">
		{#if modalApp === 'whisparr'}
			<label class="block">
				<span class="block text-sm font-medium text-surface-300 mb-1.5">Version</span>
				<select bind:value={whisparrVersion} disabled={!!editingInstance} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500 disabled:opacity-50">
					<option value="whisparr">v2</option>
					<option value="eros">v3</option>
				</select>
			</label>
		{/if}
		<Input bind:value={arrForm.name} label="Name" placeholder="My {appDisplayName(modalApp)}" />
		<Input bind:value={arrForm.api_url} label="URL" placeholder={appPlaceholderUrl(effectiveAppType)} />
		<Input bind:value={arrForm.api_key} type="password" label={editingInstance ? 'API Key (leave empty to keep current)' : 'API Key'} />
		{#if editingInstance}
			<Toggle bind:checked={arrForm.enabled} label="Enabled" />
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
		<label class="block">
			<span class="block text-sm font-medium text-surface-300 mb-1.5">Client Type</span>
			<select bind:value={dlForm.client_type} disabled={!!editingDl} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500 disabled:opacity-50">
				{#each clientTypes as ct}
					<option value={ct}>{ct}</option>
				{/each}
			</select>
		</label>
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
			<Toggle bind:checked={dlForm.enabled} label="Enabled" />
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
			<Toggle bind:checked={prowlarr.enabled} label="Enabled" />
			<Input bind:value={prowlarr.url} label="URL" placeholder="http://prowlarr:9696" hint="Base URL of your Prowlarr instance" />
			<Input bind:value={prowlarr.api_key} label="API Key" type="password" hint="Found in Prowlarr → Settings → General" />
			<Toggle bind:checked={prowlarr.sync_indexers} label="Sync Indexers" hint="Automatically sync Prowlarr indexers to connected arr apps" />
			<Input bind:value={prowlarr.timeout} type="number" label="Timeout (seconds)" hint="Connection timeout for Prowlarr API requests" />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveProwlarr} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testProwlarr}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-surface-500 text-center py-4">Loading Prowlarr settings...</p>
	{/if}
</Modal>

<!-- ── Seerr Settings Modal ──────────────────────────────── -->
<Modal bind:open={showSeerrModal} title="Seerr Settings" onclose={() => showSeerrModal = false}>
	{#if seerr}
		<div class="space-y-4">
			<Toggle bind:checked={seerr.enabled} label="Enabled" />
			<Input bind:value={seerr.url} label="URL" placeholder="http://seerr:5055" hint="Base URL of your Seerr instance" />
			<Input bind:value={seerr.api_key} label="API Key" type="password" hint="Found in Seerr → Settings → General" />
			<Input bind:value={seerr.sync_interval_minutes} type="number" label="Sync Interval (minutes)" hint="How often to poll Seerr for new requests" />
			<Toggle bind:checked={seerr.auto_approve} label="Auto-Approve Requests" hint="Automatically approve incoming media requests" />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveSeerr} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testSeerr}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-surface-500 text-center py-4">Loading Seerr settings...</p>
	{/if}
</Modal>
