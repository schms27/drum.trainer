import { getSiteInfo } from '$lib/api/site';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => ({ site: await getSiteInfo() });
