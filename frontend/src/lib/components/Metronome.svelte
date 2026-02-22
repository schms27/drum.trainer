<script lang="ts">
	import { createMetronome, type Metronome } from '$lib/midi/metronome';

	let {
		initialBpm = 120,
		trackBpm = undefined,
		onBpmChange = undefined,
		onMetronomeChange = undefined
	}: {
		initialBpm?: number;
		/** Native tempo of the track — shown as reference, enables reset button. */
		trackBpm?: number;
		/** Called whenever the user changes the BPM. */
		onBpmChange?: (bpm: number) => void;
		/** Called with the live Metronome instance when started, null when stopped. */
		onMetronomeChange?: (m: Metronome | null) => void;
	} = $props();

	let bpm = $state(Math.round(initialBpm));
	let playing = $state(false);
	let activeBeat = $state(-1);

	let metronome: Metronome | null = null;

	const nativeBpm = $derived(trackBpm !== undefined ? Math.round(trackBpm) : undefined);
	const isModified = $derived(nativeBpm !== undefined && bpm !== nativeBpm);

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
			onMetronomeChange?.(null);
		} else {
			metronome = createMetronome(bpm, { onBeat });
			metronome.start();
			playing = true;
			onMetronomeChange?.(metronome);
		}
	}

	function applyBpm(value: number) {
		bpm = Math.max(20, Math.min(300, value));
		metronome?.setBpm(bpm);
		onBpmChange?.(bpm);
	}

	function handleBpmInput(e: Event) {
		const val = parseInt((e.target as HTMLInputElement).value, 10);
		if (!isNaN(val)) applyBpm(val);
	}

	function nudge(delta: number) {
		applyBpm(bpm + delta);
	}

	function reset() {
		if (nativeBpm !== undefined) applyBpm(nativeBpm);
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

		{#if nativeBpm !== undefined}
			<span class="text-zinc-600 text-xs ml-1">♩={nativeBpm}</span>
			{#if isModified}
				<button
					onclick={reset}
					title="Reset to track tempo"
					class="text-zinc-500 hover:text-amber-400 text-xs transition"
				>↺</button>
			{/if}
		{/if}
	</div>
</div>
