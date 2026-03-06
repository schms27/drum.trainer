import type { PageLoad } from './$types';
import { fetchStats } from '$lib/api/stats';

export const load: PageLoad = async () => {
	const stats = await fetchStats();
	return { stats };
};
