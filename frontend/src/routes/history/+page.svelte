<script lang="ts">
	const { data } = $props();
	const { sessions, trackNames } = data;

	function formatDate(iso: string) {
		return new Date(iso).toLocaleString();
	}

	function formatHitRate(rate: number) {
		return (rate * 100).toFixed(1) + '%';
	}

	function formatOffset(ms: number) {
		const sign = ms >= 0 ? '+' : '';
		return sign + ms.toFixed(0) + ' ms';
	}
</script>

<div class="max-w-4xl mx-auto p-6">
	<h1 class="text-2xl font-bold text-white mb-6">Practice History</h1>

	{#if sessions.length === 0}
		<p class="text-zinc-400">No practice sessions recorded yet. Go play a track!</p>
	{:else}
		<table class="w-full text-sm text-left">
			<thead>
				<tr class="border-b border-zinc-700 text-zinc-400">
					<th class="pb-3 pr-4 font-medium">Track</th>
					<th class="pb-3 pr-4 font-medium">Date</th>
					<th class="pb-3 pr-4 font-medium text-right">Hits</th>
					<th class="pb-3 pr-4 font-medium text-right">Hit Rate</th>
					<th class="pb-3 pr-4 font-medium text-right">Avg Offset</th>
					<th class="pb-3 font-medium"></th>
				</tr>
			</thead>
			<tbody>
				{#each sessions as session}
					<tr class="border-b border-zinc-800 hover:bg-zinc-800/40 transition">
						<td class="py-3 pr-4 text-white">
							{trackNames.get(session.track_id) ?? 'Unknown'}
						</td>
						<td class="py-3 pr-4 text-zinc-300">{formatDate(session.created_at)}</td>
						<td class="py-3 pr-4 text-zinc-300 text-right">{session.hit_count}</td>
						<td class="py-3 pr-4 text-zinc-300 text-right">{formatHitRate(session.hit_rate)}</td>
						<td class="py-3 pr-4 text-zinc-300 text-right">{formatOffset(session.avg_offset_ms)}</td>
						<td class="py-3">
							<a
								href="/results/{session.id}"
								class="text-indigo-400 hover:text-indigo-300 transition font-medium"
							>
								Results →
							</a>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	{/if}
</div>
