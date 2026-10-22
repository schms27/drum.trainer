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

/** Status of the ApiError thrown when the server could not be reached at all. */
export const NETWORK_ERROR_STATUS = 0;
export const OFFLINE_MESSAGE = "You're offline, or the server can't be reached. Check your connection and try again.";

export function isNetworkError(e: unknown): boolean {
	return e instanceof ApiError && e.status === NETWORK_ERROR_STATUS;
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
	let res: Response;
	try {
		res = await fetch(`${BASE}${path}`, {
			...init,
			headers: { 'Content-Type': 'application/json', ...init?.headers },
		});
	} catch (e) {
		// fetch only rejects when no response arrived (offline, DNS, connection
		// refused); an aborted request is passed on unchanged.
		if (e instanceof DOMException && e.name === 'AbortError') throw e;
		throw new ApiError(NETWORK_ERROR_STATUS, OFFLINE_MESSAGE);
	}

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
