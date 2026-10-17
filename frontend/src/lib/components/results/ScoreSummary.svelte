<script lang="ts">
	import type { ScoreResult } from '$lib/types/session';
	import {
		describeTiming,
		formatOffset,
		formatPercent,
		hitRateTone,
		timingTone,
		toneText
	} from '$lib/score-format';

	let {
		score,
		previousHitRate = null,
		bestHitRate = null
	}: {
		score: ScoreResult;
		/** Hit rate of the previous take on the same track. */
		previousHitRate?: number | null;
		/** Best hit rate on this track before this take. */
		bestHitRate?: number | null;
	} = $props();

	const delta = $derived(previousHitRate === null ? null : Math.round((score.hit_rate - previousHitRate) * 100));

	const targetCount = $derived(score.by_instrument.reduce((n, b) => n + b.target_count, 0));
	const matchedCount = $derived(score.by_instrument.reduce((n, b) => n + b.matched_count, 0));
	const extraCount = $derived(score.ghost_hits?.length ?? 0);
</script>

<div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Notes hit</p>
		<p class="text-3xl font-bold mt-1 tabular-nums {toneText[hitRateTone(score.hit_rate)]}">
			{formatPercent(score.hit_rate)}
		</p>
		<p class="text-xs text-fg-subtle mt-1">{matchedCount} of {targetCount} notes</p>
		{#if delta !== null}
			<p class="text-xs mt-1 {delta > 0 ? 'text-good' : delta < 0 ? 'text-bad' : 'text-fg-subtle'}">
				{delta > 0 ? '▲' : delta < 0 ? '▼' : '='}
				{delta === 0 ? 'same as' : `${Math.abs(delta)} ${Math.abs(delta) === 1 ? 'point' : 'points'} vs`} last take
			</p>
		{/if}
		{#if bestHitRate !== null}
			<p class="text-xs text-fg-subtle">
				{score.hit_rate > bestHitRate ? 'New best!' : `Best: ${formatPercent(bestHitRate)}`}
			</p>
		{/if}
	</div>

	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Average timing</p>
		<p class="text-3xl font-bold mt-1 tabular-nums {toneText[timingTone(score.avg_offset_ms)]}">
			{formatOffset(score.avg_offset_ms)}
		</p>
		<p class="text-xs text-fg-subtle mt-1">{describeTiming(score.avg_offset_ms)}</p>
	</div>

	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Timing spread</p>
		<p class="text-3xl font-bold mt-1 tabular-nums text-fg">±{Math.round(score.std_offset_ms)} ms</p>
		<p class="text-xs text-fg-subtle mt-1">How much your timing varies — lower is steadier</p>
	</div>

	<div class="bg-card border border-line rounded-xl p-5">
		<p class="text-xs font-medium text-fg-muted">Extra hits</p>
		<p class="text-3xl font-bold mt-1 tabular-nums text-fg">{extraCount}</p>
		<p class="text-xs text-fg-subtle mt-1">Hits with no matching note in the track</p>
	</div>
</div>
