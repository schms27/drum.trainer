<script lang="ts">
	import ScoreSummary from '$lib/components/results/ScoreSummary.svelte';
	import InstrumentTable from '$lib/components/results/InstrumentTable.svelte';
	import TimingHistogram from '$lib/components/results/TimingHistogram.svelte';
	import TimelineView from '$lib/components/results/TimelineView.svelte';
	import PlayButton from '$lib/components/PlayButton.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const { sessionResult, track } = data;
	const { session, score } = sessionResult;

	// Convert matched hits to the common PlayableHit shape for playback.
	const recordingHits = score.matched_hits.map((h) => ({
		note_number: h.note_number,
		time_ms: h.recorded_time_ms,
		velocity: 100
	}));
</script>

<div class="flex flex-col gap-8">
	<div>
		<a href="/session/{session.track_id}" class="text-zinc-500 hover:text-zinc-300 text-sm transition">
			← Practice again
		</a>
		<h1 class="text-2xl font-bold text-white mt-2">Results — {track.name}</h1>
		<p class="text-zinc-400 text-sm mt-1">
			{session.hit_count} hits recorded · Session #{session.id}
		</p>
		<div class="flex gap-2 mt-3">
			<PlayButton hits={track.notes} label="Target Track" />
			<PlayButton hits={recordingHits} label="My Recording" />
		</div>
	</div>

	<!-- Summary cards -->
	<ScoreSummary {score} />

	<!-- Per-instrument table -->
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
		<h2 class="text-sm font-semibold text-zinc-300 mb-4">Per-Instrument Breakdown</h2>
		<InstrumentTable breakdown={score.by_instrument} />
	</div>

	<!-- Timing histogram -->
	<TimingHistogram hits={score.matched_hits} />

	<!-- Timeline -->
	<TimelineView
		targetNotes={track.notes}
		matchedHits={score.matched_hits}
		ghostHits={score.ghost_hits}
		breakdown={score.by_instrument}
		durationMs={track.duration_ms}
		bpm={track.bpm}
		timeSigNumerator={track.time_sig_numerator}
	/>
</div>
