import { listTracks } from '$lib/api/tracks';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	return { tracks: await listTracks() };
};
