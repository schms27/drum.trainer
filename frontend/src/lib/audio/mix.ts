// Playback levels for the session page's mixer. Levels are 0–1 as the slider
// shows them; levelToGain turns them into Web Audio gain.

export interface MixLevels {
	metronome: number;
	/** Level per backing track id. */
	pilots: Record<number, number>;
}

/** Built-in level when neither the track nor the kit profile has one. */
export const DEFAULT_LEVEL = 1;

/**
 * Gain for a slider level. Squared, so the lower half of the slider is
 * usable: loudness is perceived roughly logarithmically, and a linear gain
 * crams all the audible change into the top of the range.
 */
export function levelToGain(level: number): number {
	const v = Math.max(0, Math.min(1, level));
	return v * v;
}

/**
 * The levels to start with: the track's own mix, else the kit profile's
 * defaults, else full level.
 */
export function resolveMix(
	track: { metronome_volume: number | null },
	pilots: { id: number; volume: number | null }[],
	kit: { metronome_volume: number | null; pilot_volume: number | null } | null
): MixLevels {
	return {
		metronome: track.metronome_volume ?? kit?.metronome_volume ?? DEFAULT_LEVEL,
		pilots: Object.fromEntries(pilots.map((p) => [p.id, p.volume ?? kit?.pilot_volume ?? DEFAULT_LEVEL]))
	};
}
