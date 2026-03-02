<script lang="ts">
	import type { NoteEvent } from '$lib/types/track';
	import type { RecordedHit } from '$lib/types/session';

	const GM_GROUP: Record<number, string> = {
		35: 'Bass Drum', 36: 'Bass Drum',
		37: 'Snare', 38: 'Snare', 40: 'Snare',
		42: 'Hi-Hat', 44: 'Hi-Hat', 46: 'Hi-Hat',
		49: 'Crash', 57: 'Crash',
		51: 'Ride', 53: 'Ride', 59: 'Ride',
		41: 'Tom', 43: 'Tom', 45: 'Tom', 47: 'Tom', 48: 'Tom', 50: 'Tom'
	};

	// Lane order (top → bottom) and their colors.
	const LANES = ['Bass Drum', 'Snare', 'Hi-Hat', 'Crash', 'Ride', 'Tom'] as const;
	type Lane = (typeof LANES)[number];

	const COLORS: Record<Lane, string> = {
		'Bass Drum': '#f59e0b',
		Snare: '#3b82f6',
		'Hi-Hat': '#22c55e',
		Crash: '#ef4444',
		Ride: '#a855f7',
		Tom: '#eab308'
	};

	// Layout constants (all in px).
	const ROW_H = 22; // height of one instrument lane
	const MARK_H = 8; // height of a hit marker bar within a lane
	const TICK_AREA_H = 12; // space at bottom for time labels
	const SVG_H = LANES.length * ROW_H + TICK_AREA_H;

	// Within each row:  expected = top half,  actual = bottom half.
	const EXPECTED_Y_OFF = 1; // from row top
	const ACTUAL_Y_OFF = ROW_H - MARK_H - 1; // from row top

	const PLAYHEAD_RATIO = 0.28;

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

	let timelineEl: HTMLDivElement | null = $state(null);
	let width = $state(600);

	$effect(() => {
		if (!timelineEl) return;
		width = timelineEl.getBoundingClientRect().width || 600;
		const ro = new ResizeObserver(() => {
			width = timelineEl!.getBoundingClientRect().width || 600;
		});
		ro.observe(timelineEl);
		return () => ro.disconnect();
	});

	function laneIndex(group: string): number {
		return (LANES as readonly string[]).indexOf(group);
	}

	function noteGroup(noteNumber: number): string {
		return GM_GROUP[noteNumber] ?? '';
	}

	const playheadX = $derived(width * PLAYHEAD_RATIO);
	const leftMs = $derived(currentTimeMs - PLAYHEAD_RATIO * windowMs);
	const rightMs = $derived(leftMs + windowMs);

	const scaledExpected = $derived(
		timeScale === 1
			? expectedHits
			: expectedHits.map((h) => ({ ...h, time_ms: h.time_ms * timeScale }))
	);

	const visibleExpected = $derived(
		scaledExpected.filter((h) => h.time_ms >= leftMs - 20 && h.time_ms <= rightMs + 20)
	);

	const visibleActual = $derived(
		actualHits.filter((h) => h.time_ms >= leftMs - 20 && h.time_ms <= rightMs + 20)
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

	function rowY(lane: number): number {
		return lane * ROW_H;
	}
</script>

<div class="bg-zinc-900 rounded-xl border border-zinc-700 overflow-hidden">
	<div class="px-4 pt-3 pb-2">
		<h3 class="text-xs text-zinc-500 uppercase tracking-wider">Timeline</h3>
	</div>

	<div class="flex" style="height:{SVG_H}px;">
		<!-- Instrument label column -->
		<div class="flex-shrink-0 flex flex-col border-r border-zinc-800" style="width:72px;">
			{#each LANES as lane, i}
				<div
					class="flex items-center justify-end pr-2 text-[10px] font-mono select-none"
					style="height:{ROW_H}px; color:{COLORS[lane]}; background:{i % 2 === 0
						? '#0f0f11'
						: '#111114'};"
				>
					{lane}
				</div>
			{/each}
			<!-- Spacer for tick label row -->
			<div style="height:{TICK_AREA_H}px;"></div>
		</div>

		<!-- Timeline SVG (edge-fade via CSS mask) -->
		<div
			bind:this={timelineEl}
			class="flex-1 overflow-hidden"
			style="mask-image: linear-gradient(to right, transparent 0%, black 4%, black 96%, transparent 100%);
			       -webkit-mask-image: linear-gradient(to right, transparent 0%, black 4%, black 96%, transparent 100%);"
		>
			<svg {width} height={SVG_H} style="display:block;">
				<!-- Lane backgrounds (alternating) -->
				{#each LANES as _lane, i}
					<rect
						x={0}
						y={rowY(i)}
						{width}
						height={ROW_H}
						fill={i % 2 === 0 ? '#0f0f11' : '#111114'}
					/>
				{/each}

				<!-- Past region overlay -->
				<rect x={0} y={0} width={playheadX} height={LANES.length * ROW_H} fill="#000" opacity="0.2" />

				<!-- Lane separator lines -->
				{#each LANES as _lane, i}
					<line
						x1={0}
						y1={rowY(i + 1) - 0.5}
						x2={width}
						y2={rowY(i + 1) - 0.5}
						stroke="#1c1c1f"
						stroke-width="1"
					/>
				{/each}

				<!-- Second-grid lines -->
				{#each gridTicks as tick}
					{@const x = timeToX(tick)}
					<line x1={x} y1={0} x2={x} y2={LANES.length * ROW_H} stroke="#27272a" stroke-width="1" />
					<text
						x={x + 2}
						y={SVG_H - 2}
						fill="#52525b"
						font-size="8"
						font-family="monospace"
						dominant-baseline="auto">{(tick / 1000).toFixed(0)}s</text
					>
				{/each}

				<!-- Expected hits -->
				{#each visibleExpected as hit}
					{@const li = laneIndex(hit.instrument)}
					{#if li >= 0}
						{@const x = timeToX(hit.time_ms)}
						{@const y = rowY(li) + EXPECTED_Y_OFF}
						<rect
							x={x - 1}
							{y}
							width={2}
							height={MARK_H}
							fill={COLORS[hit.instrument as Lane] ?? '#71717a'}
							opacity="0.45"
							rx="0.5"
						/>
					{/if}
				{/each}

				<!-- Actual hits -->
				{#each visibleActual as hit}
					{@const group = noteGroup(hit.note_number)}
					{@const li = laneIndex(group)}
					{#if li >= 0}
						{@const x = timeToX(hit.time_ms)}
						{@const y = rowY(li) + ACTUAL_Y_OFF}
						<rect
							x={x - 1.5}
							{y}
							width={3}
							height={MARK_H}
							fill={COLORS[group as Lane] ?? '#71717a'}
							opacity={0.35 + (hit.velocity / 127) * 0.65}
							rx="0.5"
						/>
					{/if}
				{/each}

				<!-- Playhead -->
				<line
					x1={playheadX}
					y1={0}
					x2={playheadX}
					y2={LANES.length * ROW_H}
					stroke="#ef4444"
					stroke-width="1.5"
					opacity="0.8"
				/>
				<!-- Playhead tick at top -->
				<polygon
					points="{playheadX - 4},0 {playheadX + 4},0 {playheadX},6"
					fill="#ef4444"
					opacity="0.8"
				/>
			</svg>
		</div>
	</div>

	<!-- Legend -->
	<div class="flex items-center gap-4 px-4 py-2 border-t border-zinc-800">
		<div class="flex items-center gap-1.5 text-[10px] text-zinc-500 font-mono">
			<span class="inline-block w-4 h-2 rounded-sm bg-zinc-500 opacity-45"></span>
			expected
		</div>
		<div class="flex items-center gap-1.5 text-[10px] text-zinc-500 font-mono">
			<span class="inline-block w-4 h-2 rounded-sm bg-zinc-300"></span>
			played
		</div>
	</div>
</div>
