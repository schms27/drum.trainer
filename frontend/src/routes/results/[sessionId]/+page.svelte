<script lang="ts">
	import ScoreSummary from '$lib/components/results/ScoreSummary.svelte';
	import InstrumentTable from '$lib/components/results/InstrumentTable.svelte';
	import TimingHistogram from '$lib/components/results/TimingHistogram.svelte';
	import TimelineView from '$lib/components/results/TimelineView.svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const { sessionResult, track } = data;
	const { session, score } = sessionResult;
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
		breakdown={score.by_instrument}
		durationMs={track.duration_ms}
	/>
</div>
