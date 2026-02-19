// @ts-nocheck
import { getSession } from '$lib/api/sessions';
import { getTrack } from '$lib/api/tracks';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = async ({ params }: Parameters<PageLoad>[0]) => {
	const id = parseInt(params.sessionId, 10);
	if (isNaN(id)) error(400, 'Invalid session ID');

	try {
		const sessionResult = await getSession(id);
		const track = await getTrack(sessionResult.session.track_id);
		return { sessionResult, track };
	} catch (e: unknown) {
		const status = (e as { status?: number })?.status ?? 500;
		error(status, status === 404 ? 'Session not found' : 'Failed to load results');
	}
};
