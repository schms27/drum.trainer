<script lang="ts">
	import type { WeekStat } from '$lib/api/stats';

	type Bar = { week_start: string; value: number };

	let {
		data,
		valueKey,
		label,
		color = '#6366f1',
		unit = ''
	}: {
		data: WeekStat[];
		valueKey: keyof WeekStat;
		label: string;
		color?: string;
		unit?: string;
	} = $props();

	// Re-shape into a simple {week_start, value} array so the template stays readable.
	const bars: Bar[] = $derived(
		data.map((w) => ({ week_start: w.week_start, value: Number(w[valueKey]) }))
	);

	const W = 740;
	const H = 160;
	const PAD = { top: 12, right: 8, bottom: 28, left: 36 };

	const chartW = W - PAD.left - PAD.right;
	const chartH = H - PAD.top - PAD.bottom;

	const maxVal = $derived(Math.max(...bars.map((b) => b.value), 1));
	const barW = $derived(chartW / bars.length);
	const gap = $derived(Math.max(1, barW * 0.18));

	function bx(i: number) {
		return PAD.left + i * barW + gap / 2;
	}
	function bh(v: number) {
		return (v / maxVal) * chartH;
	}
	function by(v: number) {
		return PAD.top + chartH - bh(v);
	}

	// Y axis: three ticks at 0, mid, max.
	const yTicks = $derived([
		{ v: 0, y: PAD.top + chartH },
		{ v: maxVal / 2, y: PAD.top + chartH / 2 },
		{ v: maxVal, y: PAD.top }
	]);

	// X axis: show month name at the first bar of each new month.
	const xLabels = $derived(
		bars.reduce<{ x: number; label: string }[]>((acc, bar, i) => {
			// Parse as local date to avoid UTC midnight shifting to previous day.
			const [y, m, d] = bar.week_start.split('-').map(Number);
			const date = new Date(y, m - 1, d);
			const monthLabel = date.toLocaleString('default', { month: 'short' });
			const prev = acc.at(-1);
			if (!prev || prev.label !== monthLabel) {
				acc.push({ x: bx(i) + (barW - gap) / 2, label: monthLabel });
			}
			return acc;
		}, [])
	);

	// Tooltip
	type Hovered = Bar & { i: number };
	let hovered: Hovered | null = $state(null);

	function formatValue(v: number) {
		if (unit === 'min') return v.toFixed(1) + ' min';
		return String(v);
	}

	function formatWeek(s: string) {
		const [y, m, d] = s.split('-').map(Number);
		return new Date(y, m - 1, d).toLocaleDateString('default', {
			month: 'short',
			day: 'numeric',
			year: 'numeric'
		});
	}

	// Tooltip SVG position clamped so it never overflows the right edge.
	const ttW = 108;
	const ttH = 38;
	const tipPos = $derived.by(() => {
		const h = hovered;
		if (!h) return null;
		return {
			x: Math.min(bx(h.i) + (barW - gap) / 2 - ttW / 2, W - PAD.right - ttW),
			y: Math.max(PAD.top, by(h.value) - ttH - 6)
		};
	});
</script>

<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
	<h2 class="text-sm font-semibold text-zinc-300 mb-4">{label}</h2>

	<svg viewBox="0 0 {W} {H}" class="w-full" style="overflow: visible;" aria-label={label}>
		<!-- Grid + Y axis labels -->
		{#each yTicks as tick}
			<line
				x1={PAD.left}
				y1={tick.y}
				x2={W - PAD.right}
				y2={tick.y}
				stroke="#3f3f46"
				stroke-width="1"
			/>
			<text x={PAD.left - 5} y={tick.y + 4} text-anchor="end" fill="#71717a" font-size="10">
				{Math.round(tick.v)}
			</text>
		{/each}

		<!-- Bars -->
		{#each bars as bar, i}
			{@const isCurrentWeek = i === bars.length - 1}
			<!-- Invisible full-height hit target for hover -->
			<rect
				x={bx(i)}
				y={PAD.top}
				width={barW - gap}
				height={chartH}
				fill="transparent"
				onmouseenter={() => (hovered = { ...bar, i })}
				onmouseleave={() => (hovered = null)}
			/>
			{#if bar.value > 0}
				<rect
					x={bx(i)}
					y={by(bar.value)}
					width={barW - gap}
					height={bh(bar.value)}
					fill={isCurrentWeek ? '#818cf8' : color}
					rx="2"
					opacity={hovered?.i === i ? 1 : 0.75}
				/>
			{/if}
		{/each}

		<!-- X axis month labels -->
		{#each xLabels as xl}
			<text x={xl.x} y={H - 4} text-anchor="middle" fill="#71717a" font-size="10">
				{xl.label}
			</text>
		{/each}

		<!-- Tooltip -->
		{#if hovered && tipPos}
			<rect x={tipPos.x} y={tipPos.y} width={ttW} height={ttH} rx="4" fill="#18181b" stroke="#3f3f46" />
			<text
				x={tipPos.x + ttW / 2}
				y={tipPos.y + 14}
				text-anchor="middle"
				fill="#a1a1aa"
				font-size="10"
			>{formatWeek(hovered.week_start)}</text>
			<text
				x={tipPos.x + ttW / 2}
				y={tipPos.y + 28}
				text-anchor="middle"
				fill={color}
				font-size="12"
				font-weight="600"
			>{formatValue(hovered.value)}</text>
		{/if}
	</svg>
</div>
