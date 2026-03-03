<script lang="ts">
	import { createRecorder, type Recorder } from '$lib/midi/recorder';
	import type { RecordedHit } from '$lib/types/session';
	import type { NoteEvent } from '$lib/types/track';
	import LiveEventFeed from './LiveEventFeed.svelte';
	import HitTimeline from './HitTimeline.svelte';

	let {
		deviceId,
		onFinished,
		expectedHits = [],
		timeScale = 1,
		bpm = 120,
		timeSigNumerator = 4,
		durationMs = 0
	}: {
		deviceId: string;
		onFinished: (hits: RecordedHit[], startedAt: string, endedAt: string) => void;
		expectedHits?: NoteEvent[];
		timeScale?: number;
		bpm?: number;
		timeSigNumerator?: number;
		durationMs?: number;
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

	// Wall-clock ms when the first hit starts the clock — used to advance the timeline.
	let recordingStartWallMs = 0;
	let currentTimeMs = $state(0);

	async function start() {
		if (!deviceId) {
			error = 'Select a MIDI device first.';
			return;
		}
		error = '';
		try {
			recorder = await createRecorder(deviceId, {
				onStarted() {
					startedAt = new Date().toISOString();
					recordingStartWallMs = Date.now();
					recorderState = 'recording';
				}
			});
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to open MIDI device';
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

		// If no hit was ever received the clock never started — just reset.
		if (!startedAt) {
			liveHits = [];
			recorderState = 'idle';
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
	}
</script>

<div class="flex flex-col gap-4">
	<div class="flex gap-3 items-center">
		{#if recorderState === 'idle'}
			<button
				onclick={start}
				disabled={!deviceId}
				class="px-5 py-2 bg-green-600 hover:bg-green-500 disabled:bg-zinc-700 disabled:text-zinc-500 text-white rounded-lg font-medium transition"
			>
				▶ Start Recording
			</button>
		{:else if recorderState === 'waiting'}
			<button
				onclick={stop}
				class="px-5 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg font-medium transition"
			>
				⏹ Stop
			</button>
			<span class="text-yellow-400 text-sm animate-pulse">◉ Waiting for first hit…</span>
		{:else if recorderState === 'recording'}
			<button
				onclick={stop}
				class="px-5 py-2 bg-red-600 hover:bg-red-500 text-white rounded-lg font-medium transition"
			>
				⏹ Stop
			</button>
			<span class="text-red-400 text-sm animate-pulse">● Recording…</span>
		{:else}
			<span class="text-zinc-400 text-sm">Submitting…</span>
		{/if}

		{#if recorderState !== 'idle'}
			<button
				onclick={reset}
				class="px-4 py-2 bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded-lg text-sm transition"
			>
				Reset
			</button>
		{/if}
	</div>

	{#if recorderState === 'recording' && effectiveDurationMs > 0}
		{@const pct = Math.min((currentTimeMs / effectiveDurationMs) * 100, 100)}
		<div class="flex flex-col gap-1">
			<div class="flex justify-between text-xs text-zinc-400">
				<span>{formatTime(currentTimeMs)} / {formatTime(effectiveDurationMs)}</span>
				<span>{Math.round(pct)}%</span>
			</div>
			<div class="h-2 bg-zinc-700 rounded-full overflow-hidden">
				<div
					class="h-full bg-indigo-500 rounded-full transition-[width] duration-100"
					style="width: {pct}%"
				></div>
			</div>
		</div>
	{/if}

	{#if error}
		<p class="text-red-400 text-sm">{error}</p>
	{/if}

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
