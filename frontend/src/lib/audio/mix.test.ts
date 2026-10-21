import { describe, expect, it } from 'vitest';
import { levelToGain, resolveMix } from './mix';

describe('mix', () => {
	it('squares levels into gain and clamps them', () => {
		expect(levelToGain(1)).toBe(1);
		expect(levelToGain(0.5)).toBe(0.25);
		expect(levelToGain(0)).toBe(0);
		expect(levelToGain(2)).toBe(1);
		expect(levelToGain(-1)).toBe(0);
	});

	it('prefers the track mix, then the kit defaults, then full level', () => {
		const kit = { metronome_volume: 0.4, pilot_volume: 0.7 };
		const pilots = [
			{ id: 1, volume: 0.2 },
			{ id: 2, volume: null }
		];
		expect(resolveMix({ metronome_volume: 0.9 }, pilots, kit)).toEqual({ metronome: 0.9, pilots: { 1: 0.2, 2: 0.7 } });
		expect(resolveMix({ metronome_volume: null }, pilots, kit)).toEqual({ metronome: 0.4, pilots: { 1: 0.2, 2: 0.7 } });
		expect(resolveMix({ metronome_volume: null }, pilots, null)).toEqual({ metronome: 1, pilots: { 1: 0.2, 2: 1 } });
		expect(resolveMix({ metronome_volume: 0 }, [], kit).metronome).toBe(0);
	});
});
