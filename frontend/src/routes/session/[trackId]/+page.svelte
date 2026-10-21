<script lang="ts">
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import { Gauge } from '@lucide/svelte';
	import MidiDeviceSelector from '$lib/components/recorder/MidiDeviceSelector.svelte';
	import PadCheck from '$lib/components/recorder/PadCheck.svelte';
	import CalibrateDialog from '$lib/components/recorder/CalibrateDialog.svelte';
	import RecordingControls from '$lib/components/recorder/RecordingControls.svelte';
	import MixerControls from '$lib/components/session/MixerControls.svelte';
	import PlayButton from '$lib/components/PlayButton.svelte';
	import Metronome from '$lib/components/Metronome.svelte';
	import { submitSession } from '$lib/api/sessions';
	import { setPracticeBpm, setTrackMix } from '$lib/api/tracks';
	import { setKitMix } from '$lib/api/kits';
	import { levelToGain, resolveMix } from '$lib/audio/mix';
	import { sessionMix } from '$lib/audio/session-mix.svelte';
	import { pilotAudioUrl } from '$lib/api/pilots';
	import { PilotPlayer } from '$lib/audio/PilotPlayer';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import { INSTRUMENTS } from '$lib/drums';
	import { practiceSettings } from '$lib/practice-settings.svelte';
	import { kitInstrument } from '$lib/api/kits';
	import type { RecordedHit } from '$lib/types/session';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	const { track, pilots, activeKit } = data;
	// Recorded notes are mapped through the active kit profile (or General MIDI),
	// matching how the backend scores the take.
	const hitGroup = (note: number) => kitInstrument(activeKit, note);

	// ── Mixer ───────────────────────────────────────────────────────────────
	// Levels start from the track's saved mix, else the kit profile's defaults.
	// svelte-ignore state_referenced_locally
	Object.assign(sessionMix, resolveMix(track, pilots, activeKit));
	// svelte-ignore state_referenced_locally
	let kitDefaults = $state(activeKit ? { metronome_volume: activeKit.metronome_volume, pilot_volume: activeKit.pilot_volume } : null);

	let saveMixTimer: ReturnType<typeof setTimeout> | null = null;
	/** Saves the mix for this track shortly after the last change. */
	function handleMixChange() {
		if (saveMixTimer) clearTimeout(saveMixTimer);
		saveMixTimer = setTimeout(() => {
			saveMixTimer = null;
			setTrackMix(
				track.id,
				sessionMix.metronome,
				pilots.map((p) => ({ id: p.id, volume: sessionMix.pilots[p.id] ?? null }))
			).catch(() => {
				// Not saved, but still applied for this visit.
			});
		}, 600);
	}

	async function saveMixAsKitDefault() {
		if (!activeKit) return;
		// Backing tracks share one default; with several, use their average.
		const levels = pilots.map((p) => sessionMix.pilots[p.id] ?? 1);
		const pilotDefault = levels.length ? levels.reduce((a, b) => a + b, 0) / levels.length : (kitDefaults?.pilot_volume ?? null);
		const saved = await setKitMix(activeKit.id, sessionMix.metronome, pilotDefault);
		kitDefaults = { metronome_volume: saved.metronome_volume, pilot_volume: saved.pilot_volume };
	}

	/** Forgets the track's own mix, going back to the kit profile's defaults. */
	function resetMix() {
		const levels = resolveMix({ metronome_volume: null }, pilots.map((p) => ({ id: p.id, volume: null })), kitDefaults);
		sessionMix.metronome = levels.metronome;
		sessionMix.pilots = levels.pilots;
		for (const p of pilots) pilotPlayer?.setVolume(p.id, levelToGain(levels.pilots[p.id]));
		if (saveMixTimer) clearTimeout(saveMixTimer);
		setTrackMix(track.id, null, pilots.map((p) => ({ id: p.id, volume: null }))).catch(() => {});
	}

	// ── Backing tracks ("pilots" in the API) ────────────────────────────────
	let pilotPlayer = $state<PilotPlayer | null>(null);
	let pilotLoadError = $state('');
	let pilotRawBuffers: ArrayBuffer[] = [];
	let pilotsReady = $state(pilots.length === 0); // true when pre-fetch done (or no pilots)
	let recordingCtx: AudioContext | null = null;
	/** Start requested before the audio finished decoding. */
	let pendingPilotStart: { audioTime: number; offsetSec: number } | null = null;

	onMount(async () => {
		if (pilots.length === 0) return;
		try {
			pilotRawBuffers = await Promise.all(
				pilots.map(async (p) => {
					const resp = await fetch(pilotAudioUrl(p.track_id, p.id));
					if (!resp.ok) throw new Error(`Failed to fetch backing track: ${resp.status}`);
					return resp.arrayBuffer();
				})
			);
			pilotsReady = true;
		} catch (e: unknown) {
			pilotLoadError = e instanceof Error ? e.message : 'Failed to load backing tracks';
		}
	});

	async function handleRecordingStart(audioCtx: AudioContext) {
		recordingCtx = audioCtx;
		pendingPilotStart = null;
		if (pilots.length === 0) return;
		pilotLoadError = '';
		try {
			const player = new PilotPlayer(audioCtx);
			player.setInitialVolumes(
				Object.fromEntries(pilots.map((p) => [p.id, levelToGain(sessionMix.pilots[p.id] ?? 1)]))
			);
			await player.loadFromBytes(
				pilots.map((p) => ({ id: p.id, trackId: p.track_id, name: p.name })),
				pilotRawBuffers
			);
			if (recordingCtx !== audioCtx) return; // the take was cancelled meanwhile
			pilotPlayer = player;
			// The count-in or first hit may have asked to start while we were decoding.
			const pending = takePendingPilotStart();
			if (pending) startPilots(pending.audioTime, pending.offsetSec);
		} catch (e: unknown) {
			pilotLoadError = e instanceof Error ? e.message : 'Failed to load backing tracks';
		}
	}

	function takePendingPilotStart() {
		const pending = pendingPilotStart;
		pendingPilotStart = null;
		return pending;
	}

	/** Starts backing tracks at `audioTime`, `offsetSec` into the audio (original tempo). */
	function startPilots(audioTime: number, offsetSec: number) {
		if (!pilotPlayer || !recordingCtx) {
			pendingPilotStart = { audioTime, offsetSec };
			return;
		}
		pendingPilotStart = null;
		const rate = currentBpm / track.bpm;
		// If decoding took longer than the count-in, join in at the right position.
		const late = recordingCtx.currentTime - audioTime;
		if (late > 0) pilotPlayer.startAll(recordingCtx.currentTime, rate, offsetSec + late * rate);
		else pilotPlayer.startAll(audioTime, rate, offsetSec);
	}

	// In first-note mode the first hit is assumed to be the first target note,
	// which may not be at 0 ms if the MIDI file has a lead-in. Notes arrive sorted by time.
	const firstTargetMs = track.notes[0]?.time_ms ?? 0;

	function handleRecordingStop() {
		recordingCtx = null;
		pendingPilotStart = null;
		pilotPlayer?.stopAll();
		pilotPlayer?.destroy();
		pilotPlayer = null;
	}

	// ── Tempo and recording ─────────────────────────────────────────────────
	let selectedDeviceId = $state('');
	let submitting = $state(false);
	let submitError = $state('');
	let recordingActive = $state(false);
	let calibrateOpen = $state(false);
	let activeMetronome = $state<import('$lib/midi/metronome').Metronome | null>(null);
	// A ?bpm= link (e.g. "Try slower" on the results page) overrides the saved tempo.
	const requestedBpm = Number(page.url.searchParams.get('bpm'));
	let currentBpm = $state(
		Number.isInteger(requestedBpm) && requestedBpm >= 20 && requestedBpm <= 300
			? requestedBpm
			: Math.round(track.practice_bpm ?? track.bpm)
	);

	const previewSpeed = $derived(currentBpm / track.bpm);
	const latency = $derived(selectedDeviceId ? practiceSettings.latencyFor(selectedDeviceId) : null);

	let saveBpmTimer: ReturnType<typeof setTimeout> | null = null;

	function handleBpmChange(newBpm: number) {
		currentBpm = newBpm;
		if (saveBpmTimer) clearTimeout(saveBpmTimer);
		saveBpmTimer = setTimeout(() => {
			// Save null when the user has reset back to the track's native tempo.
			const value = newBpm === Math.round(track.bpm) ? null : newBpm;
			setPracticeBpm(track.id, value);
		}, 600);
	}

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		return `${Math.floor(secs / 60)}:${String(secs % 60).padStart(2, '0')}`;
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

