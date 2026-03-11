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

	interface Schedule {
		id: string;
		app_type: string;
		action: string;
		days: string[];
		hour: number;
		minute: number;
		enabled: boolean;
	}

	interface ScheduleExecution {
		id: number;
		schedule_id: string;
		executed_at: string;
		result: string | null;
	}

	const appTypes = ['sonarr', 'radarr', 'lidarr', 'readarr', 'whisparr', 'eros'] as const;
	const actions = ['lurk_missing', 'lurk_upgrade', 'lurk_all', 'clean_queue'] as const;
	const dayOptions = ['monday', 'tuesday', 'wednesday', 'thursday', 'friday', 'saturday', 'sunday'] as const;

	let schedules = $state<Schedule[]>([]);
	let history = $state<ScheduleExecution[]>([]);
	let loading = $state(true);
	let showHistory = $state(false);

	// Modal state
	let showModal = $state(false);
	let editing = $state<Schedule | null>(null);
	let form = $state({ app_type: 'sonarr', action: 'lurk_missing', days: [] as string[], hour: 0, minute: 0, enabled: true });
	let saving = $state(false);
	let deleteConfirm = $state<string | null>(null);

	async function load() {
		loading = true;
		try {
			schedules = await api.get<Schedule[]>('/schedules');
		} catch {
			schedules = [];
		}
		loading = false;
	}

	async function loadHistory() {
		try {
			history = await api.get<ScheduleExecution[]>('/schedules/history?limit=50');
			showHistory = true;
		} catch {
			toasts.error('Failed to load history');
		}
	}

	function openAdd() {
		editing = null;
		form = { app_type: 'sonarr', action: 'lurk_missing', days: [], hour: 0, minute: 0, enabled: true };
		showModal = true;
	}

	function openEdit(sched: Schedule) {
		editing = sched;
		form = { app_type: sched.app_type, action: sched.action, days: [...sched.days], hour: sched.hour, minute: sched.minute, enabled: sched.enabled };
		showModal = true;
	}

	function toggleDay(day: string) {
		if (form.days.includes(day)) {
			form.days = form.days.filter(d => d !== day);
		} else {
			form.days = [...form.days, day];
		}
	}

	async function saveSchedule() {
		saving = true;
		try {
			const payload = { ...form };
			if (editing) {
				await api.put(`/schedules/${editing.id}`, payload);
				toasts.success('Schedule updated');
			} else {
				await api.post('/schedules', payload);
				toasts.success('Schedule created');
			}
			showModal = false;
			await load();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to save');
		}
		saving = false;
	}

	async function deleteSchedule(id: string) {
		try {
			await api.del(`/schedules/${id}`);
			toasts.success('Schedule deleted');
			deleteConfirm = null;
			await load();
		} catch (e) {
			toasts.error(e instanceof Error ? e.message : 'Failed to delete');
		}
	}

	$effect(() => { load(); });

	function formatTime(h: number, m: number): string {
		return `${String(h).padStart(2, '0')}:${String(m).padStart(2, '0')}`;
	}

	function formatAction(action: string): string {
		return action.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
	}
</script>

<svelte:head><title>Scheduling - Lurkarr</title></svelte:head>

