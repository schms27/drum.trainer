import { redirect } from '@sveltejs/kit';
import type { PageLoad } from './$types';

// History and Stats were merged into /progress.
export const load: PageLoad = () => {
	redirect(308, '/progress');
};
