<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import DataTable from '$lib/components/ui/DataTable.svelte';
	import { Bell, Plus, Trash2 } from 'lucide-svelte';

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

	const eventGroups = [
		{ label: 'Lurking', events: ['lurk.started', 'lurk.completed', 'lurk.error'] },
		{ label: 'Queue', events: ['queue.blocklisted', 'queue.stalled', 'queue.imported'] },
		{ label: 'Health', events: ['health.degraded', 'health.restored'] },
		{ label: 'System', events: ['system.startup', 'system.shutdown'] }
	] as const;

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
	let formTitleTemplate = $state('');
	let formBodyTemplate = $state('');

	const selectedProvider = $derived(providerTypes.find(p => p.value === formType));

	// Reset config fields when provider type changes
	let prevFormType = $state(formType);
	$effect(() => {
		if (formType !== prevFormType) {
			prevFormType = formType;
			const fields = providerTypes.find(p => p.value === formType)?.fields ?? [];
			formConfig = Object.fromEntries(fields.map(f => [f, '']));
		}
	});

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

	const fieldPlaceholders: Record<string, string> = {
		webhook_url: 'https://discord.com/api/webhooks/...',
		bot_token: '123456:ABC-DEF...',
		chat_id: '-1001234567890',
		user_key: 'uQiRzpo4DXghDmr9QzzfQu27cmVRsG',
		api_token: 'azGDORePK8gMaC0QOYAMyEEuzJnyUi',
		url: 'http://hostname:port',
		token: 'tk_...',
		topic: 'lurkarr',
		smtp_host: 'smtp.gmail.com',
		smtp_port: '587',
		username: 'user@example.com',
		password: '',
		from: 'lurkarr@example.com',
		to: 'you@example.com',
		method: 'POST',
		headers: '{"Content-Type": "application/json"}'
	};

	const fieldHints: Record<string, string> = {
		webhook_url: 'From channel settings → Integrations → Webhooks',
		bot_token: 'From @BotFather on Telegram',
		chat_id: 'User, group, or channel ID',
		topic: 'ntfy topic name to publish to',
		smtp_port: '587 for STARTTLS, 465 for SSL',
	};

	async function load() {
		loading = true;
		try {
			providers = await api.get<NotificationProvider[]>('/notifications/providers');
		} catch {
			providers = [];
		}
		loading = false;
	}

	function openAdd() {
		editing = null;
		formType = 'discord';
		formName = '';
		formEnabled = true;
		const defaultFields = providerTypes.find(p => p.value === 'discord')?.fields ?? [];
		formConfig = Object.fromEntries(defaultFields.map(f => [f, '']));
		formEvents = [...allEvents];
		formTitleTemplate = '';
		formBodyTemplate = '';
		showModal = true;
	}

	function openEdit(p: NotificationProvider) {
		editing = p;
		formType = p.type;
		formName = p.name;
		formEnabled = p.enabled;
		const fields = providerTypes.find(pt => pt.value === p.type)?.fields ?? [];
		formConfig = Object.fromEntries(fields.map(f => [f, p.config[f] ?? '']));
		formEvents = [...p.events];
		formTitleTemplate = p.config['title_template'] ?? '';
		formBodyTemplate = p.config['body_template'] ?? '';
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
			const config: Record<string, string> = { ...formConfig };
			if (formTitleTemplate.trim()) config['title_template'] = formTitleTemplate.trim();
			if (formBodyTemplate.trim()) config['body_template'] = formBodyTemplate.trim();

			const body = {
				type: formType,
				name: formName,
				enabled: formEnabled,
				config,
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

	// Tab state
	type ActiveTab = 'providers' | 'history';
	let activeTab = $state<ActiveTab>('providers');

	// History state
	interface HistoryEntry {
		id: string;
		provider_type: string;
		provider_name: string;
		event_type: string;
		title: string;
		message: string;
		app_type: string;
		instance: string;
		status: string;
		error: string;
		duration_ms: number;
		created_at: string;
	}

	let historyItems = $state<HistoryEntry[]>([]);
	let historyLoading = $state(false);

	async function loadHistory() {
		historyLoading = true;
		try {
			historyItems = await api.get<HistoryEntry[]>('/notifications/history');
		} catch {
			historyItems = [];
		}
		historyLoading = false;
	}

	function statusVariant(status: string): 'success' | 'error' {
		return status === 'sent' ? 'success' : 'error';
	}

	const historyColumns = [
		{ key: 'title' as const, label: 'Title', sortable: true },
		{ key: 'provider_name' as const, label: 'Provider', sortable: true },
		{ key: 'event_type' as const, label: 'Event', sortable: true },
		{ key: 'status' as const, label: 'Status', sortable: true },
		{ key: 'created_at' as const, label: 'Date', sortable: true }
	];

	function onTabChange(tab: string) {
		activeTab = tab as ActiveTab;
		if (tab === 'history' && historyItems.length === 0) {
			loadHistory();
		}
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Notifications - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Notifications" description="Configure notification providers and event subscriptions.">
		{#snippet actions()}
			{#if activeTab === 'providers'}
				<Button size="sm" onclick={openAdd}>
					<Plus class="h-4 w-4" />
					Add Provider
				</Button>
			{/if}
		{/snippet}
	</PageHeader>

	<Tabs
		tabs={[
			{ value: 'providers', label: 'Providers' },
			{ value: 'history', label: 'History' }
		]}
		bind:value={activeTab}
		onchange={onTabChange}
	/>

	{#if activeTab === 'providers'}
	{#if loading}
		<Skeleton rows={3} height="h-20" />
	{:else if providers.length === 0}
		<EmptyState icon={Bell} title="No notification providers" description="Add your first notification provider to start receiving alerts.">
			{#snippet actions()}
				<Button size="sm" onclick={openAdd}>
					<Plus class="h-4 w-4" />
					Add Your First Provider
				</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<div class="space-y-3">
			{#each providers as provider (provider.id)}
				<Card>
					<div class="flex flex-col sm:flex-row sm:items-center gap-3">
						<div class="flex-1 min-w-0">
							<div class="flex items-center gap-2 mb-1">
								<span class="font-medium text-foreground truncate">{provider.name}</span>
								<Badge variant={provider.enabled ? 'success' : 'default'}>
									{provider.enabled ? 'Active' : 'Disabled'}
								</Badge>
								<Badge>{providerLabel(provider.type)}</Badge>
							</div>
							<p class="text-xs text-muted-foreground">
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
									<Trash2 class="w-4 h-4 text-red-400" />
								</Button>
							{/if}
						</div>
					</div>
				</Card>
			{/each}
		</div>
	{/if}
	{/if}

	{#if activeTab === 'history'}
		{#if historyLoading}
			<Skeleton rows={5} height="h-10" />
		{:else if historyItems.length === 0}
			<EmptyState icon={Bell} title="No notification history" description="Notification delivery attempts will appear here as they happen." />
		{:else}
			<DataTable data={historyItems} columns={historyColumns}>
				{#snippet row(item: HistoryEntry)}
					<td class="px-4 py-2 text-sm truncate max-w-[240px]" title={item.title}>{item.title}</td>
					<td class="px-4 py-2 text-sm">
						<Badge>{item.provider_name || item.provider_type}</Badge>
					</td>
					<td class="px-4 py-2 text-sm font-mono text-xs">{item.event_type}</td>
					<td class="px-4 py-2 text-sm">
						<Badge variant={statusVariant(item.status)}>{item.status}</Badge>
						{#if item.error}
							<span class="block text-[10px] text-red-400 mt-0.5 truncate max-w-[200px]" title={item.error}>{item.error}</span>
						{/if}
					</td>
					<td class="px-4 py-2 text-xs text-muted-foreground whitespace-nowrap">
						{new Date(item.created_at).toLocaleString()}
					</td>
				{/snippet}
			</DataTable>
		{/if}
	{/if}
</div>

<!-- Add/Edit Modal -->
<Modal bind:open={showModal} title={editing ? 'Edit Provider' : 'Add Provider'} onclose={() => showModal = false}>
	<div class="space-y-4">
		{#if !editing}
			<Select bind:value={formType} label="Provider Type">
				{#each providerTypes as pt}
					<option value={pt.value}>{pt.label}</option>
				{/each}
			</Select>
		{/if}

		<Input bind:value={formName} label="Name" placeholder="My {providerLabel(formType)} notifications" />
		<Toggle bind:checked={formEnabled} label="Enabled" />

		<!-- Provider-specific config fields -->
		{#if selectedProvider}
			<div class="border-t border-border pt-4">
				<h3 class="text-sm font-medium text-muted-foreground mb-3">Configuration</h3>
				<div class="space-y-3">
					{#if formType === 'email'}
						<!-- Email: group SMTP connection and addresses -->
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							<Input bind:value={formConfig['smtp_host']} label="SMTP Host" />
							<Input bind:value={formConfig['smtp_port']} label="SMTP Port" />
						</div>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							<Input bind:value={formConfig['username']} label="Username" />
							<Input bind:value={formConfig['password']} label="Password" type="password" />
						</div>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
							<Input bind:value={formConfig['from']} label="From Address" />
							<Input bind:value={formConfig['to']} label="To Address" />
						</div>
					{:else}
						{#each selectedProvider.fields as field}
							<Input
								bind:value={formConfig[field]}
								label={fieldLabels[field] ?? field}
								type={sensitiveFields.has(field) ? 'password' : 'text'}
								placeholder={fieldPlaceholders[field] ?? ''}
								hint={fieldHints[field] ?? ''}
							/>
						{/each}
					{/if}
				</div>
			</div>
		{/if}

		<!-- Templates -->
		<div class="border-t border-border pt-4">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-sm font-medium text-muted-foreground">Templates</h3>
			</div>
			<p class="text-xs text-muted-foreground mb-3">
				Customise notification text with Go templates. Available: {'{{.Title}}'}, {'{{.Message}}'}, {'{{.AppType}}'}, {'{{.Instance}}'}, {'{{.Type}}'}, {'{{index .Fields "key"}}'}.
			</p>
			<div class="space-y-3">
				<div>
					<label for="title-tpl" class="block text-sm font-medium text-foreground mb-1">Title Template</label>
					<input id="title-tpl" bind:value={formTitleTemplate} placeholder="Leave blank for default" class="w-full rounded-md bg-muted border border-border px-3 py-2 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring" />
				</div>
				<div>
					<label for="body-tpl" class="block text-sm font-medium text-foreground mb-1">Body Template</label>
					<textarea id="body-tpl" bind:value={formBodyTemplate} placeholder="Leave blank for default" rows={3} class="w-full rounded-md bg-muted border border-border px-3 py-2 text-sm font-mono text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring resize-y"></textarea>
				</div>
			</div>
		</div>

		<!-- Events -->
		<div class="border-t border-border pt-4">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-sm font-medium text-muted-foreground">Events</h3>
				<button
					onclick={() => formEvents = formEvents.length === allEvents.length ? [] : [...allEvents]}
					class="text-xs text-primary hover:text-primary/80"
				>
					{formEvents.length === allEvents.length ? 'Deselect all' : 'Select all'}
				</button>
			</div>
			<div class="space-y-3">
				{#each eventGroups as group}
					<div>
						<span class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-1 block">{group.label}</span>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-1">
							{#each group.events as event}
								<label class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm cursor-pointer transition-colors hover:bg-muted {formEvents.includes(event) ? 'text-foreground' : 'text-muted-foreground'}">
									<input type="checkbox" checked={formEvents.includes(event)} onchange={() => toggleEvent(event)} class="rounded border-border text-primary focus:ring-ring bg-muted" />
									<span class="font-mono text-xs">{event}</span>
								</label>
							{/each}
						</div>
					</div>
				{/each}
			</div>
		</div>

		<div class="flex justify-end gap-2 pt-2">
			<Button variant="secondary" onclick={() => showModal = false}>Cancel</Button>
			<Button onclick={save} loading={saving}>{editing ? 'Update' : 'Create'}</Button>
		</div>
	</div>
</Modal>
