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

export function loginUrl(provider: 'google' | 'github'): string {
	return `/api/v1/auth/${provider}/login`;
}
