<script lang="ts">
	import type { MatchedHit, GhostHit, InstrumentBreakdown } from '$lib/types/session';
	import type { NoteEvent } from '$lib/types/track';
	import { timingTone, toneColor, formatOffset, ON_TIME_MS, CLOSE_MS } from '$lib/score-format';

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
	const CROSS = 4;

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

	// Target notes that were matched, keyed by note number and time. Every
	// other target note was missed and gets a cross instead of a circle.
	const matchedKeys = $derived(new Set(safeHits.map((h) => `${h.note_number}@${h.target_time_ms}`)));
	const isMissed = (note: NoteEvent) => !matchedKeys.has(`${note.note_number}@${note.time_ms}`);

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
</script>

<div class="bg-card border border-line rounded-xl p-5 overflow-x-auto">
	<h2 class="text-sm font-semibold text-fg mb-4">Note by note</h2>
	{#if targetNotes.length === 0}
		<p class="text-fg-subtle text-sm">No data.</p>
	{:else}
		<svg viewBox="0 0 {WIDTH} {svgH}" class="w-full min-w-[400px]" role="img" aria-label="Every note in the track and how you played it">
			<!-- Row backgrounds -->
			{#each groups as group, i (group)}
				<rect
					x={0}
					y={MARGIN.top + i * ROW_H}
					width={WIDTH}
					height={ROW_H}
					style:fill={i % 2 === 0 ? 'var(--chart-row)' : 'var(--chart-row-alt)'}
				/>
				<text x={LABEL_W - 8} y={rowY(i) + 4} text-anchor="end" style:fill="var(--fg-muted)" font-size="11">{group}</text>
				<line
					x1={LABEL_W}
					y1={MARGIN.top + (i + 1) * ROW_H}
					x2={WIDTH}
					y2={MARGIN.top + (i + 1) * ROW_H}
					style:stroke="var(--chart-grid)"
				/>
			{/each}

			<!-- Beat / bar grid lines (drawn before notes so they're behind) -->
			{#each beatLines as line (line.ms)}
				{@const x = xOf(line.ms)}
				{#if x >= LABEL_W && x <= WIDTH}
					<line
						x1={x}
						y1={MARGIN.top}
						x2={x}
						y2={MARGIN.top + groups.length * ROW_H}
						style:stroke={line.isBar ? 'var(--chart-grid-strong)' : 'var(--chart-grid)'}
						stroke-width={line.isBar ? 1 : 0.5}
					/>
				{/if}
			{/each}

			<!-- Target notes: open circles; missed notes: crosses -->
			{#each targetNotes as note}
				{@const gi = groups.indexOf(note.instrument)}
				{#if gi >= 0}
					{@const cx = xOf(note.time_ms)}
					{@const cy = rowY(gi)}
					{#if isMissed(note)}
						<g style:stroke="var(--bad)" stroke-width="2" stroke-linecap="round">
							<line x1={cx - CROSS} y1={cy - CROSS} x2={cx + CROSS} y2={cy + CROSS} />
							<line x1={cx - CROSS} y1={cy + CROSS} x2={cx + CROSS} y2={cy - CROSS} />
							<title>{note.instrument} — missed</title>
						</g>
					{:else}
						<circle {cx} {cy} r={DOT_R} fill="none" style:stroke="var(--chart-target)" stroke-width="1.5" />
					{/if}
				{/if}
			{/each}

			<!-- Matched hits: filled circles, coloured by offset -->
			{#each safeHits as hit}
				{@const gi = groups.indexOf(hit.instrument)}
				{#if gi >= 0}
					<circle cx={xOf(hit.recorded_time_ms)} cy={rowY(gi)} r={DOT_R - 1} style:fill={toneColor[timingTone(hit.offset_ms)]}>
						<title>{hit.instrument} — {formatOffset(hit.offset_ms)} ({hit.offset_ms > 0 ? 'late' : hit.offset_ms < 0 ? 'early' : 'exact'})</title>
					</circle>
				{/if}
			{/each}

			<!-- Extra hits: hollow squares -->
			{#each safeGhosts as ghost}
				{@const gi = groups.indexOf(ghost.instrument)}
				{#if gi >= 0}
					<rect
						x={xOf(ghost.recorded_time_ms) - DOT_R + 1}
						y={rowY(gi) - DOT_R + 1}
						width={(DOT_R - 1) * 2}
						height={(DOT_R - 1) * 2}
						fill="none"
						style:stroke="var(--bad)"
						stroke-width="1.5"
					>
						<title>{ghost.instrument} — extra hit (no matching note)</title>
					</rect>
				{/if}
			{/each}

			<!-- Label column divider -->
			<line x1={LABEL_W} y1={0} x2={LABEL_W} y2={svgH} style:stroke="var(--chart-grid-strong)" />
		</svg>

		<div class="flex flex-wrap gap-x-5 gap-y-1 mt-3 text-xs text-fg-muted justify-center">
			<span class="inline-flex items-center gap-1">
				<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="4" fill="none" style:stroke="var(--chart-target)" stroke-width="1.5" /></svg>Note
			</span>
			<span class="inline-flex items-center gap-1">
				<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="3" style:fill="var(--good)" /></svg>On time (within {ON_TIME_MS} ms)
			</span>
			<span class="inline-flex items-center gap-1">
				<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="3" style:fill="var(--warn)" /></svg>Close (within {CLOSE_MS} ms)
			</span>
			<span class="inline-flex items-center gap-1">
				<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><circle cx="5" cy="5" r="3" style:fill="var(--bad)" /></svg>Early or late
			</span>
			<span class="inline-flex items-center gap-1">
				<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"
					><g style:stroke="var(--bad)" stroke-width="2" stroke-linecap="round"><line x1="2" y1="2" x2="8" y2="8" /><line x1="2" y1="8" x2="8" y2="2" /></g></svg
				>Missed
			</span>
			<span class="inline-flex items-center gap-1">
				<svg class="w-3 h-3" viewBox="0 0 10 10" aria-hidden="true"><rect x="2" y="2" width="6" height="6" fill="none" style:stroke="var(--bad)" stroke-width="1.5" /></svg>Extra hit
			</span>
		</div>
		<p class="text-xs text-fg-subtle mt-2 text-center">Hover a mark for details.</p>
	{/if}
</div>
