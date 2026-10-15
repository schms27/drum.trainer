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
