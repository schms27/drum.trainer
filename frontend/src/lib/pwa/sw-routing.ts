// Request routing for the service worker (src/service-worker.ts), kept as a
// pure function so it can be unit-tested.
//
// - cache-first:   content-hashed build output and the precached static files
// - network-first: page navigations; fall back to the cached app shell offline
// - passthrough:   everything else, the service worker doesn't touch it. This
//                  covers all of /api (OAuth redirects, uploads, backing-track
//                  audio, JSON), non-GET requests and other origins.

export type Strategy = 'cache-first' | 'network-first' | 'passthrough';

export type RouteInput = {
	url: URL;
	method: string;
	mode: RequestMode;
	/** The service worker's own origin. */
	origin: string;
	/** Paths of the precached static files (`files` from $service-worker). */
	staticFiles: ReadonlySet<string>;
};

export function routeRequest({ url, method, mode, origin, staticFiles }: RouteInput): Strategy {
	if (method !== 'GET' || url.origin !== origin) return 'passthrough';
	if (url.pathname === '/api' || url.pathname.startsWith('/api/')) return 'passthrough';
	if (url.pathname.startsWith('/_app/immutable/')) return 'cache-first';
	if (staticFiles.has(url.pathname)) return 'cache-first';
	if (mode === 'navigate') return 'network-first';
	return 'passthrough';
}
