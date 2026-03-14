<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';
	import Badge from '$lib/components/ui/Badge.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Skeleton from '$lib/components/ui/Skeleton.svelte';
	import EmptyState from '$lib/components/ui/EmptyState.svelte';
	import DataTable, { type Column } from '$lib/components/ui/DataTable.svelte';
	import * as T from '$lib/components/ui/table';
	import { Plus, Users, ShieldCheck, Trash2 } from 'lucide-svelte';

	const toasts = getToasts();

	interface UserEntry {
		id: string;
		username: string;
		auth_provider: string;
		is_admin: boolean;
		has_2fa: boolean;
		created_at: string;
	}

	let users = $state<UserEntry[]>([]);
	let loading = $state(true);

	// Create user
	let showCreate = $state(false);
	let newUsername = $state('');
	let newPassword = $state('');
	let newIsAdmin = $state(false);
	let creating = $state(false);

	// Reset password
	let showReset = $state(false);
	let resetUserId = $state('');
	let resetUsername = $state('');
	let resetPassword = $state('');
	let resetting = $state(false);

	async function load() {
		loading = true;
		try {
			users = await api.get<UserEntry[]>('/admin/users');
		} catch {
			toasts.error('Failed to load users');
		}
		loading = false;
	}

	async function createUser() {
		if (!newUsername || !newPassword) {
			toasts.error('Username and password required');
			return;
		}
		creating = true;
		try {
			await api.post('/admin/users', { username: newUsername, password: newPassword, is_admin: newIsAdmin });
			toasts.success(`User "${newUsername}" created`);
			showCreate = false;
			newUsername = '';
			newPassword = '';
			newIsAdmin = false;
			await load();
		} catch (e: any) {
			toasts.error(e?.message || 'Failed to create user');
		}
		creating = false;
	}

	let confirmDelete = $state<{ id: string; username: string } | null>(null);

	async function deleteUser(id: string, username: string) {
		confirmDelete = { id, username };
	}

	async function confirmDeleteUser() {
		if (!confirmDelete) return;
		try {
			await api.del(`/admin/users/${confirmDelete.id}`);
			toasts.success(`User "${confirmDelete.username}" deleted`);
			confirmDelete = null;
			await load();
		} catch (e: any) {
			toasts.error(e?.message || 'Failed to delete user');
		}
	}

	async function toggleAdmin(id: string, currentAdmin: boolean) {
		try {
			await api.post(`/admin/users/${id}/toggle-admin`, { is_admin: !currentAdmin });
			toasts.success('Admin status updated');
			await load();
		} catch (e: any) {
			toasts.error(e?.message || 'Failed to toggle admin');
		}
	}

	function openResetPassword(id: string, username: string) {
		resetUserId = id;
		resetUsername = username;
		resetPassword = '';
		showReset = true;
	}

	async function resetUserPassword() {
		if (!resetPassword) {
			toasts.error('Password required');
			return;
		}
		resetting = true;
		try {
			await api.post(`/admin/users/${resetUserId}/reset-password`, { password: resetPassword });
			toasts.success(`Password reset for "${resetUsername}"`);
			showReset = false;
		} catch (e: any) {
			toasts.error(e?.message || 'Failed to reset password');
		}
		resetting = false;
	}

	$effect(() => { load(); });

	const userColumns: Column<UserEntry>[] = [
		{ key: 'username', header: 'Username', sortable: true },
		{ key: 'auth_provider', header: 'Provider', sortable: true },
		{ key: 'is_admin', header: 'Role', sortable: true },
		{ key: 'has_2fa', header: '2FA' },
		{ key: 'created_at', header: 'Created', sortable: true },
		{ key: '_actions', header: 'Actions', headerClass: 'text-right' }
	];
</script>

