const BASE = '/api';

class APIError extends Error {
	status: number;
	constructor(status: number, message: string) {
		super(message);
		this.status = status;
	}
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const opts: RequestInit = {
		method,
		headers: { 'Content-Type': 'application/json' },
		credentials: 'same-origin'
	};
	if (body) opts.body = JSON.stringify(body);

	const res = await fetch(`${BASE}${path}`, opts);
	if (!res.ok) {
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
