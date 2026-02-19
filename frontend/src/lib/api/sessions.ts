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

export async function listSessions(trackId?: number): Promise<SessionSummary[]> {
	const qs = trackId ? `?track_id=${trackId}` : '';
	const data = await apiFetch<{ sessions: SessionSummary[] }>(`/sessions${qs}`);
	return data.sessions;
}
