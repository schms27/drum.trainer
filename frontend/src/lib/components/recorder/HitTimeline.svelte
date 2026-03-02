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

	const LANES = ['Bass Drum', 'Snare', 'Hi-Hat', 'Crash', 'Ride', 'Tom'] as const;

	// SVG coordinate space — matches the results TimelineView so the two look identical.
	const SVG_W = 680;
	const LABEL_W = 90;
	const TRACK_W = SVG_W - LABEL_W;
	const ROW_H = 32;
	const DOT_R = 5;
	const MARGIN_TOP = 8;
	const MARGIN_BOTTOM = 20;
	const SVG_H = MARGIN_TOP + LANES.length * ROW_H + MARGIN_BOTTOM;

	// Playhead sits 28 % from the left of the track area.
	const PLAYHEAD_RATIO = 0.28;
	const PLAYHEAD_X = LABEL_W + TRACK_W * PLAYHEAD_RATIO;

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

	// Scale expected hit times to the current practice tempo.
	const scaledExpected = $derived(
		timeScale === 1
			? expectedHits
			: expectedHits.map((h) => ({ ...h, time_ms: h.time_ms * timeScale }))
	);

	const leftMs = $derived(currentTimeMs - PLAYHEAD_RATIO * windowMs);
	const rightMs = $derived(leftMs + windowMs);

	// Buffer so circles don't pop in/out abruptly at the edges.
	const EDGE_BUFFER_MS = 150;

	const visibleExpected = $derived(
		scaledExpected.filter(
			(h) => h.time_ms >= leftMs - EDGE_BUFFER_MS && h.time_ms <= rightMs + EDGE_BUFFER_MS
		)
	);

	// Pre-compute each actual hit's lane index and accuracy colour once,
	// matching it against the nearest expected hit of the same instrument.
	const coloredActuals = $derived(
		actualHits.flatMap((hit) => {
			const group = GM_GROUP[hit.note_number] ?? '';
			const li = (LANES as readonly string[]).indexOf(group);
			if (li < 0) return [];

			let bestDelta = Infinity;
			for (const h of scaledExpected) {
				if (h.instrument !== group) continue;
				const d = Math.abs(h.time_ms - hit.time_ms);
				if (d < bestDelta) bestDelta = d;
			}

			const color =
				bestDelta > 200
					? '#f87171' // ghost note — no nearby target
					: bestDelta <= 20
						? '#4ade80'
						: bestDelta <= 50
							? '#facc15'
							: '#f87171';

			return [{ time_ms: hit.time_ms, li, color }];
		})
	);

	const visibleActuals = $derived(
		coloredActuals.filter(
			(h) => h.time_ms >= leftMs - EDGE_BUFFER_MS && h.time_ms <= rightMs + EDGE_BUFFER_MS
		)
	);

	const gridTicks = $derived(
		(() => {
			const first = Math.ceil(Math.max(0, leftMs) / 1000) * 1000;
			const ticks: number[] = [];
			for (let t = first; t <= rightMs; t += 1000) ticks.push(t);
			return ticks;
		})()
	);

	function timeToX(timeMs: number): number {
		return PLAYHEAD_X + ((timeMs - currentTimeMs) / windowMs) * TRACK_W;
	}

	function rowY(i: number): number {
		return MARGIN_TOP + i * ROW_H + ROW_H / 2;
	}
</script>

