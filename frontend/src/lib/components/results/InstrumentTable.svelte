<script lang="ts">
	import type { InstrumentBreakdown } from '$lib/types/session';

	let { breakdown }: { breakdown: InstrumentBreakdown[] } = $props();

	function pct(v: number) {
		return (v * 100).toFixed(1) + '%';
	}

	function hitColor(r: number) {
		if (r >= 0.9) return 'text-green-400';
		if (r >= 0.7) return 'text-yellow-400';
		return 'text-red-400';
	}

	function offsetStr(v: number | null) {
		if (v === null) return '—';
		return (v > 0 ? '+' : '') + v.toFixed(1) + ' ms';
	}

	function offsetColor(v: number | null) {
		if (v === null) return 'text-zinc-500';
		const abs = Math.abs(v);
		if (abs <= 20) return 'text-green-400';
		if (abs <= 50) return 'text-yellow-400';
		return 'text-red-400';
	}
</script>

<div class="overflow-x-auto">
	<table class="w-full text-sm">
		<thead>
			<tr class="text-zinc-500 text-xs uppercase tracking-wider border-b border-zinc-700">
				<th class="text-left py-2 pr-4">Instrument</th>
				<th class="text-right py-2 px-3">Target</th>
				<th class="text-right py-2 px-3">Hit</th>
				<th class="text-right py-2 px-3">Miss</th>
				<th class="text-right py-2 px-3">Ghost</th>
				<th class="text-right py-2 px-3">Hit Rate</th>
				<th class="text-right py-2 pl-3">Avg Offset</th>
			</tr>
		</thead>
		<tbody>
			{#each breakdown as row (row.instrument)}
				<tr class="border-b border-zinc-800 hover:bg-zinc-800/50 transition">
					<td class="py-2.5 pr-4 text-white font-medium">{row.instrument}</td>
					<td class="text-right py-2.5 px-3 text-zinc-400">{row.target_count}</td>
					<td class="text-right py-2.5 px-3 text-zinc-300">{row.matched_count}</td>
					<td class="text-right py-2.5 px-3 text-red-400">{row.missed_count}</td>
					<td class="text-right py-2.5 px-3 text-zinc-400">{row.ghost_count}</td>
					<td class="text-right py-2.5 px-3 font-medium {hitColor(row.hit_rate)}">
						{pct(row.hit_rate)}
					</td>
					<td class="text-right py-2.5 pl-3 font-mono {offsetColor(row.avg_offset_ms)}">
						{offsetStr(row.avg_offset_ms)}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