<svelte:head><title>User Management - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<PageHeader title="User Management" description="Create, manage, and remove user accounts.">
		{#snippet actions()}
			<Button size="sm" onclick={() => showCreate = true}>
				<Plus class="h-4 w-4" />
				Add User
			</Button>
		{/snippet}
	</PageHeader>

	{#if loading}
		<Skeleton rows={3} height="h-14" />
	{:else if users.length === 0}
		<EmptyState icon={Users} title="No users found" description="Create the first user account to get started.">
			{#snippet actions()}
				<Button size="sm" onclick={() => showCreate = true}>
					<Plus class="h-4 w-4" />
					Add User
				</Button>
			{/snippet}
		</EmptyState>
	{:else}
		<DataTable data={users} columns={userColumns} searchable searchPlaceholder="Search users..." noun="users">
			{#snippet row(u)}
				<T.Row>
					<T.Cell class="font-medium">{u.username}</T.Cell>
					<T.Cell><Badge variant="default">{u.auth_provider}</Badge></T.Cell>
					<T.Cell>
						{#if u.is_admin}
							<Badge variant="warning">Admin</Badge>
						{:else}
							<span class="text-muted-foreground text-xs">User</span>
						{/if}
					</T.Cell>
					<T.Cell>
						{#if u.has_2fa}
							<ShieldCheck class="h-4 w-4 text-green-400" />
						{:else}
							<span class="text-muted-foreground/50">—</span>
						{/if}
					</T.Cell>
					<T.Cell class="text-xs text-muted-foreground">{new Date(u.created_at).toLocaleDateString()}</T.Cell>
					<T.Cell class="text-right">
						<div class="flex items-center justify-end gap-1">
							<Button variant="ghost" size="sm" onclick={() => toggleAdmin(u.id, u.is_admin)}>
								{u.is_admin ? 'Demote' : 'Promote'}
							</Button>
							<Button variant="ghost" size="sm" onclick={() => openResetPassword(u.id, u.username)}>Reset PW</Button>
							<Button variant="danger" size="sm" onclick={() => deleteUser(u.id, u.username)}>
								<Trash2 class="h-4 w-4" />
							</Button>
						</div>
					</T.Cell>
				</T.Row>
			{/snippet}
		</DataTable>
	{/if}
</div>

<!-- Delete Confirmation Modal -->
<Modal open={!!confirmDelete} title="Delete User" onclose={() => confirmDelete = null}>
	<div class="space-y-4">
		<p class="text-sm text-muted-foreground">Delete user <strong class="text-foreground">"{confirmDelete?.username}"</strong>? This cannot be undone.</p>
		<div class="flex justify-end gap-2">
			<Button variant="secondary" onclick={() => confirmDelete = null}>Cancel</Button>
			<Button variant="danger" onclick={confirmDeleteUser}>Delete</Button>
		</div>
	</div>
</Modal>

<!-- Create User Modal -->
<Modal open={showCreate} title="Create User" onclose={() => showCreate = false}>
	<div class="space-y-4">
		<Input bind:value={newUsername} label="Username" placeholder="Enter username" />
		<Input bind:value={newPassword} type="password" label="Password" placeholder="Min 8 chars, upper + lower + digit" />
		<label class="flex items-center gap-2 text-sm text-muted-foreground">
			<input type="checkbox" bind:checked={newIsAdmin} class="rounded border-border bg-muted text-primary focus:ring-ring" />
			Admin privileges
		</label>
		<div class="flex gap-2">
			<Button onclick={createUser} loading={creating}>Create User</Button>
			<Button variant="ghost" onclick={() => showCreate = false}>Cancel</Button>
		</div>
	</div>
</Modal>

<!-- Reset Password Modal -->
<Modal open={showReset} title="Reset Password — {resetUsername}" onclose={() => showReset = false}>
	<div class="space-y-4">
		<Input bind:value={resetPassword} type="password" label="New Password" placeholder="Min 8 chars, upper + lower + digit" />
		<div class="flex gap-2">
			<Button onclick={resetUserPassword} loading={resetting}>Reset Password</Button>
			<Button variant="ghost" onclick={() => showReset = false}>Cancel</Button>
		</div>
	</div>
</Modal>
