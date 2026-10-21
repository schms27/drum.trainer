// Pure helpers for the track editor. Notes are drum hits at a time in ms of
// the track's native tempo; `id` only identifies a note inside the editor.

export interface EditorNote {
	id: number;
	note_number: number;
	time_ms: number;
	velocity: number;
}

let nextId = 1;

export function withIds(notes: { note_number: number; time_ms: number; velocity: number }[]): EditorNote[] {
	return notes.map((n) => ({ id: nextId++, note_number: n.note_number, time_ms: n.time_ms, velocity: n.velocity }));
}

export function newNote(note_number: number, time_ms: number, velocity = DEFAULT_VELOCITY): EditorNote {
	return { id: nextId++, note_number, time_ms, velocity };
}

export const DEFAULT_VELOCITY = 100;

/** Notes closer than this on the same note number count as the same hit. */
export const SAME_HIT_MS = 1;

export function sortNotes(notes: EditorNote[]): EditorNote[] {
	return [...notes].sort((a, b) => a.time_ms - b.time_ms || a.note_number - b.note_number);
}

export function countByNote(notes: { note_number: number }[]): Map<number, number> {
	const counts = new Map<number, number>();
	for (const n of notes) counts.set(n.note_number, (counts.get(n.note_number) ?? 0) + 1);
	return counts;
}

/**
 * Merges hits that ended up on the same note number at the same time (for
 * example after mapping hi-hat and ride to one cymbal), keeping the loudest.
 */
export function mergeDuplicates(notes: EditorNote[]): EditorNote[] {
	const sorted = sortNotes(notes);
	const out: EditorNote[] = [];
	const lastByNote = new Map<number, EditorNote>();
	for (const n of sorted) {
		const prev = lastByNote.get(n.note_number);
		if (prev && n.time_ms - prev.time_ms < SAME_HIT_MS) {
			prev.velocity = Math.max(prev.velocity, n.velocity);
			continue;
		}
		const copy = { ...n };
		out.push(copy);
		lastByNote.set(n.note_number, copy);
	}
	return out;
}

/** What to do with every hit of a note number: map it to another one, or delete it. */
export type NoteAction = number | 'delete';

/** Applies note-number replacements and deletions to the whole track. */
export function applyNoteActions(notes: EditorNote[], actions: Map<number, NoteAction>): EditorNote[] {
	const mapped: EditorNote[] = [];
	for (const n of notes) {
		const action = actions.get(n.note_number);
		if (action === 'delete') continue;
		mapped.push(action === undefined ? n : { ...n, note_number: action });
	}
	return mergeDuplicates(mapped);
}

/** Grid divisions, as a fraction of a beat (a quarter note). */
export const DIVISIONS = [
	{ value: '1/4', label: '1/4', beats: 1 },
	{ value: '1/8', label: '1/8', beats: 1 / 2 },
	{ value: '1/16', label: '1/16', beats: 1 / 4 },
	{ value: '1/32', label: '1/32', beats: 1 / 8 },
	{ value: '1/8T', label: '1/8 triplet', beats: 1 / 3 },
	{ value: '1/16T', label: '1/16 triplet', beats: 1 / 6 }
] as const;

export type Division = (typeof DIVISIONS)[number]['value'];

export function stepMs(bpm: number, division: Division): number {
	const beats = DIVISIONS.find((d) => d.value === division)?.beats ?? 1 / 4;
	return (60000 / bpm) * beats;
}

export function snap(ms: number, step: number): number {
	return Math.max(0, Math.round(ms / step) * step);
}

/** Moves notes by `deltaMs` and `deltaRows` rows of `rows` (note numbers, top to bottom). */
export function moveNotes(
	notes: EditorNote[],
	ids: Set<number>,
	deltaMs: number,
	deltaRows: number,
	rows: number[]
): EditorNote[] {
	const moved = notes.map((n) => {
		if (!ids.has(n.id)) return n;
		const row = rows.indexOf(n.note_number);
		const target = row < 0 ? n.note_number : rows[Math.max(0, Math.min(rows.length - 1, row + deltaRows))];
		return { ...n, note_number: target, time_ms: Math.max(0, n.time_ms + deltaMs) };
	});
	return mergeDuplicates(moved);
}

/** Bars needed to hold `durationMs` (at least one). */
export function barsFor(durationMs: number, bpm: number, beatsPerBar: number): number {
	return Math.max(1, Math.ceil(durationMs / ((60000 / bpm) * beatsPerBar) - 1e-6));
}

/**
 * Repeats bars `from`–`to` (1-based, inclusive) `times` more times right
 * after bar `to`, pushing later hits back. Returns the notes and how many
 * bars were added.
 */
export function repeatBars(
	notes: EditorNote[],
	barMs: number,
	from: number,
	to: number,
	times: number
): { notes: EditorNote[]; addedBars: number } {
	if (from < 1 || to < from || times < 1) return { notes, addedBars: 0 };
	const start = (from - 1) * barMs;
	const end = to * barMs;
	const span = end - start;
	const shift = span * times;
	// A tiny epsilon keeps hits stored a hair before a bar line in their bar.
	const inRange = (n: EditorNote) => n.time_ms >= start - 1e-6 && n.time_ms < end - 1e-6;
	const out: EditorNote[] = notes.map((n) => (n.time_ms >= end - 1e-6 ? { ...n, time_ms: n.time_ms + shift } : n));
	for (let k = 1; k <= times; k++) {
		for (const n of notes.filter(inRange)) out.push(newNote(n.note_number, n.time_ms + span * k, n.velocity));
	}
	return { notes: sortNotes(out), addedBars: (to - from + 1) * times };
}

/** Undo/redo history of note lists. */
export class History {
	private past: EditorNote[][] = [];
	private future: EditorNote[][] = [];

	constructor(private limit = 200) {}

	/** Records `previous` before a change. */
	push(previous: EditorNote[]): void {
		this.past.push(previous);
		if (this.past.length > this.limit) this.past.shift();
		this.future = [];
	}

	undo(current: EditorNote[]): EditorNote[] | null {
		const prev = this.past.pop();
		if (!prev) return null;
		this.future.push(current);
		return prev;
	}

	redo(current: EditorNote[]): EditorNote[] | null {
		const next = this.future.pop();
		if (!next) return null;
		this.past.push(current);
		return next;
	}

	get canUndo(): boolean {
		return this.past.length > 0;
	}

	get canRedo(): boolean {
		return this.future.length > 0;
	}
}
