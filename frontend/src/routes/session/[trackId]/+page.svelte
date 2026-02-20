<script lang="ts">
	import { goto } from '$app/navigation';
	import MidiDeviceSelector from '$lib/components/recorder/MidiDeviceSelector.svelte';
	import RecordingControls from '$lib/components/recorder/RecordingControls.svelte';
	import PlayButton from '$lib/components/PlayButton.svelte';
	import { submitSession } from '$lib/api/sessions';
	import type { RecordedHit } from '$lib/types/session';
	import type { PageData } from './$types';

	let { data }: { data: PageData } = $props();
	const { track } = data;

	let selectedDeviceId = $state('');
	let submitting = $state(false);
	let submitError = $state('');

	function formatDuration(ms: number) {
		const secs = Math.floor(ms / 1000);
		const m = Math.floor(secs / 60);
		const s = secs % 60;
		return `${m}:${String(s).padStart(2, '0')}`;
	}

	const noteCounts = $derived(
		track.notes.reduce<Record<string, number>>((acc, n) => {
			acc[n.instrument] = (acc[n.instrument] ?? 0) + 1;
			return acc;
		}, {})
	);

	async function handleFinished(hits: RecordedHit[], startedAt: string, endedAt: string) {
		submitting = true;
		submitError = '';
		try {
			const result = await submitSession({
				track_id: track.id,
				started_at: startedAt,
				ended_at: endedAt,
				hits
			});
			goto(`/results/${result.session.id}`);
		} catch (e: unknown) {
			submitError = e instanceof Error ? e.message : 'Failed to submit session';
			submitting = false;
		}
	}
</script>

<div class="flex flex-col gap-8">
	<!-- Track info -->
	<div>
		<a href="/" class="text-zinc-500 hover:text-zinc-300 text-sm transition">← Back to tracks</a>
		<div class="flex items-start justify-between mt-2 gap-4">
			<div>
				<h1 class="text-2xl font-bold text-white">{track.name}</h1>
				<p class="text-zinc-400 text-sm mt-1">
					{track.note_count} notes · {formatDuration(track.duration_ms)} · {track.filename}
				</p>
			</div>
			<PlayButton hits={track.notes} label="Preview Track" />
		</div>
	</div>

	<!-- Device selector -->
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5 max-w-sm">
		<MidiDeviceSelector bind:selectedId={selectedDeviceId} />
	</div>

	<!-- Recording controls + live feed -->
	<div>
		<h2 class="text-lg font-semibold text-white mb-4">Recording</h2>
		{#if submitting}
			<p class="text-zinc-400 text-sm animate-pulse">Calculating your score…</p>
		{:else}
			<RecordingControls deviceId={selectedDeviceId} onFinished={handleFinished} />
		{/if}
		{#if submitError}
			<p class="text-red-400 text-sm mt-3">{submitError}</p>
		{/if}
	</div>

	<!-- Track overview (note count per instrument) -->
	<div class="bg-zinc-800 border border-zinc-700 rounded-xl p-5">
		<h3 class="text-sm font-semibold text-zinc-300 mb-3">Target Track Contents</h3>
		<div class="flex flex-wrap gap-3 text-sm">
			{#each Object.entries(noteCounts).sort((a, b) => b[1] - a[1]) as [instr, count]}
				<span class="bg-zinc-700 text-zinc-300 rounded-full px-3 py-1">
					{instr}: {count}
				</span>
			{/each}
		</div>
	</div>
</div>
