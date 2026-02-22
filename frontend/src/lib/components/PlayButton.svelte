<script lang="ts">
	import { createPlayer, type PlayableHit } from '$lib/midi/player';
	import type { Metronome } from '$lib/midi/metronome';

	let {
		hits,
		label = 'Play',
		metronome = undefined
	}: {
		hits: PlayableHit[];
		label?: string;
		/** When provided, playback starts on the next bar-1 beat of this metronome. */
		metronome?: Metronome;
	} = $props();

	let playing = $state(false);
	let waiting = $state(false);
	let player: ReturnType<typeof createPlayer> | null = null;
	let waitTimer: ReturnType<typeof setTimeout> | null = null;

	function startPlayer() {
		player = createPlayer(hits);
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
	onclick={toggle}
	disabled={hits.length === 0}
	class="inline-flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-sm font-medium transition
		   bg-zinc-700 hover:bg-zinc-600 disabled:opacity-40 disabled:cursor-not-allowed text-zinc-200"
>
	{#if playing}
		<span class="text-red-400">⏹</span> Stop
	{:else if waiting}
		<span class="text-amber-400 animate-pulse">◉</span> Syncing…
	{:else}
		<span>▶</span>
		{label}
	{/if}
</button>
