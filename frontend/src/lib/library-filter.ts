// Search, tag filter and sort for the library. Runs in the browser: the whole
// library is already loaded, so results update instantly while typing.

import type { Track } from '$lib/types/track';

export type LibrarySort = 'newest' | 'name' | 'recent' | 'most';

export const SORTS: { value: LibrarySort; label: string }[] = [
	{ value: 'newest', label: 'Newest first' },
	{ value: 'name', label: 'Name A–Z' },
	{ value: 'recent', label: 'Recently practised' },
	{ value: 'most', label: 'Most practised' }
];

export interface LibraryFilter {
	q: string;
	/** Selected tag ids; a track needs all of them. */
	tagIds: number[];
	sort: LibrarySort;
}

/** Lower case without accents, so "Beyonce" finds "Beyoncé". */
export function fold(s: string): string {
	return s.normalize('NFD').replace(/\p{M}/gu, '').toLowerCase();
}

/** Splits a query into plain words and `tag:` terms (`tag:"gig set"` for spaces). */
export function parseQuery(q: string): { words: string[]; tags: string[] } {
	const words: string[] = [];
	const tags: string[] = [];
	for (const m of q.matchAll(/(tag:)?(?:"([^"]*)"|(\S+))/gi)) {
		const text = fold((m[2] ?? m[3] ?? '').trim());
		if (!text) continue;
		(m[1] ? tags : words).push(text);
	}
	return { words, tags };
}

function matches(track: Track, words: string[], tagTerms: string[], tagIds: number[]): boolean {
	const tags = track.tags ?? [];
	const tagNames = tags.map((t) => fold(t.name));
	if (!tagIds.every((id) => tags.some((t) => t.id === id))) return false;
	if (!tagTerms.every((term) => tagNames.includes(term))) return false;
	const haystack = [fold(track.name), fold(track.filename ?? ''), ...tagNames].join('\n');
	return words.every((w) => haystack.includes(w));
}

const byName = (a: Track, b: Track) => a.name.localeCompare(b.name, undefined, { sensitivity: 'base', numeric: true });

export function filterTracks(tracks: Track[], filter: LibraryFilter): Track[] {
	const { words, tags } = parseQuery(filter.q);
	const out = tracks.filter((t) => matches(t, words, tags, filter.tagIds));
	const time = (s: string | null | undefined) => (s ? Date.parse(s) : -Infinity);
	switch (filter.sort) {
		case 'name':
			return out.sort(byName);
		case 'recent':
			return out.sort((a, b) => time(b.stats?.last_practiced_at) - time(a.stats?.last_practiced_at) || byName(a, b));
		case 'most':
			return out.sort((a, b) => (b.stats?.sessions ?? 0) - (a.stats?.sessions ?? 0) || byName(a, b));
		default:
			return out.sort((a, b) => time(b.created_at) - time(a.created_at) || b.id - a.id);
	}
}

/** Splits `text` into parts, marking where any of the (folded) words match. */
export function highlight(text: string, words: string[]): { text: string; match: boolean }[] {
	if (!words.length) return [{ text, match: false }];
	// Fold character by character, remembering where each folded char came from.
	let folded = '';
	const origin: number[] = [];
	for (let i = 0; i < text.length; i++) {
		const f = fold(text[i]);
		for (let k = 0; k < f.length; k++) origin.push(i);
		folded += f;
	}
	const marked = new Array<boolean>(text.length).fill(false);
	for (const w of words) {
		for (let at = folded.indexOf(w); at >= 0; at = folded.indexOf(w, at + 1)) {
			for (let k = at; k < at + w.length; k++) marked[origin[k]] = true;
		}
	}
	const parts: { text: string; match: boolean }[] = [];
	for (let i = 0; i < text.length; i++) {
		const last = parts.at(-1);
		if (last && last.match === marked[i]) last.text += text[i];
		else parts.push({ text: text[i], match: marked[i] });
	}
	return parts;
}

/** Reads the filter from the URL (`?q=…&tag=1&tag=2&sort=name`). */
export function filterFromParams(params: URLSearchParams): LibraryFilter {
	const sort = params.get('sort');
	return {
		q: params.get('q') ?? '',
		tagIds: params.getAll('tag').map(Number).filter((n) => Number.isInteger(n) && n > 0),
		sort: SORTS.some((s) => s.value === sort) ? (sort as LibrarySort) : 'newest'
	};
}

export function filterToParams(filter: LibraryFilter): URLSearchParams {
	const params = new URLSearchParams();
	if (filter.q.trim()) params.set('q', filter.q);
	for (const id of filter.tagIds) params.append('tag', String(id));
	if (filter.sort !== 'newest') params.set('sort', filter.sort);
	return params;
}