{#if !recordingActive}
	<PageHeader
		title={track.name}
		subtitle="{formatDuration(track.duration_ms)} · {Math.round(track.bpm)} BPM · {track.note_count} notes · {instrumentsUsed.join(', ')}"
		back={{ href: `/tracks/${track.id}`, label: 'Track details' }}
	>
		{#snippet actions()}
			<PlayButton hits={track.notes} label="Listen to track" metronome={activeMetronome} speed={previewSpeed} />
		{/snippet}
	</PageHeader>

	<div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
		<Card title="1. Connect your kit" description="Your e-drum kit, connected by USB.">
			<div class="flex flex-col gap-3">
				<MidiDeviceSelector bind:selectedId={selectedDeviceId} />
				{#if selectedDeviceId}
					<PadCheck deviceId={selectedDeviceId} kit={activeKit} />
					<p class="text-sm text-fg-muted">
						Pads: <span class="text-fg">{activeKit ? activeKit.name : 'standard General MIDI'}</span> ·
						<a href="/kit" class="text-accent-text hover:underline">{activeKit ? 'change' : 'set up my kit'}</a>
					</p>
					<div class="flex flex-wrap items-center gap-x-3 gap-y-1 text-sm">
						<span class="text-fg-muted">
							Timing correction:
							<span class="text-fg tabular-nums">{latency === null ? 'not calibrated' : `${latency} ms`}</span>
						</span>
						<Button variant="ghost" size="sm" onclick={() => (calibrateOpen = true)}>
							<Gauge size={14} aria-hidden="true" />
							{latency === null ? 'Calibrate' : 'Recalibrate'}
						</Button>
					</div>
				{/if}
			</div>
		</Card>

		<Card title="2. Choose your tempo" description="Practise slower first — your score is adjusted to the tempo you pick.">
			<Metronome
				initialBpm={currentBpm}
				trackBpm={track.bpm}
				beatsPerBar={track.time_sig_numerator}
				locked={recordingActive}
				onBpmChange={handleBpmChange}
				onMetronomeChange={(m) => (activeMetronome = m)}
			/>
		</Card>
	</div>
{:else}
	<h1 class="text-xl font-bold text-fg">{track.name} <span class="text-fg-muted font-normal text-base">· {currentBpm} BPM</span></h1>
{/if}

<Card
	title="Volume"
	description={pilots.length > 0 ? 'Metronome and backing tracks that play along while you record.' : 'Metronome and count-in.'}
>
	<MixerControls
		{pilots}
		player={pilotPlayer}
		ready={pilotsReady}
		kitName={activeKit?.name ?? null}
		onChange={handleMixChange}
		onSaveAsKitDefault={activeKit ? saveMixAsKitDefault : undefined}
		onResetToKitDefault={resetMix}
	/>
	{#if pilotLoadError}
		<div class="mt-3"><Notice variant="error">{pilotLoadError}</Notice></div>
	{/if}
</Card>

<section class="flex flex-col gap-4" aria-labelledby="record-heading">
	{#if !recordingActive}<h2 id="record-heading" class="text-lg font-semibold text-fg">3. Record</h2>{/if}
	{#if submitError}
		<Notice variant="error">Your take could not be saved: {submitError}</Notice>
	{/if}
	{#if submitting}
		<div class="bg-card border border-line rounded-xl p-5 text-sm text-fg-muted" role="status">Scoring your take…</div>
	{:else}
		<RecordingControls
			deviceId={selectedDeviceId}
			bind:active={recordingActive}
			onFinished={handleFinished}
			onRecordingStart={handleRecordingStart}
			onFirstHit={(t) => startPilots(t, firstTargetMs / 1000)}
			onCountIn={(t) => startPilots(t, 0)}
			onRecordingStop={handleRecordingStop}
			expectedHits={track.notes}
			timeScale={track.bpm / currentBpm}
			bpm={track.bpm}
			practiceBpm={currentBpm}
			timeSigNumerator={track.time_sig_numerator}
			durationMs={track.duration_ms}
			firstHitTimeMs={firstTargetMs * (track.bpm / currentBpm)}
			{hitGroup}
		/>
	{/if}
</section>

<CalibrateDialog bind:open={calibrateOpen} deviceId={selectedDeviceId} />
