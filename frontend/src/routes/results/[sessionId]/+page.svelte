<script lang="ts">
	import ScoreSummary from '$lib/components/results/ScoreSummary.svelte';
	import InstrumentTable from '$lib/components/results/InstrumentTable.svelte';
	import TimingHistogram from '$lib/components/results/TimingHistogram.svelte';
	import TimelineView from '$lib/components/results/TimelineView.svelte';
	import PlayButton from '$lib/components/PlayButton.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import { RotateCcw } from '@lucide/svelte';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	// svelte-ignore state_referenced_locally
	const { sessionResult, track } = data;
	const { session, score } = sessionResult;

	// Convert matched hits to the common PlayableHit shape for playback.
	const recordingHits = score.matched_hits.map((h) => ({
		note_number: h.note_number,
		time_ms: h.recorded_time_ms,
		velocity: 100
	}));
</script>

<svelte:head><title>Results: {track.name} — Drum Trainer</title></svelte:head>

<PageHeader
	title={track.name}
	subtitle="Played {new Date(session.created_at).toLocaleString(undefined, { dateStyle: 'medium', timeStyle: 'short' })} · {session.hit_count} hits recorded"
	back={{ href: '/history', label: 'Progress' }}
>
	{#snippet actions()}
		<Button href="/session/{session.track_id}" variant="primary">
			<RotateCcw size={16} aria-hidden="true" /> Practice again
		</Button>
	{/snippet}
</PageHeader>

<ScoreSummary {score} />

<div class="flex flex-wrap items-center gap-2 text-sm text-fg-muted">
	<span>Compare by ear:</span>
	<PlayButton hits={track.notes} label="The track" />
	<PlayButton hits={recordingHits} label="Your take" />
</div>

<Card title="By instrument">
	<InstrumentTable breakdown={score.by_instrument} />
</Card>

<TimingHistogram hits={score.matched_hits} />

<TimelineView
	targetNotes={track.notes}
	matchedHits={score.matched_hits}
	ghostHits={score.ghost_hits}
	breakdown={score.by_instrument}
	durationMs={track.duration_ms}
	bpm={track.bpm}
	timeSigNumerator={track.time_sig_numerator}
/>
