const BASE = '/api';

let csrfToken = '';

class APIError extends Error {
	status: number;
	constructor(status: number, message: string) {
		super(message);
		this.status = status;
	}
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const headers: Record<string, string> = { 'Content-Type': 'application/json' };
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
		// Redirect to login on 401 (unless already there).
		if (res.status === 401 && !window.location.pathname.startsWith('/login')) {
			window.location.href = '/login';
		}
		const data = await res.json().catch(() => ({ error: res.statusText }));
		throw new APIError(res.status, data.error || res.statusText);
	}
	return res.json();
}

export const api = {
	get: <T>(path: string) => request<T>('GET', path),
	post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
	put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
	del: <T>(path: string) => request<T>('DELETE', path)
};

export { APIError };
