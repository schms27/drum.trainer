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

/** "today", "yesterday", "3 days ago", "2 weeks ago" … in the user's locale. */
export function formatRelativeDate(iso: string, now = new Date()): string {
	const then = new Date(iso);
	const startOfDay = (d: Date) => new Date(d.getFullYear(), d.getMonth(), d.getDate()).getTime();
	const days = Math.round((startOfDay(now) - startOfDay(then)) / 86_400_000);
	const rtf = new Intl.RelativeTimeFormat(undefined, { numeric: 'auto' });
	if (days < 7) return rtf.format(-days, 'day');
	if (days < 30) return rtf.format(-Math.floor(days / 7), 'week');
	if (days < 365) return rtf.format(-Math.floor(days / 30), 'month');
	return rtf.format(-Math.floor(days / 365), 'year');
}

export interface TakeSummaryInput {
	hitRate: number;
	matched: number;
	target: number;
	avgOffsetMs: number;
	/** Instrument with the most missed notes, if any were missed. */
	weakest?: { instrument: string; missed: number } | null;
	/** Hit rate of the previous take on this track. */
	previousHitRate?: number | null;
	/** Best hit rate on this track before this take. */
	bestHitRate?: number | null;
}

/** A few plain-language sentences describing a take. */
export function summarizeTake(t: TakeSummaryInput): string {
	const parts: string[] = [];
	const timing = describeTiming(t.avgOffsetMs).split(' — ')[0];
	parts.push(
		timing === 'right on the beat'
			? `You hit ${t.matched} of ${t.target} notes and were right on the beat.`
			: `You hit ${t.matched} of ${t.target} notes and played ${timing} on average.`
	);
	if (t.bestHitRate != null && t.hitRate > t.bestHitRate) {
		parts.push('New personal best!');
	} else if (t.previousHitRate != null) {
		const delta = Math.round((t.hitRate - t.previousHitRate) * 100);
		if (delta > 0) parts.push(`That's ${delta} ${delta === 1 ? 'point' : 'points'} better than your last take.`);
		else if (delta < 0) parts.push(`That's ${-delta} ${delta === -1 ? 'point' : 'points'} below your last take.`);
		else parts.push('Same as your last take.');
	}
	if (t.weakest && t.weakest.missed > 0) {
		parts.push(`Most missed: ${t.weakest.instrument} (${t.weakest.missed}).`);
	}
	return parts.join(' ');
}
