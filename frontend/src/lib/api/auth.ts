import { apiFetch } from './client';

export type AuthUser = {
	id: number;
	provider: string;
	email: string;
	name: string;
	avatar_url: string;
};

export async function fetchCurrentUser(): Promise<AuthUser> {
	return apiFetch<AuthUser>('/auth/me');
}

export async function logout(): Promise<void> {
	await apiFetch<void>('/auth/logout', { method: 'POST' });
}

/** Sign-in URL; `next` is a same-origin path to return to afterwards. */
export function loginUrl(provider: 'google' | 'github', next?: string | null): string {
	const base = `/api/v1/auth/${provider}/login`;
	return next ? `${base}?next=${encodeURIComponent(next)}` : base;
}
