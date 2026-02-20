<script lang="ts">
	import type { MatchedHit } from '$lib/types/session';
	import * as d3 from 'd3';

	let { hits = [] }: { hits: MatchedHit[] | null } = $props();
	const safeHits = $derived(hits ?? []);

	const WIDTH = 560;
	const HEIGHT = 200;
	const MARGIN = { top: 16, right: 16, bottom: 36, left: 40 };
	const DOMAIN: [number, number] = [-200, 200];
	const BIN_COUNT = 20;

	const innerW = WIDTH - MARGIN.left - MARGIN.right;
	const innerH = HEIGHT - MARGIN.top - MARGIN.bottom;

	const offsets = $derived(safeHits.map((h) => h.offset_ms));

	const x = $derived(d3.scaleLinear().domain(DOMAIN).range([0, innerW]));

	const bins = $derived(
		d3
			.bin()
			.domain(DOMAIN)
			.thresholds(BIN_COUNT)(offsets)
	);

	const yMax = $derived(d3.max(bins, (b) => b.length) ?? 1);
	const y = $derived(d3.scaleLinear().domain([0, yMax]).range([innerH, 0]).nice());

	const xTicks = $derived(x.ticks(5));
	const yTicks = $derived(y.ticks(4));

	function barColor(x0: number | undefined, x1: number | undefined): string {
		const center = ((x0 ?? 0) + (x1 ?? 0)) / 2;
		const abs = Math.abs(center);
		if (abs <= 20) return '#4ade80'; // green
		if (abs <= 50) return '#facc15'; // yellow
		return '#f87171'; // red
	}
</script>

<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
	<h3 class="text-sm font-semibold text-zinc-300 mb-4">Timing Distribution</h3>
	{#if safeHits.length === 0}
		<p class="text-zinc-500 text-sm">No matched hits to display.</p>
	{:else}
		<svg viewBox="0 0 {WIDTH} {HEIGHT}" class="w-full" aria-label="Timing histogram">
			<g transform="translate({MARGIN.left},{MARGIN.top})">
				<!-- Bars -->
				{#each bins as bin}
					{#if bin.x0 !== undefined && bin.x1 !== undefined}
						<rect
							x={x(bin.x0) + 1}
							y={y(bin.length)}
							width={Math.max(0, x(bin.x1) - x(bin.x0) - 2)}
							height={innerH - y(bin.length)}
							fill={barColor(bin.x0, bin.x1)}
							opacity="0.85"
						/>
					{/if}
				{/each}

				<!-- Zero line -->
				<line
					x1={x(0)} y1={0}
					x2={x(0)} y2={innerH}
					stroke="#71717a" stroke-width="1" stroke-dasharray="4 3"
				/>

				<!-- X axis -->
				<line x1={0} y1={innerH} x2={innerW} y2={innerH} stroke="#52525b" />
				{#each xTicks as tick}
					<text
						x={x(tick)}
						y={innerH + 18}
						text-anchor="middle"
						fill="#a1a1aa"
						font-size="11"
					>{tick > 0 ? `+${tick}` : tick}ms</text>
				{/each}

				<!-- Y axis -->
				{#each yTicks as tick}
					<line x1={0} y1={y(tick)} x2={innerW} y2={y(tick)} stroke="#3f3f46" />
					<text x={-6} y={y(tick) + 4} text-anchor="end" fill="#71717a" font-size="10">
						{tick}
					</text>
				{/each}

				<!-- Axis labels -->
				<text
					x={innerW / 2}
					y={innerH + 34}
					text-anchor="middle"
					fill="#71717a"
					font-size="11"
				>Timing offset (ms)</text>
			</g>
		</svg>
		<div class="flex gap-4 mt-2 justify-center text-xs text-zinc-500">
			<span><span class="inline-block w-3 h-3 rounded-sm bg-green-400 mr-1"></span>±20 ms</span>
			<span><span class="inline-block w-3 h-3 rounded-sm bg-yellow-400 mr-1"></span>±50 ms</span>
			<span><span class="inline-block w-3 h-3 rounded-sm bg-red-400 mr-1"></span>&gt;50 ms</span>
		</div>
	{/if}
</div>
