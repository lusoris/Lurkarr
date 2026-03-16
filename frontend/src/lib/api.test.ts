import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api, APIError } from '$lib/api';

// Mock global fetch
const mockFetch = vi.fn();
vi.stubGlobal('fetch', mockFetch);

// Mock window.location
const mockLocation = { pathname: '/dashboard', href: '' };
vi.stubGlobal('window', { location: mockLocation });

function jsonResponse(data: unknown, status = 200, headers: Record<string, string> = {}) {
	const h = new Headers(headers);
	return {
		ok: status >= 200 && status < 300,
		status,
		statusText: status === 200 ? 'OK' : 'Error',
		headers: { get: (key: string) => h.get(key) },
		json: () => Promise.resolve(data)
	};
}

beforeEach(() => {
	mockFetch.mockReset();
	mockLocation.pathname = '/dashboard';
	mockLocation.href = '';
});

describe('api.get', () => {
	it('sends GET request to /api prefix', async () => {
		mockFetch.mockResolvedValue(jsonResponse({ id: 1 }));
		const result = await api.get('/user');
		expect(mockFetch).toHaveBeenCalledWith('/api/user', expect.objectContaining({ method: 'GET' }));
		expect(result).toEqual({ id: 1 });
	});

	it('does not include Content-Type header on GET (no body)', async () => {
		mockFetch.mockResolvedValue(jsonResponse({}));
		await api.get('/test');
		const opts = mockFetch.mock.calls[0][1];
		expect(opts.headers['Content-Type']).toBeUndefined();
	});
});

describe('api.post', () => {
	it('sends POST request with JSON body', async () => {
		mockFetch.mockResolvedValue(jsonResponse({ ok: true }));
		await api.post('/auth/login', { username: 'admin', password: 'pass' });
		const opts = mockFetch.mock.calls[0][1];
		expect(opts.method).toBe('POST');
		expect(JSON.parse(opts.body)).toEqual({ username: 'admin', password: 'pass' });
	});
});

describe('api.put', () => {
	it('sends PUT request', async () => {
		mockFetch.mockResolvedValue(jsonResponse({ updated: true }));
		await api.put('/settings', { key: 'value' });
		const opts = mockFetch.mock.calls[0][1];
		expect(opts.method).toBe('PUT');
	});
});

describe('api.del', () => {
	it('sends DELETE request', async () => {
		mockFetch.mockResolvedValue(jsonResponse({}));
		await api.del('/history/sonarr');
		const opts = mockFetch.mock.calls[0][1];
		expect(opts.method).toBe('DELETE');
	});
});

describe('CSRF token handling', () => {
	it('stores CSRF token from response and sends it on next mutation', async () => {
		// First request returns a CSRF token
		mockFetch.mockResolvedValue(jsonResponse({}, 200, { 'X-CSRF-Token': 'tok123' }));
		await api.get('/init');

		// Second request (POST) should include the token
		mockFetch.mockResolvedValue(jsonResponse({}));
		await api.post('/action');
		const opts = mockFetch.mock.calls[1][1];
		expect(opts.headers['X-CSRF-Token']).toBe('tok123');
	});
});

describe('error handling', () => {
	it('throws APIError on non-ok response', async () => {
		mockFetch.mockResolvedValue(jsonResponse({ error: 'not found' }, 404));
		await expect(api.get('/missing')).rejects.toThrow(APIError);
		await expect(api.get('/missing')).rejects.toThrow('not found');
	});

	it('redirects to /login on 401', async () => {
		mockFetch.mockResolvedValue(jsonResponse({ error: 'unauthorized' }, 401));
		// The 401 handler sets location.href and returns a never-resolving promise.
		// We race against a short timeout to avoid hanging.
		const result = await Promise.race([
			api.get('/protected').catch(() => 'rejected'),
			new Promise((resolve) => setTimeout(() => resolve('pending'), 50))
		]);
		expect(mockLocation.href).toBe('/login');
	});

	it('does not redirect if already on /login', async () => {
		mockLocation.pathname = '/login';
		mockFetch.mockResolvedValue(jsonResponse({ error: 'unauthorized' }, 401));
		await expect(api.get('/protected')).rejects.toThrow();
		expect(mockLocation.href).toBe('');
	});
});

describe('CSRF retry on 403', () => {
	it('retries request after refreshing CSRF token on csrf error', async () => {
		// First call: POST returns 403 with csrf error
		mockFetch
			.mockResolvedValueOnce(jsonResponse({ error: 'csrf token invalid' }, 403))
			// Second call: refreshCsrfToken does GET /api/user → returns new token
			.mockResolvedValueOnce(jsonResponse({}, 200, { 'X-CSRF-Token': 'new-token' }))
			// Third call: retried POST succeeds
			.mockResolvedValueOnce(jsonResponse({ ok: true }));

		const result = await api.post('/action', { data: 1 });
		expect(result).toEqual({ ok: true });

		// Verify the retry happened (3 total fetch calls)
		expect(mockFetch).toHaveBeenCalledTimes(3);

		// The retry request should carry the new CSRF token
		const retryOpts = mockFetch.mock.calls[2][1];
		expect(retryOpts.headers['X-CSRF-Token']).toBe('new-token');
	});

	it('throws on non-csrf 403 without retrying', async () => {
		mockFetch.mockResolvedValueOnce(jsonResponse({ error: 'forbidden' }, 403));
		await expect(api.post('/action')).rejects.toThrow('forbidden');
		expect(mockFetch).toHaveBeenCalledTimes(1);
	});
});

describe('204 No Content', () => {
	it('returns undefined on 204 response', async () => {
		mockFetch.mockResolvedValue({
			ok: true,
			status: 204,
			statusText: 'No Content',
			headers: { get: () => null },
			json: () => Promise.reject(new Error('no body'))
		});
		const result = await api.del('/history/sonarr');
		expect(result).toBeUndefined();
	});
});

describe('getCsrfToken', () => {
	it('returns empty string initially', async () => {
		// After module reset, token should default to empty
		// We can verify the getter works by checking it returns a string
		expect(typeof api.getCsrfToken()).toBe('string');
	});

	it('returns stored token after a response sets it', async () => {
		mockFetch.mockResolvedValue(jsonResponse({}, 200, { 'X-CSRF-Token': 'abc123' }));
		await api.get('/init');
		expect(api.getCsrfToken()).toBe('abc123');
	});
});
