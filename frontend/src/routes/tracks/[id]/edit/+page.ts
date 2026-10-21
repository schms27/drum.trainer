import { error, redirect } from '@sveltejs/kit';
import { getTrack } from '$lib/api/tracks';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = Number(params.id);
	if (!Number.isInteger(id)) error(400, 'Invalid track ID');
	let track;
	try {
		track = await getTrack(id);
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Track not found' : 'Failed to load track');
	}
	// Removed tracks have no notes; the track page explains why.
	if (track.disabled) redirect(303, `/tracks/${id}`);
	return { track };
};
