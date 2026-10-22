// Run all load functions in the browser (SPA mode).
// Relative API fetches like /api/v1/... only work client-side where the
// Vite dev proxy (and the production reverse proxy) can forward them to the backend.
export const ssr = false;

import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { fetchCurrentUser } from '$lib/api/auth';
import { ApiError, isNetworkError } from '$lib/api/client';

export const load: LayoutLoad = async ({ url }) => {
	// The login page is public — skip the auth check to avoid an infinite redirect.
	if (url.pathname === '/login') {
		return { user: null, offline: false };
	}

	try {
		const user = await fetchCurrentUser();
		return { user, offline: false };
	} catch (e) {
		// No connection: we don't know who is signed in, so don't send them to
		// /login. The page's own load fails the same way and +error.svelte
		// explains it.
		if (isNetworkError(e)) return { user: null, offline: true };
		if (e instanceof ApiError && e.status === 401) {
			// Share links and the terms page work without an account.
			if (isPublicPath(url.pathname)) return { user: null, offline: false };
			throw redirect(303, '/login');
		}
		throw e;
	}
};

function isPublicPath(path: string): boolean {
	return path.startsWith('/s/') || ['/explore', '/terms', '/copyright', '/report'].includes(path);
}
