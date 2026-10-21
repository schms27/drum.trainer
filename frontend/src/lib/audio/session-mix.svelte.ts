// Live mixer levels of the session page, read by the metronome while it
// schedules clicks and by the mixer controls. One session page is open at a
// time, so a single shared state is enough.

import { DEFAULT_LEVEL, levelToGain, type MixLevels } from './mix';

export const sessionMix = $state<MixLevels>({ metronome: DEFAULT_LEVEL, pilots: {} });

/** Gain for the next metronome click. */
export function metronomeGain(): number {
	return levelToGain(sessionMix.metronome);
}