<div class="space-y-4">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-surface-50">Schedules</h1>
		<div class="flex gap-2">
			<Button size="sm" variant="secondary" onclick={loadHistory}>History</Button>
			<Button size="sm" onclick={openAdd}>Add Schedule</Button>
		</div>
	</div>

	{#if loading}
		<div class="space-y-2">
			{#each Array(3) as _}
				<div class="h-16 rounded-lg bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if schedules.length === 0}
		<Card>
			<p class="text-sm text-surface-500 text-center py-8">No schedules configured</p>
		</Card>
	{:else}
		<div class="space-y-2">
			{#each schedules as sched}
				<Card class="flex items-center justify-between">
					<div class="flex items-center gap-4">
						<Badge variant={sched.enabled ? 'success' : 'error'}>
							{sched.enabled ? 'Active' : 'Inactive'}
						</Badge>
						<div>
							<span class="font-medium text-surface-100 capitalize">{sched.app_type}</span>
							<span class="text-surface-500 mx-2">&middot;</span>
							<span class="text-surface-300">{formatAction(sched.action)}</span>
						</div>
					</div>
					<div class="flex items-center gap-3">
						<div class="text-right">
							<span class="font-mono text-surface-200">{formatTime(sched.hour, sched.minute)}</span>
							<p class="text-xs text-surface-500">{sched.days.length > 0 ? sched.days.join(', ') : 'Every day'}</p>
						</div>
						<Button size="sm" variant="ghost" onclick={() => openEdit(sched)}>Edit</Button>
						{#if deleteConfirm === sched.id}
							<Button size="sm" variant="danger" onclick={() => deleteSchedule(sched.id)}>Confirm</Button>
							<Button size="sm" variant="ghost" onclick={() => deleteConfirm = null}>Cancel</Button>
						{:else}
							<Button size="sm" variant="danger" onclick={() => deleteConfirm = sched.id}>Delete</Button>
						{/if}
					</div>
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Add/Edit Schedule Modal -->
<Modal bind:open={showModal} title={editing ? 'Edit Schedule' : 'Add Schedule'} onclose={() => showModal = false}>
	<form onsubmit={saveSchedule} class="space-y-4">
		<label class="block">
			<span class="block text-sm font-medium text-surface-300 mb-1.5">App Type</span>
			<select bind:value={form.app_type} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
				{#each appTypes as app}
					<option value={app} class="capitalize">{app}</option>
				{/each}
			</select>
		</label>
		<label class="block">
			<span class="block text-sm font-medium text-surface-300 mb-1.5">Action</span>
			<select bind:value={form.action} class="w-full rounded-lg border border-surface-700 bg-surface-900 text-surface-100 px-3 py-2 text-sm focus:outline-none focus:ring-1 focus:border-lurk-500 focus:ring-lurk-500">
				{#each actions as action}
					<option value={action}>{formatAction(action)}</option>
				{/each}
			</select>
		</label>
		<div class="grid grid-cols-2 gap-4">
			<Input bind:value={form.hour} type="number" label="Hour (0-23)" />
			<Input bind:value={form.minute} type="number" label="Minute (0-59)" />
		</div>
		<div>
			<span class="block text-sm font-medium text-surface-300 mb-2">Days (empty = every day)</span>
			<div class="flex flex-wrap gap-2">
				{#each dayOptions as day}
					<button
						type="button"
						onclick={() => toggleDay(day)}
						class="px-3 py-1.5 rounded-lg text-xs font-medium transition-colors
							{form.days.includes(day) ? 'bg-lurk-600 text-white' : 'bg-surface-800 text-surface-400 hover:bg-surface-700'}"
					>{day.slice(0, 3)}</button>
				{/each}
			</div>
		</div>
		<Toggle bind:checked={form.enabled} label="Enabled" />
		<div class="flex gap-2 pt-2">
			<Button type="submit" loading={saving}>{editing ? 'Update' : 'Create'}</Button>
		</div>
	</form>
</Modal>

<!-- History Modal -->
<Modal bind:open={showHistory} title="Schedule History" onclose={() => showHistory = false}>
	{#if history.length === 0}
		<p class="text-sm text-surface-500 text-center py-4">No execution history</p>
	{:else}
		<div class="max-h-96 overflow-y-auto space-y-2">
			{#each history as exec}
				<div class="flex items-center justify-between rounded-lg bg-surface-800/50 px-3 py-2 text-sm">
					<span class="text-surface-400 text-xs">{new Date(exec.executed_at).toLocaleString()}</span>
					<Badge variant={exec.result === 'success' ? 'success' : exec.result ? 'error' : 'default'}>
						{exec.result ?? 'pending'}
					</Badge>
				</div>
			{/each}
		</div>
	{/if}
</Modal>
