// Starter patterns for a track created from scratch, so the grid isn't blank.

export type PatternId = 'empty' | 'rock' | 'eighths' | 'four-on-the-floor';

export const PATTERNS: { id: PatternId; label: string }[] = [
	{ id: 'empty', label: 'Empty' },
	{ id: 'rock', label: 'Basic rock beat' },
	{ id: 'eighths', label: '8th-note hi-hat' },
	{ id: 'four-on-the-floor', label: 'Four on the floor' }
];

const KICK = 36;
const SNARE = 38;
const HH_CLOSED = 42;
const HH_OPEN = 46;

/**
 * Notes of a pattern repeated over `bars` bars. Backbeats fall on every
 * second beat (2 and 4 in 4/4).
 */
export function patternNotes(
	id: PatternId,
	bpm: number,
	beatsPerBar: number,
	bars: number
): { note_number: number; time_ms: number; velocity: number }[] {
	const beatMs = 60000 / bpm;
	const out: { note_number: number; time_ms: number; velocity: number }[] = [];
	const add = (note: number, beat: number, velocity: number) =>
		out.push({ note_number: note, time_ms: Math.round(beat * beatMs * 1000) / 1000, velocity });
	for (let bar = 0; bar < bars; bar++) {
		for (let b = 0; b < beatsPerBar; b++) {
			const beat = bar * beatsPerBar + b;
			const backbeat = b % 2 === 1;
			switch (id) {
				case 'rock':
					add(backbeat ? SNARE : KICK, beat, 100);
					add(HH_CLOSED, beat, 90);
					add(HH_CLOSED, beat + 0.5, 70);
					break;
				case 'eighths':
					add(HH_CLOSED, beat, 90);
					add(HH_CLOSED, beat + 0.5, 70);
					break;
				case 'four-on-the-floor':
					add(KICK, beat, 110);
					if (backbeat) add(SNARE, beat, 100);
					add(HH_OPEN, beat + 0.5, 80);
					break;
			}
		}
	}
	return out;
}
