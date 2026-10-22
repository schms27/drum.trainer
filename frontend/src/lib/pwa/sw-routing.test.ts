import { describe, expect, it } from 'vitest';
import { routeRequest } from './sw-routing';

const origin = 'https://drums.example';
const staticFiles = new Set(['/favicon.svg', '/manifest.webmanifest', '/icons/icon-192.png']);

function route(path: string, opts: { method?: string; mode?: RequestMode; origin?: string } = {}) {
	return routeRequest({
		url: new URL(path, opts.origin ?? origin),
		method: opts.method ?? 'GET',
		mode: opts.mode ?? 'cors',
		origin,
		staticFiles
	});
}

describe('routeRequest', () => {
	it('serves hashed build output from the cache', () => {
		expect(route('/_app/immutable/entry/start.abc123.js')).toBe('cache-first');
		expect(route('/_app/immutable/assets/0.def456.css', { mode: 'no-cors' })).toBe('cache-first');
	});

	it('serves precached static files from the cache', () => {
		expect(route('/favicon.svg')).toBe('cache-first');
		expect(route('/manifest.webmanifest')).toBe('cache-first');
		expect(route('/icons/icon-192.png')).toBe('cache-first');
	});

	it('uses network-first for page navigations', () => {
		expect(route('/', { mode: 'navigate' })).toBe('network-first');
		expect(route('/session/12', { mode: 'navigate' })).toBe('network-first');
		expect(route('/s/abc?x=1', { mode: 'navigate' })).toBe('network-first');
	});

	it('never touches the API, including OAuth navigations', () => {
		expect(route('/api/v1/tracks')).toBe('passthrough');
		expect(route('/api/v1/auth/google/login?next=%2F', { mode: 'navigate' })).toBe('passthrough');
		expect(route('/api/v1/auth/github/callback?code=x&state=y', { mode: 'navigate' })).toBe('passthrough');
		expect(route('/api/v1/tracks/3/pilots/1/audio')).toBe('passthrough');
		expect(route('/api', { mode: 'navigate' })).toBe('passthrough');
	});

	it('does not mistake look-alike paths for the API', () => {
		expect(route('/apiary', { mode: 'navigate' })).toBe('network-first');
	});

	it('never touches non-GET requests', () => {
		expect(route('/api/v1/sessions', { method: 'POST' })).toBe('passthrough');
		expect(route('/_app/immutable/x.js', { method: 'HEAD' })).toBe('passthrough');
		expect(route('/', { method: 'POST', mode: 'navigate' })).toBe('passthrough');
	});

	it('receives files shared to the app', () => {
		expect(route('/share-target', { method: 'POST', mode: 'navigate' })).toBe('share-target');
		expect(route('/share-target', { mode: 'navigate' })).toBe('network-first');
		expect(route('/share-target', { method: 'POST', origin: 'https://evil.example' })).toBe('passthrough');
	});

	it('never touches other origins', () => {
		expect(route('/_app/immutable/x.js', { origin: 'https://cdn.example' })).toBe('passthrough');
		expect(route('/', { origin: 'https://accounts.google.com', mode: 'navigate' })).toBe('passthrough');
	});

	it('leaves the version check and other requests alone', () => {
		expect(route('/_app/version.json')).toBe('passthrough');
		expect(route('/logo-unknown.png')).toBe('passthrough');
	});
});
