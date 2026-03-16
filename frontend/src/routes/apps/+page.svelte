<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, visibleAppTypes, appDisplayName, appLogo, appWebsite, appTabLabel, appColor, appAccentBorder, appBgColor, appPlaceholderUrl, type AppType } from '$lib';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import { getInstances } from '$lib/stores/instances.svelte';
	import { onMount, untrack } from 'svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import * as DropdownMenu from '$lib/components/ui/dropdown-menu';
	import * as ScrollArea from '$lib/components/ui/scroll-area';
	import { Plus, Cable as CableIcon, Layers, Trash2, GripVertical } from 'lucide-svelte';
	import type { AppInstance, ProwlarrSettings, SeerrSettings, BazarrSettings, KapowarrSettings, ShokoSettings, DownloadClientInstance, HealthInfo, InstanceGroup, InstanceGroupMember } from '$lib/types';
	import HealthBadge from '$lib/components/ui/HealthBadge.svelte';

	const toasts = getToasts();
	const instanceStore = getInstances();

	const clientTypes = ['qbittorrent', 'transmission', 'deluge', 'rtorrent', 'sabnzbd', 'nzbget'] as const;

	const clientDefaults: Record<string, { url: string; port: number }> = {
		qbittorrent: { url: 'http://qbittorrent', port: 8080 },
		transmission: { url: 'http://transmission', port: 9091 },
		deluge: { url: 'http://deluge', port: 8112 },
		rtorrent: { url: 'http://rtorrent', port: 8080 },
		sabnzbd: { url: 'http://sabnzbd', port: 8080 },
		nzbget: { url: 'http://nzbget', port: 6789 }
	};

	// ── Arr instances state (from shared store) ─────────────
	const instances = $derived(instanceStore.cache);
	let healthStatus = $state<Record<string, HealthInfo>>({});
	let loading = $derived(instanceStore.loading);
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
	let bazarr = $state<BazarrSettings | null>(null);
	let kapowarr = $state<KapowarrSettings | null>(null);
	let shoko = $state<ShokoSettings | null>(null);
	let prowlarrHealth = $state<HealthInfo | null>(null);
	let seerrHealth = $state<HealthInfo | null>(null);
	let bazarrHealth = $state<HealthInfo | null>(null);
	let kapowarrHealth = $state<HealthInfo | null>(null);
	let shokoHealth = $state<HealthInfo | null>(null);
	let showProwlarrModal = $state(false);
	let showSeerrModal = $state(false);
	let showBazarrModal = $state(false);
	let showKapowarrModal = $state(false);
	let showShokoModal = $state(false);
	let prowlarrIndexers = $state<{ id: number; name: string; enable: boolean; protocol: string; priority: number; fields?: { name: string; value: any }[] }[]>([]);
	let indexersLoading = $state(false);

	// ── Download client instances state ─────────────────────
	let dlClients = $state<DownloadClientInstance[]>([]);
	let dlHealthStatus = $state<Record<string, HealthInfo>>({});
	let showDlModal = $state(false);
	let editingDl = $state<DownloadClientInstance | null>(null);
	let dlForm = $state({ name: '', client_type: 'sabnzbd' as string, url: '', api_key: '', username: '', password: '', category: '', timeout: 30, enabled: true });

	// ── Instance Groups state ───────────────────────────────
	let instanceGroups = $state<InstanceGroup[]>([]);
	let showGroupModal = $state(false);
	let editingGroup = $state<InstanceGroup | null>(null);
	let groupForm = $state({ name: '', app_type: 'radarr' as string, mode: 'quality_hierarchy' });
	let groupMembers = $state<{ instance_id: string; quality_rank: number; is_independent: boolean }[]>([]);
	let savingGroup = $state(false);
	let deleteGroupConfirm = $state<string | null>(null);

	const groupModes = [
		{ value: 'quality_hierarchy', label: 'Quality Hierarchy', description: 'Rank-1 instance keeps the file; lower-ranked duplicates are removed.' },
		{ value: 'overlap_detect', label: 'Overlap Detect', description: 'Flags media present in multiple instances without automatic removal.' },
		{ value: 'split_season', label: 'Split Season', description: 'Splits seasons across instances using configured rules.' }
	] as const;

	// App types that support groups (same-type instances)
	const groupableAppTypes = ['sonarr', 'radarr', 'lidarr', 'readarr'] as const;

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
	const hasAnyGroups = $derived(instanceGroups.length > 0);
	const hasAnything = $derived(hasAnyArrApps || hasAnyDlClients || hasAnyGroups);

	// ── Loaders ─────────────────────────────────────────────
	async function loadAll() {
		await instanceStore.refetch();
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
					.then(h => { prowlarrHealth = h; if (h.status === 'ok') loadProwlarrIndexers(); })
					.catch(() => prowlarrHealth = { status: 'offline' });
			}
		}).catch(() => { prowlarr = { url: '', api_key: '', enabled: false, sync_indexers: false, timeout: 30 }; });
		api.get<SeerrSettings>('/seerr/settings').then(r => {
			seerr = r;
			if (r?.enabled && r.url) {
				api.post<{ status: string; version: string }>('/seerr/test', { url: r.url, api_key: r.api_key })
					.then(h => seerrHealth = h)
					.catch(() => seerrHealth = { status: 'offline' });
			}
		}).catch(() => { seerr = { id: '', url: '', api_key: '', enabled: false, sync_interval_minutes: 60, auto_approve: false, cleanup_enabled: false, cleanup_after_days: 30 }; });
		api.get<BazarrSettings>('/bazarr/settings').then(r => {
			bazarr = r;
			if (r?.enabled && r.url && r.api_key) {
				api.post<{ status: string; version: string }>('/bazarr/test', { url: r.url, api_key: r.api_key })
					.then(h => bazarrHealth = h)
					.catch(() => bazarrHealth = { status: 'offline' });
			}
		}).catch(() => { bazarr = { id: 0, url: '', api_key: '', enabled: false, timeout: 30 }; });
		api.get<KapowarrSettings>('/kapowarr/settings').then(r => {
			kapowarr = r;
			if (r?.enabled && r.url && r.api_key) {
				api.post<{ status: string; version: string }>('/kapowarr/test', { url: r.url, api_key: r.api_key })
					.then(h => kapowarrHealth = h)
					.catch(() => kapowarrHealth = { status: 'offline' });
			}
		}).catch(() => { kapowarr = { id: 0, url: '', api_key: '', enabled: false, timeout: 30 }; });
		api.get<ShokoSettings>('/shoko/settings').then(r => {
			shoko = r;
			if (r?.enabled && r.url && r.api_key) {
				api.post<{ status: string; version: string }>('/shoko/test', { url: r.url, api_key: r.api_key })
					.then(h => shokoHealth = h)
					.catch(() => shokoHealth = { status: 'offline' });
			}
		}).catch(() => { shoko = { id: 0, url: '', api_key: '', enabled: false, timeout: 30 }; });
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
			loadProwlarrIndexers();
		} catch { toasts.error('Prowlarr connection failed'); }
	}

	async function loadProwlarrIndexers() {
		indexersLoading = true;
		try {
			prowlarrIndexers = await api.get('/prowlarr/indexers');
		} catch { prowlarrIndexers = []; }
		indexersLoading = false;
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
			const res = await api.post<{ status: string; version: string }>('/seerr/test', { url: seerr?.url, api_key: seerr?.api_key });
			toasts.success(`Seerr connected — v${res.version}`);
		} catch { toasts.error('Seerr connection failed'); }
	}

	async function saveBazarr() {
		if (!bazarr) return;
		saving = true;
		try {
			await api.put('/bazarr/settings', bazarr);
			toasts.success('Bazarr settings saved');
		} catch { toasts.error('Failed to save Bazarr settings'); }
		saving = false;
	}

	async function testBazarr() {
		try {
			const res = await api.post<{ status: string; version: string }>('/bazarr/test', { url: bazarr?.url, api_key: bazarr?.api_key });
			toasts.success(`Bazarr connected — v${res.version}`);
		} catch { toasts.error('Bazarr connection failed'); }
	}

	async function saveKapowarr() {
		if (!kapowarr) return;
		saving = true;
		try {
			await api.put('/kapowarr/settings', kapowarr);
			toasts.success('Kapowarr settings saved');
		} catch { toasts.error('Failed to save Kapowarr settings'); }
		saving = false;
	}

	async function testKapowarr() {
		try {
			const res = await api.post<{ status: string; version: string }>('/kapowarr/test', { url: kapowarr?.url, api_key: kapowarr?.api_key });
			toasts.success(`Kapowarr connected — v${res.version}`);
		} catch { toasts.error('Kapowarr connection failed'); }
	}

	async function saveShoko() {
		if (!shoko) return;
		saving = true;
		try {
			await api.put('/shoko/settings', shoko);
			toasts.success('Shoko settings saved');
		} catch { toasts.error('Failed to save Shoko settings'); }
		saving = false;
	}

	async function testShoko() {
		try {
			const res = await api.post<{ status: string; version: string }>('/shoko/test', { url: shoko?.url, api_key: shoko?.api_key });
			toasts.success(`Shoko connected — v${res.version}`);
		} catch { toasts.error('Shoko connection failed'); }
	}

	// ── Download client actions ─────────────────────────────
	function openAddDl(clientType?: string) {
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
					timeout: dlForm.timeout,
					enabled: dlForm.enabled
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

	// ── Instance Group actions ───────────────────────────────
	async function loadInstanceGroups() {
		const allGroups: InstanceGroup[] = [];
		await Promise.allSettled(
			groupableAppTypes.map(async (app) => {
				const groups = await api.get<InstanceGroup[]>(`/instance-groups/${app}`);
				allGroups.push(...groups);
			})
		);
		instanceGroups = allGroups;
	}

	function openAddGroup() {
		editingGroup = null;
		groupForm = { name: '', app_type: 'radarr', mode: 'quality_hierarchy' };
		groupMembers = [];
		showGroupModal = true;
	}

	function openEditGroup(group: InstanceGroup) {
		editingGroup = group;
		deleteGroupConfirm = null;
		groupForm = { name: group.name, app_type: group.app_type, mode: group.mode };
		groupMembers = (group.members ?? []).map(m => ({
			instance_id: m.instance_id,
			quality_rank: m.quality_rank,
			is_independent: m.is_independent
		}));
		showGroupModal = true;
	}

	function addMemberToGroup() {
		const appInstances = getInstancesForGroupApp();
		const available = appInstances.filter(i => !groupMembers.some(m => m.instance_id === i.id));
		if (available.length === 0) return;
		groupMembers = [...groupMembers, {
			instance_id: available[0].id,
			quality_rank: groupMembers.length + 1,
			is_independent: false
		}];
	}

	function removeMember(idx: number) {
		groupMembers = groupMembers.filter((_, i) => i !== idx);
		// Re-rank
		groupMembers = groupMembers.map((m, i) => ({ ...m, quality_rank: i + 1 }));
	}

	function getInstancesForGroupApp(): AppInstance[] {
		const app = editingGroup?.app_type ?? groupForm.app_type;
		// For whisparr, combine both v2 and v3
		if (app === 'whisparr') {
			return [...(instances['whisparr'] ?? []), ...(instances['eros'] ?? [])];
		}
		return instances[app] ?? [];
	}

	function instanceNameById(id: string): string {
		const all = Object.values(instances).flat();
		return all.find(i => i.id === id)?.name ?? id.slice(0, 8);
	}

	async function saveGroup() {
		savingGroup = true;
		try {
			if (editingGroup) {
				await api.put(`/instance-groups/by-id/${editingGroup.id}`, {
					name: groupForm.name,
					mode: groupForm.mode
				});
				if (groupMembers.length > 0) {
					await api.put(`/instance-groups/by-id/${editingGroup.id}/members`, {
						members: groupMembers
					});
				}
				toasts.success('Instance group updated');
			} else {
				const group = await api.post<InstanceGroup>(`/instance-groups/${groupForm.app_type}`, {
					name: groupForm.name
				});
				// Set mode if not default
				if (groupForm.mode !== 'quality_hierarchy') {
					await api.put(`/instance-groups/by-id/${group.id}`, { mode: groupForm.mode });
				}
				// Set members
				if (groupMembers.length > 0) {
					await api.put(`/instance-groups/by-id/${group.id}/members`, {
						members: groupMembers
					});
				}
				toasts.success('Instance group created');
			}
			showGroupModal = false;
			await loadInstanceGroups();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save group');
		}
		savingGroup = false;
	}

	async function deleteGroup(id: string) {
		try {
			await api.del(`/instance-groups/by-id/${id}`);
			toasts.success('Instance group deleted');
			deleteGroupConfirm = null;
			showGroupModal = false;
			await loadInstanceGroups();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to delete group');
		}
	}

	// ── Effects ─────────────────────────────────────────────
	onMount(() => { loadServices(); loadDlClients(); loadInstanceGroups(); });
	$effect(() => { checkAllHealth(instances); });
</script>

<svelte:head><title>Connections - Lurkarr</title></svelte:head>

<div class="space-y-8">
	<!-- ── Header with Add Dropdown ──────────────────────── -->
	<PageHeader title="Connections" description="Manage your Arr apps, download clients, and services.">
		{#snippet actions()}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger class="inline-flex shrink-0 items-center justify-center gap-1.5 rounded-md text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90 h-8 px-3 cursor-pointer [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4">
						<Plus class="h-4 w-4" />
						Add Connection
					</DropdownMenu.Trigger>
					<DropdownMenu.Content class="w-56" align="end">
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Arr Apps</DropdownMenu.GroupHeading>
							{#each visibleAppTypes as app}
								{@const logo = appLogo(app)}
								<DropdownMenu.Item onclick={() => openAddArr(app)}>
									{#if logo}<img src={logo} alt="" class="w-4 h-4 rounded shrink-0" />{/if}
									<span>{appDisplayName(app)}</span>
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Group>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Download Clients</DropdownMenu.GroupHeading>
							{#each clientTypes as ct}
								{@const logo = appLogo(ct)}
								<DropdownMenu.Item onclick={() => openAddDl(ct)}>
									{#if logo}<img src={logo} alt="" class="w-4 h-4 rounded shrink-0" />{/if}
									<span>{appDisplayName(ct)}</span>
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Group>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Services</DropdownMenu.GroupHeading>
							<DropdownMenu.Item onclick={() => { showProwlarrModal = true; }}>
								<img src={appLogo('prowlarr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Prowlarr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showSeerrModal = true; }}>
								<img src={appLogo('seerr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Seerr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showBazarrModal = true; }}>
								<img src={appLogo('bazarr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Bazarr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showKapowarrModal = true; }}>
								<img src={appLogo('kapowarr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Kapowarr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showShokoModal = true; }}>
								<img src={appLogo('shoko')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Shoko</span>
							</DropdownMenu.Item>
						</DropdownMenu.Group>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Instance Groups</DropdownMenu.GroupHeading>
							<DropdownMenu.Item onclick={openAddGroup}>
								<Layers class="w-4 h-4 shrink-0 text-muted-foreground" />
								<span>Instance Group</span>
							</DropdownMenu.Item>
						</DropdownMenu.Group>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
				<HelpDrawer page="apps" />
		{/snippet}
	</PageHeader>

	{#if loading}
		<!-- ── Skeleton loader ──────────────────────────────── -->
		<Skeleton rows={6} height="h-28" />
	{:else if !hasAnything}
		<!-- ── Empty state ──────────────────────────────────── -->
		<EmptyState icon={CableIcon} title="No connections yet" description="Add your first Arr app, download client, or service to get started with Lurkarr.">
			{#snippet actions()}
				<DropdownMenu.Root>
					<DropdownMenu.Trigger class="inline-flex shrink-0 items-center justify-center gap-2 rounded-md text-sm font-medium bg-primary text-primary-foreground shadow-xs hover:bg-primary/90 h-9 px-4 cursor-pointer [&_svg]:pointer-events-none [&_svg]:shrink-0 [&_svg:not([class*='size-'])]:size-4">
						<Plus class="h-4 w-4" />
						Add Connection
					</DropdownMenu.Trigger>
					<DropdownMenu.Content class="w-56">
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Arr Apps</DropdownMenu.GroupHeading>
							{#each visibleAppTypes as app}
								{@const logo = appLogo(app)}
								<DropdownMenu.Item onclick={() => openAddArr(app)}>
									{#if logo}<img src={logo} alt="" class="w-4 h-4 rounded shrink-0" />{/if}
									<span>{appDisplayName(app)}</span>
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Group>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Download Clients</DropdownMenu.GroupHeading>
							{#each clientTypes as ct}
								{@const logo = appLogo(ct)}
								<DropdownMenu.Item onclick={() => openAddDl(ct)}>
									{#if logo}<img src={logo} alt="" class="w-4 h-4 rounded shrink-0" />{/if}
									<span>{appDisplayName(ct)}</span>
								</DropdownMenu.Item>
							{/each}
						</DropdownMenu.Group>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Services</DropdownMenu.GroupHeading>
							<DropdownMenu.Item onclick={() => { showProwlarrModal = true; }}>
								<img src={appLogo('prowlarr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Prowlarr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showSeerrModal = true; }}>
								<img src={appLogo('seerr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Seerr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showBazarrModal = true; }}>
								<img src={appLogo('bazarr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Bazarr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showKapowarrModal = true; }}>
								<img src={appLogo('kapowarr')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Kapowarr</span>
							</DropdownMenu.Item>
							<DropdownMenu.Item onclick={() => { showShokoModal = true; }}>
								<img src={appLogo('shoko')} alt="" class="w-4 h-4 rounded shrink-0" />
								<span>Shoko</span>
							</DropdownMenu.Item>
						</DropdownMenu.Group>
						<DropdownMenu.Separator />
						<DropdownMenu.Group>
							<DropdownMenu.GroupHeading>Instance Groups</DropdownMenu.GroupHeading>
							<DropdownMenu.Item onclick={openAddGroup}>
								<Layers class="w-4 h-4 shrink-0 text-muted-foreground" />
								<span>Instance Group</span>
							</DropdownMenu.Item>
						</DropdownMenu.Group>
					</DropdownMenu.Content>
				</DropdownMenu.Root>
			{/snippet}
		</EmptyState>
	{:else}
		<!-- ── Arr Apps (only populated) ────────────────────── -->
		{#if hasAnyArrApps}
			<section>
				<h3 class="text-sm font-semibold text-foreground mb-3">Arr Apps</h3>
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
								<HealthBadge health={healthStatus[inst.id]} />
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
				<h3 class="text-sm font-semibold text-foreground mb-3">Download Clients</h3>
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
							{#if dl.enabled}
								<HealthBadge health={dlHealthStatus[dl.id]} />
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
		<h3 class="text-sm font-semibold text-foreground mb-3">Services</h3>
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
			<!-- Prowlarr card -->
			<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder('prowlarr')}" onclick={() => showProwlarrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('prowlarr')} alt="Prowlarr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-foreground">Prowlarr</span>
					</div>
					{#if prowlarr?.enabled}
						<HealthBadge health={prowlarrHealth} />
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
					{#if seerr?.enabled}
						<HealthBadge health={seerrHealth} />
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

			<!-- Bazarr card -->
			<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder('bazarr')}" onclick={() => showBazarrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('bazarr')} alt="Bazarr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-foreground">Bazarr</span>
					</div>
					{#if bazarr?.enabled}
						<HealthBadge health={bazarrHealth} />
					{:else}
						<Badge variant={bazarr?.enabled ? 'success' : 'error'}>
							{bazarr?.enabled ? 'Enabled' : 'Disabled'}
						</Badge>
					{/if}
				</div>
				{#if bazarr?.url}
					<p class="text-xs text-muted-foreground truncate mb-1">{bazarr.url}</p>
				{:else}
					<p class="text-xs text-muted-foreground/50 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-muted-foreground/50">Subtitle manager</p>
			</Card>

			<!-- Kapowarr card -->
			<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder('kapowarr')}" onclick={() => showKapowarrModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('kapowarr')} alt="Kapowarr" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-foreground">Kapowarr</span>
					</div>
					{#if kapowarr?.enabled}
						<HealthBadge health={kapowarrHealth} />
					{:else}
						<Badge variant={kapowarr?.enabled ? 'success' : 'error'}>
							{kapowarr?.enabled ? 'Enabled' : 'Disabled'}
						</Badge>
					{/if}
				</div>
				{#if kapowarr?.url}
					<p class="text-xs text-muted-foreground truncate mb-1">{kapowarr.url}</p>
				{:else}
					<p class="text-xs text-muted-foreground/50 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-muted-foreground/50">Comic book manager</p>
			</Card>

			<!-- Shoko card -->
			<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder('shoko')}" onclick={() => showShokoModal = true}>
				<div class="flex items-start justify-between gap-2 mb-2">
					<div class="flex items-center gap-2">
						<img src={appLogo('shoko')} alt="Shoko" class="w-5 h-5 rounded" />
						<span class="font-medium text-sm text-foreground">Shoko</span>
					</div>
					{#if shoko?.enabled}
						<HealthBadge health={shokoHealth} />
					{:else}
						<Badge variant={shoko?.enabled ? 'success' : 'error'}>
							{shoko?.enabled ? 'Enabled' : 'Disabled'}
						</Badge>
					{/if}
				</div>
				{#if shoko?.url}
					<p class="text-xs text-muted-foreground truncate mb-1">{shoko.url}</p>
				{:else}
					<p class="text-xs text-muted-foreground/50 mb-1">Not configured</p>
				{/if}
				<p class="text-[10px] text-muted-foreground/50">Anime library manager</p>
			</Card>
		</div>
	</section>

	<!-- ── Instance Groups ───────────────────────────────── -->
	<section>
		<div class="flex items-center justify-between mb-3">
			<h3 class="text-sm font-semibold text-foreground">Instance Groups</h3>
			<Button size="sm" variant="ghost" onclick={openAddGroup}>
				<Plus class="h-3.5 w-3.5 mr-1" />
				Add Group
			</Button>
		</div>
		{#if hasAnyGroups}
			<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-3">
				{#each instanceGroups as group}
					{@const modeInfo = groupModes.find(m => m.value === group.mode)}
					<Card class="!p-4 cursor-pointer hover:border-muted-foreground transition-colors border-l-2 {appAccentBorder(group.app_type)}" onclick={() => openEditGroup(group)}>
						<div class="flex items-start justify-between gap-2 mb-2">
							<div class="flex items-center gap-2 min-w-0">
								<Layers class="w-4 h-4 shrink-0 text-muted-foreground" />
								<span class="font-medium text-sm text-foreground truncate">{group.name}</span>
							</div>
							<Badge variant="info">{appDisplayName(group.app_type)}</Badge>
						</div>
						<div class="flex items-center gap-2 mb-2">
							<Badge variant={group.mode === 'quality_hierarchy' ? 'success' : 'default'}>
								{modeInfo?.label ?? group.mode}
							</Badge>
							<span class="text-[10px] text-muted-foreground">{group.members?.length ?? 0} member{(group.members?.length ?? 0) !== 1 ? 's' : ''}</span>
						</div>
						{#if group.members && group.members.length > 0}
							<div class="flex flex-wrap gap-1">
								{#each group.members.sort((a, b) => a.quality_rank - b.quality_rank) as member}
									<span class="text-[10px] px-1.5 py-0.5 rounded bg-muted/50 text-muted-foreground">
										#{member.quality_rank} {member.instance_name ?? member.instance_id.slice(0, 8)}
									</span>
								{/each}
							</div>
						{:else}
							<p class="text-[10px] text-muted-foreground/50 italic">No members — click to configure</p>
						{/if}
					</Card>
				{/each}
			</div>
		{:else}
			<Card class="!p-6 text-center">
				<Layers class="h-8 w-8 text-muted-foreground/40 mx-auto mb-2" />
				<p class="text-sm text-muted-foreground mb-1">No instance groups yet</p>
				<p class="text-xs text-muted-foreground/60">Group multiple instances of the same app to enable cross-instance dedup on the <a href="/dedup" class="underline hover:text-foreground">Dedup page</a>.</p>
			</Card>
		{/if}
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
				{#if prowlarrHealth?.status === 'ok'}
					<Button variant="ghost" onclick={loadProwlarrIndexers} loading={indexersLoading}>Refresh Indexers</Button>
				{/if}
			</div>

			{#if prowlarrIndexers.length > 0}
				<div class="pt-2 border-t border-border">
					<h4 class="text-xs font-semibold text-muted-foreground uppercase tracking-wide mb-2">Indexers ({prowlarrIndexers.length})</h4>
					<ScrollArea.Root class="max-h-48">
						<div class="space-y-1">
							{#each prowlarrIndexers as idx}
							<div class="flex items-center justify-between px-2 py-1 rounded text-sm bg-muted/30">
								<span class="truncate {idx.enable ? 'text-foreground' : 'text-muted-foreground line-through'}">{idx.name}</span>
								<span class="flex items-center gap-2 shrink-0">
									<span class="text-[10px] uppercase text-muted-foreground">{idx.protocol}</span>
									<span class="w-2 h-2 rounded-full {idx.enable ? 'bg-emerald-500' : 'bg-muted-foreground/40'}"></span>
								</span>
							</div>
						{/each}
						</div>
					</ScrollArea.Root>
				</div>
			{/if}
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

<!-- ── Bazarr Settings Modal ─────────────────────────────── -->
<Modal bind:open={showBazarrModal} title="Bazarr Settings" onclose={() => showBazarrModal = false}>
	{#if bazarr}
		<div class="space-y-4">
			<Toggle bind:checked={bazarr.enabled} label="Enabled" color={appBgColor('bazarr')} />
			<Input bind:value={bazarr.url} label="URL" placeholder="http://bazarr:6767" hint="Base URL of your Bazarr instance" />
			<Input bind:value={bazarr.api_key} label="API Key" type="password" hint="Found in Bazarr config.ini under [auth] → apikey" />
			<Input bind:value={bazarr.timeout} type="number" label="Timeout (seconds)" hint="Connection timeout for Bazarr API requests" />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveBazarr} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testBazarr}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground text-center py-4">Loading Bazarr settings...</p>
	{/if}
</Modal>

<!-- ── Kapowarr Settings Modal ────────────────────────────── -->
<Modal bind:open={showKapowarrModal} title="Kapowarr Settings" onclose={() => showKapowarrModal = false}>
	{#if kapowarr}
		<div class="space-y-4">
			<Toggle bind:checked={kapowarr.enabled} label="Enabled" color={appBgColor('kapowarr')} />
			<Input bind:value={kapowarr.url} label="URL" placeholder="http://kapowarr:5656" hint="Base URL of your Kapowarr instance" />
			<Input bind:value={kapowarr.api_key} label="API Key" type="password" hint="Found in Kapowarr Settings → General → API Key" />
			<Input bind:value={kapowarr.timeout} type="number" label="Timeout (seconds)" hint="Connection timeout for Kapowarr API requests" />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveKapowarr} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testKapowarr}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground text-center py-4">Loading Kapowarr settings...</p>
	{/if}
</Modal>

<!-- ── Shoko Settings Modal ───────────────────────────────── -->
<Modal bind:open={showShokoModal} title="Shoko Settings" onclose={() => showShokoModal = false}>
	{#if shoko}
		<div class="space-y-4">
			<Toggle bind:checked={shoko.enabled} label="Enabled" color={appBgColor('shoko')} />
			<Input bind:value={shoko.url} label="URL" placeholder="http://shoko:8111" hint="Base URL of your Shoko Server instance" />
			<Input bind:value={shoko.api_key} label="API Key" type="password" hint="Generate in Shoko Desktop → Settings → API Keys" />
			<Input bind:value={shoko.timeout} type="number" label="Timeout (seconds)" hint="Connection timeout for Shoko API requests" />
			<div class="flex gap-2 pt-2">
				<Button onclick={saveShoko} loading={saving}>Save</Button>
				<Button variant="secondary" onclick={testShoko}>Test Connection</Button>
			</div>
		</div>
	{:else}
		<p class="text-sm text-muted-foreground text-center py-4">Loading Shoko settings...</p>
	{/if}
</Modal>

<!-- ── Instance Group Modal ──────────────────────────────── -->
<Modal bind:open={showGroupModal} title={editingGroup ? `Edit Group: ${editingGroup.name}` : 'Create Instance Group'} onclose={() => showGroupModal = false}>
	<form onsubmit={(e: Event) => { e.preventDefault(); saveGroup(); }} class="space-y-4">
		<Input bind:value={groupForm.name} label="Group Name" placeholder="e.g. Movies Quality Stack" />
		<Select bind:value={groupForm.app_type} label="App Type" disabled={!!editingGroup}>
			{#each groupableAppTypes as app}
				<option value={app}>{appDisplayName(app)}</option>
			{/each}
		</Select>
		<Select bind:value={groupForm.mode} label="Mode">
			{#each groupModes as mode}
				<option value={mode.value}>{mode.label}</option>
			{/each}
		</Select>
		<p class="text-xs text-muted-foreground">{groupModes.find(m => m.value === groupForm.mode)?.description}</p>

		<!-- Members -->
		<div class="space-y-2">
			<div class="flex items-center justify-between">
				<span class="text-sm font-medium text-foreground">Members</span>
				<Button type="button" size="sm" variant="ghost" onclick={addMemberToGroup} disabled={getInstancesForGroupApp().length <= groupMembers.length}>
					<Plus class="h-3.5 w-3.5 mr-1" />Add Instance
				</Button>
			</div>
			{#if groupMembers.length === 0}
				<p class="text-xs text-muted-foreground/60 italic py-2">No members yet. Add instances that should be grouped together.</p>
			{/if}
			{#each groupMembers as member, idx}
				{@const appInstances = getInstancesForGroupApp()}
				<div class="flex items-center gap-2 p-2 rounded-md bg-muted/30 border border-border">
					<GripVertical class="h-4 w-4 text-muted-foreground/40 shrink-0" />
					<span class="text-xs font-mono text-muted-foreground shrink-0 w-6">#{member.quality_rank}</span>
					<Select bind:value={member.instance_id} class="flex-1">
						{#each appInstances as inst}
							<option value={inst.id} disabled={groupMembers.some(m => m.instance_id === inst.id && m !== member)}>
								{inst.name}
							</option>
						{/each}
					</Select>
					<Checkbox bind:checked={member.is_independent} label="Indie" class="shrink-0" />
					<Button type="button" size="icon" variant="ghost" class="h-auto w-auto p-1 text-muted-foreground hover:text-destructive" onclick={() => removeMember(idx)}>
						<Trash2 class="h-3.5 w-3.5" />
					</Button>
				</div>
			{/each}
			{#if groupMembers.length > 0 && groupForm.mode === 'quality_hierarchy'}
				<p class="text-[10px] text-muted-foreground">Rank 1 = highest quality. Lower ranks are considered duplicates.</p>
			{/if}
		</div>

		<div class="flex items-center gap-2 pt-2">
			<Button type="submit" loading={savingGroup}>{editingGroup ? 'Update Group' : 'Create Group'}</Button>
			{#if editingGroup}
				<div class="ml-auto">
					{#if deleteGroupConfirm === editingGroup.id}
						<Button size="sm" variant="danger" onclick={() => { if (editingGroup) deleteGroup(editingGroup.id); }}>Confirm Delete</Button>
						<Button size="sm" variant="ghost" onclick={() => deleteGroupConfirm = null}>Cancel</Button>
					{:else}
						<Button size="sm" variant="danger" onclick={() => { if (editingGroup) deleteGroupConfirm = editingGroup.id; }}>Delete</Button>
					{/if}
				</div>
			{/if}
		</div>
	</form>
</Modal>
