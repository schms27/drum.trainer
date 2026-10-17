<script lang="ts">
	import type { InstrumentBreakdown } from '$lib/types/session';
	import { formatOffset, formatPercent, hitRateTone, timingTone, toneText } from '$lib/score-format';

	let { breakdown }: { breakdown: InstrumentBreakdown[] } = $props();
</script>

<div class="overflow-x-auto -mx-5 px-5">
	<table class="w-full text-sm">
		<thead>
			<tr class="text-fg-muted text-xs border-b border-line">
				<th scope="col" class="text-left font-medium py-2 pr-4">Instrument</th>
				<th scope="col" class="text-right font-medium py-2 px-3">Notes hit</th>
				<th scope="col" class="text-right font-medium py-2 px-3">Missed</th>
				<th scope="col" class="text-right font-medium py-2 px-3">Extra</th>
				<th scope="col" class="text-right font-medium py-2 pl-3">Average timing</th>
			</tr>
		</thead>
		<tbody>
			{#each breakdown as row (row.instrument)}
				<tr class="border-b border-line-subtle last:border-0">
					<th scope="row" class="text-left py-2.5 pr-4 text-fg font-medium">{row.instrument}</th>
					<td class="text-right py-2.5 px-3 tabular-nums whitespace-nowrap">
						<span class="font-medium {toneText[hitRateTone(row.hit_rate)]}">{formatPercent(row.hit_rate)}</span>
						<span class="text-fg-subtle text-xs ml-1">{row.matched_count}/{row.target_count}</span>
					</td>
					<td class="text-right py-2.5 px-3 tabular-nums {row.missed_count > 0 ? 'text-bad' : 'text-fg-subtle'}">
						{row.missed_count}
					</td>
					<td class="text-right py-2.5 px-3 tabular-nums text-fg-muted">{row.ghost_count}</td>
					<td class="text-right py-2.5 pl-3 tabular-nums whitespace-nowrap">
						{#if row.avg_offset_ms === null}
							<span class="text-fg-subtle">—</span>
						{:else}
							<span class={toneText[timingTone(row.avg_offset_ms)]}>{formatOffset(row.avg_offset_ms)}</span>
						{/if}
					</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>
