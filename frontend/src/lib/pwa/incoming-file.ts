// MIDI files handed to the installed app by the operating system:
// - desktop "Open with" / double-click: manifest `file_handlers`, read from `launchQueue`
// - Android "Share": manifest `share_target`, received by the service worker
// Both park the file in a cache, so it survives the sign-in redirect, and the
// library page takes it from there and uploads it.

export const INCOMING_CACHE = 'incoming-files';
/** Fired on window when a launched file has been parked. */
export const INCOMING_EVENT = 'drum-trainer:incoming-file';

const KEY = '/incoming/midi';
/** A file not picked up within this time (e.g. sign-in abandoned) is dropped. */
const MAX_AGE_MS = 60 * 60 * 1000;

export async function stashIncomingFile(file: File, storage: CacheStorage = caches, now = Date.now()) {
	const cache = await storage.open(INCOMING_CACHE);
	await cache.put(
		KEY,
		new Response(file, {
			headers: {
				'content-type': file.type || 'audio/midi',
				'x-file-name': encodeURIComponent(file.name),
				'x-stored-at': String(now)
			}
		})
	);
}

/** Returns the parked file once (it is removed), or null. */
export async function takeIncomingFile(
	storage: CacheStorage | undefined = typeof caches === 'undefined' ? undefined : caches,
	now = Date.now()
): Promise<File | null> {
	if (!storage) return null;
	const cache = await storage.open(INCOMING_CACHE);
	const res = await cache.match(KEY);
	if (!res) return null;
	await cache.delete(KEY);
	const storedAt = Number(res.headers.get('x-stored-at'));
	if (!(now - storedAt < MAX_AGE_MS)) return null;
	const name = decodeURIComponent(res.headers.get('x-file-name') ?? '') || 'shared.mid';
	return new File([await res.blob()], name, { type: res.headers.get('content-type') ?? 'audio/midi' });
}

// launchQueue isn't in the TypeScript DOM types yet.
type LaunchParams = { files?: readonly FileSystemFileHandle[] };
type LaunchQueue = { setConsumer(consumer: (params: LaunchParams) => void): void };

/** Parks a file the app was launched with. Call once, early, on every page. */
export function captureLaunchedFile() {
	const queue = (window as Window & { launchQueue?: LaunchQueue }).launchQueue;
	if (!queue) return;
	queue.setConsumer(async ({ files }) => {
		const handle = files?.[0];
		if (!handle) return;
		await stashIncomingFile(await handle.getFile());
		window.dispatchEvent(new Event(INCOMING_EVENT));
	});
}
