<script lang="ts">
	import type { MatchedHit } from '$lib/types/session';
	import * as d3 from 'd3';
	import { CLOSE_MS, ON_TIME_MS, timingTone, toneColor } from '$lib/score-format';

	let { hits = [] }: { hits: MatchedHit[] | null } = $props();
	const safeHits = $derived(hits ?? []);

	const WIDTH = 560;
	const HEIGHT = 200;
	const MARGIN = { top: 16, right: 32, bottom: 40, left: 40 };
	const BIN_COUNT = 20;

	const innerW = WIDTH - MARGIN.left - MARGIN.right;
	const innerH = HEIGHT - MARGIN.top - MARGIN.bottom;

	const offsets = $derived(safeHits.map((h) => h.offset_ms));

	// Zoom the x axis to the data: a symmetric range just wide enough for the
	// largest offset, in 25 ms steps, between ±50 ms and the ±200 ms match window.
	const extent = $derived(
		Math.min(200, Math.max(50, Math.ceil((d3.max(offsets, (o) => Math.abs(o)) ?? 0) / 25) * 25))
	);
	const domain = $derived<[number, number]>([-extent, extent]);

	const x = $derived(d3.scaleLinear().domain(domain).range([0, innerW]));

	const bins = $derived(
		d3
			.bin()
			.domain(domain)
			.thresholds(d3.range(-extent, extent, (2 * extent) / BIN_COUNT))(offsets)
	);

	const yMax = $derived(d3.max(bins, (b: d3.Bin<number, number>) => b.length) ?? 1);
	const y = $derived(d3.scaleLinear().domain([0, yMax]).range([innerH, 0]).nice());

	const xTicks = $derived(x.ticks(5));
	const yTicks = $derived(y.ticks(4).filter((t) => Number.isInteger(t)));

	function barColor(x0: number | undefined, x1: number | undefined): string {
		return toneColor[timingTone(((x0 ?? 0) + (x1 ?? 0)) / 2)];
	}

	const early = $derived(offsets.filter((o) => o < -ON_TIME_MS).length);
	const late = $derived(offsets.filter((o) => o > ON_TIME_MS).length);
</script>

<div class="bg-card border border-line rounded-xl p-5">
	<div class="flex flex-wrap items-baseline justify-between gap-2 mb-4">
		<h2 class="text-sm font-semibold text-fg">Early or late?</h2>
		{#if safeHits.length > 0}
			<p class="text-xs text-fg-muted">
				{early} early · {safeHits.length - early - late} on time · {late} late
			</p>
		{/if}
	</div>
	{#if safeHits.length === 0}
		<p class="text-fg-subtle text-sm">No notes were hit, so there is no timing to show.</p>
	{:else}
		<svg viewBox="0 0 {WIDTH} {HEIGHT}" class="w-full" role="img" aria-label="How early or late each hit was">
			<g transform="translate({MARGIN.left},{MARGIN.top})">
				<!-- Y grid -->
				{#each yTicks as tick}
					<line x1={0} y1={y(tick)} x2={innerW} y2={y(tick)} style:stroke="var(--chart-grid)" />
					<text x={-6} y={y(tick) + 4} text-anchor="end" style:fill="var(--fg-subtle)" font-size="10">{tick}</text>
				{/each}

				<!-- Bars -->
				{#each bins as bin}
					{#if bin.x0 !== undefined && bin.x1 !== undefined && bin.length > 0}
						<rect
							x={x(bin.x0) + 1}
							y={y(bin.length)}
							width={Math.max(0, x(bin.x1) - x(bin.x0) - 2)}
							height={innerH - y(bin.length)}
							rx="2"
							style:fill={barColor(bin.x0, bin.x1)}
						>
							<title>{bin.length} hits between {Math.round(bin.x0)} and {Math.round(bin.x1)} ms</title>
						</rect>
					{/if}
				{/each}

				<!-- Zero line -->
				<line x1={x(0)} y1={0} x2={x(0)} y2={innerH} style:stroke="var(--fg-subtle)" stroke-dasharray="4 3" />

				<!-- X axis -->
				<line x1={0} y1={innerH} x2={innerW} y2={innerH} style:stroke="var(--chart-grid-strong)" />
				{#each xTicks as tick}
					<text x={x(tick)} y={innerH + 16} text-anchor="middle" style:fill="var(--fg-muted)" font-size="11"
						>{tick > 0 ? `+${tick}` : tick === 0 ? '0' : `−${-tick}`} ms</text
					>
				{/each}
				<text x={0} y={innerH + 34} text-anchor="start" style:fill="var(--fg-subtle)" font-size="11">← early</text>
				<text x={innerW} y={innerH + 34} text-anchor="end" style:fill="var(--fg-subtle)" font-size="11">late →</text>
			</g>
		</svg>
		<div class="flex flex-wrap gap-x-4 gap-y-1 mt-2 justify-center text-xs text-fg-muted">
			<span class="inline-flex items-center gap-1"><span class="w-3 h-3 rounded-sm bg-good"></span>On time (within {ON_TIME_MS} ms)</span>
			<span class="inline-flex items-center gap-1"><span class="w-3 h-3 rounded-sm bg-warn"></span>Close (within {CLOSE_MS} ms)</span>
			<span class="inline-flex items-center gap-1"><span class="w-3 h-3 rounded-sm bg-bad"></span>Off by more than {CLOSE_MS} ms</span>
		</div>
	{/if}
</div>
