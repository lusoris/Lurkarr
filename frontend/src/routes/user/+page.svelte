<script lang="ts">
	import { api } from '$lib/api';
	import { getToasts } from '$lib/stores/toast.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Input from '$lib/components/ui/Input.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const toasts = getToasts();

	interface User {
		id: string;
		username: string;
		created_at: string;
	}

	let user = $state<User | null>(null);
	let newUsername = $state('');
	let currentPassword = $state('');
	let newPassword = $state('');
	let saving = $state(false);

	async function load() {
		try {
			user = await api.get<User>('/user');
			newUsername = user?.username ?? '';
		} catch { /* handled */ }
	}

	async function updateUsername() {
		saving = true;
		try {
			await api.post('/user/username', { username: newUsername });
			toasts.success('Username updated');
			await load();
		} catch {
			toasts.error('Failed to update username');
		}
		saving = false;
	}

	async function updatePassword() {
		if (!currentPassword || !newPassword) {
			toasts.error('Both fields required');
			return;
		}
		saving = true;
		try {
			await api.post('/user/password', { current_password: currentPassword, new_password: newPassword });
			toasts.success('Password updated');
			currentPassword = '';
			newPassword = '';
		} catch {
			toasts.error('Failed to update password');
		}
		saving = false;
	}

	$effect(() => { load(); });
</script>

<svelte:head><title>Profile - Lurkarr</title></svelte:head>

<div class="space-y-6">
	<h1 class="text-2xl font-bold text-surface-50">Profile</h1>

	{#if user}
		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Username</h2>
			<div class="space-y-3">
				<Input bind:value={newUsername} label="Username" />
				<Button onclick={updateUsername} loading={saving}>Update Username</Button>
			</div>
		</Card>

		<Card>
			<h2 class="text-lg font-semibold text-surface-200 mb-4">Change Password</h2>
			<div class="space-y-3">
				<Input bind:value={currentPassword} type="password" label="Current Password" />
				<Input bind:value={newPassword} type="password" label="New Password" />
				<Button onclick={updatePassword} loading={saving}>Update Password</Button>
			</div>
		</Card>

		<Card>
			<p class="text-xs text-surface-500">
				Account created: {new Date(user.created_at).toLocaleDateString()}
			</p>
		</Card>
	{:else}
		<Card>
			<p class="text-sm text-surface-500 text-center py-4">Loading profile...</p>
		</Card>
	{/if}
</div>
