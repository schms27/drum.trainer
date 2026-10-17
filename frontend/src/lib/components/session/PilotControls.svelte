<script lang="ts">
	import { Volume2, VolumeX } from '@lucide/svelte';
	import type { PilotPlayer } from '$lib/audio/PilotPlayer';
	import type { PilotTrack } from '$lib/api/pilots';

	// Backing tracks ("pilot tracks" in the API): audio that plays along while recording.
	let {
		pilots,
		player,
		ready = true
	}: {
		pilots: PilotTrack[];
		player: PilotPlayer | null;
		ready?: boolean;
	} = $props();

	let mutedIds = $state<Set<number>>(new Set());

	function toggleMute(id: number) {
		if (!player) return;
		const nowMuted = !mutedIds.has(id);
		player.setMuted(id, nowMuted);
		const next = new Set(mutedIds);
		if (nowMuted) next.add(id);
		else next.delete(id);
		mutedIds = next;
	}
</script>

<div class="flex flex-col gap-2">
	{#each pilots as pilot (pilot.id)}
		{@const muted = mutedIds.has(pilot.id)}
		<div class="flex items-center gap-3">
			<button
				type="button"
				onclick={() => toggleMute(pilot.id)}
				disabled={!player}
				aria-pressed={muted}
				aria-label={muted ? `Unmute ${pilot.name}` : `Mute ${pilot.name}`}
				class="w-9 h-9 rounded-lg flex items-center justify-center transition-colors cursor-pointer
					disabled:opacity-40 disabled:cursor-not-allowed
					{muted ? 'bg-control text-fg-subtle hover:bg-control-hover' : 'bg-accent-soft text-accent-text'}"
			>
				{#if muted}<VolumeX size={16} aria-hidden="true" />{:else}<Volume2 size={16} aria-hidden="true" />{/if}
			</button>
			<span class="flex-1 min-w-0 text-sm text-fg truncate">{pilot.name}</span>
			{#if !ready}
				<span class="text-xs text-fg-subtle">Loading…</span>
			{:else if muted}
				<span class="text-xs text-fg-subtle">Muted</span>
			{/if}
		</div>
	{/each}
	<p class="text-xs text-fg-subtle mt-1">
		{#if !ready}
			Loading audio…
		{:else if !player}
			Starts together with your first note. You can mute tracks while recording.
		{:else}
			Playing along with your recording.
		{/if}
	</p>
</div>
