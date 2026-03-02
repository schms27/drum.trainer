<script lang="ts">
	import type { NoteEvent } from '$lib/types/track';
	import type { RecordedHit } from '$lib/types/session';

	// Coarse instrument group from MIDI note number
	const GM_GROUP: Record<number, string> = {
		35: 'Bass Drum', 36: 'Bass Drum',
		37: 'Snare', 38: 'Snare', 40: 'Snare',
		42: 'Hi-Hat', 44: 'Hi-Hat', 46: 'Hi-Hat',
		49: 'Crash', 57: 'Crash',
		51: 'Ride', 53: 'Ride', 59: 'Ride',
		41: 'Tom', 43: 'Tom', 45: 'Tom', 47: 'Tom', 48: 'Tom', 50: 'Tom'
	};

	const GROUP_COLORS: Record<string, string> = {
		'Bass Drum': '#f59e0b',
		Snare: '#3b82f6',
		'Hi-Hat': '#22c55e',
		Crash: '#ef4444',
		Ride: '#a855f7',
		Tom: '#eab308'
	};

	let {
		expectedHits,
		actualHits,
		currentTimeMs = 0,
		timeScale = 1,
		windowMs = 8000
	}: {
		expectedHits: NoteEvent[];
		actualHits: RecordedHit[];
		currentTimeMs: number;
		timeScale?: number;
		windowMs?: number;
	} = $props();

	// Playhead sits 28% from the left — shows ~2.2s of past, ~5.8s of future.
	const PLAYHEAD_RATIO = 0.28;
	const HEIGHT = 110;
	const AXIS_Y = 54;

	let container: HTMLDivElement | null = $state(null);
	let width = $state(600);

	$effect(() => {
		if (!container) return;
		width = container.getBoundingClientRect().width || 600;
		const ro = new ResizeObserver(() => {
			width = container!.getBoundingClientRect().width || 600;
		});
		ro.observe(container);
		return () => ro.disconnect();
	});

	function groupColor(instrument: string): string {
		for (const [key, color] of Object.entries(GROUP_COLORS)) {
			if (instrument.includes(key)) return color;
		}
		return '#71717a';
	}

	function noteGroupColor(noteNumber: number): string {
		return groupColor(GM_GROUP[noteNumber] ?? '');
	}

	// Scale expected times to the current practice tempo.
	const scaledExpected = $derived(
		timeScale === 1
			? expectedHits
			: expectedHits.map((h) => ({ ...h, time_ms: h.time_ms * timeScale }))
	);

	const playheadX = $derived(width * PLAYHEAD_RATIO);
	const leftMs = $derived(currentTimeMs - PLAYHEAD_RATIO * windowMs);
	const rightMs = $derived(leftMs + windowMs);

	const visibleExpected = $derived(
		scaledExpected.filter((h) => h.time_ms >= leftMs - 50 && h.time_ms <= rightMs + 50)
	);

	const visibleActual = $derived(
		actualHits.filter((h) => h.time_ms >= leftMs - 50 && h.time_ms <= rightMs + 50)
	);

	const gridTicks = $derived((() => {
		const first = Math.ceil(Math.max(0, leftMs) / 1000) * 1000;
		const ticks: number[] = [];
		for (let t = first; t <= rightMs; t += 1000) ticks.push(t);
		return ticks;
	})());

	function timeToX(timeMs: number): number {
		return playheadX + ((timeMs - currentTimeMs) / windowMs) * width;
	}
</script>

<div class="bg-zinc-900 rounded-xl border border-zinc-700 overflow-hidden">
	<!-- Header -->
	<div class="flex flex-wrap items-center justify-between gap-2 px-4 pt-3 pb-2">
		<h3 class="text-xs text-zinc-500 uppercase tracking-wider">Timeline</h3>
		<div class="flex flex-wrap gap-x-3 gap-y-1">
			{#each Object.entries(GROUP_COLORS) as [name, color]}
				<span class="flex items-center gap-1 text-xs text-zinc-400">
					<span class="w-2 h-2 rounded-sm" style="background:{color}"></span>
					{name}
				</span>
			{/each}
		</div>
	</div>

	<!-- Row labels -->
	<div class="flex text-[10px] text-zinc-600 px-4 pb-1 gap-6 font-mono select-none">
		<span>▲ expected</span>
		<span>▼ played</span>
	</div>

	<!-- SVG canvas (edge-fade via CSS mask) -->
	<div
		bind:this={container}
		class="w-full"
		style="mask-image: linear-gradient(to right, transparent 0%, black 5%, black 95%, transparent 100%);
		       -webkit-mask-image: linear-gradient(to right, transparent 0%, black 5%, black 95%, transparent 100%);"
	>
		<svg {width} height={HEIGHT} style="display:block;">
			<!-- Dark background -->
			<rect x={0} y={0} {width} height={HEIGHT} fill="#0f0f11" />

			<!-- Past region (subtle shade) -->
			<rect x={0} y={0} width={playheadX} height={HEIGHT} fill="#000" opacity="0.25" />

			<!-- Second-grid lines -->
			{#each gridTicks as tick}
				{@const x = timeToX(tick)}
				<line x1={x} y1={0} x2={x} y2={HEIGHT} stroke="#27272a" stroke-width="1" />
				<text
					x={x + 3}
					y={AXIS_Y - 6}
					fill="#3f3f46"
					font-size="8"
					font-family="monospace"
					dominant-baseline="auto">{(tick / 1000).toFixed(0)}s</text
				>
			{/each}

			<!-- Axis line -->
			<line x1={0} y1={AXIS_Y} x2={width} y2={AXIS_Y} stroke="#3f3f46" stroke-width="1" />

			<!-- Expected hits — above axis -->
			{#each visibleExpected as hit}
				{@const x = timeToX(hit.time_ms)}
				<rect
					x={x - 1.5}
					y={10}
					width={3}
					height={AXIS_Y - 16}
					fill={groupColor(hit.instrument)}
					opacity="0.75"
					rx="1"
				/>
			{/each}

			<!-- Actual hits — below axis -->
			{#each visibleActual as hit}
				{@const x = timeToX(hit.time_ms)}
				<rect
					x={x - 1.5}
					y={AXIS_Y + 4}
					width={3}
					height={HEIGHT - AXIS_Y - 10}
					fill={noteGroupColor(hit.note_number)}
					opacity={0.35 + (hit.velocity / 127) * 0.65}
					rx="1"
				/>
			{/each}

			<!-- Playhead -->
			<line
				x1={playheadX}
				y1={0}
				x2={playheadX}
				y2={HEIGHT}
				stroke="#ef4444"
				stroke-width="2"
				opacity="0.85"
			/>
			<!-- Playhead arrowhead pointing to axis -->
			<polygon
				points="{playheadX - 5},{AXIS_Y - 12} {playheadX + 5},{AXIS_Y - 12} {playheadX},{AXIS_Y - 2}"
				fill="#ef4444"
				opacity="0.85"
			/>
		</svg>
	</div>
</div>
