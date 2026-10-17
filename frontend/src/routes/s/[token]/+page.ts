import { error } from '@sveltejs/kit';
import { getSharedTrack } from '$lib/api/sharing';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	try {
		return { token: params.token, shared: await getSharedTrack(params.token) };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(
			status,
			status === 404 ? 'This share link doesn’t exist or has been turned off.' : 'Failed to load the shared track'
		);
	}
};
