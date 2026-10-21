import { describe, expect, it } from 'vitest';
import type { Track } from '$lib/types/track';
import { filterFromParams, filterToParams, filterTracks, fold, highlight, parseQuery } from './library-filter';

const warm = { id: 1, name: 'Warm-up', color: 'teal' };
const gig = { id: 2, name: 'Gig set', color: 'pink' };

function track(id: number, name: string, extra: Partial<Track> = {}): Track {
	return {
		id,
		name,
		filename: `${name}.mid`,
		created_at: `2026-10-${String(10 + id).padStart(2, '0')}T10:00:00Z`,
		tags: [],
		stats: { sessions: 0, best_hit_rate: null, last_hit_rate: null, last_practiced_at: null, recent_hit_rates: [] },
		...extra
	} as Track;
}

const tracks = [
	track(1, 'Crazy in Love – Beyoncé', { tags: [gig] }),
	track(2, 'Back in Black', { tags: [warm, gig], filename: 'acdc_bib.mid' }),
	track(3, 'paradiddles', {
		tags: [warm],
		stats: { sessions: 9, best_hit_rate: 1, last_hit_rate: 1, last_practiced_at: '2026-10-20T10:00:00Z', recent_hit_rates: [] }
	}),
	track(4, 'Billie Jean', {
		stats: { sessions: 2, best_hit_rate: 1, last_hit_rate: 1, last_practiced_at: '2026-10-21T10:00:00Z', recent_hit_rates: [] }
	})
];
const names = (list: Track[]) => list.map((t) => t.name);
const run = (q: string, extra: { tagIds?: number[]; sort?: 'newest' | 'name' | 'recent' | 'most' } = {}) =>
	names(filterTracks(tracks, { q, tagIds: extra.tagIds ?? [], sort: extra.sort ?? 'newest' }));

describe('library filter', () => {
	it('folds case and accents', () => {
		expect(fold('Beyoncé')).toBe('beyonce');
		expect(run('beyonce')).toEqual(['Crazy in Love – Beyoncé']);
	});

	it('needs every word, in any order, across name, filename and tags', () => {
		expect(run('black back')).toEqual(['Back in Black']);
		expect(run('acdc')).toEqual(['Back in Black']);
		expect(run('warm')).toEqual(['paradiddles', 'Back in Black']);
		expect(run('in love jean')).toEqual([]);
	});

	it('filters by tag ids and tag: terms, all of them', () => {
		expect(run('', { tagIds: [1] })).toEqual(['paradiddles', 'Back in Black']);
		expect(run('', { tagIds: [1, 2] })).toEqual(['Back in Black']);
		expect(run('tag:"gig set"')).toEqual(['Back in Black', 'Crazy in Love – Beyoncé']);
		expect(run('tag:warm-up back')).toEqual(['Back in Black']);
		expect(parseQuery('tag:"Gig Set" back')).toEqual({ words: ['back'], tags: ['gig set'] });
	});

	it('sorts', () => {
		expect(run('')).toEqual(['Billie Jean', 'paradiddles', 'Back in Black', 'Crazy in Love – Beyoncé']);
		expect(run('', { sort: 'name' })).toEqual(['Back in Black', 'Billie Jean', 'Crazy in Love – Beyoncé', 'paradiddles']);
		expect(run('', { sort: 'recent' }).slice(0, 2)).toEqual(['Billie Jean', 'paradiddles']);
		expect(run('', { sort: 'most' }).slice(0, 2)).toEqual(['paradiddles', 'Billie Jean']);
	});

	it('highlights matches in the original text', () => {
		expect(highlight('Beyoncé live', ['yonce'])).toEqual([
			{ text: 'Be', match: false },
			{ text: 'yoncé', match: true },
			{ text: ' live', match: false }
		]);
		expect(highlight('abc', [])).toEqual([{ text: 'abc', match: false }]);
	});

	it('round-trips through the URL', () => {
		const filter = { q: 'back in', tagIds: [1, 2], sort: 'name' as const };
		expect(filterFromParams(filterToParams(filter))).toEqual(filter);
		expect(filterToParams({ q: '', tagIds: [], sort: 'newest' }).toString()).toBe('');
		expect(filterFromParams(new URLSearchParams('sort=bogus&tag=x')).sort).toBe('newest');
	});
});
