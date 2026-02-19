<script lang="ts">
	import type { ScoreResult } from '$lib/types/session';

	let { score }: { score: ScoreResult } = $props();

	function pct(v: number) {
		return (v * 100).toFixed(1) + '%';
	}

	function offsetLabel(ms: number) {
		if (Math.abs(ms) < 0.5) return '±0 ms';
		return (ms > 0 ? '+' : '') + ms.toFixed(1) + ' ms';
	}

	function offsetColor(ms: number) {
		const abs = Math.abs(ms);
		if (abs <= 20) return 'text-green-400';
		if (abs <= 50) return 'text-yellow-400';
		return 'text-red-400';
	}

	function hitRateColor(r: number) {
		if (r >= 0.9) return 'text-green-400';
		if (r >= 0.7) return 'text-yellow-400';
		return 'text-red-400';
	}
</script>

<div class="grid grid-cols-2 sm:grid-cols-4 gap-4">
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
		<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Hit Rate</p>
		<p class="text-3xl font-bold {hitRateColor(score.hit_rate)}">{pct(score.hit_rate)}</p>
	</div>
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
		<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Avg Timing</p>
		<p class="text-3xl font-bold {offsetColor(score.avg_offset_ms)}">
			{offsetLabel(score.avg_offset_ms)}
		</p>
		<p class="text-zinc-500 text-xs mt-1">
			{score.avg_offset_ms < 0 ? 'playing early' : score.avg_offset_ms > 0 ? 'playing late' : 'perfect'}
		</p>
	</div>
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
		<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Consistency</p>
		<p class="text-3xl font-bold text-white">±{score.std_offset_ms.toFixed(1)} ms</p>
		<p class="text-zinc-500 text-xs mt-1">std deviation</p>
	</div>
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 text-center">
		<p class="text-zinc-400 text-xs uppercase tracking-wider mb-1">Ghost Notes</p>
		<p class="text-3xl font-bold text-zinc-300">{pct(score.ghost_rate)}</p>
	</div>
</div>
