<script lang="ts">
	import { createPlayer, type PlayableHit } from '$lib/midi/player';
	import type { Metronome } from '$lib/midi/metronome';
	import { Play, Square, LoaderCircle } from '@lucide/svelte';

	let {
		hits,
		label = 'Play',
		metronome = undefined,
		speed = 1
	}: {
		hits: PlayableHit[];
		label?: string;
		/** When provided, playback starts on the next bar-1 beat of this metronome. */
		metronome?: Metronome | null;
		/** Playback speed ratio relative to the track's native tempo. 0.5 = half speed, 2 = double. */
		speed?: number;
	} = $props();

	let playing = $state(false);
	let waiting = $state(false);
	let player: ReturnType<typeof createPlayer> | null = null;
	let waitTimer: ReturnType<typeof setTimeout> | null = null;

	function startPlayer() {
		player = createPlayer(hits, speed);
		player.play(() => {
			player = null;
			playing = false;
		});
		playing = true;
		waiting = false;
	}

	function toggle() {
		if (playing || waiting) {
			if (waitTimer) { clearTimeout(waitTimer); waitTimer = null; }
			player?.stop();
			player = null;
			playing = false;
			waiting = false;
			return;
		}

		const delayMs = metronome?.getNextBeat1DelayMs() ?? null;
		// Only defer if the metronome is running and beat 1 is more than 50 ms away.
		if (delayMs !== null && delayMs > 50) {
			waiting = true;
			// Start the AudioContext slightly early so it's ready when the beat lands.
			waitTimer = setTimeout(() => {
				waitTimer = null;
				startPlayer();
			}, Math.max(0, delayMs - 30));
		} else {
			startPlayer();
		}
	}
</script>

<button
	type="button"
	onclick={toggle}
	disabled={hits.length === 0}
	aria-pressed={playing}
	class="inline-flex items-center gap-2 h-9 px-3 rounded-lg text-sm font-medium transition-colors cursor-pointer
		bg-control text-fg hover:bg-control-hover disabled:opacity-40 disabled:cursor-not-allowed"
>
	{#if playing}
		<Square size={14} class="text-bad" aria-hidden="true" /> Stop
	{:else if waiting}
		<LoaderCircle size={14} class="text-warn animate-spin" aria-hidden="true" /> Waiting for bar 1…
	{:else}
		<Play size={14} aria-hidden="true" /> {label}
	{/if}
</button>
