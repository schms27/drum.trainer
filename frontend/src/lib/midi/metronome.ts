export interface MetronomeOptions {
	beatsPerBar?: number;
	/** Called just before each click sounds. beatIndex is 0-based within the bar. */
	onBeat?: (beatIndex: number, accent: boolean) => void;
}

export interface Metronome {
	readonly playing: boolean;
	start(): void;
	stop(): void;
	setBpm(bpm: number): void;
}

export function createMetronome(initialBpm: number, options?: MetronomeOptions): Metronome {
	const beatsPerBar = options?.beatsPerBar ?? 4;

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

		const osc = ctx.createOscillator();
		const env = ctx.createGain();
		osc.type = 'triangle';
		osc.frequency.value = accent ? 1200 : 900;
		env.gain.setValueAtTime(accent ? 0.6 : 0.35, time);
		env.gain.exponentialRampToValueAtTime(0.001, time + 0.04);
		osc.connect(env);
		env.connect(ctx.destination);
		osc.start(time);
		osc.stop(time + 0.045);

		if (options?.onBeat) {
			const delayMs = Math.max(0, (time - ctx.currentTime) * 1000);
			const idx = beatIndex % beatsPerBar;
			setTimeout(() => options.onBeat!(idx, accent), delayMs);
		}
	}

	function runScheduler(): void {
		if (!ctx) return;
		while (nextBeatTime < ctx.currentTime + SCHEDULE_AHEAD_S) {
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
			ctx = new AudioContext();
			beatIndex = 0;
			nextBeatTime = ctx.currentTime + 0.05;
			runScheduler();
		},
		stop() {
			if (timerId) {
				clearTimeout(timerId);
				timerId = null;
			}
			ctx?.close();
			ctx = null;
		},
		setBpm(newBpm: number) {
			bpm = newBpm;
		}
	};
}
