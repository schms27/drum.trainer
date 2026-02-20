<script lang="ts">
	import { createPlayer, type PlayableHit } from '$lib/midi/player';

	let {
		hits,
		label = 'Play'
	}: {
		hits: PlayableHit[];
		label?: string;
	} = $props();

	let playing = $state(false);
	let player: ReturnType<typeof createPlayer> | null = null;

	function toggle() {
		if (playing) {
			player?.stop();
			player = null;
			playing = false;
		} else {
			player = createPlayer(hits);
			player.play(() => {
				player = null;
				playing = false;
			});
			playing = true;
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
	{:else}
		<span>▶</span>
		{label}
	{/if}
</button>
