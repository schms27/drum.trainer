import { listKits } from '$lib/api/kits';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
	return { kits: await listKits() };
};
