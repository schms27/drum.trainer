import { error } from '@sveltejs/kit';
import { getTrack } from '$lib/api/tracks';
import { listPilots } from '$lib/api/pilots';
import { listSessions } from '$lib/api/sessions';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = Number(params.id);
	if (!Number.isInteger(id)) error(400, 'Invalid track ID');

	try {
		const [track, pilots, takes] = await Promise.all([
			getTrack(id),
			listPilots(id),
			listSessions({ trackId: id, limit: 100 })
		]);
		return { track, pilots, takes: takes.sessions, hasMoreTakes: takes.hasMore };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Track not found' : 'Failed to load track');
	}
};
