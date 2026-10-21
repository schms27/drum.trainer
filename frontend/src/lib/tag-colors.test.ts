import { describe, expect, it } from 'vitest';
import { TAG_COLOR_KEYS, chipStyle, tagHue } from './tag-colors';

describe('tag colours', () => {
	it('matches the colours the API accepts (backend models.TagColors)', () => {
		expect(TAG_COLOR_KEYS).toEqual(['gray', 'red', 'orange', 'amber', 'green', 'teal', 'sky', 'blue', 'violet', 'pink']);
	});

	it('falls back to gray for unknown colours', () => {
		expect(tagHue('nope')).toBe(tagHue('gray'));
		expect(chipStyle('teal')).toContain(tagHue('teal'));
	});
});
