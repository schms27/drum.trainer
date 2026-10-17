import { describe, expect, it } from 'vitest';
import {
	describeTiming,
	formatOffset,
	formatPercent,
	formatRelativeDate,
	hitRateTone,
	summarizeTake,
	timingTone
} from './score-format';

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

describe('formatRelativeDate', () => {
	const now = new Date(2026, 9, 16, 20, 0);
	it('uses days, then weeks and months', () => {
		expect(formatRelativeDate(new Date(2026, 9, 16, 8).toISOString(), now)).toMatch(/today/i);
		expect(formatRelativeDate(new Date(2026, 9, 15, 23).toISOString(), now)).toMatch(/yesterday/i);
		expect(formatRelativeDate(new Date(2026, 9, 13).toISOString(), now)).toMatch(/3 days/);
		expect(formatRelativeDate(new Date(2026, 9, 1).toISOString(), now)).toMatch(/2 weeks/);
		expect(formatRelativeDate(new Date(2026, 6, 1).toISOString(), now)).toMatch(/3 months/);
	});
});

describe('summarizeTake', () => {
	const base = { hitRate: 0.9, matched: 45, target: 50, avgOffsetMs: 2 };

	it('describes notes and timing', () => {
		expect(summarizeTake(base)).toBe('You hit 45 of 50 notes and were right on the beat.');
		expect(summarizeTake({ ...base, avgOffsetMs: 12 })).toBe('You hit 45 of 50 notes and played slightly late on average.');
	});

	it('compares with the previous take', () => {
		expect(summarizeTake({ ...base, previousHitRate: 0.85, bestHitRate: 0.95 })).toContain('5 points better');
		expect(summarizeTake({ ...base, previousHitRate: 0.91, bestHitRate: 0.95 })).toContain('1 point below');
		expect(summarizeTake({ ...base, previousHitRate: 0.9, bestHitRate: 0.95 })).toContain('Same as');
	});

	it('celebrates a new best', () => {
		expect(summarizeTake({ ...base, previousHitRate: 0.8, bestHitRate: 0.85 })).toContain('New personal best!');
	});

	it('names the weakest instrument', () => {
		expect(summarizeTake({ ...base, weakest: { instrument: 'Hi-Hat', missed: 4 } })).toContain('Most missed: Hi-Hat (4).');
	});
});
