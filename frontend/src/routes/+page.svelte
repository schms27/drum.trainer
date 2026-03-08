<script lang="ts">
	import MidiUpload from '$lib/components/upload/MidiUpload.svelte';
	import TrackList from '$lib/components/tracks/TrackList.svelte';
	import type { TrackWithNotes } from '$lib/types/track';

	let refreshTrigger = $state(0);
	let drumChannel = $state(10);

	function handleUploaded(_track: TrackWithNotes) {
		refreshTrigger++;
	}
</script>

<svelte:head><title>Tracks — Drum Trainer</title></svelte:head>

<div class="flex flex-col gap-10">
	<div>
		<h1 class="text-2xl font-bold text-white mb-2">Upload a Target Track</h1>
		<p class="text-zinc-400 text-sm mb-4">
			Upload a standard MIDI file (.mid) to use as your practice target.
		</p>
		<div class="flex items-center gap-2 mb-5">
			<label for="drum-channel" class="text-zinc-300 text-sm">Extract drum notes from channel</label>
			<select
				id="drum-channel"
				bind:value={drumChannel}
				class="bg-zinc-700 border border-zinc-600 text-white text-sm rounded px-2 py-1 focus:outline-none focus:border-indigo-500"
			>
				{#each Array.from({ length: 16 }, (_, i) => i + 1) as ch}
					<option value={ch}>{ch}{ch === 10 ? ' (GM drums)' : ''}</option>
				{/each}
			</select>
		</div>
		<MidiUpload onUploaded={handleUploaded} channel={drumChannel} />
	</div>

	<div>
		<h2 class="text-xl font-semibold text-white mb-4">Your Tracks</h2>
		<TrackList {refreshTrigger} />
	</div>
</div>
