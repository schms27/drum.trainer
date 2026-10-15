import { describe, expect, it } from 'vitest';
import { GM_GROUP, INSTRUMENTS, LANES, drumGroup } from './drums';

describe('drums', () => {
	it('maps every GM note to a known instrument', () => {
		for (const group of Object.values(GM_GROUP)) {
			expect(INSTRUMENTS).toContain(group);
		}
	});

	it('falls back to Other for unmapped notes', () => {
		expect(drumGroup(36)).toBe('Bass Drum');
		expect(drumGroup(127)).toBe('Other');
	});

	it('has a lane for every instrument except Other', () => {
		expect(LANES).not.toContain('Other');
		expect(LANES).toHaveLength(INSTRUMENTS.length - 1);
	});
});
