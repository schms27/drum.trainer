import { listSessions } from '$lib/api/sessions';
import { listTracks } from '$lib/api/tracks';

export const load = async () => {
	const [page, tracks] = await Promise.all([listSessions(), listTracks()]);
	const trackNames = new Map(tracks.map((t) => [t.id, t.name]));
	return { sessions: page.sessions, hasMore: page.hasMore, trackNames };
};
