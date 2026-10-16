import { describe, expect, it } from 'vitest';
import { describeTiming, formatOffset, formatPercent, hitRateTone, timingTone } from './score-format';

describe('score-format', () => {
	it('classifies timing offsets', () => {
		expect(timingTone(0)).toBe('good');
		expect(timingTone(-20)).toBe('good');
		expect(timingTone(35)).toBe('warn');
		expect(timingTone(-51)).toBe('bad');
	});

	it('classifies hit rates', () => {
		expect(hitRateTone(0.95)).toBe('good');
		expect(hitRateTone(0.8)).toBe('warn');
		expect(hitRateTone(0.5)).toBe('bad');
	});

	it('formats offsets with a sign', () => {
		expect(formatOffset(7.9)).toBe('+8 ms');
		expect(formatOffset(-12.2)).toBe('−12 ms');
		expect(formatOffset(0.3)).toBe('±0 ms');
		expect(formatOffset(6.25, 1)).toBe('+6.3 ms');
	});

	it('formats percentages', () => {
		expect(formatPercent(0.875)).toBe('88%');
	});

	it('describes average timing in plain words', () => {
		expect(describeTiming(2)).toBe('right on the beat');
		expect(describeTiming(8)).toBe('slightly late');
		expect(describeTiming(-15)).toBe('slightly early');
		expect(describeTiming(30)).toContain('late');
	});
});
