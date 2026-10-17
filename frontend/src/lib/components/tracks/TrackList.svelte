<script lang="ts">
	import type { Track } from '$lib/types/track';
	import TrackCard from './TrackCard.svelte';

	let { tracks, onDeleted }: { tracks: Track[]; onDeleted: (id: number) => void } = $props();
</script>

<div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-4">
	{#each tracks as track (track.id)}
		<TrackCard {track} {onDeleted} />
	{/each}
</div>
