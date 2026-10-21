import { apiFetch } from './client';

export interface PilotTrack {
	id: number;
	track_id: number;
	name: string;
	filename: string;
	file_size: number;
	mime_type: string;
	/** Saved level (0–1); null = the kit profile's default. */
	volume: number | null;
	created_at: string;
}

export function pilotAudioUrl(trackId: number, pilotId: number): string {
	return `/api/v1/tracks/${trackId}/pilots/${pilotId}/audio`;
}

export async function listPilots(trackId: number): Promise<PilotTrack[]> {
	const data = await apiFetch<{ pilots: PilotTrack[] }>(`/tracks/${trackId}/pilots`);
	return data.pilots;
}

export async function uploadPilot(
	trackId: number,
	name: string,
	file: File,
	onProgress?: (pct: number) => void
): Promise<PilotTrack> {
	return new Promise((resolve, reject) => {
		const form = new FormData();
		form.append('name', name);
		form.append('file', file);

		const xhr = new XMLHttpRequest();
		xhr.open('POST', `/api/v1/tracks/${trackId}/pilots`);

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

export async function deletePilot(trackId: number, pilotId: number): Promise<void> {
	await apiFetch(`/tracks/${trackId}/pilots/${pilotId}`, { method: 'DELETE' });
}
