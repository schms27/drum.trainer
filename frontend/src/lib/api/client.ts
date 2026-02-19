const BASE = '/api/v1';

export class ApiError extends Error {
	constructor(
		public status: number,
		message: string
	) {
		super(message);
		this.name = 'ApiError';
	}
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
	const res = await fetch(`${BASE}${path}`, {
		headers: { 'Content-Type': 'application/json', ...init?.headers },
		...init
	});

	if (!res.ok) {
		let message = `HTTP ${res.status}`;
		try {
			const body = await res.json();
			if (body?.error) message = body.error;
		} catch {
			// ignore parse errors
		}
		throw new ApiError(res.status, message);
	}

	// 204 No Content
	if (res.status === 204) return undefined as T;
	return res.json();
}
