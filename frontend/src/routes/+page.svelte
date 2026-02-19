<script lang="ts">
	import MidiUpload from '$lib/components/upload/MidiUpload.svelte';
	import TrackList from '$lib/components/tracks/TrackList.svelte';
	import type { TrackWithNotes } from '$lib/types/track';

	let refreshTrigger = $state(0);

	function handleUploaded(_track: TrackWithNotes) {
		refreshTrigger++;
	}
</script>

<div class="flex flex-col gap-10">
	<div>
		<h1 class="text-2xl font-bold text-white mb-2">Upload a Target Track</h1>
		<p class="text-zinc-400 text-sm mb-5">
			Upload a standard MIDI file (.mid) to use as your practice target. The app will extract all
			drum notes from channel 10.
		</p>
		<MidiUpload onUploaded={handleUploaded} />
	</div>

	<div>
		<h2 class="text-xl font-semibold text-white mb-4">Your Tracks</h2>
		<TrackList {refreshTrigger} />
	</div>
</div>
