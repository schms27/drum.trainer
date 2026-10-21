import { describe, expect, it } from 'vitest';
import { patternNotes } from './patterns';

describe('patternNotes', () => {
	it('builds a rock beat with backbeats on 2 and 4', () => {
		const notes = patternNotes('rock', 120, 4, 1);
		const at = (ms: number) => notes.filter((n) => n.time_ms === ms).map((n) => n.note_number).sort();
		expect(at(0)).toEqual([36, 42]);
		expect(at(500)).toEqual([38, 42]);
		expect(at(250)).toEqual([42]);
		expect(notes).toHaveLength(4 * 3);
	});

	it('repeats over bars and is empty for the empty pattern', () => {
		expect(patternNotes('eighths', 120, 3, 2)).toHaveLength(2 * 3 * 2);
		expect(patternNotes('empty', 120, 4, 4)).toEqual([]);
	});
});
