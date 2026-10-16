<script lang="ts">
	import type { RecordedHit } from '$lib/types/session';
	import { noteLabel } from '$lib/drums';

	let { hits }: { hits: RecordedHit[] } = $props();

	// Show only the last 8 hits, newest first.
	const displayed = $derived(hits.slice(-8).reverse());
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
						<span class="text-fg">{noteLabel(hit.note_number)}</span>
						<span class="text-fg-subtle">velocity {hit.velocity}</span>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</details>
