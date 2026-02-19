<script lang="ts">
	import { createRecorder, type Recorder } from '$lib/midi/recorder';
	import type { RecordedHit } from '$lib/types/session';
	import LiveEventFeed from './LiveEventFeed.svelte';

	let {
		deviceId,
		onFinished
	}: {
		deviceId: string;
		onFinished: (hits: RecordedHit[], startedAt: string, endedAt: string) => void;
	} = $props();

	type State = 'idle' | 'recording' | 'stopping';

	let recorderState = $state<State>('idle');
	let recorder: Recorder | null = null;
	let liveHits = $state<RecordedHit[]>([]);
	let error = $state('');
	let startedAt = '';
	let pollInterval: ReturnType<typeof setInterval>;

	async function start() {
		if (!deviceId) {
			error = 'Select a MIDI device first.';
			return;
		}
		error = '';
		try {
			recorder = await createRecorder(deviceId);
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to open MIDI device';
			return;
		}

		startedAt = new Date().toISOString();
		recorder.start();
		recorderState = 'recording';

		// Poll for live display every 80ms.
		pollInterval = setInterval(() => {
			if (recorder) liveHits = recorder.hits;
		}, 80);
	}

	function stop() {
		if (!recorder) return;
		clearInterval(pollInterval);
		recorderState = 'stopping';
		const hits = recorder.stop();
		liveHits = hits;
		const endedAt = new Date().toISOString();
		recorder.destroy();
		recorder = null;
		onFinished(hits, startedAt, endedAt);
	}

	function reset() {
		clearInterval(pollInterval);
		recorder?.destroy();
		recorder = null;
		liveHits = [];
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

	{#if error}
		<p class="text-red-400 text-sm">{error}</p>
	{/if}

	<LiveEventFeed hits={liveHits} />
</div>
