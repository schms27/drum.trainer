import type { SessionSummary, SessionWithScore, SubmitSessionRequest } from '$lib/types/session';
import { apiFetch } from './client';

export async function submitSession(req: SubmitSessionRequest): Promise<SessionWithScore> {
	return apiFetch<SessionWithScore>('/sessions', {
		method: 'POST',
		body: JSON.stringify(req)
	});
}

export async function getSession(id: number): Promise<SessionWithScore> {
	return apiFetch<SessionWithScore>(`/sessions/${id}`);
}

export interface SessionPage {
	sessions: SessionSummary[];
	hasMore: boolean;
}

export async function listSessions(opts: { trackId?: number; limit?: number; before?: number } = {}): Promise<SessionPage> {
	const params = new URLSearchParams();
	if (opts.trackId) params.set('track_id', String(opts.trackId));
	if (opts.limit) params.set('limit', String(opts.limit));
	if (opts.before) params.set('before', String(opts.before));
	const qs = params.size ? `?${params}` : '';
	const data = await apiFetch<{ sessions: SessionSummary[]; has_more: boolean }>(`/sessions${qs}`);
	return { sessions: data.sessions, hasMore: data.has_more };
}
