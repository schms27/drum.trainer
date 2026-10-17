import { fetchStats } from '$lib/api/stats';
import { listTracks } from '$lib/api/tracks';
import { listSessions } from '$lib/api/sessions';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const requested = Number(url.searchParams.get('track'));
	const trackFilter = Number.isInteger(requested) && requested > 0 ? requested : null;
	const [stats, tracks, takes] = await Promise.all([
		fetchStats(),
		listTracks(),
		listSessions({ trackId: trackFilter ?? undefined })
	]);
	return { stats, tracks, takes: takes.sessions, hasMore: takes.hasMore, trackFilter };
};
