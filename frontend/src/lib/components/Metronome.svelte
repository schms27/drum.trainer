<script lang="ts">
	import { Minus, Plus, RotateCcw, Square, Timer } from '@lucide/svelte';
	import { createMetronome, type Metronome } from '$lib/midi/metronome';

	let {
		initialBpm = 120,
		trackBpm = undefined,
		beatsPerBar = 4,
		locked = false,
		onBpmChange = undefined,
		onMetronomeChange = undefined
	}: {
		initialBpm?: number;
		/** Native tempo of the track — shown as reference, enables reset button. */
		trackBpm?: number;
		/** Beats per bar (time signature numerator) — sets the accent and beat lights. */
		beatsPerBar?: number;
		/** While true (e.g. during a take) the click is stopped and the tempo can't change. */
		locked?: boolean;
		/** Called whenever the user changes the BPM. */
		onBpmChange?: (bpm: number) => void;
		/** Called with the live Metronome instance when started, null when stopped. */
		onMetronomeChange?: (m: Metronome | null) => void;
	} = $props();

	// svelte-ignore state_referenced_locally
	let bpm = $state(Math.round(initialBpm));
	let playing = $state(false);
	let activeBeat = $state(-1);
	let metronome: Metronome | null = null;

	const nativeBpm = $derived(trackBpm !== undefined ? Math.round(trackBpm) : undefined);
	const isModified = $derived(nativeBpm !== undefined && bpm !== nativeBpm);
	const speedPct = $derived(nativeBpm ? Math.round((bpm / nativeBpm) * 100) : 100);

	const PRESETS = [50, 75, 90, 100];

	// A take has its own click on the recording clock — stop this one so the
	// two never play at once.
	$effect(() => {
		if (locked && playing) toggle();
	});

	// Never leave a click running after the component is gone.
	$effect(() => () => {
		metronome?.stop();
		metronome = null;
	});

	function onBeat(beatIndex: number) {
		activeBeat = beatIndex;
		setTimeout(() => (activeBeat = -1), 80);
	}

	function toggle() {
		if (playing) {
			metronome?.stop();
			metronome = null;
			playing = false;
			activeBeat = -1;
			onMetronomeChange?.(null);
		} else {
			metronome = createMetronome(bpm, { onBeat, beatsPerBar });
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
</script>

<div class="flex flex-wrap items-center gap-x-6 gap-y-4">
	<!-- Tempo -->
	<div class="flex items-center gap-2">
		<button
			type="button"
			onclick={() => applyBpm(bpm - 1)}
			disabled={locked}
			aria-label="Slower"
			class="w-9 h-9 rounded-lg bg-control hover:bg-control-hover text-fg flex items-center justify-center cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
		>
			<Minus size={16} aria-hidden="true" />
		</button>
		<label class="flex items-baseline gap-1.5">
			<span class="sr-only">Practice tempo in BPM</span>
			<input
				type="number"
				min="20"
				max="300"
				value={bpm}
				oninput={handleBpmInput}
				disabled={locked}
				class="w-16 h-9 text-center bg-control border border-line rounded-lg text-fg text-base font-mono tabular-nums
					focus:outline-none focus:ring-2 focus:ring-accent
					[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none"
			/>
			<span class="text-sm text-fg-muted">BPM</span>
		</label>
		<button
			type="button"
			onclick={() => applyBpm(bpm + 1)}
			disabled={locked}
			aria-label="Faster"
			class="w-9 h-9 rounded-lg bg-control hover:bg-control-hover text-fg flex items-center justify-center cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
		>
			<Plus size={16} aria-hidden="true" />
		</button>
	</div>

	{#if nativeBpm !== undefined}
		<div class="flex items-center gap-2 text-sm">
			<span class="text-fg-muted">
				{#if isModified}
					{speedPct}% of the original {nativeBpm} BPM
				{:else}
					Original tempo
				{/if}
			</span>
			{#if isModified}
				<button
					type="button"
					onclick={() => applyBpm(nativeBpm)}
					class="inline-flex items-center gap-1 text-accent-text hover:underline cursor-pointer"
				>
					<RotateCcw size={13} aria-hidden="true" /> Reset
				</button>
			{/if}
		</div>
	{/if}

	{#if nativeBpm !== undefined}
		<div class="flex flex-wrap items-center gap-1.5 w-full" role="group" aria-label="Speed presets">
			<span class="text-sm text-fg-muted mr-1">Speed</span>
			{#each PRESETS as pct (pct)}
				{@const target = Math.round((nativeBpm * pct) / 100)}
				<button
					type="button"
					onclick={() => applyBpm(target)}
					disabled={locked}
					aria-pressed={bpm === target}
					class="h-8 px-3 rounded-lg text-sm tabular-nums transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed
						{bpm === target ? 'bg-accent text-accent-fg' : 'bg-control text-fg hover:bg-control-hover'}"
				>
					{pct}%
				</button>
			{/each}
		</div>
	{/if}

	<!-- Metronome -->
	<div class="flex items-center gap-3 sm:ml-auto">
		<div class="flex gap-1.5" aria-hidden="true">
			{#each { length: beatsPerBar } as _, i}
				<div
					class="w-2.5 h-2.5 rounded-full transition-colors duration-75
						{activeBeat === i ? (i === 0 ? 'bg-warn-solid' : 'bg-fg') : 'bg-control-hover'}"
				></div>
			{/each}
		</div>
		<button
			type="button"
			onclick={toggle}
			disabled={locked}
			aria-pressed={playing}
			class="inline-flex items-center gap-2 h-9 px-3 rounded-lg text-sm font-medium transition-colors cursor-pointer
				disabled:opacity-40 disabled:cursor-not-allowed
				{playing ? 'bg-warn-solid text-white hover:opacity-90' : 'bg-control text-fg hover:bg-control-hover'}"
		>
			{#if playing}
				<Square size={14} aria-hidden="true" /> Stop click
			{:else}
				<Timer size={14} aria-hidden="true" /> Metronome
			{/if}
		</button>
	</div>
</div>
