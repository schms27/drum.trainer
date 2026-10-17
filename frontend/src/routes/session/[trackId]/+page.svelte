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
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import { INSTRUMENTS } from '$lib/drums';
	import type { RecordedHit } from '$lib/types/session';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
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

	// Instruments used by the track, in kit order, for the page subtitle.
	const instrumentsUsed = (INSTRUMENTS as readonly string[]).filter((i) =>
		track.notes.some((n) => n.instrument === i)
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

<PageHeader
	title={track.name}
	subtitle="{formatDuration(track.duration_ms)} · {Math.round(track.bpm)} BPM · {track.note_count} notes · {instrumentsUsed.join(', ')}"
	back={{ href: '/', label: 'Library' }}
>
	{#snippet actions()}
		<PlayButton hits={track.notes} label="Listen to track" metronome={activeMetronome} speed={previewSpeed} />
	{/snippet}
</PageHeader>

<div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
	<Card title="1. Connect your kit" description="Your e-drum kit, connected by USB.">
		<MidiDeviceSelector bind:selectedId={selectedDeviceId} />
	</Card>

	<Card title="2. Choose your tempo" description="Practise slower first — your score is adjusted to the tempo you pick.">
		<Metronome
			initialBpm={track.practice_bpm ?? track.bpm}
			trackBpm={track.bpm}
			beatsPerBar={track.time_sig_numerator}
			onBpmChange={handleBpmChange}
			onMetronomeChange={(m) => (activeMetronome = m)}
		/>
	</Card>
</div>

{#if pilots.length > 0}
	<Card title="Backing tracks" description="Audio that plays along while you record.">
		<PilotControls {pilots} player={pilotPlayer} ready={pilotsReady} />
		{#if pilotLoadError}
			<div class="mt-3"><Notice variant="error">{pilotLoadError}</Notice></div>
		{/if}
	</Card>
{/if}

<section class="flex flex-col gap-4" aria-labelledby="record-heading">
	<h2 id="record-heading" class="text-lg font-semibold text-fg">3. Record</h2>
	{#if submitError}
		<Notice variant="error">Your take could not be saved: {submitError}</Notice>
	{/if}
	{#if submitting}
		<div class="bg-card border border-line rounded-xl p-5 text-sm text-fg-muted" role="status">Scoring your take…</div>
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
</section>
