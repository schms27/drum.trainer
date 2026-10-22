import { describe, expect, it } from 'vitest';
import { INCOMING_CACHE, stashIncomingFile, takeIncomingFile } from './incoming-file';

/** Minimal in-memory CacheStorage: one cache, string keys. */
function fakeStorage() {
	const caches = new Map<string, Map<string, Response>>();
	const storage = {
		async open(name: string) {
			if (!caches.has(name)) caches.set(name, new Map());
			const entries = caches.get(name)!;
			return {
				async put(key: string, res: Response) {
					entries.set(key, res);
				},
				async match(key: string) {
					return entries.get(key)?.clone();
				},
				async delete(key: string) {
					return entries.delete(key);
				}
			};
		}
	};
	return { storage: storage as unknown as CacheStorage, caches };
}

const midi = () => new File([new Uint8Array([0x4d, 0x54, 0x68, 0x64])], 'Groove – Bär.mid', { type: 'audio/midi' });

describe('incoming file', () => {
	it('round-trips name, type and content, and is taken only once', async () => {
		const { storage, caches } = fakeStorage();
		await stashIncomingFile(midi(), storage, 1000);
		expect(caches.has(INCOMING_CACHE)).toBe(true);

		const file = await takeIncomingFile(storage, 2000);
		expect(file?.name).toBe('Groove – Bär.mid');
		expect(file?.type).toBe('audio/midi');
		expect([...new Uint8Array(await file!.arrayBuffer())]).toEqual([0x4d, 0x54, 0x68, 0x64]);

		expect(await takeIncomingFile(storage, 3000)).toBeNull();
	});

	it('keeps only the latest file', async () => {
		const { storage } = fakeStorage();
		await stashIncomingFile(new File(['a'], 'a.mid'), storage, 1000);
		await stashIncomingFile(new File(['b'], 'b.mid'), storage, 1000);
		expect((await takeIncomingFile(storage, 2000))?.name).toBe('b.mid');
	});

	it('drops a file that waited more than an hour', async () => {
		const { storage } = fakeStorage();
		await stashIncomingFile(midi(), storage, 0);
		expect(await takeIncomingFile(storage, 60 * 60 * 1000 + 1)).toBeNull();
		expect(await takeIncomingFile(storage, 0)).toBeNull();
	});

	it('returns null when nothing is parked or caches are unavailable', async () => {
		expect(await takeIncomingFile(fakeStorage().storage)).toBeNull();
		expect(await takeIncomingFile(undefined)).toBeNull();
	});
});
