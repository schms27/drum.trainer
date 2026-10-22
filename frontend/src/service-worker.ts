/// <reference no-default-lib="true"/>
/// <reference lib="esnext" />
/// <reference lib="webworker" />
/// <reference types="@sveltejs/kit" />

// App-shell service worker. SvelteKit registers it automatically.
// Which requests it handles is decided in $lib/pwa/sw-routing.ts; the API is
// never touched, so sign-in, uploads and all data always go to the server.

import { build, files, version } from '$service-worker';
import { routeRequest } from '$lib/pwa/sw-routing';
import { stashIncomingFile } from '$lib/pwa/incoming-file';

const sw = self as unknown as ServiceWorkerGlobalScope;

const CACHE = `shell-${version}`;
// Every SPA route is served the same index.html; it's cached under this key
// and used when a navigation fails because the network is down.
const SHELL = '/';
const ASSETS = [...build, ...files];
const staticFiles = new Set(files);

sw.addEventListener('install', (event) => {
	event.waitUntil(
		(async () => {
			const cache = await caches.open(CACHE);
			await cache.addAll([...ASSETS, SHELL]);
			// Take over right away. Navigations are network-first, so open pages
			// pick up the new version on their next reload (see UpdateNotice).
			await sw.skipWaiting();
		})()
	);
});

sw.addEventListener('activate', (event) => {
	event.waitUntil(
		(async () => {
			// Only old app-shell caches; a shared file waiting in `incoming-files` stays.
			for (const key of await caches.keys()) {
				if (key.startsWith('shell-') && key !== CACHE) await caches.delete(key);
			}
			await sw.clients.claim();
		})()
	);
});

sw.addEventListener('fetch', (event) => {
	const { request } = event;
	const strategy = routeRequest({
		url: new URL(request.url),
		method: request.method,
		mode: request.mode,
		origin: sw.location.origin,
		staticFiles
	});
	if (strategy === 'passthrough') return;
	if (strategy === 'share-target') event.respondWith(receiveSharedFile(request));
	else event.respondWith(strategy === 'cache-first' ? cacheFirst(request) : networkFirst(request));
});

// Android "Share → Drum Trainer": park the file and open the library, which
// uploads it (after sign-in, if needed).
async function receiveSharedFile(request: Request): Promise<Response> {
	try {
		const file = (await request.formData()).get('file');
		if (file instanceof File) await stashIncomingFile(file);
	} catch {
		// Malformed share — just open the app.
	}
	return Response.redirect(new URL('/', sw.location.origin).href, 303);
}

async function cacheFirst(request: Request): Promise<Response> {
	const cache = await caches.open(CACHE);
	const cached = await cache.match(request);
	if (cached) return cached;
	const response = await fetch(request);
	// Hashed files never change, so anything not precached (e.g. built by a
	// newer deploy) can be kept once fetched.
	if (response.ok && new URL(request.url).pathname.startsWith('/_app/immutable/')) {
		await cache.put(request, response.clone());
	}
	return response;
}

async function networkFirst(request: Request): Promise<Response> {
	try {
		const response = await fetch(request);
		// Keep the shell fresh: any successful same-origin HTML page is index.html.
		if (
			response.ok &&
			response.type === 'basic' &&
			!response.redirected &&
			response.headers.get('content-type')?.includes('text/html')
		) {
			const cache = await caches.open(CACHE);
			await cache.put(SHELL, response.clone());
		}
		return response;
	} catch (err) {
		const shell = await caches.match(SHELL, { cacheName: CACHE });
		if (shell) return shell;
		throw err;
	}
}
