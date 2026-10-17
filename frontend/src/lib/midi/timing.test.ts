import { describe, expect, it } from 'vitest';
import { audioTimeToPerfMs, computeCalibration, median } from './timing';

describe('audioTimeToPerfMs', () => {
	it('uses getOutputTimestamp when available', () => {
		const ctx = {
			currentTime: 10,
			getOutputTimestamp: () => ({ contextTime: 9.95, performanceTime: 5000 })
		};
		// 2 s after contextTime 9.95 → 5000 + 2050 ms
		expect(audioTimeToPerfMs(ctx, 12, 9999)).toBeCloseTo(7050);
	});

	it('falls back to currentTime and now', () => {
		const ctx = { currentTime: 10 };
		expect(audioTimeToPerfMs(ctx, 10.5, 1000)).toBeCloseTo(1500);
	});

	it('falls back when the output timestamp is not ready yet', () => {
		const ctx = { currentTime: 1, getOutputTimestamp: () => ({ contextTime: 0, performanceTime: 0 }) };
		expect(audioTimeToPerfMs(ctx, 2, 300)).toBeCloseTo(1300);
	});
});

describe('median', () => {
	it('handles odd and even lengths', () => {
		expect(median([3, 1, 2])).toBe(2);
		expect(median([4, 1, 2, 3])).toBe(2.5);
	});
});

describe('computeCalibration', () => {
	const beatMs = 600;
	const clicks = Array.from({ length: 16 }, (_, i) => 1000 + i * beatMs);

	it('returns the typical offset, ignoring the warm-up clicks', () => {
		// Warm-up taps are way off; the rest arrive 25 ms late, ±3 ms.
		const taps = clicks.map((c, i) => (i < 4 ? c + 200 : c + 25 + (i % 2 ? 3 : -3)));
		const result = computeCalibration(taps, clicks, beatMs);
		expect(result).not.toBeNull();
		expect(result!.offsetMs).toBe(25);
		expect(result!.spreadMs).toBe(3);
		expect(result!.count).toBe(12);
	});

	it('pairs each click with at most one tap', () => {
		const taps = clicks.slice(4).flatMap((c) => [c + 10, c + 12]);
		const result = computeCalibration(taps, clicks, beatMs);
		expect(result!.count).toBe(12);
	});

	it('returns null when there are too few taps', () => {
		expect(computeCalibration([clicks[5] + 10, clicks[6] + 10], clicks, beatMs)).toBeNull();
	});
});
