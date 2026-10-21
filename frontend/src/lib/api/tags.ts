import type { TagRef } from '$lib/types/track';
import { apiFetch } from './client';

export interface Tag extends TagRef {
	/** How many of the user's tracks carry the tag. */
	track_count: number;
}

export async function listTags(): Promise<Tag[]> {
	const data = await apiFetch<{ tags: Tag[] }>('/tags');
	return data.tags;
}

export async function createTag(name: string, color: string): Promise<Tag> {
	return apiFetch<Tag>('/tags', { method: 'POST', body: JSON.stringify({ name, color }) });
}

export async function updateTag(id: number, changes: { name?: string; color?: string }): Promise<void> {
	await apiFetch(`/tags/${id}`, { method: 'PATCH', body: JSON.stringify(changes) });
}

export async function deleteTag(id: number): Promise<void> {
	await apiFetch(`/tags/${id}`, { method: 'DELETE' });
}

/** Replaces a track's tags; returns them as saved, sorted by name. */
export async function setTrackTags(trackId: number, tagIds: number[]): Promise<TagRef[]> {
	const data = await apiFetch<{ tags: TagRef[] }>(`/tracks/${trackId}/tags`, {
		method: 'PUT',
		body: JSON.stringify({ tag_ids: tagIds })
	});
	return data.tags;
}
