<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import Tabs from '$lib/components/ui/Tabs.svelte';
	import DataTable from '$lib/components/ui/DataTable.svelte';
	import Checkbox from '$lib/components/ui/Checkbox.svelte';
	import ConfirmAction from '$lib/components/ui/ConfirmAction.svelte';
	import { Textarea } from '$lib/components/ui/textarea';
	import { Label } from '$lib/components/ui/label';
	import { Bell, Plus, Trash2 } from 'lucide-svelte';
	import type { NotificationProvider } from '$lib/types';

	const toasts = getToasts();

	const providerTypes = [
		{ value: 'discord', label: 'Discord', fields: ['webhook_url'] },
		{ value: 'telegram', label: 'Telegram', fields: ['bot_token', 'chat_id'] },
		{ value: 'pushover', label: 'Pushover', fields: ['user_key', 'api_token'] },
		{ value: 'gotify', label: 'Gotify', fields: ['server_url', 'app_token'] },
		{ value: 'ntfy', label: 'ntfy', fields: ['server_url', 'topic', 'token'] },
		{ value: 'apprise', label: 'Apprise', fields: ['server_url', 'urls', 'tag'] },
		{ value: 'email', label: 'Email', fields: ['host', 'port', 'username', 'password', 'from', 'to'] },
		{ value: 'webhook', label: 'Webhook', fields: ['url', 'method', 'headers'] }
	] as const;

	const allEvents = [
		'lurk_started', 'lurk_completed',
		'queue_item_removed', 'download_stuck',
		'scheduler_action', 'error'
	];

	const eventGroups = [
		{ label: 'Lurking', events: ['lurk_started', 'lurk_completed'] },
		{ label: 'Queue', events: ['queue_item_removed', 'download_stuck'] },
		{ label: 'System', events: ['scheduler_action', 'error'] }
	] as const;

	let providers = $state<NotificationProvider[]>([]);
	let loading = $state(true);
	let showModal = $state(false);
	let editing = $state<NotificationProvider | null>(null);
	let saving = $state(false);
	let testing = $state<string | null>(null);

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
		app_token: 'App Token',
		server_url: 'Server URL',
		url: 'URL',
		token: 'Token',
		topic: 'Topic',
		urls: 'Notification URLs',
		tag: 'Tag',
		host: 'SMTP Host',
		port: 'SMTP Port',
		username: 'Username',
		password: 'Password',
		from: 'From Address',
		to: 'To Address(es)',
		method: 'HTTP Method',
		headers: 'Headers (JSON)'
	};

	const sensitiveFields = new Set(['bot_token', 'api_token', 'app_token', 'token', 'password', 'webhook_url']);

	const fieldPlaceholders: Record<string, string> = {
		webhook_url: 'https://discord.com/api/webhooks/...',
		bot_token: '123456:ABC-DEF...',
		chat_id: '-1001234567890',
		user_key: 'uQiRzpo4DXghDmr9QzzfQu27cmVRsG',
		api_token: 'azGDORePK8gMaC0QOYAMyEEuzJnyUi',
		app_token: 'AKW3F4_...',
		server_url: 'http://hostname:port',
		url: 'https://your-endpoint.example.com/webhook',
		token: 'tk_...',
		topic: 'lurkarr',
		urls: 'mailto://user:pass@gmail.com, slack://token/channel',
		tag: 'all',
		host: 'smtp.gmail.com',
		port: '587',
		username: 'user@example.com',
		password: '',
		from: 'lurkarr@example.com',
		to: 'you@example.com, admin@example.com',
		method: 'POST',
		headers: '{"Content-Type": "application/json"}'
	};

	const fieldHints: Record<string, string> = {
		webhook_url: 'From channel settings → Integrations → Webhooks',
		bot_token: 'From @BotFather on Telegram',
		chat_id: 'User, group, or channel ID',
		topic: 'ntfy topic name to publish to',
		urls: 'Comma-separated Apprise notification URLs',
		tag: 'Optional Apprise tag filter',
		port: '587 for STARTTLS, 465 for SSL',
		to: 'Comma-separated recipient addresses',
		headers: 'JSON object of custom HTTP headers',
	};

	// Template presets per provider type.
	const templatePresets: Record<string, { label: string; title: string; body: string }[]> = {
		_default: [
			{ label: 'None (use default)', title: '', body: '' },
			{ label: 'Detailed', title: '[{{.AppType}}] {{.Title}}', body: '{{.Message}}\nInstance: {{.Instance}}' },
			{ label: 'Compact', title: '{{.Title}}', body: '{{.AppType}}/{{.Instance}}: {{.Message}}' },
		],
		discord: [
			{ label: 'None (use default)', title: '', body: '' },
			{ label: 'Detailed', title: '[{{.AppType}}] {{.Title}}', body: '{{.Message}}\n\n**Instance:** {{.Instance}}' },
			{ label: 'Compact', title: '{{.Title}}', body: '`{{.AppType}}`/`{{.Instance}}` — {{.Message}}' },
		],
		telegram: [
			{ label: 'None (use default)', title: '', body: '' },
			{ label: 'Detailed', title: '[{{.AppType}}] {{.Title}}', body: '<b>{{.Title}}</b>\n{{.Message}}\n<i>Instance:</i> {{.Instance}}' },
			{ label: 'Compact', title: '{{.Title}}', body: '<code>{{.AppType}}</code>/{{.Instance}}: {{.Message}}' },
		],
		email: [
			{ label: 'None (use default)', title: '', body: '' },
			{ label: 'Detailed', title: '[Lurkarr] {{.AppType}} — {{.Title}}', body: '{{.Title}}\n\n{{.Message}}\n\nApp: {{.AppType}}\nInstance: {{.Instance}}\nEvent: {{.Type}}' },
			{ label: 'Compact', title: '[Lurkarr] {{.Title}}', body: '{{.AppType}}/{{.Instance}}: {{.Message}}' },
		],
	};

	function getPresets() {
		return templatePresets[formType] ?? templatePresets['_default'];
	}

	function applyPreset(idx: number) {
		const preset = getPresets()[idx];
		if (preset) {
			formTitleTemplate = preset.title;
			formBodyTemplate = preset.body;
		}
	}

	const eventLabels: Record<string, string> = {
		lurk_started: 'Lurk Started',
		lurk_completed: 'Lurk Completed',
		queue_item_removed: 'Queue Item Removed',
		download_stuck: 'Download Stuck',
		scheduler_action: 'Scheduler Action',
		error: 'Error',
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
		formConfig = Object.fromEntries(fields.map(f => {
			const v = p.config[f];
			// Convert arrays back to comma-separated strings for form inputs
			if (Array.isArray(v)) return [f, v.join(', ')];
			return [f, v != null ? String(v) : ''];
		}));
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
			const config: Record<string, any> = { ...formConfig };
			if (formTitleTemplate.trim()) config['title_template'] = formTitleTemplate.trim();
			if (formBodyTemplate.trim()) config['body_template'] = formBodyTemplate.trim();

			// Convert types to match backend expectations
			if (formType === 'email') {
				if (typeof config['to'] === 'string') {
					config['to'] = (config['to'] as string).split(',').map((s: string) => s.trim()).filter(Boolean);
				}
				if (config['port']) config['port'] = Number(config['port']);
			}
			if (formType === 'apprise' && typeof config['urls'] === 'string') {
				config['urls'] = (config['urls'] as string).split(',').map((s: string) => s.trim()).filter(Boolean);
			}
			if (formType === 'webhook' && typeof config['headers'] === 'string' && config['headers'].trim()) {
				try { config['headers'] = JSON.parse(config['headers'] as string); } catch { /* keep as string, backend will ignore */ }
			}

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
		{ key: 'title' as const, header: 'Title', sortable: true },
		{ key: 'provider_name' as const, header: 'Provider', sortable: true },
		{ key: 'event_type' as const, header: 'Event', sortable: true },
		{ key: 'status' as const, header: 'Status', sortable: true },
		{ key: 'created_at' as const, header: 'Date', sortable: true }
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
			<HelpDrawer page="notifications" />
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
							<Button size="sm" variant="secondary" loading={testing === provider.id} onclick={() => testProvider(provider.id)}>Test</Button>
							<Button size="sm" variant="ghost" onclick={() => openEdit(provider)}>Edit</Button>
							<ConfirmAction onconfirm={() => deleteProvider(provider.id)}>
								<Button size="sm" variant="ghost">
									<Trash2 class="w-4 h-4 text-destructive" />
								</Button>
							</ConfirmAction>
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
					<td class="px-4 py-2 font-mono text-xs">{item.event_type}</td>
					<td class="px-4 py-2 text-sm">
						<Badge variant={statusVariant(item.status)}>{item.status}</Badge>
						{#if item.error}
							<span class="block text-[10px] text-destructive mt-0.5 truncate max-w-[200px]" title={item.error}>{item.error}</span>
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
				<h3 class="text-sm font-semibold text-foreground mb-3">Configuration</h3>
				<div class="space-y-3">
					{#if formType === 'email'}
						<!-- Email: group SMTP connection and addresses -->
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<Input bind:value={formConfig['host']} label="SMTP Host" placeholder={fieldPlaceholders['host']} />
						<Input bind:value={formConfig['port']} label="SMTP Port" placeholder={fieldPlaceholders['port']} hint={fieldHints['port']} />
					</div>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<Input bind:value={formConfig['username']} label="Username" />
						<Input bind:value={formConfig['password']} label="Password" type="password" />
					</div>
					<div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
						<Input bind:value={formConfig['from']} label="From Address" placeholder={fieldPlaceholders['from']} />
						<Input bind:value={formConfig['to']} label="To Address(es)" placeholder={fieldPlaceholders['to']} hint={fieldHints['to']} />
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
				<h3 class="text-sm font-semibold text-foreground">Templates</h3>
				<Select value="0" onchange={(e: Event) => applyPreset(Number((e.target as HTMLSelectElement).value))} label="" class="w-40 h-8 text-xs">
					{#each getPresets() as preset, i}
						<option value={i}>{preset.label}</option>
					{/each}
				</Select>
			</div>
			<p class="text-xs text-muted-foreground mb-3">
				Customise notification text with Go templates. Available: {'{{.Title}}'}, {'{{.Message}}'}, {'{{.AppType}}'}, {'{{.Instance}}'}, {'{{.Type}}'}, {'{{index .Fields "key"}}'}.
			</p>
			<div class="space-y-3">
				<Input bind:value={formTitleTemplate} label="Title Template" placeholder="Leave blank for default" class="font-mono" />
				<div class="space-y-1.5">
					<Label for="body-tpl">Body Template</Label>
					<Textarea id="body-tpl" bind:value={formBodyTemplate} placeholder="Leave blank for default" rows={3} class="font-mono resize-y" />
				</div>
			</div>
		</div>

		<!-- Events -->
		<div class="border-t border-border pt-4">
			<div class="flex items-center justify-between mb-3">
				<h3 class="text-sm font-semibold text-foreground">Events</h3>
				<Button size="sm" variant="link" class="h-auto p-0 text-xs" onclick={() => formEvents = formEvents.length === allEvents.length ? [] : [...allEvents]}>
					{formEvents.length === allEvents.length ? 'Deselect all' : 'Select all'}
				</Button>
			</div>
			<div class="space-y-3">
				{#each eventGroups as group}
					<div>
						<span class="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-1 block">{group.label}</span>
						<div class="grid grid-cols-1 sm:grid-cols-2 gap-1">
							{#each group.events as event}
							<label class="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm cursor-pointer transition-colors hover:bg-muted {formEvents.includes(event) ? 'text-foreground' : 'text-muted-foreground'}">
									<Checkbox checked={formEvents.includes(event)} onchange={() => toggleEvent(event)} />
									<span class="text-xs">{eventLabels[event] ?? event}</span>
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

<ScrollToTop />
