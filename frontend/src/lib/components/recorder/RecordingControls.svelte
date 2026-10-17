<script lang="ts">
	import { createRecorder, type Recorder } from '$lib/midi/recorder';
	import type { RecordedHit } from '$lib/types/session';
	import type { NoteEvent } from '$lib/types/track';
	import { CircleDot, Square, RotateCcw } from '@lucide/svelte';
	import LiveEventFeed from './LiveEventFeed.svelte';
	import HitTimeline from './HitTimeline.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	let {
		deviceId,
		onFinished,
		onRecordingStart,
		onFirstHit,
		onRecordingStop,
		expectedHits = [],
		timeScale = 1,
		bpm = 120,
		timeSigNumerator = 4,
		durationMs = 0,
		firstHitTimeMs = 0
	}: {
		deviceId: string;
		onFinished: (hits: RecordedHit[], startedAt: string, endedAt: string) => void;
		onRecordingStart?: (audioCtx: AudioContext) => void;
		onFirstHit?: (audioCtxTime: number) => void;
		onRecordingStop?: () => void;
		expectedHits?: NoteEvent[];
		timeScale?: number;
		bpm?: number;
		timeSigNumerator?: number;
		durationMs?: number;
		/** Time (ms, practice tempo) the first hit is stamped with — the first target note. */
		firstHitTimeMs?: number;
	} = $props();

	// Duration scaled to the current practice tempo.
	const effectiveDurationMs = $derived(durationMs * timeScale);

	function formatTime(ms: number): string {
		const secs = Math.floor(ms / 1000);
		const m = Math.floor(secs / 60);
		const s = secs % 60;
		return `${m}:${String(s).padStart(2, '0')}`;
	}

	type State = 'idle' | 'waiting' | 'recording' | 'stopping';

	let recorderState = $state<State>('idle');
	let recorder: Recorder | null = null;
	let liveHits = $state<RecordedHit[]>([]);
	let error = $state('');
	let startedAt = '';
	let pollInterval: ReturnType<typeof setInterval>;
	let audioCtx: AudioContext | null = null;

	// Wall-clock ms when the first hit starts the clock — used to advance the timeline.
	let recordingStartWallMs = 0;
	let currentTimeMs = $state(0);

	async function start() {
		if (!deviceId) {
			error = 'Select a MIDI device first.';
			return;
		}
		error = '';

		// Create a fresh AudioContext for this recording session.
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
			onRecordingStop?.();
			audioCtx.close();
			audioCtx = null;
			return;
		}

		recorder.arm();
		recorderState = 'waiting';

		// Poll for live display every 80ms — also advances the timeline clock.
		pollInterval = setInterval(() => {
			if (recorder) liveHits = recorder.hits;
			if (recorderState === 'recording' && recordingStartWallMs > 0) {
				currentTimeMs = Date.now() - recordingStartWallMs;
			}
		}, 80);
	}

	function stop() {
		if (!recorder) return;
		clearInterval(pollInterval);
		const hits = recorder.stop();
		recorder.destroy();
		recorder = null;
		onRecordingStop?.();

		// If no hit was ever received the clock never started — just reset.
		if (!startedAt) {
			liveHits = [];
			recorderState = 'idle';
			audioCtx?.close();
			audioCtx = null;
			return;
		}

		liveHits = hits;
		recorderState = 'stopping';
		const endedAt = new Date().toISOString();
		onFinished(hits, startedAt, endedAt);
	}

	function reset() {
		clearInterval(pollInterval);
		recorder?.destroy();
		recorder = null;
		liveHits = [];
		startedAt = '';
		currentTimeMs = 0;
		recordingStartWallMs = 0;
		recorderState = 'idle';
		error = '';
		onRecordingStop?.();
		audioCtx?.close();
		audioCtx = null;
	}
</script>

<div class="flex flex-col gap-4">
	<div class="bg-card border border-line rounded-xl p-5 flex flex-col gap-4">
		<div class="flex flex-wrap items-center gap-3">
			{#if recorderState === 'idle'}
				<Button variant="success" size="lg" onclick={start} disabled={!deviceId}>
					<CircleDot size={18} aria-hidden="true" /> Start
				</Button>
				{#if deviceId}
					<p class="text-sm text-fg-muted">Press Start, then play. Recording begins with your first note.</p>
				{:else}
					<p class="text-sm text-fg-muted">Connect your kit (step 1) to start recording.</p>
				{/if}
			{:else if recorderState === 'waiting' || recorderState === 'recording'}
				<Button variant="danger" size="lg" onclick={stop}>
					<Square size={16} aria-hidden="true" /> Stop
				</Button>
				<Button variant="ghost" onclick={reset}>
					<RotateCcw size={16} aria-hidden="true" /> Discard
				</Button>
				<p class="flex items-center gap-2 text-sm font-medium {recorderState === 'waiting' ? 'text-warn' : 'text-bad'}" role="status">
					<span class="w-2.5 h-2.5 rounded-full animate-pulse {recorderState === 'waiting' ? 'bg-warn-solid' : 'bg-bad-solid'}" aria-hidden="true"></span>
					{recorderState === 'waiting' ? 'Ready — play the first note to start recording' : 'Recording… press Stop when you are done'}
				</p>
			{:else}
				<p class="text-sm text-fg-muted" role="status">Scoring your take…</p>
			{/if}
		</div>

		{#if recorderState === 'recording' && effectiveDurationMs > 0}
			{@const pct = Math.min((currentTimeMs / effectiveDurationMs) * 100, 100)}
			<div class="flex flex-col gap-1">
				<div class="flex justify-between text-xs text-fg-muted tabular-nums">
					<span>{formatTime(currentTimeMs)} / {formatTime(effectiveDurationMs)}</span>
					<span>{Math.round(pct)}%</span>
				</div>
				<div class="h-2 bg-control rounded-full overflow-hidden" role="progressbar" aria-label="Track progress" aria-valuenow={Math.round(pct)} aria-valuemin={0} aria-valuemax={100}>
					<div class="h-full bg-accent rounded-full transition-[width] duration-100" style="width: {pct}%"></div>
				</div>
			</div>
		{/if}

		{#if error}
			<Notice variant="error">{error}</Notice>
		{/if}
	</div>

	<HitTimeline
		{expectedHits}
		actualHits={liveHits}
		{currentTimeMs}
		{timeScale}
		{bpm}
		{timeSigNumerator}
	/>

	<LiveEventFeed hits={liveHits} />
</div>
