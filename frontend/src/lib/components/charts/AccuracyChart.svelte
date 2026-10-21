<script lang="ts">
	import { formatPercent } from '$lib/score-format';
	import { describeTempo, tempoKind } from '$lib/tempo';

	// Notes hit per take over time. `takes` must be oldest first.
	let {
		takes,
		label = 'Notes hit per take'
	}: {
		takes: { id: number; hit_rate: number; created_at: string; tempo_percent?: number | null; practice_bpm?: number | null; notes_revision?: number | null }[];
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

	// Takes away from the original tempo are drawn hollow, so a jump in the
	// line can be told apart from a change of tempo.
	const offTempo = (t: (typeof takes)[number]) => {
		const k = tempoKind(t.tempo_percent);
		return k === 'slower' || k === 'faster';
	};
	const showLegend = $derived(takes.some(offTempo));

	// Where the track's notes were edited between two takes: a dashed line
	// halfway between them.
	const edits = $derived(
		takes.flatMap((t, i) => {
			const prev = takes[i - 1]?.notes_revision;
			return i > 0 && prev != null && t.notes_revision != null && prev !== t.notes_revision ? [i] : [];
		})
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

	{#each edits as i (i)}
		{@const ex = (x(i - 1) + x(i)) / 2}
		<line x1={ex} x2={ex} y1={PAD.top} y2={H - PAD.bottom} style:stroke="var(--fg-subtle)" stroke-dasharray="4 3" />
		<text x={ex + 4} y={PAD.top + 10} font-size="10" style:fill="var(--fg-subtle)">notes edited</text>
	{/each}

	{#if takes.length > 1}
		<path d={path} fill="none" style:stroke="var(--accent-text)" stroke-width="2" stroke-linejoin="round" />
	{/if}
	{#each takes as take, i (take.id)}
		<a
			href="/results/{take.id}"
			aria-label="Take on {shortDate(take.created_at)}: {formatPercent(take.hit_rate)}, {describeTempo(take.tempo_percent, take.practice_bpm)}"
		>
			<circle
				cx={x(i)}
				cy={y(take.hit_rate)}
				r={take.hit_rate === best ? 5 : 3.5}
				style:fill={offTempo(take) ? 'var(--card)' : 'var(--accent-text)'}
				style:stroke="var(--accent-text)"
				stroke-width={offTempo(take) ? 2 : 0}
			>
				<title>{shortDate(take.created_at)}: {formatPercent(take.hit_rate)} of notes hit · {describeTempo(take.tempo_percent, take.practice_bpm)}</title>
			</circle>
		</a>
	{/each}

	{#each xLabels as i (i)}
		<text x={x(i)} y={H - 8} text-anchor={i === 0 && takes.length > 1 ? 'start' : i === takes.length - 1 && takes.length > 1 ? 'end' : 'middle'} font-size="11" style:fill="var(--fg-subtle)">
			{shortDate(takes[i].created_at)}
		</text>
	{/each}
</svg>

{#if showLegend}
	<p class="flex flex-wrap gap-4 text-xs text-fg-subtle mt-1" aria-hidden="true">
		<span class="inline-flex items-center gap-1.5">
			<svg width="10" height="10"><circle cx="5" cy="5" r="4" style:fill="var(--accent-text)" /></svg> Original tempo
		</span>
		<span class="inline-flex items-center gap-1.5">
			<svg width="10" height="10"><circle cx="5" cy="5" r="3.5" style:fill="var(--card)" style:stroke="var(--accent-text)" stroke-width="1.5" /></svg>
			Slower or faster (hover for the tempo)
		</span>
	</p>
{/if}
