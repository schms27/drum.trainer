import type { PublicTrack, RightsBasis, SharedTrack, SharedVisibility } from '$lib/types/track';
import { apiFetch } from './client';

export interface ShareState {
	visibility: SharedVisibility;
	share_token: string;
}

/** Shares a track (or changes how it's shared); the caller has shown and checked the attestation. */
export async function shareTrack(
	trackId: number,
	rightsBasis: RightsBasis,
	visibility: SharedVisibility
): Promise<ShareState> {
	return apiFetch<ShareState>(`/tracks/${trackId}/sharing`, {
		method: 'PUT',
		body: JSON.stringify({ visibility, rights_basis: rightsBasis, accept_terms: true })
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

/** Rights bases that allow listing a track in Explore. Mirrors models.ListableRightsBases. */
export const LISTABLE_RIGHTS: readonly RightsBasis[] = ['original', 'licensed', 'public_domain'];

export type ExploreSort = 'newest' | 'popular';

export interface ExplorePage {
	tracks: PublicTrack[];
	page: number;
	has_more: boolean;
}

export async function getExplore(query: string, sort: ExploreSort, page: number): Promise<ExplorePage> {
	const params = new URLSearchParams({ sort, page: String(page) });
	if (query) params.set('q', query);
	return apiFetch<ExplorePage>(`/explore?${params}`);
}
