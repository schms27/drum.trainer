export interface MetronomeOptions {
	beatsPerBar?: number;
	/** Called just before each click sounds. beatIndex is 0-based within the bar. */
	onBeat?: (beatIndex: number, accent: boolean) => void;
	/**
	 * Use an existing AudioContext instead of creating one. The context is then
	 * owned by the caller and is not closed by stop(). Lets a recording share
	 * one clock between the count-in, the click and backing tracks.
	 */
	context?: AudioContext;
	/** AudioContext time of the first click (default: just after start()). */
	startTime?: number;
	/** Bar position of the first click, 0-based (default 0, the accented beat 1). */
	firstBeat?: number;
	/** Stop automatically after this many clicks (e.g. one bar of count-in). */
	maxBeats?: number;
	/** Gain (0–1) for each click, read as it is scheduled so changes apply right away. */
	gain?: () => number;
}

export interface Metronome {
	readonly playing: boolean;
	start(): void;
	stop(): void;
	setBpm(bpm: number): void;
	/**
	 * Milliseconds until the next bar-1 accent beat, or null if stopped.
	 * Use this to delay playback so it starts in phase with the metronome.
	 */
	getNextBeat1DelayMs(): number | null;
}

/** Schedule a single click sound at the given AudioContext time. */
export function playClick(ctx: AudioContext, time: number, accent: boolean, gain = 1): void {
	if (gain <= 0) return; // muted
	const osc = ctx.createOscillator();
	const env = ctx.createGain();
	osc.type = 'triangle';
	osc.frequency.value = accent ? 1200 : 900;
	env.gain.setValueAtTime((accent ? 0.6 : 0.35) * gain, time);
	env.gain.exponentialRampToValueAtTime(0.001, time + 0.04);
	osc.connect(env);
	env.connect(ctx.destination);
	osc.start(time);
	osc.stop(time + 0.045);
}

export function createMetronome(initialBpm: number, options?: MetronomeOptions): Metronome {
	const beatsPerBar = options?.beatsPerBar ?? 4;
	const ownsContext = !options?.context;
	const firstBeat = options?.firstBeat ?? 0;

	let ctx: AudioContext | null = null;
	let bpm = initialBpm;
	let nextBeatTime = 0;
	let beatIndex = 0;
	let timerId: ReturnType<typeof setTimeout> | null = null;

	// Lookahead scheduler: wake every 25 ms and schedule beats
	// up to 100 ms into the future. Gives ~75 ms to react while
	// keeping the schedule buffer small enough to feel responsive.
	const LOOKAHEAD_MS = 25;
	const SCHEDULE_AHEAD_S = 0.1;

	function scheduleClick(time: number, accent: boolean): void {
		if (!ctx) return;
		playClick(ctx, time, accent, options?.gain?.() ?? 1);

		if (options?.onBeat) {
			const delayMs = Math.max(0, (time - ctx.currentTime) * 1000);
			const idx = beatIndex % beatsPerBar;
			setTimeout(() => options.onBeat!(idx, accent), delayMs);
		}
	}

	function stop(): void {
		if (timerId) {
			clearTimeout(timerId);
			timerId = null;
		}
		if (ownsContext) ctx?.close();
		ctx = null;
	}

	function runScheduler(): void {
		if (!ctx) return;
		while (nextBeatTime < ctx.currentTime + SCHEDULE_AHEAD_S) {
			if (options?.maxBeats !== undefined && beatIndex - firstBeat >= options.maxBeats) {
				stop();
				return;
			}
			const accent = beatIndex % beatsPerBar === 0;
			scheduleClick(nextBeatTime, accent);
			nextBeatTime += 60 / bpm;
			beatIndex++;
		}
		timerId = setTimeout(runScheduler, LOOKAHEAD_MS);
	}

	return {
		get playing() {
			return ctx !== null;
		},
		start() {
			if (ctx) return;
			ctx = options?.context ?? new AudioContext();
			beatIndex = firstBeat;
			nextBeatTime = options?.startTime ?? ctx.currentTime + 0.05;
			runScheduler();
		},
		stop,
		setBpm(newBpm: number) {
			bpm = newBpm;
		},
		getNextBeat1DelayMs() {
			if (!ctx) return null;
			// How many beats until the next bar-1 accent?
			const phase = beatIndex % beatsPerBar;
			const beatsUntilBeat1 = (beatsPerBar - phase) % beatsPerBar;
			const secs = nextBeatTime - ctx.currentTime + beatsUntilBeat1 * (60 / bpm);
			return secs * 1000;
		}
	};
}
