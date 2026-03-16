<script lang="ts">
	import { api } from '$lib/api';
	import { appTypes, appDisplayName, appTabLabel } from '$lib';
	import ScrollToTop from '$lib/components/ScrollToTop.svelte';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Toggle from '$lib/components/ui/Toggle.svelte';
	import Select from '$lib/components/ui/Select.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import HelpDrawer from '$lib/components/HelpDrawer.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import * as ScrollArea from '$lib/components/ui/scroll-area';
	import * as ToggleGroup from '$lib/components/ui/toggle-group';
	import { CalendarDays, Plus, Clock } from '@lucide/svelte';
	import type { Schedule, ScheduleExecution } from '$lib/types';

	const toasts = getToasts();

	const actions = [
		{ value: 'lurk_missing', label: 'Lurk Missing', desc: 'Search for missing media' },
		{ value: 'lurk_upgrade', label: 'Lurk Upgrades', desc: 'Search for quality upgrades' },
		{ value: 'lurk_all', label: 'Lurk All', desc: 'Search missing + upgrades' },
		{ value: 'clean_queue', label: 'Clean Queue', desc: 'Run queue cleaner pass' },
		{ value: 'enable', label: 'Enable Instances', desc: 'Enable all instances' },
		{ value: 'disable', label: 'Disable Instances', desc: 'Disable all instances' },
	] as const;
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

	function recentExecs(schedId: string): ScheduleExecution[] {
		return history.filter(e => e.schedule_id === schedId).slice(0, 3);
	}

	async function load() {
		loading = true;
		try {
			const [s, h] = await Promise.all([
				api.get<Schedule[]>('/schedules'),
				api.get<ScheduleExecution[]>('/schedules/history?limit=50')
			]);
			schedules = s ?? [];
			history = h ?? [];
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
		const found = actions.find(a => a.value === action);
		if (found) return found.label;
		if (action.startsWith('api-')) return `API Cap: ${action.slice(4)}/hr`;
		return action.replace(/_/g, ' ').replace(/\b\w/g, c => c.toUpperCase());
	}
</script>

<svelte:head><title>Scheduling - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="Schedules">
		{#snippet actions()}
			<Button size="sm" variant="secondary" onclick={loadHistory}>
				<Clock class="h-3.5 w-3.5" />
				History
			</Button>
			<Button size="sm" onclick={openAdd}>
				<Plus class="h-3.5 w-3.5" />
				Add Schedule
			</Button>
			<HelpDrawer page="scheduling" />
		{/snippet}
	</PageHeader>

	{#if loading}
		<Skeleton rows={3} height="h-16" />
	{:else if schedules.length === 0}
		<EmptyState icon={CalendarDays} title="No schedules configured" description="Create a schedule to automate your lurking and queue cleaning.">
			{#snippet actions()}
				<Button size="sm" onclick={openAdd}>Add Schedule</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<div class="space-y-2">
			{#each schedules as sched}
				{@const recent = recentExecs(sched.id)}
				<Card>
					<div class="flex flex-col sm:flex-row sm:items-center justify-between gap-3">
						<div class="flex items-center gap-4 flex-wrap">
							<Badge variant={sched.enabled ? 'success' : 'error'}>
								{sched.enabled ? 'Active' : 'Inactive'}
							</Badge>
							<div>
								<span class="font-medium text-foreground">{sched.app_type === 'global' ? 'Global (All Apps)' : appTabLabel(sched.app_type)}</span>
								<span class="text-muted-foreground mx-2">&middot;</span>
								<span class="text-muted-foreground">{formatAction(sched.action)}</span>
							</div>
						</div>
						<div class="flex items-center gap-3">
							<div class="text-right">
								<span class="font-mono text-foreground">{formatTime(sched.hour, sched.minute)}</span>
								<p class="text-xs text-muted-foreground">{sched.days.length > 0 ? sched.days.join(', ') : 'Every day'}</p>
							</div>
							<Button size="sm" variant="ghost" onclick={() => openEdit(sched)}>Edit</Button>
							{#if deleteConfirm === sched.id}
								<Button size="sm" variant="danger" onclick={() => deleteSchedule(sched.id)}>Confirm</Button>
								<Button size="sm" variant="ghost" onclick={() => deleteConfirm = null}>Cancel</Button>
							{:else}
								<Button size="sm" variant="danger" onclick={() => deleteConfirm = sched.id}>Delete</Button>
							{/if}
						</div>
					</div>
					{#if recent.length > 0}
						<div class="mt-2 pt-2 border-t border-border/50 flex flex-wrap gap-3 text-[11px] text-muted-foreground">
							<span>Recent:</span>
							{#each recent as exec}
								<span class="flex items-center gap-1">
									<span class="inline-block w-1.5 h-1.5 rounded-full {exec.result === 'success' ? 'bg-emerald-400' : exec.result ? 'bg-destructive' : 'bg-muted-foreground'}"></span>
									{new Date(exec.executed_at).toLocaleDateString(undefined, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })}
								</span>
							{/each}
						</div>
					{/if}
				</Card>
			{/each}
		</div>
	{/if}
</div>

<!-- Add/Edit Schedule Modal -->
<Modal bind:open={showModal} title={editing ? 'Edit Schedule' : 'Add Schedule'} onclose={() => showModal = false}>
	<form onsubmit={(e: Event) => { e.preventDefault(); saveSchedule(); }} class="space-y-4">
		<Select bind:value={form.app_type} label="App Type">
			<option value="global">Global (All Apps)</option>
			{#each appTypes as app}
				<option value={app}>{appTabLabel(app)}</option>
			{/each}
		</Select>
		<Select bind:value={form.action} label="Action">
			{#each actions as action}
				<option value={action.value}>{action.label}</option>
			{/each}
		</Select>
		{#if actions.find(a => a.value === form.action)?.desc}
			<p class="text-xs text-muted-foreground -mt-2">{actions.find(a => a.value === form.action)?.desc}</p>
		{/if}
		<div class="grid grid-cols-2 gap-4">
			<Input bind:value={form.hour} type="number" label="Hour (0-23)" min={0} max={23} />
			<Input bind:value={form.minute} type="number" label="Minute (0-59)" min={0} max={59} />
		</div>
		<div>
			<span class="block text-sm font-medium text-muted-foreground mb-2">Days (empty = every day)</span>
			<ToggleGroup.Root type="multiple" bind:value={form.days} variant="outline" size="sm" class="flex flex-wrap gap-1">
				{#each dayOptions as day}
					<ToggleGroup.Item value={day} class="px-3 py-1.5 text-xs">{day.slice(0, 3)}</ToggleGroup.Item>
				{/each}
			</ToggleGroup.Root>
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
		<p class="text-sm text-muted-foreground text-center py-4">No execution history</p>
	{:else}
		<ScrollArea.Root class="max-h-96">
			<div class="space-y-2">
				{#each history as exec}
					<div class="flex items-center justify-between rounded-lg bg-muted/50 px-3 py-2 text-sm">
						<span class="text-muted-foreground text-xs">{new Date(exec.executed_at).toLocaleString()}</span>
						<Badge variant={exec.result === 'success' ? 'success' : exec.result ? 'error' : 'default'}>
							{exec.result ?? 'pending'}
						</Badge>
					</div>
				{/each}
			</div>
		</ScrollArea.Root>
	{/if}
</Modal>

<ScrollToTop />
