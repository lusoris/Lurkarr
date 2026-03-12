<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';

	const toasts = getToasts();

	interface NotificationProvider {
		id: string;
		type: string;
		name: string;
		enabled: boolean;
		config: Record<string, string>;
		events: string[];
		created_at: string;
		updated_at: string;
	}

	const providerTypes = [
		{ value: 'discord', label: 'Discord', fields: ['webhook_url'] },
		{ value: 'telegram', label: 'Telegram', fields: ['bot_token', 'chat_id'] },
		{ value: 'pushover', label: 'Pushover', fields: ['user_key', 'api_token'] },
		{ value: 'gotify', label: 'Gotify', fields: ['url', 'token'] },
		{ value: 'ntfy', label: 'ntfy', fields: ['url', 'topic', 'token'] },
		{ value: 'apprise', label: 'Apprise', fields: ['url'] },
		{ value: 'email', label: 'Email', fields: ['smtp_host', 'smtp_port', 'username', 'password', 'from', 'to'] },
		{ value: 'webhook', label: 'Webhook', fields: ['url', 'method', 'headers'] }
	] as const;

	const allEvents = [
		'lurk.started', 'lurk.completed', 'lurk.error',
		'queue.blocklisted', 'queue.stalled', 'queue.imported',
		'health.degraded', 'health.restored',
		'system.startup', 'system.shutdown'
	];

	let providers = $state<NotificationProvider[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editing = $state<NotificationProvider | null>(null);
	let saving = $state(false);
	let testing = $state<string | null>(null);
	let deleteConfirm = $state<string | null>(null);

	// Form state
	let formType = $state('discord');
	let formName = $state('');
	let formEnabled = $state(true);
	let formConfig = $state<Record<string, string>>({});
	let formEvents = $state<string[]>([]);

	const selectedProvider = $derived(providerTypes.find(p => p.value === formType));

	const fieldLabels: Record<string, string> = {
		webhook_url: 'Webhook URL',
		bot_token: 'Bot Token',
		chat_id: 'Chat ID',
		user_key: 'User Key',
		api_token: 'API Token',
		url: 'URL',
		token: 'Token',
		topic: 'Topic',
		smtp_host: 'SMTP Host',
		smtp_port: 'SMTP Port',
		username: 'Username',
		password: 'Password',
		from: 'From Address',
		to: 'To Address',
		method: 'HTTP Method',
		headers: 'Headers (JSON)'
	};

	const sensitiveFields = new Set(['bot_token', 'api_token', 'token', 'password', 'webhook_url']);

	async function load() {
		loading = true;
		try {
			providers = await api.get<NotificationProvider[]>('/notifications/providers');
		} catch {
			toasts.error('Failed to load notification providers');
		}
		loading = false;
	}

	function openAdd() {
		editing = null;
		formType = 'discord';
		formName = '';
		formEnabled = true;
		formConfig = {};
		formEvents = [...allEvents];
		showModal = true;
	}

	function openEdit(p: NotificationProvider) {
		editing = p;
		formType = p.type;
		formName = p.name;
		formEnabled = p.enabled;
		formConfig = { ...p.config };
		formEvents = [...p.events];
		showModal = true;
	}

	function toggleEvent(event: string) {
		if (formEvents.includes(event)) {
			formEvents = formEvents.filter(e => e !== event);
		} else {
			formEvents = [...formEvents, event];
		}
	}

	async function save() {
		saving = true;
		try {
			const body = {
				type: formType,
				name: formName,
				enabled: formEnabled,
				config: formConfig,
				events: formEvents
			};

			if (editing) {
				await api.put(`/notifications/providers/${editing.id}`, body);
				toasts.success('Provider updated');
			} else {
				await api.post('/notifications/providers', body);
				toasts.success('Provider created');
			}
			showModal = false;
			await load();
		} catch {
			toasts.error(editing ? 'Failed to update provider' : 'Failed to create provider');
		}
		saving = false;
	}

	async function testProvider(id: string) {
		testing = id;
		try {
			await api.post(`/notifications/providers/${id}/test`);
			toasts.success('Test notification sent');
		} catch {
			toasts.error('Test notification failed');
		}
		testing = null;
	}

	async function deleteProvider(id: string) {
		try {
			await api.del(`/notifications/providers/${id}`);
			toasts.success('Provider deleted');
			deleteConfirm = null;
			await load();
		} catch {
			toasts.error('Failed to delete provider');
		}
	}

	function providerLabel(type: string): string {
		return providerTypes.find(p => p.value === type)?.label ?? type;
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Notifications - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-surface-50">Notifications</h1>
		<Button size="sm" onclick={openAdd}>Add Provider</Button>
	</div>

	{#if loading}
		<div class="space-y-4">
			{#each Array(3) as _}
				<div class="h-20 rounded-xl bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if providers.length === 0}
		<Card>
			<div class="text-center py-8">
				<svg class="w-12 h-12 mx-auto text-surface-600 mb-3" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
					<path stroke-linecap="round" stroke-linejoin="round" d="M14.857 17.082a23.848 23.848 0 005.454-1.31A8.967 8.967 0 0118 9.75V9A6 6 0 006 9v.75a8.967 8.967 0 01-2.312 6.022c1.733.64 3.56 1.085 5.455 1.31m5.714 0a24.255 24.255 0 01-5.714 0m5.714 0a3 3 0 11-5.714 0" />
				</svg>
				<p class="text-sm text-surface-500">No notification providers configured</p>
				<Button size="sm" class="mt-4" onclick={openAdd}>Add Your First Provider</Button>
			</div>
		</Card>
	{:else}
		<div class="space-y-3">
			{#each providers as provider (provider.id)}
				<Card>
					<div class="flex flex-col sm:flex-row sm:items-center gap-3">
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-1">
								<span class="font-medium text-surface-100 truncate">{provider.name}</span>
								<Badge variant={provider.enabled ? 'success' : 'default'}>
									{provider.enabled ? 'Active' : 'Disabled'}
								</Badge>
								<Badge>{providerLabel(provider.type)}</Badge>
							</div>
							<p class="text-xs text-surface-500">
								{provider.events.length} event{provider.events.length !== 1 ? 's' : ''} subscribed
							</p>
						</div>
						<div class="flex items-center gap-2 shrink-0">
							{#if deleteConfirm === provider.id}
								<span class="text-xs text-red-400 mr-1">Delete?</span>
								<Button size="sm" variant="danger" onclick={() => deleteProvider(provider.id)}>Yes</Button>
								<Button size="sm" variant="ghost" onclick={() => deleteConfirm = null}>No</Button>
							{:else}
								<Button size="sm" variant="secondary" loading={testing === provider.id} onclick={() => testProvider(provider.id)}>Test</Button>
								<Button size="sm" variant="ghost" onclick={() => openEdit(provider)}>Edit</Button>
								<Button size="sm" variant="ghost" onclick={() => deleteConfirm = provider.id}>
									<svg class="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
										<path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0" />
									</svg>
								</Button>
							{/if}
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Add/Edit Modal -->
<Modal bind:open={showModal} title={editing ? 'Edit Provider' : 'Add Provider'} onclose={() => showModal = false}>
	<div class="space-y-4">
		{#if !editing}
			<label class="block">
				<span class="block text-sm font-medium text-surface-300 mb-1.5">Provider Type</span>
				<select
					bind:value={formType}
					onchange={() => formConfig = {}}
					class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500"
				>
					{#each providerTypes as pt}
						<option value={pt.value}>{pt.label}</option>
					{/each}
				</select>
			</label>
		{/if}

		<Input bind:value={formName} label="Name" placeholder="My Discord webhook" />
		<Toggle bind:checked={formEnabled} label="Enabled" />

		<!-- Provider-specific config fields -->
		{#if selectedProvider}
			<div class="border-t border-surface-800 pt-4">
				<h3 class="text-sm font-medium text-surface-300 mb-3">Configuration</h3>
				<div class="space-y-3">
					{#each selectedProvider.fields as field}
						<Input
							bind:value={formConfig[field]}
							label={fieldLabels[field] ?? field}
							type={sensitiveFields.has(field) ? 'password' : 'text'}
							placeholder={field === 'method' ? 'POST' : ''}
						/>
					{/each}
				</div>
			</div>
		{/if}

		<!-- Events -->
		<div class="border-t border-surface-800 pt-4">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-sm font-medium text-surface-300">Events</h3>
				<button
					onclick={() => formEvents = formEvents.length === allEvents.length ? [] : [...allEvents]}
					class="text-xs text-lurk-400 hover:text-lurk-300"
				>
					{formEvents.length === allEvents.length ? 'Deselect all' : 'Select all'}
				</button>
			</div>
			<div class="grid grid-cols-1 sm:grid-cols-2 gap-2">
				{#each allEvents as event}
					<label class="flex items-center gap-2 rounded-lg px-3 py-2 text-sm cursor-pointer transition-colors hover:bg-surface-800 {formEvents.includes(event) ? 'text-surface-100' : 'text-surface-500'}">
						<input type="checkbox" checked={formEvents.includes(event)} onchange={() => toggleEvent(event)} class="rounded border-surface-600 text-lurk-600 focus:ring-lurk-500 bg-surface-800" />
						<span class="font-mono text-xs">{event}</span>
					</label>
				{/each}
			</div>
		</div>

		<div class="flex justify-end gap-2 pt-2">
			<Button variant="secondary" onclick={() => showModal = false}>Cancel</Button>
			<Button onclick={save} loading={saving}>{editing ? 'Update' : 'Create'}</Button>
		</div>
	</div>
</Modal>
