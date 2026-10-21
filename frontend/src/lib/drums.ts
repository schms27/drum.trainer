// Shared drum-kit vocabulary. Mirrors backend/midi/drummap.go — keep the two
// in sync (DrumMap ↔ GM_GROUP, Instruments ↔ INSTRUMENTS).

/** GM channel-10 note number → instrument group. */
export const GM_GROUP: Readonly<Record<number, string>> = {
	35: 'Bass Drum', 36: 'Bass Drum',
	37: 'Snare', 38: 'Snare', 40: 'Snare',
	42: 'Hi-Hat', 44: 'Hi-Hat', 46: 'Hi-Hat',
	49: 'Crash', 57: 'Crash',
	51: 'Ride', 53: 'Ride', 59: 'Ride',
	50: 'Tom (high)',
	47: 'Tom (mid)', 48: 'Tom (mid)',
	43: 'Tom (low)', 45: 'Tom (low)',
	41: 'Tom (floor)'
};

/** Every instrument group a note can be mapped to, in display order. */
export const INSTRUMENTS = [
	'Bass Drum', 'Snare', 'Hi-Hat', 'Crash', 'Ride',
	'Tom (high)', 'Tom (mid)', 'Tom (low)', 'Tom (floor)',
	'Other'
] as const;

export type Instrument = (typeof INSTRUMENTS)[number];

/** Timeline lanes: all instruments except the catch-all "Other". */
export const LANES = INSTRUMENTS.filter((i) => i !== 'Other');

/** Instrument group for a GM note number ("Other" if unmapped). */
export function drumGroup(note: number): string {
	return GM_GROUP[note] ?? 'Other';
}

/** Note-level names, more specific than the instrument groups (e.g. rim vs. snare). */
const NOTE_NAMES: Readonly<Record<number, string>> = {
	35: 'Bass Drum', 36: 'Bass Drum',
	37: 'Side Stick', 38: 'Snare', 40: 'Snare (rim)',
	42: 'Hi-Hat (closed)', 44: 'Hi-Hat (pedal)', 46: 'Hi-Hat (open)',
	49: 'Crash 1', 57: 'Crash 2',
	51: 'Ride', 53: 'Ride Bell', 59: 'Ride 2',
	50: 'Tom (high)',
	47: 'Tom (mid)', 48: 'Tom (mid)',
	43: 'Tom (low)', 45: 'Tom (low)',
	41: 'Tom (floor)'
};

/** Note numbers that have a name, ascending. */
export const NAMED_NOTES: readonly number[] = Object.keys(NOTE_NAMES)
	.map(Number)
	.sort((a, b) => a - b);

/** Human-readable name for a MIDI note from a drum kit. */
export function noteLabel(note: number): string {
	return NOTE_NAMES[note] ?? `Note ${note}`;
}
