<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Modal from '$lib/components/ui/Modal.svelte';

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

	async function deleteUser(id: string, username: string) {
		if (!confirm(`Delete user "${username}"? This cannot be undone.`)) return;
		try {
			await api.del(`/admin/users/${id}`);
			toasts.success(`User "${username}" deleted`);
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
</script>

<svelte:head><title>User Management - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<div class="flex items-center justify-between">
		<h1 class="text-2xl font-bold text-surface-50">User Management</h1>
		<Button onclick={() => showCreate = true}>
			<svg class="w-4 h-4" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M12 4.5v15m7.5-7.5h-15"/></svg>
			Add User
		</Button>
	</div>

	{#if loading}
		<div class="space-y-3">
			{#each Array(3) as _}
				<div class="h-14 rounded-xl bg-surface-800/50 animate-pulse"></div>
			{/each}
		</div>
	{:else if users.length === 0}
		<Card>
			<p class="text-sm text-surface-500 text-center py-4">No users found</p>
		</Card>
	{:else}
		<Card>
			<div class="overflow-x-auto">
				<table class="w-full text-sm">
					<thead>
						<tr class="border-b border-surface-700 text-left text-surface-400">
							<th class="pb-3 pr-4 font-medium">Username</th>
							<th class="pb-3 pr-4 font-medium">Provider</th>
							<th class="pb-3 pr-4 font-medium">Role</th>
							<th class="pb-3 pr-4 font-medium">2FA</th>
							<th class="pb-3 pr-4 font-medium">Created</th>
							<th class="pb-3 font-medium text-right">Actions</th>
						</tr>
					</thead>
					<tbody class="text-surface-200">
						{#each users as u}
							<tr class="border-b border-surface-800/50 last:border-0">
								<td class="py-3 pr-4 font-medium">{u.username}</td>
								<td class="py-3 pr-4">
									<span class="px-2 py-0.5 rounded text-xs bg-surface-800 text-surface-300">{u.auth_provider}</span>
								</td>
								<td class="py-3 pr-4">
									{#if u.is_admin}
										<span class="px-2 py-0.5 rounded text-xs bg-lurk-600/20 text-lurk-400 font-medium">Admin</span>
									{:else}
										<span class="text-surface-500 text-xs">User</span>
									{/if}
								</td>
								<td class="py-3 pr-4">
									{#if u.has_2fa}
										<svg class="w-4 h-4 text-green-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M9 12.75L11.25 15 15 9.75m-3-7.036A11.959 11.959 0 013.598 6 11.99 11.99 0 003 9.749c0 5.592 3.824 10.29 9 11.623 5.176-1.332 9-6.03 9-11.622 0-1.31-.21-2.571-.598-3.751h-.152c-3.196 0-6.1-1.248-8.25-3.285z"/></svg>
									{:else}
										<span class="text-surface-600">—</span>
									{/if}
								</td>
								<td class="py-3 pr-4 text-xs text-surface-400">{new Date(u.created_at).toLocaleDateString()}</td>
								<td class="py-3 text-right">
									<div class="flex items-center justify-end gap-1">
										<Button variant="ghost" size="sm" onclick={() => toggleAdmin(u.id, u.is_admin)}>
											{u.is_admin ? 'Demote' : 'Promote'}
										</Button>
										<Button variant="ghost" size="sm" onclick={() => openResetPassword(u.id, u.username)}>Reset PW</Button>
										<Button variant="ghost" size="sm" onclick={() => deleteUser(u.id, u.username)}>
											<svg class="w-4 h-4 text-red-400" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor"><path stroke-linecap="round" stroke-linejoin="round" d="M14.74 9l-.346 9m-4.788 0L9.26 9m9.968-3.21c.342.052.682.107 1.022.166m-1.022-.165L18.16 19.673a2.25 2.25 0 01-2.244 2.077H8.084a2.25 2.25 0 01-2.244-2.077L4.772 5.79m14.456 0a48.108 48.108 0 00-3.478-.397m-12 .562c.34-.059.68-.114 1.022-.165m0 0a48.11 48.11 0 013.478-.397m7.5 0v-.916c0-1.18-.91-2.164-2.09-2.201a51.964 51.964 0 00-3.32 0c-1.18.037-2.09 1.022-2.09 2.201v.916m7.5 0a48.667 48.667 0 00-7.5 0"/></svg>
										</Button>
									</div>
								</td>
							</tr>
						{/each}
					</tbody>
				</table>
			</div>
		</Card>
	{/if}
</div>

<!-- Create User Modal -->
<Modal open={showCreate} title="Create User" onclose={() => showCreate = false}>
	<div class="space-y-4">
		<Input bind:value={newUsername} label="Username" placeholder="Enter username" />
		<Input bind:value={newPassword} type="password" label="Password" placeholder="Min 8 chars, upper + lower + digit" />
		<label class="flex items-center gap-2 text-sm text-surface-300">
			<input type="checkbox" bind:checked={newIsAdmin} class="rounded border-surface-600 bg-surface-800 text-lurk-500 focus:ring-lurk-500" />
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
