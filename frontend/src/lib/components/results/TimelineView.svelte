<script lang="ts">
	import type { MatchedHit, GhostHit, InstrumentBreakdown } from '$lib/types/session';
	import type { NoteEvent } from '$lib/types/track';

	let {
		targetNotes,
		matchedHits,
		ghostHits,
		breakdown,
		durationMs,
		bpm = 120,
		timeSigNumerator = 4
	}: {
		targetNotes: NoteEvent[];
		matchedHits: MatchedHit[] | null;
		ghostHits: GhostHit[] | null;
		breakdown: InstrumentBreakdown[] | null;
		durationMs: number;
		bpm?: number;
		timeSigNumerator?: number;
	} = $props();
	const safeHits = $derived(matchedHits ?? []);
	const safeGhosts = $derived(ghostHits ?? []);
	const safeBreakdown = $derived(breakdown ?? []);

	const WIDTH = 680;
	const ROW_H = 32;
	const LABEL_W = 90;
	const MARGIN = { top: 8, bottom: 16 };
	const DOT_R = 5;

	const groups = $derived(safeBreakdown.map((b) => b.instrument));
	const svgH = $derived(MARGIN.top + groups.length * ROW_H + MARGIN.bottom);
	const trackW = $derived(WIDTH - LABEL_W);

	// Time scale: ms → px
	const msPerPx = $derived(durationMs > 0 ? durationMs / trackW : 1);

	function xOf(ms: number) {
		return LABEL_W + ms / msPerPx;
	}

	function rowY(i: number) {
		return MARGIN.top + i * ROW_H + ROW_H / 2;
	}

	// Beat/bar grid: beat every 60000/bpm ms, bar every 4 beats (4/4 assumed).
	const beatMs = $derived(bpm > 0 ? 60000 / bpm : 0);
	const beatLines = $derived(
		(() => {
			if (beatMs <= 0 || durationMs <= 0) return [];
			const lines: { ms: number; isBar: boolean }[] = [];
			for (let n = 1; n * beatMs <= durationMs + beatMs; n++) {
				lines.push({ ms: n * beatMs, isBar: n % timeSigNumerator === 0 });
			}
			return lines;
		})()
	);

	// Color of the matched dot by offset magnitude
	function offsetDotColor(offset: number) {
		const a = Math.abs(offset);
		if (a <= 20) return '#4ade80';
		if (a <= 50) return '#facc15';
		return '#f87171';
	}
</script>

<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 overflow-x-auto">
	<h3 class="text-sm font-semibold text-zinc-300 mb-4">Timeline</h3>
	{#if targetNotes.length === 0}
		<p class="text-zinc-500 text-sm">No data.</p>
	{:else}
		<svg viewBox="0 0 {WIDTH} {svgH}" class="w-full min-w-[400px]" aria-label="Hit timeline">
			<!-- Row backgrounds -->
			{#each groups as group, i}
				<rect
					x={0} y={MARGIN.top + i * ROW_H}
					width={WIDTH} height={ROW_H}
					fill={i % 2 === 0 ? '#27272a' : '#1c1c1e'}
				/>
				<text x={LABEL_W - 8} y={rowY(i) + 4} text-anchor="end" fill="#a1a1aa" font-size="11"
					>{group}</text
				>
				<line
					x1={LABEL_W} y1={MARGIN.top + (i + 1) * ROW_H}
					x2={WIDTH} y2={MARGIN.top + (i + 1) * ROW_H}
					stroke="#3f3f46"
				/>
			{/each}

			<!-- Beat / bar grid lines (drawn before notes so they're behind) -->
			{#each beatLines as line}
				{@const x = xOf(line.ms)}
				{#if x >= LABEL_W && x <= WIDTH}
					<line
						x1={x} y1={MARGIN.top}
						x2={x} y2={MARGIN.top + groups.length * ROW_H}
						stroke={line.isBar ? '#52525b' : '#3f3f46'}
						stroke-width={line.isBar ? 1 : 0.5}
						opacity={line.isBar ? 0.9 : 0.5}
					/>
				{/if}
			{/each}

			<!-- Target notes: open circles -->
			{#each targetNotes as note}
				{@const gi = groups.indexOf(note.instrument)}
				{#if gi >= 0}
					<circle
						cx={xOf(note.time_ms)}
						cy={rowY(gi)}
						r={DOT_R}
						fill="none"
						stroke="#6366f1"
						stroke-width="1.5"
						opacity="0.7"
					/>
				{/if}
			{/each}

			<!-- Matched hits: filled circles, colored by offset -->
			{#each safeHits as hit}
				{@const gi = groups.indexOf(hit.instrument)}
				{#if gi >= 0}
					<circle
						cx={xOf(hit.recorded_time_ms)}
						cy={rowY(gi)}
						r={DOT_R - 1}
						fill={offsetDotColor(hit.offset_ms)}
						opacity="0.9"
					>
						<title>{hit.instrument} | offset: {hit.offset_ms.toFixed(1)}ms</title>
					</circle>
				{/if}
			{/each}

			<!-- Ghost hits: hollow red circles -->
			{#each safeGhosts as ghost}
				{@const gi = groups.indexOf(ghost.instrument)}
				{#if gi >= 0}
					<circle
						cx={xOf(ghost.recorded_time_ms)}
						cy={rowY(gi)}
						r={DOT_R - 1}
						fill="none"
						stroke="#f87171"
						stroke-width="1.5"
						opacity="0.85"
					>
						<title>{ghost.instrument} | ghost note</title>
					</circle>
				{/if}
			{/each}

			<!-- Label column divider -->
			<line x1={LABEL_W} y1={0} x2={LABEL_W} y2={svgH} stroke="#52525b" />
		</svg>

		<div class="flex flex-wrap gap-5 mt-3 text-xs text-zinc-500 justify-center">
			<span>
				<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
					<circle cx="5" cy="5" r="4" fill="none" stroke="#6366f1" stroke-width="1.5" />
				</svg>Target
			</span>
			<span>
				<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
					<circle cx="5" cy="5" r="3" fill="#4ade80" />
				</svg>Hit (±20ms)
			</span>
			<span>
				<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
					<circle cx="5" cy="5" r="3" fill="#facc15" />
				</svg>Hit (±50ms)
			</span>
			<span>
				<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
					<circle cx="5" cy="5" r="3" fill="#f87171" />
				</svg>Late/Early
			</span>
			<span>
				<svg class="inline w-3 h-3 mr-1" viewBox="0 0 10 10">
					<circle cx="5" cy="5" r="3" fill="none" stroke="#f87171" stroke-width="1.5" />
				</svg>Ghost
			</span>
		</div>
	{/if}
</div>
