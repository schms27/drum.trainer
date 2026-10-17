import { getTrack } from '$lib/api/tracks';
import { listPilots } from '$lib/api/pilots';
import { listKits } from '$lib/api/kits';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = parseInt(params.trackId, 10);
	if (isNaN(id)) error(400, 'Invalid track ID');

	try {
		const [track, pilots, kits] = await Promise.all([getTrack(id), listPilots(id), listKits()]);
		return { track, pilots, activeKit: kits.find((k) => k.active) ?? null };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Track not found' : 'Failed to load track');
	}
};
