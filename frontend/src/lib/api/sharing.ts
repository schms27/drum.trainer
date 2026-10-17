import type { RightsBasis, SharedTrack } from '$lib/types/track';
import { apiFetch } from './client';

export interface ShareState {
	visibility: 'unlisted';
	share_token: string;
}

export async function shareTrack(trackId: number, rightsBasis: RightsBasis): Promise<ShareState> {
	return apiFetch<ShareState>(`/tracks/${trackId}/sharing`, {
		method: 'PUT',
		body: JSON.stringify({ rights_basis: rightsBasis, accept_terms: true })
	});
}

export async function rotateShareLink(trackId: number): Promise<ShareState> {
	return apiFetch<ShareState>(`/tracks/${trackId}/sharing/rotate`, { method: 'POST' });
}

export async function unshareTrack(trackId: number): Promise<void> {
	await apiFetch(`/tracks/${trackId}/sharing`, { method: 'DELETE' });
}

export async function getSharedTrack(token: string): Promise<SharedTrack> {
	return apiFetch<SharedTrack>(`/shared/${encodeURIComponent(token)}`);
}

/** Adds the shared track to the signed-in user's library; returns the copy's id. */
export async function copySharedTrack(token: string): Promise<number> {
	const data = await apiFetch<{ track_id: number }>(`/shared/${encodeURIComponent(token)}/copy`, {
		method: 'POST'
	});
	return data.track_id;
}

/** Absolute URL of a share link, for copying to the clipboard. */
export function shareUrl(token: string): string {
	return `${location.origin}/s/${token}`;
}

export const RIGHTS_LABELS: Record<RightsBasis, string> = {
	original: 'I wrote this groove myself',
	licensed: 'I have a licence or permission to share it',
	public_domain: 'The music is in the public domain',
	transcription: 'It is my transcription of someone else’s song'
};
