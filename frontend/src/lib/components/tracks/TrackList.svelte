<script lang="ts">
	import { listTracks } from '$lib/api/tracks';
	import type { Track } from '$lib/types/track';
	import TrackCard from './TrackCard.svelte';

	let {
		refreshTrigger = 0
	}: {
		refreshTrigger?: number;
	} = $props();

	let tracks = $state<Track[]>([]);
	let loading = $state(true);
	let error = $state('');

	async function load() {
		loading = true;
		error = '';
		try {
			tracks = await listTracks();
		} catch (e: unknown) {
			error = e instanceof Error ? e.message : 'Failed to load tracks';
		} finally {
			loading = false;
		}
	}

	$effect(() => {
		// Re-runs whenever refreshTrigger changes.
		refreshTrigger;
		load();
	});

	function handleDeleted(id: number) {
		tracks = tracks.filter((t) => t.id !== id);
	}
</script>

<div>
	{#if loading}
		<p class="text-zinc-500 text-sm">Loading tracks…</p>
	{:else if error}
		<p class="text-red-400 text-sm">{error}</p>
	{:else if tracks.length === 0}
		<p class="text-zinc-500 text-sm">No tracks yet — upload a MIDI file above.</p>
	{:else}
		<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
			{#each tracks as track (track.id)}
				<TrackCard {track} onDeleted={handleDeleted} />
			{/each}
		</div>
	{/if}
</div>
