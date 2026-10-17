import { getSession } from '$lib/api/sessions';
import { getTrack } from '$lib/api/tracks';
import { listSessions } from '$lib/api/sessions';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = parseInt(params.sessionId, 10);
	if (isNaN(id)) error(400, 'Invalid session ID');

	try {
		const sessionResult = await getSession(id);
		const trackId = sessionResult.session.track_id;
		const [track, history] = await Promise.all([getTrack(trackId), listSessions({ trackId, limit: 100 })]);
		// Earlier takes on the same track, newest first — for "compared with last time".
		const earlier = history.sessions.filter((s) => s.id < id);
		const previous = earlier[0] ?? null;
		const bestBefore = earlier.length ? Math.max(...earlier.map((s) => s.hit_rate)) : null;
		return { sessionResult, track, previous, bestBefore };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Session not found' : 'Failed to load results');
	}
};
