import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { holdScreenWakeLock } from './wake-lock';

class FakeSentinel {
	released = false;
	release = vi.fn(async () => {
		this.released = true;
	});
}

let visibility: DocumentVisibilityState;
let listeners: Set<() => void>;
let sentinels: FakeSentinel[];
let request: ReturnType<typeof vi.fn>;

const flush = () => new Promise((r) => setTimeout(r, 0));

function setVisibility(v: DocumentVisibilityState) {
	visibility = v;
	for (const l of listeners) l();
}

beforeEach(() => {
	visibility = 'visible';
	listeners = new Set();
	sentinels = [];
	request = vi.fn(async () => {
		const s = new FakeSentinel();
		sentinels.push(s);
		return s;
	});
	vi.stubGlobal('navigator', { wakeLock: { request } });
	vi.stubGlobal('document', {
		get visibilityState() {
			return visibility;
		},
		addEventListener: (_: string, l: () => void) => listeners.add(l),
		removeEventListener: (_: string, l: () => void) => listeners.delete(l)
	});
});

afterEach(() => {
	vi.unstubAllGlobals();
});

describe('holdScreenWakeLock', () => {
	it('requests a screen lock and releases it', async () => {
		const release = holdScreenWakeLock();
		await flush();
		expect(request).toHaveBeenCalledWith('screen');
		release();
		expect(sentinels[0].release).toHaveBeenCalled();
		expect(listeners.size).toBe(0);
	});

	it('requests it again when the page becomes visible after the browser dropped it', async () => {
		const release = holdScreenWakeLock();
		await flush();
		sentinels[0].released = true; // the browser drops it when the tab is hidden
		setVisibility('hidden');
		await flush();
		expect(request).toHaveBeenCalledTimes(1);
		setVisibility('visible');
		await flush();
		expect(request).toHaveBeenCalledTimes(2);
		release();
	});

	it('does not request a second lock while one is held', async () => {
		const release = holdScreenWakeLock();
		await flush();
		setVisibility('visible');
		await flush();
		expect(request).toHaveBeenCalledTimes(1);
		release();
	});

	it('releases a lock that arrives after it was let go', async () => {
		const release = holdScreenWakeLock();
		release();
		await flush();
		expect(sentinels[0].release).toHaveBeenCalled();
	});

	it('ignores a refused request', async () => {
		request.mockRejectedValueOnce(new DOMException('denied', 'NotAllowedError'));
		const release = holdScreenWakeLock();
		await flush();
		expect(() => release()).not.toThrow();
	});

	it('does nothing where the API is missing', () => {
		vi.stubGlobal('navigator', {});
		const release = holdScreenWakeLock();
		expect(listeners.size).toBe(0);
		expect(() => release()).not.toThrow();
	});
});
