<script lang="ts">
	import type { RecordedHit } from '$lib/types/session';

	const GM_DRUM_MAP: Record<number, string> = {
		35: 'Bass Drum', 36: 'Bass Drum',
		37: 'Side Stick', 38: 'Snare', 40: 'Snare',
		42: 'Hi-Hat (closed)', 44: 'Hi-Hat (pedal)', 46: 'Hi-Hat (open)',
		49: 'Crash 1', 57: 'Crash 2',
		51: 'Ride', 53: 'Ride Bell', 59: 'Ride 2',
		50: 'Tom (high)',
		47: 'Tom (mid)', 48: 'Tom (mid)',
		43: 'Tom (low)', 45: 'Tom (low)',
		41: 'Tom (floor)'
	};

	let {
		hits
	}: {
		hits: RecordedHit[];
	} = $props();

	// Show only the last 20 hits, newest first.
	const displayed = $derived(hits.slice(-20).reverse());

	function label(note: number) {
		return GM_DRUM_MAP[note] ?? `Note ${note}`;
	}

	function velColor(vel: number) {
		if (vel > 90) return 'text-green-400';
		if (vel > 50) return 'text-yellow-400';
		return 'text-zinc-400';
	}
</script>

<div class="bg-zinc-800 rounded-xl p-4 border border-zinc-700 min-h-[160px]">
	<h3 class="text-xs text-zinc-500 uppercase tracking-wider mb-3">Live Feed</h3>
	{#if displayed.length === 0}
		<p class="text-zinc-600 text-sm">No hits yet…</p>
	{:else}
		<ul class="space-y-1 text-sm font-mono">
			{#each displayed as hit, i (i)}
				<li class="flex gap-3 items-center opacity-{i === 0 ? 100 : Math.max(40, 100 - i * 8)}">
					<span class="text-zinc-500 w-16 text-right">{hit.time_ms.toFixed(0)}ms</span>
					<span class="text-white">{label(hit.note_number)}</span>
					<span class={velColor(hit.velocity)}>vel {hit.velocity}</span>
				</li>
			{/each}
		</ul>
	{/if}
</div>
