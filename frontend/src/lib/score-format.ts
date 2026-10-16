// Shared wording and colour thresholds for presenting scores, so the results
// page, history and live timeline all describe accuracy the same way.

export type Tone = 'good' | 'warn' | 'bad';

/** Timing within this many ms counts as "on time". */
export const ON_TIME_MS = 20;
/** Timing within this many ms counts as "close". */
export const CLOSE_MS = 50;

export function timingTone(offsetMs: number): Tone {
	const abs = Math.abs(offsetMs);
	if (abs <= ON_TIME_MS) return 'good';
	if (abs <= CLOSE_MS) return 'warn';
	return 'bad';
}

export function hitRateTone(rate: number): Tone {
	if (rate >= 0.9) return 'good';
	if (rate >= 0.7) return 'warn';
	return 'bad';
}

/** Tailwind text colour class for a tone. */
export const toneText: Record<Tone, string> = {
	good: 'text-good',
	warn: 'text-warn',
	bad: 'text-bad'
};

/** CSS colour (for SVG) for a tone. */
export const toneColor: Record<Tone, string> = {
	good: 'var(--good)',
	warn: 'var(--warn)',
	bad: 'var(--bad)'
};

export function formatPercent(rate: number): string {
	return `${Math.round(rate * 100)}%`;
}

/** "+8 ms" / "−12 ms" / "±0 ms" */
export function formatOffset(ms: number, digits = 0): string {
	const rounded = Number(ms.toFixed(digits));
	if (rounded === 0) return '±0 ms';
	return `${rounded > 0 ? '+' : '−'}${Math.abs(rounded).toFixed(digits)} ms`;
}

/** Plain-language description of an average timing offset. */
export function describeTiming(avgOffsetMs: number): string {
	const abs = Math.abs(avgOffsetMs);
	if (abs < 5) return 'right on the beat';
	const direction = avgOffsetMs > 0 ? 'late' : 'early';
	if (abs <= ON_TIME_MS) return `slightly ${direction}`;
	return `${direction} — try to ${avgOffsetMs > 0 ? 'anticipate the beat' : 'relax and wait for the beat'}`;
}
