import { describe, expect, it } from 'vitest';
import { describeTempo, formatTempo, matchesTempo, parseTempoFilter, tempoKind, tempoRange } from './tempo';

describe('tempo', () => {
	it('classifies takes by tempo', () => {
		expect(tempoKind(100)).toBe('original');
		expect(tempoKind(80)).toBe('slower');
		expect(tempoKind(110)).toBe('faster');
		expect(tempoKind(null)).toBeNull();
	});

	it('only matches unknown tempos when not filtering', () => {
		expect(matchesTempo(null, 'all')).toBe(true);
		expect(matchesTempo(null, 'original')).toBe(false);
		expect(matchesTempo(80, 'slower')).toBe(true);
		expect(matchesTempo(80, 'faster')).toBe(false);
	});

	it('maps filters to API bounds', () => {
		expect(tempoRange('all')).toEqual({});
		expect(tempoRange('original')).toEqual({ min: 100, max: 100 });
		expect(tempoRange('slower')).toEqual({ max: 99 });
		expect(tempoRange('faster')).toEqual({ min: 101 });
	});

	it('parses unknown filter values as all', () => {
		expect(parseTempoFilter('slower')).toBe('slower');
		expect(parseTempoFilter('bogus')).toBe('all');
		expect(parseTempoFilter(null)).toBe('all');
	});

	it('formats tempos', () => {
		expect(formatTempo(80)).toBe('80 %');
		expect(formatTempo(null)).toBe('–');
		expect(describeTempo(80, 96)).toBe('80 % of the original tempo (96 BPM)');
		expect(describeTempo(100)).toBe('Original tempo');
		expect(describeTempo(null)).toBe('Tempo not recorded');
	});
});
