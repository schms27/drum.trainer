<script lang="ts">
	import type { PilotPlayer } from '$lib/audio/PilotPlayer';
	import type { PilotTrack } from '$lib/api/pilots';

	let {
		pilots,
		player,
		ready = true
	}: {
		pilots: PilotTrack[];
		player: PilotPlayer | null;
		ready?: boolean;
	} = $props();

	// Track mute state locally so the UI stays reactive.
	let mutedIds = $state<Set<number>>(new Set());

	function toggleMute(id: number) {
		if (!player) return;
		const nowMuted = !mutedIds.has(id);
		player.setMuted(id, nowMuted);
		mutedIds = new Set(mutedIds);
		if (nowMuted) {
			mutedIds.add(id);
		} else {
			mutedIds.delete(id);
		}
		mutedIds = new Set(mutedIds); // trigger reactivity
	}

	function formatSize(bytes: number): string {
		if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(0)} KB`;
		return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
	}
</script>

{#if pilots.length > 0}
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
		<h3 class="text-sm font-semibold text-zinc-300 mb-3">Pilot Tracks</h3>
		<div class="flex flex-col gap-2">
			{#each pilots as pilot (pilot.id)}
				{@const muted = mutedIds.has(pilot.id)}
				<div class="flex items-center gap-3">
					<button
						onclick={() => toggleMute(pilot.id)}
						disabled={!player}
						class="w-8 h-8 rounded-lg flex items-center justify-center text-sm transition
							{muted
							? 'bg-zinc-700 text-zinc-500 hover:bg-zinc-600'
							: 'bg-indigo-600 text-white hover:bg-indigo-500'}
							disabled:opacity-40 disabled:cursor-not-allowed"
						title={muted ? 'Unmute' : 'Mute'}
					>
						{muted ? '🔇' : '🔊'}
					</button>
					<div class="flex-1 min-w-0">
						<p class="text-sm text-zinc-200 truncate">{pilot.name}</p>
						<p class="text-xs text-zinc-500">{formatSize(pilot.file_size)}</p>
					</div>
					{#if !ready}
						<span class="text-xs text-zinc-500">Loading…</span>
					{:else if muted}
						<span class="text-xs text-zinc-500">Muted</span>
					{/if}
				</div>
			{/each}
		</div>
		{#if !ready}
			<p class="text-xs text-zinc-500 mt-3">Loading audio…</p>
		{:else if !player}
			<p class="text-xs text-zinc-500 mt-3">Pilot tracks will play in sync with your first hit.</p>
		{/if}
	</div>
{/if}
