// Clock helpers for lining up MIDI input with Web Audio output.
//
// MIDI events carry `event.timeStamp` on the performance.now() clock, while
// clicks and backing tracks are scheduled on an AudioContext clock. To score a
// take that starts after a count-in, both must be expressed on one clock.

/** Minimal slice of AudioContext used here (keeps the helpers testable). */
export interface AudioClock {
	currentTime: number;
	getOutputTimestamp?: () => AudioTimestamp;
}

/**
 * Converts an AudioContext time (seconds) to the performance.now() clock (ms),
 * i.e. the moment that sample is expected to be heard. Uses
 * getOutputTimestamp() when available, which accounts for output latency.
 */
export function audioTimeToPerfMs(ctx: AudioClock, audioTime: number, now = performance.now()): number {
	const ts = ctx.getOutputTimestamp?.();
	if (ts && ts.contextTime !== undefined && ts.performanceTime !== undefined && ts.performanceTime > 0) {
		return ts.performanceTime + (audioTime - ts.contextTime) * 1000;
	}
	return now + (audioTime - ctx.currentTime) * 1000;
}

export function median(values: number[]): number {
	if (values.length === 0) return NaN;
	const sorted = [...values].sort((a, b) => a - b);
	const mid = Math.floor(sorted.length / 2);
	return sorted.length % 2 ? sorted[mid] : (sorted[mid - 1] + sorted[mid]) / 2;
}

export interface CalibrationResult {
	/** Typical delay between hearing a click and the hit arriving (ms). */
	offsetMs: number;
	/** How consistent the taps were (median absolute deviation, ms). */
	spreadMs: number;
	/** Number of taps used. */
	count: number;
}

/**
 * Estimates input latency from a tap-along test. Each tap is paired with the
 * nearest click (within half a beat); the first `skipClicks` clicks are
 * ignored while the player settles in. Returns null with fewer than
 * `minTaps` usable taps.
 */
export function computeCalibration(
	tapTimes: number[],
	clickTimes: number[],
	beatMs: number,
	skipClicks = 4,
	minTaps = 6
): CalibrationResult | null {
	const clicks = clickTimes.slice(skipClicks);
	if (clicks.length === 0) return null;
	const used = new Set<number>();
	const offsets: number[] = [];
	for (const tap of tapTimes) {
		let best = -1;
		let bestDist = Infinity;
		clicks.forEach((c, i) => {
			const d = Math.abs(tap - c);
			if (d < bestDist) {
				bestDist = d;
				best = i;
			}
		});
		if (best < 0 || bestDist > beatMs / 2 || used.has(best)) continue;
		used.add(best);
		offsets.push(tap - clicks[best]);
	}
	if (offsets.length < minTaps) return null;
	const offset = median(offsets);
	return {
		offsetMs: Math.round(offset),
		spreadMs: Math.round(median(offsets.map((o) => Math.abs(o - offset)))),
		count: offsets.length
	};
}
