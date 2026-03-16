const BASE = '/api';

let csrfToken = '';

class APIError extends Error {
	status: number;
	constructor(status: number, message: string) {
		super(message);
		this.status = status;
	}
}

async function request<T>(method: string, path: string, body?: unknown, _retry = false, _attempt = 0): Promise<T> {
	const headers: Record<string, string> = {};
	if (body) headers['Content-Type'] = 'application/json';
	if (csrfToken && method !== 'GET') {
		headers['X-CSRF-Token'] = csrfToken;
	}

	const opts: RequestInit = {
		method,
		headers,
		credentials: 'same-origin'
	};
	if (body) opts.body = JSON.stringify(body);

	const res = await fetch(`${BASE}${path}`, opts);

	// Update CSRF token from response header.
	const token = res.headers.get('X-CSRF-Token');
	if (token) csrfToken = token;

	if (!res.ok) {
		// Retry on 429 rate-limit with exponential backoff + jitter (up to 3 attempts).
		if (res.status === 429 && _attempt < 3) {
			const retryAfter = res.headers.get('Retry-After');
			const base = retryAfter ? parseInt(retryAfter, 10) * 1000 : 500 * 2 ** _attempt;
			const jitter = Math.random() * base;
			await new Promise((r) => setTimeout(r, base + jitter));
			return request<T>(method, path, body, _retry, _attempt + 1);
		}
		// On CSRF failure, refresh token via a GET and retry once.
		if (res.status === 403 && !_retry) {
			const data = await res.json().catch(() => ({ error: '' }));
			if (data.error?.includes('csrf')) {
				await refreshCsrfToken();
				return request<T>(method, path, body, true, _attempt);
			}
			throw new APIError(res.status, data.error || res.statusText);
		}
		// Redirect to login on 401 (unless already there).
		if (res.status === 401 && !window.location.pathname.startsWith('/login')) {
			window.location.href = '/login';
			return new Promise(() => {}) as T;
		}
		const data = await res.json().catch(() => ({ error: res.statusText }));
		throw new APIError(res.status, data.error || res.statusText);
	}
	if (res.status === 204) return undefined as T;
	return res.json();
}

async function refreshCsrfToken(): Promise<void> {
	const res = await fetch(`${BASE}/user`, { credentials: 'same-origin' });
	if (res.status === 401 && !window.location.pathname.startsWith('/login')) {
		window.location.href = '/login';
		return;
	}
	const token = res.headers.get('X-CSRF-Token');
	if (token) csrfToken = token;
}

export const api = {
	get: <T>(path: string) => request<T>('GET', path),
	post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
	put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
	del: <T>(path: string) => request<T>('DELETE', path),
	getCsrfToken: () => csrfToken
};

export { APIError };
