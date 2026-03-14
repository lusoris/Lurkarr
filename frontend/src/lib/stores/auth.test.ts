import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock the api module before importing auth store
vi.mock('$lib/api', () => {
	const mockApi = {
		get: vi.fn(),
		post: vi.fn(),
		put: vi.fn(),
		del: vi.fn()
	};
	return { api: mockApi, APIError: Error };
});

import { getAuth } from '$lib/stores/auth.svelte';
import { api } from '$lib/api';

const mockApi = api as unknown as {
	get: ReturnType<typeof vi.fn>;
	post: ReturnType<typeof vi.fn>;
};

describe('auth store', () => {
	let auth: ReturnType<typeof getAuth>;

	beforeEach(() => {
		vi.clearAllMocks();
		auth = getAuth();
	});

	describe('check', () => {
		it('sets user on successful check', async () => {
			mockApi.get.mockResolvedValue({ id: '1', username: 'admin', is_admin: true });
			await auth.check();
			expect(auth.user).toEqual({ id: '1', username: 'admin', is_admin: true });
			expect(auth.loading).toBe(false);
		});

		it('sets user to null on failed check', async () => {
			mockApi.get.mockRejectedValue(new Error('401'));
			await auth.check();
			expect(auth.user).toBeNull();
			expect(auth.loading).toBe(false);
		});
	});

	describe('login', () => {
		it('posts credentials and refreshes user', async () => {
			mockApi.post.mockResolvedValue({});
			mockApi.get.mockResolvedValue({ id: '1', username: 'admin', is_admin: true });
			await auth.login('admin', 'password123');
			expect(mockApi.post).toHaveBeenCalledWith('/auth/login', {
				username: 'admin',
				password: 'password123'
			});
			expect(mockApi.get).toHaveBeenCalledWith('/user');
		});

		it('sends totp_code when provided', async () => {
			mockApi.post.mockResolvedValue({});
			mockApi.get.mockResolvedValue({ id: '1', username: 'admin', is_admin: false });
			await auth.login('admin', 'pass', '123456');
			expect(mockApi.post).toHaveBeenCalledWith('/auth/login', {
				username: 'admin',
				password: 'pass',
				totp_code: '123456'
			});
		});

		it('sends recovery_code when provided', async () => {
			mockApi.post.mockResolvedValue({});
			mockApi.get.mockResolvedValue({ id: '1', username: 'admin', is_admin: false });
			await auth.login('admin', 'pass', undefined, 'abcd-ef12');
			expect(mockApi.post).toHaveBeenCalledWith('/auth/login', {
				username: 'admin',
				password: 'pass',
				recovery_code: 'abcd-ef12'
			});
		});
	});

	describe('logout', () => {
		it('posts logout and clears user', async () => {
			// First set up a user
			mockApi.get.mockResolvedValue({ id: '1', username: 'admin', is_admin: true });
			await auth.check();
			expect(auth.user).not.toBeNull();

			// Now logout
			mockApi.post.mockResolvedValue({});
			await auth.logout();
			expect(mockApi.post).toHaveBeenCalledWith('/auth/logout');
			expect(auth.user).toBeNull();
		});
	});
});
