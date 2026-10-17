import { error } from '@sveltejs/kit';
import { getExplore, type ExploreSort } from '$lib/api/sharing';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ url }) => {
	const query = url.searchParams.get('q')?.trim() ?? '';
	const sort: ExploreSort = url.searchParams.get('sort') === 'popular' ? 'popular' : 'newest';
	try {
		return { query, sort, first: await getExplore(query, sort, 1) };
	} catch {
		error(500, 'Failed to load Explore');
	}
};
