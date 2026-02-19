import type { Track, TrackWithNotes } from '$lib/types/track';
import { apiFetch } from './client';

export async function listTracks(): Promise<Track[]> {
	const data = await apiFetch<{ tracks: Track[] }>('/tracks');
	return data.tracks;
}

export async function getTrack(id: number): Promise<TrackWithNotes> {
	return apiFetch<TrackWithNotes>(`/tracks/${id}`);
}

export async function uploadTrack(
	file: File,
	onProgress?: (pct: number) => void
): Promise<TrackWithNotes> {
	return new Promise((resolve, reject) => {
		const form = new FormData();
		form.append('file', file);

		const xhr = new XMLHttpRequest();
		xhr.open('POST', '/api/v1/tracks');

		if (onProgress) {
			xhr.upload.addEventListener('progress', (e) => {
				if (e.lengthComputable) onProgress(Math.round((e.loaded / e.total) * 100));
			});
		}

		xhr.addEventListener('load', () => {
			if (xhr.status >= 200 && xhr.status < 300) {
				resolve(JSON.parse(xhr.responseText));
			} else {
				let message = `HTTP ${xhr.status}`;
				try {
					const body = JSON.parse(xhr.responseText);
					if (body?.error) message = body.error;
				} catch {
					// ignore
				}
				reject(new Error(message));
			}
		});

		xhr.addEventListener('error', () => reject(new Error('Network error')));
		xhr.send(form);
	});
}

export async function deleteTrack(id: number): Promise<void> {
	await apiFetch(`/tracks/${id}`, { method: 'DELETE' });
}
