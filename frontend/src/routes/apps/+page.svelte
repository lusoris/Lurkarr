<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';

	const toasts = getToasts();

	interface AppInstance {
		id: string;
		app_type: string;
		name: string;
		api_url: string;
		api_key: string;
		enabled: boolean;
	}

	const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;
	type AppType = (typeof appTypes)[number];

	let instances = $state<Record<string, AppInstance[]>>({});
	let loading = $state(true);

	// Modal state
	let showModal = $state(false);
	let editingInstance = $state<AppInstance | null>(null);
	let modalApp = $state<AppType>('sonarr');
	let form = $state({ name: '', api_url: '', api_key: '', enabled: true });
	let saving = $state(false);
	let testing = $state(false);
	let deleteConfirm = $state<string | null>(null);

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
	}

	function openAdd(app: AppType) {
		editingInstance = null;
		modalApp = app;
		form = { name: '', api_url: '', api_key: '', enabled: true };
		showModal = true;
	}

	function openEdit(inst: AppInstance) {
		editingInstance = inst;
		modalApp = inst.app_type as AppType;
		form = { name: inst.name, api_url: inst.api_url, api_key: '', enabled: inst.enabled };
		showModal = true;
	}

	async function saveInstance() {
		saving = true;
		try {
			if (editingInstance) {
				await api.put(`/instances/${editingInstance.id}`, {
					name: form.name,
					api_url: form.api_url,
					api_key: form.api_key,
					enabled: form.enabled
				});
				toasts.success('Instance updated');
			} else {
				await api.post(`/instances/${modalApp}`, {
					name: form.name,
					api_url: form.api_url,
					api_key: form.api_key
				});
				toasts.success('Instance added');
			}
			showModal = false;
			await loadAll();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
		saving = false;
	}

	async function testConnection() {
		testing = true;
		try {
			const res = await api.post<{ status: string; app: string; version: string }>('/instances/test', {
				api_url: form.api_url,
				api_key: form.api_key || editingInstance?.api_key,
				app_type: modalApp
			});
			toasts.success(`Connected — ${res.app} v${res.version}`);
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Connection failed');
		}
		testing = false;
	}

	async function deleteInstance(id: string) {
		try {
			await api.del(`/instances/${id}`);
			toasts.success('Instance deleted');
			deleteConfirm = null;
			await loadAll();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to delete');
		}
	}

	$effect(() => { loadAll(); });
</script>

<svelte:head><title>Apps - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">App Instances</h1>

	{#if loading}
		<div class="space-y-4">
			{#each Array(3) as _}
				<div class="h-24 rounded-xl bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else}
		{#each appTypes as app}
			{@const appInstances = instances[app] ?? []}
			<div>
				<div class="flex items-center justify-between mb-3">
					<h2 class="text-lg font-semibold text-surface-200 capitalize">{app}</h2>
					<Button size="sm" onclick={() => openAdd(app)}>+ Add</Button>
				</div>
				{#if appInstances.length === 0}
					<Card>
						<p class="text-sm text-surface-500">No instances configured</p>
					</Card>
				{:else}
					<div class="space-y-2">
						{#each appInstances as inst}
						<Card class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
							<div class="flex items-center gap-3 min-w-0 flex-wrap">
									<div class="min-w-0">
										<span class="font-medium text-surface-100">{inst.name}</span>
										<span class="text-xs text-surface-500 ml-2 truncate">{inst.api_url}</span>
									</div>
								</div>
								<div class="flex items-center gap-2 shrink-0">
									<Badge variant={inst.enabled ? 'success' : 'error'}>
										{inst.enabled ? 'Enabled' : 'Disabled'}
									</Badge>
									<Button size="sm" variant="ghost" onclick={() => openEdit(inst)}>Edit</Button>
									{#if deleteConfirm === inst.id}
										<Button size="sm" variant="danger" onclick={() => deleteInstance(inst.id)}>Confirm</Button>
										<Button size="sm" variant="ghost" onclick={() => deleteConfirm = null}>Cancel</Button>
									{:else}
										<Button size="sm" variant="danger" onclick={() => deleteConfirm = inst.id}>Delete</Button>
									{/if}
								</div>
							</Card>
						{/each}
					</div>
				{/if}
			</div>
		{/each}
	{/if}
</div>

<!-- Add/Edit Instance Modal -->
<Modal bind:open={showModal} title={editingInstance ? `Edit ${modalApp} Instance` : `Add ${modalApp} Instance`} onclose={() => showModal = false}>
	<form onsubmit={saveInstance} class="space-y-4">
		<Input bind:value={form.name} label="Name" placeholder="My Sonarr" />
		<Input bind:value={form.api_url} label="URL" placeholder="http://sonarr:8989" />
		<Input bind:value={form.api_key} type="password" label={editingInstance ? 'API Key (leave empty to keep current)' : 'API Key'} />
		{#if editingInstance}
			<Toggle bind:checked={form.enabled} label="Enabled" />
		{/if}
		<div class="flex gap-2 pt-2">
			<Button type="submit" loading={saving}>{editingInstance ? 'Update' : 'Add'}</Button>
			<Button variant="secondary" loading={testing} onclick={testConnection}>Test Connection</Button>
		</div>
	</form>
</Modal>
