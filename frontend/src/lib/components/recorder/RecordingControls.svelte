<script lang="ts">
	import { CircleDot, Square, RotateCcw } from '@lucide/svelte';
	import { createRecorder, type Recorder } from '$lib/midi/recorder';
	import { createMetronome, type Metronome } from '$lib/midi/metronome';
	import { metronomeGain } from '$lib/audio/session-mix.svelte';
	import { audioTimeToPerfMs } from '$lib/midi/timing';
	import { practiceSettings, type StartMode } from '$lib/practice-settings.svelte';
	import { focusMode } from '$lib/focus.svelte';
	import { holdScreenWakeLock } from '$lib/wake-lock';
	import type { RecordedHit } from '$lib/types/session';
	import type { NoteEvent } from '$lib/types/track';
	import LiveEventFeed from './LiveEventFeed.svelte';
	import HitTimeline from './HitTimeline.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	let {
		deviceId,
		onFinished,
		onRecordingStart,
		onFirstHit,
		onCountIn,
		onRecordingStop,
		active = $bindable(false),
		expectedHits = [],
		timeScale = 1,
		bpm = 120,
		practiceBpm = 120,
		timeSigNumerator = 4,
		durationMs = 0,
		firstHitTimeMs = 0,
		hitGroup
	}: {
		deviceId: string;
		onFinished: (hits: RecordedHit[], startedAt: string, endedAt: string) => void;
		/** A fresh AudioContext for this take — backing tracks are scheduled on it. */
		onRecordingStart?: (audioCtx: AudioContext) => void;
		/** First-note mode: the first hit arrived (AudioContext time). */
		onFirstHit?: (audioCtxTime: number) => void;
		/** Count-in mode: the track starts at this AudioContext time. */
		onCountIn?: (startAudioTime: number) => void;
		onRecordingStop?: () => void;
		/** True while a take is being prepared or recorded. */
		active?: boolean;
		expectedHits?: NoteEvent[];
		timeScale?: number;
		/** The track's original tempo (for the beat grid). */
		bpm?: number;
		/** The tempo being practised at (for the count-in and click). */
		practiceBpm?: number;
		timeSigNumerator?: number;
		durationMs?: number;
		/** Time (ms, practice tempo) the first hit is stamped with in first-note mode. */
		firstHitTimeMs?: number;
		/** Instrument for a note from the kit (the active kit profile, or General MIDI). */
		hitGroup?: (note: number) => string;
	} = $props();

	/** Extra time after the last note before stopping automatically. */
	const AUTO_STOP_GRACE_MS = 1500;

	type State = 'idle' | 'waiting' | 'counting' | 'recording' | 'stopping';

	let recorderState = $state<State>('idle');
	let recorder: Recorder | null = null;
	let click: Metronome | null = null;
	let liveHits = $state<RecordedHit[]>([]);
	let error = $state('');
	let notice = $state('');
	let startedAt = '';
	let pollInterval: ReturnType<typeof setInterval> | undefined;
	let audioCtx: AudioContext | null = null;

	// Wall-clock ms of time zero (first beat, or first hit) — drives the timeline.
	let recordingStartWallMs = 0;
	let currentTimeMs = $state(0);
	// Count-in mode: AudioContext time at which the track starts (beat 1).
	let trackStartAudio = 0;
	let countdown = $state(0);

	// Duration scaled to the current practice tempo.
	const effectiveDurationMs = $derived(durationMs * timeScale);
	const beatMs = $derived(60000 / practiceBpm);
	const latency = $derived(deviceId ? practiceSettings.latencyFor(deviceId) : null);

	$effect(() => {
		active = recorderState !== 'idle' && recorderState !== 'stopping';
		focusMode.active = active;
	});

	// Keep the display awake on the practice page, between takes too: nobody
	// touches the screen while drumming.
	$effect(() => {
		if (practiceSettings.keepScreenOn) return holdScreenWakeLock();
	});

	// Leaving the page mid-take must not leave the app in focus mode or the
	// MIDI listener attached.
	$effect(() => () => {
		teardown();
		focusMode.active = false;
	});

	function formatTime(ms: number): string {
		const secs = Math.floor(Math.max(0, ms) / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
	}

	async function start() {
		if (!deviceId) {
			error = 'Select a MIDI device first.';
			return;
		}
		error = '';
		notice = '';
		startedAt = '';
		liveHits = [];
		currentTimeMs = 0;
		recordingStartWallMs = 0;

		// A fresh AudioContext for this take: count-in, click and backing tracks share its clock.
		audioCtx = new AudioContext();
		await audioCtx.resume();
		onRecordingStart?.(audioCtx);

		try {
			recorder = await createRecorder(deviceId, {
				firstHitTimeMs,
				onStarted() {
					startedAt = new Date().toISOString();
					// Back-date the clock so the timeline shows the first hit at its target time.
					recordingStartWallMs = Date.now() - firstHitTimeMs;
					recorderState = 'recording';
					onFirstHit?.(audioCtx!.currentTime);
				}
			});
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to open MIDI device';
			teardown();
			return;
		}

		if (practiceSettings.startMode === 'count-in') {
			startCountIn(audioCtx, recorder);
		} else {
			recorder.arm();
			recorderState = 'waiting';
		}

		pollInterval = setInterval(tick, 50);
	}

	function startCountIn(ctx: AudioContext, rec: Recorder) {
		const beatS = 60 / practiceBpm;
		const firstClick = ctx.currentTime + 0.3;
		const trackStart = firstClick + timeSigNumerator * beatS;

		click = createMetronome(practiceBpm, {
			context: ctx,
			startTime: firstClick,
			beatsPerBar: timeSigNumerator,
			maxBeats: practiceSettings.clickDuringTake ? undefined : timeSigNumerator,
			gain: metronomeGain
		});
		click.start();

		trackStartAudio = trackStart;
		rec.armAt(audioTimeToPerfMs(ctx, trackStart), latency ?? 0);
		syncTrackStart();
		startedAt = new Date(recordingStartWallMs).toISOString();
		countdown = timeSigNumerator;
		recorderState = 'counting';
		onCountIn?.(trackStart);
	}

	/**
	 * Re-derives beat 1 on the MIDI clock. Right after the AudioContext starts
	 * the browser can't report the output delay yet, so this runs on every tick
	 * of the count-in to pick up the accurate value before the first note.
	 */
	function syncTrackStart() {
		if (!audioCtx || !recorder) return;
		const startPerf = audioTimeToPerfMs(audioCtx, trackStartAudio);
		recorder.setStartTime(startPerf);
		recordingStartWallMs = Date.now() + (startPerf - performance.now());
	}

	function tick() {
		if (recorderState === 'counting') syncTrackStart();
		if (recorder) liveHits = recorder.hits;
		if (recordingStartWallMs <= 0) return;
		currentTimeMs = Date.now() - recordingStartWallMs;
		if (recorderState === 'counting') {
			countdown = Math.max(1, Math.min(timeSigNumerator, Math.ceil(-currentTimeMs / beatMs)));
			if (currentTimeMs >= 0) recorderState = 'recording';
		}
		if (
			recorderState === 'recording' &&
			practiceSettings.autoStop &&
			effectiveDurationMs > 0 &&
			currentTimeMs > effectiveDurationMs + AUTO_STOP_GRACE_MS
		) {
			stop();
		}
	}

	/** Stops listening and releases audio, without deciding what happens to the hits. */
	function teardown() {
		clearInterval(pollInterval);
		pollInterval = undefined;
		click?.stop();
		click = null;
		recorder?.destroy();
		recorder = null;
		onRecordingStop?.();
		audioCtx?.close();
		audioCtx = null;
	}

	function stop() {
		if (!recorder) return;
		const hits = recorder.stop();
		teardown();

		if (!startedAt || hits.length === 0) {
			// Nothing to score — e.g. stopped before playing anything.
			liveHits = [];
			currentTimeMs = 0;
			recordingStartWallMs = 0;
			recorderState = 'idle';
			if (startedAt) notice = 'No notes were recorded, so nothing was saved.';
			return;
		}

		liveHits = hits;
		recorderState = 'stopping';
		onFinished(hits, startedAt, new Date().toISOString());
	}

	function reset() {
		teardown();
		liveHits = [];
		startedAt = '';
		currentTimeMs = 0;
		recordingStartWallMs = 0;
		recorderState = 'idle';
		error = '';
	}

	const startModes: { value: StartMode; label: string; hint: string }[] = [
		{ value: 'count-in', label: 'Count-in', hint: 'One bar of clicks, then the track starts on beat 1.' },
		{ value: 'first-hit', label: 'On my first note', hint: 'The track starts the moment you hit the first note.' }
	];
</script>

<div class="flex flex-col gap-4">
	{#if recorderState === 'idle' || recorderState === 'stopping'}
		<div class="bg-card border border-line rounded-xl p-5 flex flex-col gap-5">
			<div class="flex flex-wrap items-center gap-3">
				{#if recorderState === 'idle'}
					<Button variant="success" size="lg" onclick={start} disabled={!deviceId}>
						<CircleDot size={18} aria-hidden="true" /> Start
					</Button>
					<p class="text-sm text-fg-muted">
						{#if !deviceId}
							Connect your kit (step 1) to start recording.
						{:else if practiceSettings.startMode === 'count-in'}
							You'll hear one bar of clicks, then play from the first note.
						{:else}
							Press Start, then play. Recording begins with your first note.
						{/if}
					</p>
				{:else}
					<p class="text-sm text-fg-muted" role="status">Scoring your take…</p>
				{/if}
			</div>

			{#if notice}<Notice>{notice}</Notice>{/if}
			{#if error}<Notice variant="error">{error}</Notice>{/if}

			{#if recorderState === 'idle'}
				<fieldset class="flex flex-col gap-3 border-t border-line pt-4">
					<legend class="sr-only">Recording options</legend>
					<div class="flex flex-wrap items-center gap-3">
						<span class="text-sm text-fg-muted w-20">Start</span>
						<div class="inline-flex rounded-lg bg-control p-0.5" role="radiogroup" aria-label="How the take starts">
							{#each startModes as mode (mode.value)}
								{@const selected = practiceSettings.startMode === mode.value}
								<button
									type="button"
									role="radio"
									aria-checked={selected}
									title={mode.hint}
									onclick={() => (practiceSettings.startMode = mode.value)}
									class="h-8 px-3 rounded-md text-sm transition-colors cursor-pointer
										{selected ? 'bg-card text-fg shadow-sm font-medium' : 'text-fg-muted hover:text-fg'}"
								>
									{mode.label}
								</button>
							{/each}
						</div>
					</div>
					{#if practiceSettings.startMode === 'count-in'}
						<label class="flex items-center gap-2 text-sm text-fg cursor-pointer w-fit">
							<input
								type="checkbox"
								class="w-4 h-4 accent-[var(--accent)]"
								checked={practiceSettings.clickDuringTake}
								onchange={(e) => (practiceSettings.clickDuringTake = e.currentTarget.checked)}
							/>
							Keep the metronome clicking during the take
						</label>
					{/if}
					<label class="flex items-center gap-2 text-sm text-fg cursor-pointer w-fit">
						<input
							type="checkbox"
							class="w-4 h-4 accent-[var(--accent)]"
							checked={practiceSettings.autoStop}
							onchange={(e) => (practiceSettings.autoStop = e.currentTarget.checked)}
						/>
						Stop automatically at the end of the track
					</label>
					<label class="flex items-center gap-2 text-sm text-fg cursor-pointer w-fit">
						<input
							type="checkbox"
							class="w-4 h-4 accent-[var(--accent)]"
							checked={practiceSettings.keepScreenOn}
							onchange={(e) => (practiceSettings.keepScreenOn = e.currentTarget.checked)}
						/>
						Keep the screen on while practising
					</label>
					{#if practiceSettings.startMode === 'count-in' && deviceId && latency === null}
						<p class="text-xs text-fg-subtle">
							Tip: calibrate timing in step 1 so your kit's delay isn't counted as playing late.
						</p>
					{/if}
				</fieldset>
			{/if}
		</div>
	{:else}
		<!-- Live status banner: sticky so Stop is always reachable. -->
		<div
			class="sticky top-2 z-30 bg-card border-2 rounded-xl p-4 sm:p-5 shadow-lg flex flex-col gap-3
				{recorderState === 'recording' ? 'border-bad-solid' : 'border-warn-solid'}"
			role="status"
			aria-live="polite"
		>
			<div class="flex flex-wrap items-center gap-x-4 gap-y-3">
				{#if recorderState === 'counting'}
					<span class="text-5xl font-bold tabular-nums text-warn w-12 text-center" aria-hidden="true">{countdown}</span>
					<p class="text-lg font-semibold text-fg flex-1 min-w-[12rem]">Get ready — the track starts after the count-in</p>
				{:else if recorderState === 'waiting'}
					<span class="w-4 h-4 rounded-full bg-warn-solid animate-pulse" aria-hidden="true"></span>
					<p class="text-lg font-semibold text-fg flex-1 min-w-[12rem]">Ready — play the first note to start</p>
				{:else}
					<span class="w-4 h-4 rounded-full bg-bad-solid animate-pulse" aria-hidden="true"></span>
					<p class="text-lg font-semibold text-fg flex-1 min-w-[12rem]">
						Recording
						{#if practiceSettings.autoStop}<span class="text-sm font-normal text-fg-muted">— stops at the end of the track</span>{/if}
					</p>
				{/if}
				<div class="flex gap-2">
					<Button variant="ghost" onclick={reset}>
						<RotateCcw size={16} aria-hidden="true" /> Discard
					</Button>
					<Button variant="danger" size="lg" onclick={stop}>
						<Square size={16} aria-hidden="true" /> Stop
					</Button>
				</div>
			</div>

			{#if recorderState === 'recording' && effectiveDurationMs > 0}
				{@const pct = Math.max(0, Math.min((currentTimeMs / effectiveDurationMs) * 100, 100))}
				<div class="flex flex-col gap-1">
					<div class="flex justify-between text-xs text-fg-muted tabular-nums">
						<span>{formatTime(currentTimeMs)} / {formatTime(effectiveDurationMs)}</span>
						<span>{Math.round(pct)}%</span>
					</div>
					<div
						class="h-2 bg-control rounded-full overflow-hidden"
						role="progressbar"
						aria-label="Track progress"
						aria-valuenow={Math.round(pct)}
						aria-valuemin={0}
						aria-valuemax={100}
					>
						<div class="h-full bg-accent rounded-full transition-[width] duration-100" style="width: {pct}%"></div>
					</div>
				</div>
			{/if}
		</div>
	{/if}

	<HitTimeline {expectedHits} actualHits={liveHits} {currentTimeMs} {timeScale} {bpm} {timeSigNumerator} {hitGroup} />

	<LiveEventFeed hits={liveHits} />
</div>
