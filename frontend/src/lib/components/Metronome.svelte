<script lang="ts">
	import { createMetronome, type Metronome } from '$lib/midi/metronome';

	let { initialBpm = 120 }: { initialBpm?: number } = $props();

	let bpm = $state(Math.round(initialBpm));
	let playing = $state(false);
	let activeBeat = $state(-1);

	let metronome: Metronome | null = null;

	function onBeat(beatIndex: number) {
		activeBeat = beatIndex;
		setTimeout(() => {
			activeBeat = -1;
		}, 80);
	}

	function toggle() {
		if (playing) {
			metronome?.stop();
			metronome = null;
			playing = false;
			activeBeat = -1;
		} else {
			metronome = createMetronome(bpm, { onBeat });
			metronome.start();
			playing = true;
		}
	}

	function handleBpmInput(e: Event) {
		const val = parseInt((e.target as HTMLInputElement).value, 10);
		if (!isNaN(val)) {
			bpm = Math.max(20, Math.min(300, val));
			metronome?.setBpm(bpm);
		}
	}

	function nudge(delta: number) {
		bpm = Math.max(20, Math.min(300, bpm + delta));
		metronome?.setBpm(bpm);
	}
</script>

<div class="bg-zinc-800 border border-zinc-700 rounded-xl px-4 py-3 flex items-center gap-5 flex-wrap">
	<!-- Toggle -->
	<button
		onclick={toggle}
		class="px-3 py-1.5 rounded-lg text-sm font-medium transition min-w-[110px] text-center
			   {playing
			? 'bg-amber-600 hover:bg-amber-500 text-white'
			: 'bg-zinc-700 hover:bg-zinc-600 text-zinc-300'}"
	>
		{playing ? '⏹ Stop' : '♩ Metronome'}
	</button>

	<!-- Beat dots -->
	<div class="flex gap-2">
		{#each { length: 4 } as _, i}
			<div
				class="w-3 h-3 rounded-full transition-colors duration-75
					   {activeBeat === i
					? i === 0
						? 'bg-amber-400'
						: 'bg-white'
					: 'bg-zinc-600'}"
			></div>
		{/each}
	</div>

	<!-- BPM control -->
	<div class="flex items-center gap-2 ml-auto">
		<button
			onclick={() => nudge(-1)}
			class="w-7 h-7 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-base leading-none transition flex items-center justify-center"
		>−</button>
		<input
			type="number"
			min="20"
			max="300"
			value={bpm}
			oninput={handleBpmInput}
			class="w-14 text-center bg-zinc-700 border border-zinc-600 rounded px-1 py-1
				   text-white text-sm font-mono focus:outline-none focus:ring-1 focus:ring-amber-500
				   [appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
		/>
		<button
			onclick={() => nudge(1)}
			class="w-7 h-7 rounded bg-zinc-700 hover:bg-zinc-600 text-zinc-300 text-base leading-none transition flex items-center justify-center"
		>+</button>
		<span class="text-zinc-400 text-xs">BPM</span>
	</div>
</div>
