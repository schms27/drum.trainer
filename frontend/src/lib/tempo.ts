// Practice tempo of a take, as a whole percentage of the track's native tempo
// (`tempo_percent` from the API; null when the take predates tempo tracking).

export type TempoFilter = 'all' | 'original' | 'slower' | 'faster';

export const TEMPO_FILTERS: { value: TempoFilter; label: string }[] = [
	{ value: 'all', label: 'Any tempo' },
	{ value: 'original', label: 'Original tempo (100 %)' },
	{ value: 'slower', label: 'Slower' },
	{ value: 'faster', label: 'Faster' }
];

export function parseTempoFilter(value: string | null | undefined): TempoFilter {
	return TEMPO_FILTERS.some((f) => f.value === value) ? (value as TempoFilter) : 'all';
}

/** Inclusive percent bounds for the API's tempo_min / tempo_max. */
export function tempoRange(filter: TempoFilter): { min?: number; max?: number } {
	switch (filter) {
		case 'original':
			return { min: 100, max: 100 };
		case 'slower':
			return { max: 99 };
		case 'faster':
			return { min: 101 };
		default:
			return {};
	}
}

export type TempoKind = 'original' | 'slower' | 'faster';

export function tempoKind(percent: number | null | undefined): TempoKind | null {
	if (percent == null) return null;
	return percent === 100 ? 'original' : percent < 100 ? 'slower' : 'faster';
}

export function matchesTempo(percent: number | null | undefined, filter: TempoFilter): boolean {
	return filter === 'all' || tempoKind(percent) === filter;
}

export function formatTempo(percent: number | null | undefined): string {
	return percent == null ? '–' : `${percent} %`;
}

/** Longer description for tooltips and screen readers. */
export function describeTempo(percent: number | null | undefined, practiceBpm?: number | null): string {
	if (percent == null) return 'Tempo not recorded';
	const bpm = practiceBpm ? ` (${Math.round(practiceBpm)} BPM)` : '';
	return percent === 100 ? `Original tempo${bpm}` : `${percent} % of the original tempo${bpm}`;
}