<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 overflow-x-auto">
	<h3 class="text-sm font-semibold text-zinc-300 mb-4">Timeline</h3>

	<svg viewBox="0 0 {SVG_W} {SVG_H}" class="w-full min-w-[400px]" aria-label="Live hit timeline">
		<defs>
			<!-- Fade the track area at both horizontal edges so circles don't pop in/out abruptly. -->
			<linearGradient id="htl-fade" x1="0" y1="0" x2="1" y2="0">
				<stop offset="0" stop-color="white" stop-opacity="0" />
				<stop offset="0.07" stop-color="white" stop-opacity="1" />
				<stop offset="0.93" stop-color="white" stop-opacity="1" />
				<stop offset="1" stop-color="white" stop-opacity="0" />
			</linearGradient>
			<mask id="htl-mask">
				<!-- Label column is always fully visible. -->
				<rect x={0} y={0} width={LABEL_W} height={SVG_H} fill="white" />
				<!-- Track area fades at both edges. -->
				<rect x={LABEL_W} y={0} width={TRACK_W} height={SVG_H} fill="url(#htl-fade)" />
			</mask>
		</defs>

		<!-- ── Row backgrounds & labels (not masked so they're always crisp) ── -->
		{#each LANES as lane, i}
			<rect
				x={0}
				y={MARGIN_TOP + i * ROW_H}
				width={SVG_W}
				height={ROW_H}
				fill={i % 2 === 0 ? '#27272a' : '#1c1c1e'}
			/>
			<text x={LABEL_W - 8} y={rowY(i) + 4} text-anchor="end" fill="#a1a1aa" font-size="11"
				>{lane}</text
			>
			<line
				x1={LABEL_W}
				y1={MARGIN_TOP + (i + 1) * ROW_H}
				x2={SVG_W}
				y2={MARGIN_TOP + (i + 1) * ROW_H}
				stroke="#3f3f46"
			/>
		{/each}

		<!-- Label-column divider -->
		<line x1={LABEL_W} y1={0} x2={LABEL_W} y2={SVG_H} stroke="#52525b" />

		<!-- ── All time-positioned content (masked for edge fade) ── -->
		<g mask="url(#htl-mask)">
			<!-- Past-region shade -->
			<rect
				x={LABEL_W}
				y={MARGIN_TOP}
				width={PLAYHEAD_X - LABEL_W}
				height={LANES.length * ROW_H}
				fill="#000"
				opacity="0.2"
			/>

			<!-- Second grid lines + time labels -->
			{#each gridTicks as tick}
				{@const x = timeToX(tick)}
				<line
					x1={x}
					y1={MARGIN_TOP}
					x2={x}
					y2={MARGIN_TOP + LANES.length * ROW_H}
					stroke="#3f3f46"
					stroke-width="0.5"
				/>
				<text x={x} y={SVG_H - 4} text-anchor="middle" fill="#52525b" font-size="8"
					>{(tick / 1000).toFixed(0)}s</text
				>
			{/each}

			<!-- Target notes — open circles (indigo), same style as results view -->
			{#each visibleExpected as hit}
				{@const gi = (LANES as readonly string[]).indexOf(hit.instrument)}
				{#if gi >= 0}
					<circle
						cx={timeToX(hit.time_ms)}
						cy={rowY(gi)}
						r={DOT_R}
						fill="none"
						stroke="#6366f1"
						stroke-width="1.5"
						opacity="0.7"
					/>
				{/if}
			{/each}

			<!-- Actual hits — filled circles, coloured by timing accuracy -->
			{#each visibleActuals as hit}
				<circle cx={timeToX(hit.time_ms)} cy={rowY(hit.li)} r={DOT_R - 1} fill={hit.color} opacity="0.9" />
			{/each}

			<!-- Playhead -->
			<line
				x1={PLAYHEAD_X}
				y1={MARGIN_TOP}
				x2={PLAYHEAD_X}
				y2={MARGIN_TOP + LANES.length * ROW_H}
				stroke="#ef4444"
				stroke-width="1.5"
				opacity="0.85"
			/>
			<polygon
				points="{PLAYHEAD_X - 4},{MARGIN_TOP} {PLAYHEAD_X + 4},{MARGIN_TOP} {PLAYHEAD_X},{MARGIN_TOP + 8}"
				fill="#ef4444"
				opacity="0.85"
			/>
		</g>
	</svg>

	<!-- Legend — mirrors the results view -->
	<div class="flex flex-wrap gap-5 mt-3 text-xs text-zinc-500 justify-center">
		<span>
			<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
				<circle cx="5" cy="5" r="4" fill="none" stroke="#6366f1" stroke-width="1.5" />
			</svg>Target
		</span>
		<span>
			<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
				<circle cx="5" cy="5" r="3" fill="#4ade80" />
			</svg>Hit (±20 ms)
		</span>
		<span>
			<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
				<circle cx="5" cy="5" r="3" fill="#facc15" />
			</svg>Hit (±50 ms)
		</span>
		<span>
			<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
				<circle cx="5" cy="5" r="3" fill="#f87171" />
			</svg>Late / Early / Ghost
		</span>
	</div>
</div>
