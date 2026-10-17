import { getTrack } from '$lib/api/tracks';
import { listPilots } from '$lib/api/pilots';
import { listKits } from '$lib/api/kits';
import { error, redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = parseInt(params.trackId, 10);
	if (isNaN(id)) error(400, 'Invalid track ID');

	let data;
	try {
		const [track, pilots, kits] = await Promise.all([getTrack(id), listPilots(id), listKits()]);
		data = { track, pilots, activeKit: kits.find((k) => k.active) ?? null };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Track not found' : 'Failed to load track');
	}
	// Removed tracks have no notes; the track page explains why.
	if (data.track.disabled) redirect(303, `/tracks/${id}`);
	return data;
};
