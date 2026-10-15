<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import MidiDeviceSelector from '$lib/components/recorder/MidiDeviceSelector.svelte';
	import RecordingControls from '$lib/components/recorder/RecordingControls.svelte';
	import PilotControls from '$lib/components/session/PilotControls.svelte';
	import PlayButton from '$lib/components/PlayButton.svelte';
	import Metronome from '$lib/components/Metronome.svelte';
	import { submitSession } from '$lib/api/sessions';
	import { setPracticeBpm } from '$lib/api/tracks';
	import { pilotAudioUrl } from '$lib/api/pilots';
	import { PilotPlayer } from '$lib/audio/PilotPlayer';
	import type { RecordedHit } from '$lib/types/session';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const { track, pilots } = data;

	let pilotPlayer = $state<PilotPlayer | null>(null);
	let pilotLoadError = $state('');
	let pilotRawBuffers: ArrayBuffer[] = [];
	let pilotsReady = $state(pilots.length === 0); // true when pre-fetch done (or no pilots)

	onMount(async () => {
		if (pilots.length === 0) return;
		try {
			pilotRawBuffers = await Promise.all(
				pilots.map(async (p) => {
					const resp = await fetch(pilotAudioUrl(p.track_id, p.id));
					if (!resp.ok) throw new Error(`Failed to fetch pilot audio: ${resp.status}`);
					return resp.arrayBuffer();
				})
			);
			pilotsReady = true;
		} catch (e: unknown) {
			pilotLoadError = e instanceof Error ? e.message : 'Failed to pre-load pilot audio';
		}
	});

	async function handleRecordingStart(audioCtx: AudioContext) {
		if (pilots.length === 0) return;
		pilotLoadError = '';
		try {
			const player = new PilotPlayer(audioCtx);
			await player.loadFromBytes(
				pilots.map((p) => ({ id: p.id, trackId: p.track_id, name: p.name })),
				pilotRawBuffers
			);
			pilotPlayer = player;
			// Playback starts on first hit via handleFirstHit — not here.
		} catch (e: unknown) {
			pilotLoadError = e instanceof Error ? e.message : 'Failed to load pilot audio';
		}
	}

	// The first hit is assumed to be the first target note, which may not be at
	// 0 ms if the MIDI file has a lead-in. Notes arrive sorted by time.
	const firstTargetMs = track.notes[0]?.time_ms ?? 0;

	function handleFirstHit(contextTime: number) {
		if (!pilotPlayer) return;
		pilotPlayer.startAll(contextTime, currentBpm / track.bpm, firstTargetMs / 1000);
	}

	function handleRecordingStop() {
		pilotPlayer?.stopAll();
		pilotPlayer?.destroy();
		pilotPlayer = null;
	}

	let selectedDeviceId = $state('');
	let submitting = $state(false);
	let submitError = $state('');
	let activeMetronome = $state<import('$lib/midi/metronome').Metronome | null>(null);
	let currentBpm = $state(Math.round(track.practice_bpm ?? track.bpm));

	const previewSpeed = $derived(currentBpm / track.bpm);

	let saveBpmTimer: ReturnType<typeof setTimeout> | null = null;

	function handleBpmChange(newBpm: number) {
		currentBpm = newBpm;
		// Update pilot playback rate if playing.
		pilotPlayer?.setPlaybackRate(newBpm / track.bpm);
		if (saveBpmTimer) clearTimeout(saveBpmTimer);
		saveBpmTimer = setTimeout(() => {
			// Save null when the user has reset back to the track's native tempo.
			const value = newBpm === Math.round(track.bpm) ? null : newBpm;
			setPracticeBpm(track.id, value);
		}, 600);
	}

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		const m = Math.floor(secs / 60);
		const s = secs % 60;
		return `${m}:${String(s).padStart(2, '0')}`;
	}

	const noteCounts = $derived(
		track.notes.reduce<Record<string, number>>((acc, n) => {
			acc[n.instrument] = (acc[n.instrument] ?? 0) + 1;
			return acc;
		}, {})
	);

	async function handleFinished(hits: RecordedHit[], startedAt: string, endedAt: string) {
		// Normalize hit times from practice-tempo space back to original-tempo space
		// so the scorer compares against the unscaled MIDI target times.
		const timeScale = track.bpm / currentBpm;
		const scoringHits =
			timeScale === 1 ? hits : hits.map((h) => ({ ...h, time_ms: h.time_ms / timeScale }));
		submitting = true;
		submitError = '';
		try {
			const result = await submitSession({
				track_id: track.id,
				started_at: startedAt,
				ended_at: endedAt,
				hits: scoringHits
			});
			goto(`/results/${result.session.id}`);
		} catch (e: unknown) {
			submitError = e instanceof Error ? e.message : 'Failed to submit session';
			submitting = false;
		}
	}
</script>

<svelte:head><title>{track.name} — Drum Trainer</title></svelte:head>

<div class="flex flex-col gap-8">
	<!-- Track info -->
	<div>
		<a href="/" class="text-zinc-500 hover:text-zinc-300 text-sm transition">← Back to tracks</a>
		<div class="flex items-start justify-between mt-2 gap-4">
			<div>
				<h1 class="text-2xl font-bold text-white">{track.name}</h1>
				<p class="text-zinc-400 text-sm mt-1">
					{track.note_count} notes · {formatDuration(track.duration_ms)} · {track.filename}
				</p>
			</div>
			<PlayButton hits={track.notes} label="Preview Track" metronome={activeMetronome} speed={previewSpeed} />
		</div>
	</div>

	<!-- Device selector -->
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 max-w-sm">
		<MidiDeviceSelector bind:selectedId={selectedDeviceId} />
	</div>

	<!-- Metronome -->
	<Metronome
		initialBpm={track.practice_bpm ?? track.bpm}
		trackBpm={track.bpm}
		onBpmChange={handleBpmChange}
		onMetronomeChange={(m) => (activeMetronome = m)}
	/>

	<!-- Pilot tracks -->
	{#if pilots.length > 0}
		<PilotControls {pilots} player={pilotPlayer} ready={pilotsReady} />
		{#if pilotLoadError}
			<p class="text-red-400 text-sm -mt-4">{pilotLoadError}</p>
		{/if}
	{/if}

	<!-- Recording controls + live feed -->
	<div>
		<h2 class="text-lg font-semibold text-white mb-4">Recording</h2>
		{#if submitting}
			<p class="text-zinc-400 text-sm animate-pulse">Calculating your score…</p>
		{:else}
			<RecordingControls
				deviceId={selectedDeviceId}
				onFinished={handleFinished}
				onRecordingStart={handleRecordingStart}
				onFirstHit={handleFirstHit}
				onRecordingStop={handleRecordingStop}
				expectedHits={track.notes}
				timeScale={track.bpm / currentBpm}
				bpm={track.bpm}
				timeSigNumerator={track.time_sig_numerator}
				durationMs={track.duration_ms}
				firstHitTimeMs={firstTargetMs * (track.bpm / currentBpm)}
			/>
		{/if}
		{#if submitError}
			<p class="text-red-400 text-sm mt-3">{submitError}</p>
		{/if}
	</div>

	<!-- Track overview (note count per instrument) -->
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
		<h3 class="text-sm font-semibold text-zinc-300 mb-3">Target Track Contents</h3>
		<div class="flex flex-wrap gap-3 text-sm">
			{#each Object.entries(noteCounts).sort((a, b) => b[1] - a[1]) as [instr, count]}
				<span class="bg-zinc-700 text-zinc-300 rounded-full px-3 py-1">
					{instr}: {count}
				</span>
			{/each}
		</div>
	</div>
</div>
