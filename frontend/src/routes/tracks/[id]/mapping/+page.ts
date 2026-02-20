import { getTrack, getTrackMapping } from '$lib/api/tracks';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
	const id = Number(params.id);
	const [track, mappings] = await Promise.all([getTrack(id), getTrackMapping(id)]);
	return { track, mappings };
};
