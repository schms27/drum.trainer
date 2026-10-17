import type { NoteMapping } from '$lib/types/track';
import { drumGroup } from '$lib/drums';
import { apiFetch } from './client';

/** How the pads of the user's e-drum kit map to instruments. */
export interface KitProfile {
	id: number;
	name: string;
	mappings: NoteMapping[];
	/** The kit used to score recorded hits. */
	active: boolean;
	created_at: string;
}

export async function listKits(): Promise<KitProfile[]> {
	const data = await apiFetch<{ kits: KitProfile[] }>('/kits');
	return data.kits;
}

export async function createKit(name: string, mappings: NoteMapping[]): Promise<KitProfile> {
	return apiFetch<KitProfile>('/kits', { method: 'POST', body: JSON.stringify({ name, mappings }) });
}

export async function updateKit(id: number, name: string, mappings: NoteMapping[]): Promise<KitProfile> {
	return apiFetch<KitProfile>(`/kits/${id}`, { method: 'PUT', body: JSON.stringify({ name, mappings }) });
}

export async function deleteKit(id: number): Promise<void> {
	await apiFetch<void>(`/kits/${id}`, { method: 'DELETE' });
}

/** Selects the kit used for scoring; null means standard General MIDI numbering. */
export async function setActiveKit(id: number | null): Promise<void> {
	await apiFetch<void>('/kits/active', { method: 'PUT', body: JSON.stringify({ id }) });
}

/**
 * Instrument for a note sent by the kit: the kit profile's mapping if it has
 * one, otherwise the General MIDI default. Mirrors the backend's scoring.
 */
export function kitInstrument(kit: KitProfile | null | undefined, note: number): string {
	return kit?.mappings.find((m) => m.note_number === note)?.instrument ?? drumGroup(note);
}
