import type { Track, TrackWithNotes, NoteMapping } from '$lib/types/track';
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
	/** MIDI channel 1–16, or 'auto' to let the server find the drum part. */
	channel: number | 'auto',
	onProgress?: (pct: number) => void
): Promise<TrackWithNotes> {
	return new Promise((resolve, reject) => {
		const form = new FormData();
		form.append('file', file);
		form.append('channel', String(channel));

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

/** The MIDI file as uploaded. Only for the owner's own, non-copied tracks. */
export function trackFileUrl(trackId: number): string {
	return `/api/v1/tracks/${trackId}/file`;
}

/** The track's drum notes rendered as a new MIDI file (drum channel only). */
export function trackExportUrl(trackId: number): string {
	return `/api/v1/tracks/${trackId}/export.mid`;
}

export async function deleteTrack(id: number): Promise<void> {
	await apiFetch(`/tracks/${id}`, { method: 'DELETE' });
}

export async function setPracticeBpm(trackId: number, bpm: number | null): Promise<void> {
	await apiFetch(`/tracks/${trackId}/practice-bpm`, {
		method: 'PATCH',
		body: JSON.stringify({ bpm })
	});
}

/** Saves the session mixer for a track. null clears a level, so the kit default applies. */
export async function setTrackMix(
	trackId: number,
	metronomeVolume: number | null,
	pilots: { id: number; volume: number | null }[]
): Promise<void> {
	await apiFetch(`/tracks/${trackId}/mix`, {
		method: 'PUT',
		body: JSON.stringify({ metronome_volume: metronomeVolume, pilots })
	});
}

export async function getTrackMapping(trackId: number): Promise<NoteMapping[]> {
	const data = await apiFetch<{ mappings: NoteMapping[] }>(`/tracks/${trackId}/mapping`);
	return data.mappings;
}

export async function updateTrackMapping(
	trackId: number,
	mappings: NoteMapping[]
): Promise<NoteMapping[]> {
	const data = await apiFetch<{ mappings: NoteMapping[] }>(`/tracks/${trackId}/mapping`, {
		method: 'PUT',
		body: JSON.stringify({ mappings })
	});
	return data.mappings;
}
