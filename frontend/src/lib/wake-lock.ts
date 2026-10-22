// Screen Wake Lock: keeps the display from dimming or locking while practising.
// Browsers drop the lock whenever the page is hidden, so it is requested again
// when the page becomes visible. Does nothing where the API is unsupported.

/** Holds a screen wake lock until the returned function is called. */
export function holdScreenWakeLock(): () => void {
	if (typeof navigator === 'undefined' || !('wakeLock' in navigator)) return () => {};

	let sentinel: WakeLockSentinel | null = null;
	let requesting = false;
	let done = false;

	async function acquire() {
		if (done || requesting || document.visibilityState !== 'visible') return;
		if (sentinel && !sentinel.released) return;
		requesting = true;
		try {
			const s = await navigator.wakeLock.request('screen');
			if (done) void s.release();
			else sentinel = s;
		} catch {
			// Refused (e.g. battery saver or a permissions policy) — nothing to do.
		} finally {
			requesting = false;
		}
	}

	const onVisibilityChange = () => void acquire();
	document.addEventListener('visibilitychange', onVisibilityChange);
	void acquire();

	return () => {
		done = true;
		document.removeEventListener('visibilitychange', onVisibilityChange);
		void sentinel?.release();
		sentinel = null;
	};
}
