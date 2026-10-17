<script lang="ts">
	import { formatPercent } from '$lib/score-format';

	// Notes hit per take over time. `takes` must be oldest first.
	let {
		takes,
		label = 'Notes hit per take'
	}: {
		takes: { id: number; hit_rate: number; created_at: string }[];
		label?: string;
	} = $props();

	const W = 680;
	const H = 200;
	const PAD = { top: 12, right: 16, bottom: 28, left: 40 };
	const innerW = W - PAD.left - PAD.right;
	const innerH = H - PAD.top - PAD.bottom;
	const Y_TICKS = [0, 0.25, 0.5, 0.75, 1];

	const x = (i: number) =>
		PAD.left + (takes.length <= 1 ? innerW / 2 : (i / (takes.length - 1)) * innerW);
	const y = (rate: number) => PAD.top + (1 - rate) * innerH;

	const path = $derived(
		takes.map((t, i) => `${i ? 'L' : 'M'}${x(i).toFixed(1)},${y(t.hit_rate).toFixed(1)}`).join(' ')
	);
	const best = $derived(takes.reduce((m, t) => Math.max(m, t.hit_rate), 0));

	// Label a handful of takes on the x axis: first, last and a few in between.
	const xLabels = $derived(
		takes.length <= 1
			? takes.map((_, i) => i)
			: [...new Set([0, Math.round((takes.length - 1) / 3), Math.round((2 * (takes.length - 1)) / 3), takes.length - 1])]
	);

	function shortDate(iso: string) {
		return new Date(iso).toLocaleDateString(undefined, { day: 'numeric', month: 'short' });
	}
</script>

<svg viewBox="0 0 {W} {H}" class="w-full" role="img" aria-label="{label}: {takes.length} takes, best {formatPercent(best)}">
	{#each Y_TICKS as tick (tick)}
		<line x1={PAD.left} y1={y(tick)} x2={W - PAD.right} y2={y(tick)} style:stroke="var(--chart-grid)" />
		<text x={PAD.left - 6} y={y(tick) + 4} text-anchor="end" font-size="11" style:fill="var(--fg-subtle)">{tick * 100}%</text>
	{/each}

	{#if takes.length > 1}
		<path d={path} fill="none" style:stroke="var(--accent-text)" stroke-width="2" stroke-linejoin="round" />
	{/if}
	{#each takes as take, i (take.id)}
		<a href="/results/{take.id}" aria-label="Take on {shortDate(take.created_at)}: {formatPercent(take.hit_rate)}">
			<circle cx={x(i)} cy={y(take.hit_rate)} r={take.hit_rate === best ? 5 : 3.5} style:fill="var(--accent-text)">
				<title>{shortDate(take.created_at)}: {formatPercent(take.hit_rate)} of notes hit</title>
			</circle>
		</a>
	{/each}

	{#each xLabels as i (i)}
		<text x={x(i)} y={H - 8} text-anchor={i === 0 && takes.length > 1 ? 'start' : i === takes.length - 1 && takes.length > 1 ? 'end' : 'middle'} font-size="11" style:fill="var(--fg-subtle)">
			{shortDate(takes[i].created_at)}
		</text>
	{/each}
</svg>
