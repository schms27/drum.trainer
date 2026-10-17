<script lang="ts">
	import { Plus } from '@lucide/svelte';
	import MidiUpload from '$lib/components/upload/MidiUpload.svelte';
	import TrackList from '$lib/components/tracks/TrackList.svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import SetupChecklist from '$lib/components/onboarding/SetupChecklist.svelte';
	import type { Track, TrackWithNotes } from '$lib/types/track';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();

	let tracks = $state<Track[]>(data.tracks);
	let showUpload = $state(false);

	function handleUploaded(track: TrackWithNotes) {
		const { notes: _notes, ...rest } = track;
		tracks = [rest, ...tracks];
		showUpload = false;
	}
</script>

<svelte:head><title>Library — Drum Trainer</title></svelte:head>

{#if tracks.length === 0}
	<PageHeader title="Welcome to Drum Trainer" subtitle="Play along with a MIDI drum track on your e-drum kit and see how accurately you hit every note." />

	<ol class="grid gap-4 sm:grid-cols-3 text-sm">
		<li class="bg-card border border-line rounded-xl p-4">
			<p class="font-semibold text-fg">1. Add a track</p>
			<p class="text-fg-muted mt-1">Upload a MIDI file (.mid) with the drum part you want to practise.</p>
		</li>
		<li class="bg-card border border-line rounded-xl p-4">
			<p class="font-semibold text-fg">2. Play along</p>
			<p class="text-fg-muted mt-1">Connect your e-drum kit by USB and play the track in Chrome or Edge.</p>
		</li>
		<li class="bg-card border border-line rounded-xl p-4">
			<p class="font-semibold text-fg">3. Review</p>
			<p class="text-fg-muted mt-1">See which notes you hit, missed or played early or late.</p>
		</li>
	</ol>

	<div class="grid gap-4 md:grid-cols-[minmax(0,2fr)_minmax(0,3fr)]">
		<Card title="Check your setup" description="Recording needs a few things from your browser and kit.">
			<SetupChecklist />
		</Card>

		<Card title="Add your first track">
			<MidiUpload onUploaded={handleUploaded} />
		</Card>
	</div>
{:else}
	<PageHeader title="Library" subtitle="{tracks.length} {tracks.length === 1 ? 'track' : 'tracks'}">
		{#snippet actions()}
			<Button variant={showUpload ? 'secondary' : 'primary'} onclick={() => (showUpload = !showUpload)} aria-expanded={showUpload}>
				{#if showUpload}
					Cancel
				{:else}
					<Plus size={16} aria-hidden="true" /> Add track
				{/if}
			</Button>
		{/snippet}
	</PageHeader>

	{#if showUpload}
		<Card title="Add a track" description="Upload a MIDI file (.mid) with the drum part you want to practise.">
			<MidiUpload onUploaded={handleUploaded} />
		</Card>
	{/if}

	<TrackList {tracks} />
{/if}
