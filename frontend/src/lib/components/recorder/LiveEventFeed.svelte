<script lang="ts">
	import type { RecordedHit } from '$lib/types/session';

	// Note-level names (more specific than the instrument groups) so drummers
	// can tell e.g. a rim click from a snare hit while checking their kit.
	const NOTE_NAMES: Record<number, string> = {
		35: 'Bass Drum', 36: 'Bass Drum',
		37: 'Side Stick', 38: 'Snare', 40: 'Snare (rim)',
		42: 'Hi-Hat (closed)', 44: 'Hi-Hat (pedal)', 46: 'Hi-Hat (open)',
		49: 'Crash 1', 57: 'Crash 2',
		51: 'Ride', 53: 'Ride Bell', 59: 'Ride 2',
		50: 'Tom (high)',
		47: 'Tom (mid)', 48: 'Tom (mid)',
		43: 'Tom (low)', 45: 'Tom (low)',
		41: 'Tom (floor)'
	};

	let { hits }: { hits: RecordedHit[] } = $props();

	// Show only the last 8 hits, newest first.
	const displayed = $derived(hits.slice(-8).reverse());

	function label(note: number) {
		return NOTE_NAMES[note] ?? `Note ${note}`;
	}
</script>

<details class="group bg-card border border-line rounded-xl">
	<summary class="cursor-pointer px-5 h-12 flex items-center justify-between text-sm font-semibold text-fg">
		Incoming hits
		<span class="text-xs font-normal text-fg-subtle">
			{hits.length === 0 ? 'none yet' : `${hits.length} received`}
		</span>
	</summary>
	<div class="px-5 pb-5">
		{#if displayed.length === 0}
			<p class="text-fg-subtle text-sm">Notes from your kit appear here after you press Start.</p>
		{:else}
			<ul class="space-y-1 text-sm font-mono">
				{#each displayed as hit, i (hits.length - i)}
					<li class="flex gap-3 items-center" style="opacity: {Math.max(0.4, 1 - i * 0.08)}">
						<span class="text-fg-subtle w-20 text-right tabular-nums">{(hit.time_ms / 1000).toFixed(2)} s</span>
						<span class="text-fg">{label(hit.note_number)}</span>
						<span class="text-fg-subtle">velocity {hit.velocity}</span>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</details>
