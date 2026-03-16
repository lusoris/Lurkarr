import { api } from '$lib/api';

interface User {
	id: string;
	username: string;
	is_admin: boolean;
	has_2fa: boolean;
	auth_provider: string;
	created_at: string;
}

let user = $state<User | null>(null);
let loading = $state(true);

export function getAuth() {
	async function check() {
		loading = true;
		try {
			const fetched = await api.get<User>('/user');
			// Only update if user actually changed (avoids $effect cascades).
			if (!user || user.id !== fetched.id || user.username !== fetched.username) {
				user = fetched;
			}
		} catch {
			user = null;
		} finally {
			loading = false;
		}
	}

	async function login(username: string, password: string, totp?: string, recoveryCode?: string) {
		const body: Record<string, string> = { username, password };
		if (totp) body.totp_code = totp;
		if (recoveryCode) body.recovery_code = recoveryCode;
		await api.post('/auth/login', body);
		await check();
	}

	async function logout() {
		await api.post('/auth/logout');
		user = null;
	}

	return {
		get user() { return user; },
		get loading() { return loading; },
		check,
		login,
		logout
	};
}
