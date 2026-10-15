<script lang="ts">
	import { onMount } from 'svelte';
	import type { NoteEvent } from '$lib/types/track';
	import type { RecordedHit } from '$lib/types/session';
	import { GM_GROUP, LANES } from '$lib/drums';

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
		bpm = 120,
		timeSigNumerator = 4
	}: {
		expectedHits: NoteEvent[];
		actualHits: RecordedHit[];
		currentTimeMs: number;
		timeScale?: number;
		bpm?: number;
		timeSigNumerator?: number;
	} = $props();

	// ── Zoom & scroll state ──────────────────────────────────────────────────
	let localWindowMs = $state(8000); // visible time window (zoom level)
	let viewAnchorMs = $state(0);     // time shown at the playhead X position
	let following = $state(true);     // auto-track the live recording position

	// Sync view to live position when following; re-engage on reset.
	$effect(() => {
		if (currentTimeMs === 0) following = true;
		if (following) viewAnchorMs = currentTimeMs;
	});

	let containerEl: HTMLDivElement;

	onMount(() => {
		const handler = (e: WheelEvent) => {
			e.preventDefault();
			if (e.ctrlKey) {
				// Ctrl + scroll → zoom
				const factor = e.deltaY > 0 ? 1.3 : 1 / 1.3;
				localWindowMs = Math.max(1000, Math.min(40000, localWindowMs * factor));
			} else {
				// Scroll → pan (disengage live following)
				following = false;
				const msPerPx = localWindowMs / TRACK_W;
				viewAnchorMs = Math.max(0, viewAnchorMs + e.deltaY * msPerPx * 0.8);
			}
		};
		containerEl.addEventListener('wheel', handler, { passive: false });
		return () => containerEl.removeEventListener('wheel', handler);
	});

	function zoomIn() { localWindowMs = Math.max(1000, localWindowMs / 1.5); }
	function zoomOut() { localWindowMs = Math.min(40000, localWindowMs * 1.5); }
	function goLive() { following = true; }

	const windowLabel = $derived(
		localWindowMs >= 10000
			? `${Math.round(localWindowMs / 1000)}s`
			: `${(localWindowMs / 1000).toFixed(1)}s`
	);

	// ── Derived window bounds ────────────────────────────────────────────────
	// Scale expected hit times to the current practice tempo.
	const scaledExpected = $derived(
		timeScale === 1
			? expectedHits
			: expectedHits.map((h) => ({ ...h, time_ms: h.time_ms * timeScale }))
	);

	const leftMs = $derived(viewAnchorMs - PLAYHEAD_RATIO * localWindowMs);
	const rightMs = $derived(leftMs + localWindowMs);

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

	// Beat/bar grid in the scrolling coordinate space.
	const beatMs = $derived(bpm > 0 ? (60000 / bpm) * timeScale : 0);
	const beatTicks = $derived(
		(() => {
			if (beatMs <= 0) return [];
			const first = Math.max(0, Math.floor((leftMs - EDGE_BUFFER_MS) / beatMs));
			const last = Math.ceil((rightMs + EDGE_BUFFER_MS) / beatMs);
			const ticks: { ms: number; isBar: boolean }[] = [];
			for (let n = first; n <= last; n++) {
				ticks.push({ ms: n * beatMs, isBar: n % timeSigNumerator === 0 });
			}
			return ticks;
		})()
	);

	// Playhead moves with viewAnchorMs — stays at PLAYHEAD_X when following.
	const playheadX = $derived(
		PLAYHEAD_X + ((currentTimeMs - viewAnchorMs) / localWindowMs) * TRACK_W
	);

	function timeToX(timeMs: number): number {
		return PLAYHEAD_X + ((timeMs - viewAnchorMs) / localWindowMs) * TRACK_W;
	}

	function rowY(i: number): number {
		return MARGIN_TOP + i * ROW_H + ROW_H / 2;
	}
</script>

<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 overflow-x-auto">
	<!-- Header with zoom controls and live-follow toggle -->
	<div class="flex items-center justify-between mb-4">
		<h3 class="text-sm font-semibold text-zinc-300">Timeline</h3>
		<div class="flex items-center gap-2">
			<div class="flex items-center gap-1">
				<button
					onclick={zoomIn}
					class="w-7 h-7 flex items-center justify-center bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded text-sm font-bold transition"
					title="Zoom in"
				>−</button>
				<span class="text-xs text-zinc-500 w-9 text-center tabular-nums">{windowLabel}</span>
				<button
					onclick={zoomOut}
					class="w-7 h-7 flex items-center justify-center bg-zinc-700 hover:bg-zinc-600 text-zinc-300 rounded text-sm font-bold transition"
					title="Zoom out"
				>+</button>
			</div>
			<button
				onclick={goLive}
				class="px-2.5 py-1 rounded text-xs font-medium transition {following
					? 'bg-red-950 text-red-400 border border-red-800 cursor-default'
					: 'bg-zinc-700 hover:bg-zinc-600 text-zinc-400'}"
				title="Snap to live recording position"
			>
				{following ? '● Live' : '↻ Live'}
			</button>
		</div>
	</div>

	<div bind:this={containerEl} class="select-none">
		<svg viewBox="0 0 {SVG_W} {SVG_H}" class="w-full min-w-[400px]" aria-label="Live hit timeline">
			<defs>
				<!-- Clip time-positioned content to the track area so it never bleeds into the label column. -->
				<clipPath id="htl-track-clip">
					<rect x={LABEL_W} y={0} width={TRACK_W} height={SVG_H} />
				</clipPath>
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

			<!-- ── All time-positioned content (clipped to track area) ── -->
			<g clip-path="url(#htl-track-clip)">
				<!-- Past-region shade -->
				<rect
					x={LABEL_W}
					y={MARGIN_TOP}
					width={Math.max(0, playheadX - LABEL_W)}
					height={LANES.length * ROW_H}
					fill="#000"
					opacity="0.2"
				/>

				<!-- Beat / bar grid lines -->
				{#each beatTicks as tick}
					{@const x = timeToX(tick.ms)}
					<line
						x1={x}
						y1={MARGIN_TOP}
						x2={x}
						y2={MARGIN_TOP + LANES.length * ROW_H}
						stroke={tick.isBar ? '#52525b' : '#3f3f46'}
						stroke-width={tick.isBar ? 1 : 0.5}
						opacity={tick.isBar ? 0.9 : 0.5}
					/>
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
					x1={playheadX}
					y1={MARGIN_TOP}
					x2={playheadX}
					y2={MARGIN_TOP + LANES.length * ROW_H}
					stroke="#ef4444"
					stroke-width="1.5"
					opacity="0.85"
				/>
				<polygon
					points="{playheadX - 4},{MARGIN_TOP} {playheadX + 4},{MARGIN_TOP} {playheadX},{MARGIN_TOP + 8}"
					fill="#ef4444"
					opacity="0.85"
				/>
			</g>
		</svg>
	</div>

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
	<p class="text-xs text-zinc-600 mt-2 text-center">Scroll to pan · Ctrl+scroll to zoom</p>
</div>
