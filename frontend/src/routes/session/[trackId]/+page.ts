import { getTrack } from '$lib/api/tracks';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = parseInt(params.trackId, 10);
	if (isNaN(id)) error(400, 'Invalid track ID');

	try {
		const track = await getTrack(id);
		return { track };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Track not found' : 'Failed to load track');
	}
};
