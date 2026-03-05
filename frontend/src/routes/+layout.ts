// Run all load functions in the browser (SPA mode).
// Relative API fetches like /api/v1/... only work client-side where the
// Vite dev proxy (and the production reverse proxy) can forward them to the backend.
export const ssr = false;

import { redirect } from '@sveltejs/kit';
import type { LayoutLoad } from './$types';
import { fetchCurrentUser } from '$lib/api/auth';
import { ApiError } from '$lib/api/client';

export const load: LayoutLoad = async ({ url }) => {
	// The login page is public — skip the auth check to avoid an infinite redirect.
	if (url.pathname === '/login') {
		return { user: null };
	}

	try {
		const user = await fetchCurrentUser();
		return { user };
	} catch (e) {
		if (e instanceof ApiError && e.status === 401) {
			throw redirect(303, '/login');
		}
		throw e;
	}
};
