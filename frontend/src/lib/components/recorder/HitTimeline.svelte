<script lang="ts">
	import { onMount } from 'svelte';
	import { ZoomIn, ZoomOut, Radio } from '@lucide/svelte';
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

	// Playhead sits 28 % from the left of the track area.
	const PLAYHEAD_RATIO = 0.28;
	const PLAYHEAD_X = LABEL_W + TRACK_W * PLAYHEAD_RATIO;

	let {
		expectedHits,
		actualHits,
		currentTimeMs = 0,
		timeScale = 1,
		bpm = 120,
		timeSigNumerator = 4,
		hitGroup = (note: number) => GM_GROUP[note] ?? ''
	}: {
		expectedHits: NoteEvent[];
		actualHits: RecordedHit[];
		currentTimeMs: number;
		timeScale?: number;
		bpm?: number;
		timeSigNumerator?: number;
		/** Instrument for a note from the kit (the active kit profile, or General MIDI). */
		hitGroup?: (note: number) => string;
	} = $props();

	// Only show lanes the track uses, plus any instrument the drummer actually
	// hits (so stray hits on unused pads stay visible), in kit order.
	const lanes = $derived(
		LANES.filter(
			(lane) =>
				expectedHits.some((h) => h.instrument === lane) ||
				actualHits.some((h) => hitGroup(h.note_number) === lane)
		)
	);
	const svgH = $derived(MARGIN_TOP + lanes.length * ROW_H + MARGIN_BOTTOM);
	const laneIndex = (instrument: string) => (lanes as readonly string[]).indexOf(instrument);

	// ── Zoom & scroll state ──────────────────────────────────────────────────
	let localWindowMs = $state(8000); // visible time window (zoom level)
	let viewAnchorMs = $state(0); // time shown at the playhead X position
	let following = $state(true); // auto-track the live recording position

	// Sync view to live position when following; re-engage on reset.
	$effect(() => {
		if (currentTimeMs === 0) following = true;
		if (following) viewAnchorMs = currentTimeMs;
	});

	let containerEl: HTMLDivElement;

	onMount(() => {
		// Plain vertical scrolling is left to the page. Only Ctrl+wheel (zoom) and
		// Shift+wheel / horizontal trackpad swipes (pan) are captured.
		const handler = (e: WheelEvent) => {
			if (e.ctrlKey) {
				e.preventDefault();
				const factor = e.deltaY > 0 ? 1.3 : 1 / 1.3;
				localWindowMs = Math.max(1000, Math.min(40000, localWindowMs * factor));
				return;
			}
			const horizontal = Math.abs(e.deltaX) > Math.abs(e.deltaY);
			if (!e.shiftKey && !horizontal) return;
			e.preventDefault();
			following = false;
			const delta = horizontal ? e.deltaX : e.deltaY;
			const msPerPx = localWindowMs / TRACK_W;
			viewAnchorMs = Math.max(0, viewAnchorMs + delta * msPerPx * 0.8);
		};
		containerEl.addEventListener('wheel', handler, { passive: false });
		return () => containerEl.removeEventListener('wheel', handler);
	});

	function zoomIn() {
		localWindowMs = Math.max(1000, localWindowMs / 1.5);
	}
	function zoomOut() {
		localWindowMs = Math.min(40000, localWindowMs * 1.5);
	}
	function goLive() {
		following = true;
	}

	const windowLabel = $derived(
		localWindowMs >= 10000
			? `${Math.round(localWindowMs / 1000)} s`
			: `${(localWindowMs / 1000).toFixed(1)} s`
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
			const group = hitGroup(hit.note_number);
			const li = laneIndex(group);
			if (li < 0) return [];

			let bestDelta = Infinity;
			for (const h of scaledExpected) {
				if (h.instrument !== group) continue;
				const d = Math.abs(h.time_ms - hit.time_ms);
				if (d < bestDelta) bestDelta = d;
			}

			const color =
				bestDelta <= 20 ? 'var(--good)' : bestDelta <= 50 ? 'var(--warn)' : 'var(--bad)';

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

	const iconButton =
		'w-8 h-8 flex items-center justify-center bg-control hover:bg-control-hover text-fg rounded-lg transition-colors cursor-pointer';
</script>

<div class="bg-card border border-line rounded-xl p-5 overflow-x-auto">
	<!-- Header with zoom controls and live-follow toggle -->
	<div class="flex flex-wrap items-center justify-between gap-2 mb-4">
		<h3 class="text-sm font-semibold text-fg">What to play</h3>
		<div class="flex items-center gap-2">
			<div class="flex items-center gap-1">
				<button type="button" onclick={zoomOut} class={iconButton} aria-label="Zoom out" title="Zoom out">
					<ZoomOut size={15} aria-hidden="true" />
				</button>
				<span class="text-xs text-fg-subtle w-10 text-center tabular-nums" title="Visible time span">{windowLabel}</span>
				<button type="button" onclick={zoomIn} class={iconButton} aria-label="Zoom in" title="Zoom in">
					<ZoomIn size={15} aria-hidden="true" />
				</button>
			</div>
			<button
				type="button"
				onclick={goLive}
				aria-pressed={following}
				class="inline-flex items-center gap-1.5 h-8 px-2.5 rounded-lg text-xs font-medium transition-colors cursor-pointer
					{following ? 'bg-bad-soft text-bad' : 'bg-control hover:bg-control-hover text-fg-muted'}"
				title="Follow the current playing position"
			>
				<Radio size={13} aria-hidden="true" />
				{following ? 'Following' : 'Follow'}
			</button>
		</div>
	</div>

	{#if lanes.length === 0}
		<p class="text-sm text-fg-subtle">This track has no notes on the standard drum kit instruments.</p>
	{/if}

	<div bind:this={containerEl} class="select-none">
		<svg viewBox="0 0 {SVG_W} {svgH}" class="w-full min-w-[400px]" aria-label="Live hit timeline">
			<defs>
				<!-- Clip time-positioned content to the track area so it never bleeds into the label column. -->
				<clipPath id="htl-track-clip">
					<rect x={LABEL_W} y={0} width={TRACK_W} height={svgH} />
				</clipPath>
			</defs>

			<!-- ── Row backgrounds & labels (not masked so they're always crisp) ── -->
			{#each lanes as lane, i (lane)}
				<rect
					x={0}
					y={MARGIN_TOP + i * ROW_H}
					width={SVG_W}
					height={ROW_H}
					style:fill={i % 2 === 0 ? 'var(--chart-row)' : 'var(--chart-row-alt)'}
				/>
				<text x={LABEL_W - 8} y={rowY(i) + 4} text-anchor="end" style:fill="var(--fg-muted)" font-size="11"
					>{lane}</text
				>
				<line
					x1={LABEL_W}
					y1={MARGIN_TOP + (i + 1) * ROW_H}
					x2={SVG_W}
					y2={MARGIN_TOP + (i + 1) * ROW_H}
					style:stroke="var(--chart-grid)"
				/>
			{/each}

			<!-- Label-column divider -->
			<line x1={LABEL_W} y1={0} x2={LABEL_W} y2={svgH} style:stroke="var(--chart-grid-strong)" />

			<!-- ── All time-positioned content (clipped to track area) ── -->
			<g clip-path="url(#htl-track-clip)">
				<!-- Past-region shade -->
				<rect
					x={LABEL_W}
					y={MARGIN_TOP}
					width={Math.max(0, playheadX - LABEL_W)}
					height={lanes.length * ROW_H}
					style:fill="var(--chart-past)"
				/>

				<!-- Beat / bar grid lines -->
				{#each beatTicks as tick (tick.ms)}
					{@const x = timeToX(tick.ms)}
					<line
						x1={x}
						y1={MARGIN_TOP}
						x2={x}
						y2={MARGIN_TOP + lanes.length * ROW_H}
						style:stroke={tick.isBar ? 'var(--chart-grid-strong)' : 'var(--chart-grid)'}
						stroke-width={tick.isBar ? 1 : 0.5}
					/>
				{/each}

				<!-- Target notes — open circles, same style as results view -->
				{#each visibleExpected as hit}
					{@const gi = laneIndex(hit.instrument)}
					{#if gi >= 0}
						<circle
							cx={timeToX(hit.time_ms)}
							cy={rowY(gi)}
							r={DOT_R}
							fill="none"
							style:stroke="var(--chart-target)"
							stroke-width="1.5"
						/>
					{/if}
				{/each}

				<!-- Actual hits — filled circles, coloured by timing accuracy -->
				{#each visibleActuals as hit}
					<circle cx={timeToX(hit.time_ms)} cy={rowY(hit.li)} r={DOT_R - 1} style:fill={hit.color} />
				{/each}

				<!-- Playhead -->
				<line
					x1={playheadX}
					y1={MARGIN_TOP}
					x2={playheadX}
					y2={MARGIN_TOP + lanes.length * ROW_H}
					style:stroke="var(--bad-solid)"
					stroke-width="1.5"
				/>
				<polygon
					points="{playheadX - 4},{MARGIN_TOP} {playheadX + 4},{MARGIN_TOP} {playheadX},{MARGIN_TOP + 8}"
					style:fill="var(--bad-solid)"
				/>
			</g>
		</svg>
	</div>

	<!-- Legend — mirrors the results view -->
	<div class="flex flex-wrap gap-x-5 gap-y-1 mt-3 text-xs text-fg-muted justify-center">
		<span class="inline-flex items-center gap-1">
			<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true">
				<circle cx="5" cy="5" r="4" fill="none" style:stroke="var(--chart-target)" stroke-width="1.5" />
			</svg>Note to play
		</span>
		<span class="inline-flex items-center gap-1">
			<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="3" style:fill="var(--good)" /></svg
			>On time (within 20 ms)
		</span>
		<span class="inline-flex items-center gap-1">
			<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="3" style:fill="var(--warn)" /></svg
			>Close (within 50 ms)
		</span>
		<span class="inline-flex items-center gap-1">
			<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="3" style:fill="var(--bad)" /></svg
			>Early, late or extra
		</span>
	</div>
	<p class="text-xs text-fg-subtle mt-2 text-center">Shift + scroll to move along the track · Ctrl + scroll to zoom</p>
</div>
