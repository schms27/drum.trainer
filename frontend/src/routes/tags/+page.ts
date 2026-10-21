import { listTags } from '$lib/api/tags';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => ({ tags: await listTags() });
