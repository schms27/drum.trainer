<script lang="ts">
	import type { NoteMapping } from '$lib/types/track';
	import { updateTrackMapping } from '$lib/api/tracks';
	import { INSTRUMENTS } from '$lib/drums';
	import { Check } from '@lucide/svelte';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	let { data } = $props();

	// Local copy of mappings for editing.
	// svelte-ignore state_referenced_locally
	let mappings: NoteMapping[] = $state(data.mappings.map((m) => ({ ...m })));

	// How often each note occurs in the track, to show which rows matter most.
	// svelte-ignore state_referenced_locally
	const noteCounts = data.track.notes.reduce<Record<number, number>>((acc, n) => {
		acc[n.note_number] = (acc[n.note_number] ?? 0) + 1;
		return acc;
	}, {});

	let saving = $state(false);
	let message = $state('');
	let messageOk = $state(true);

	const GM_NAMES: Record<number, string> = {
		35: 'Acoustic Bass Drum',
		36: 'Bass Drum 1',
		37: 'Side Stick',
		38: 'Acoustic Snare',
		39: 'Hand Clap',
		40: 'Electric Snare',
		41: 'Low Floor Tom',
		42: 'Closed Hi-Hat',
		43: 'High Floor Tom',
		44: 'Pedal Hi-Hat',
		45: 'Low Tom',
		46: 'Open Hi-Hat',
		47: 'Low-Mid Tom',
		48: 'Hi-Mid Tom',
		49: 'Crash Cymbal 1',
		50: 'High Tom',
		51: 'Ride Cymbal 1',
		52: 'Chinese Cymbal',
		53: 'Ride Bell',
		54: 'Tambourine',
		55: 'Splash Cymbal',
		56: 'Cowbell',
		57: 'Crash Cymbal 2',
		58: 'Vibraslap',
		59: 'Ride Cymbal 2',
		60: 'High Bongo',
		61: 'Low Bongo',
		62: 'Mute Hi Conga',
		63: 'Open Hi Conga',
		64: 'Low Conga',
		65: 'High Timbale',
		66: 'Low Timbale'
	};


	async function handleSave() {
		saving = true;
		message = '';
		try {
			const updated = await updateTrackMapping(data.track.id, mappings);
			mappings = updated.map((m) => ({ ...m }));
			message = 'Mapping saved.';
			messageOk = true;
		} catch (e) {
			message = e instanceof Error ? e.message : 'Save failed.';
			messageOk = false;
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head><title>Pad mapping: {data.track.name} — Drum Trainer</title></svelte:head>

<PageHeader
	title="Pad mapping"
	subtitle={data.track.name}
	back={{ href: `/tracks/${data.track.id}`, label: data.track.name }}
/>

<Notice>
	Every MIDI note in this file belongs to an instrument. Your hits are only compared with notes of the same
	instrument, so a bass drum hit never counts for the snare. The defaults follow the General MIDI standard —
	change them only if this file uses different note numbers. To map the pads of your kit instead, use
	<a href="/kit" class="text-accent-text hover:underline">My kit</a>.
</Notice>

<div class="bg-card border border-line rounded-xl overflow-hidden">
	<table class="w-full text-sm">
		<thead>
			<tr class="border-b border-line text-fg-muted text-xs">
				<th scope="col" class="text-left font-medium px-4 py-3">MIDI note</th>
				<th scope="col" class="text-right font-medium px-4 py-3">In track</th>
				<th scope="col" class="text-left font-medium px-4 py-3">Instrument</th>
			</tr>
		</thead>
		<tbody>
			{#each mappings as mapping, i (mapping.note_number)}
				<tr class="border-b border-line-subtle last:border-0">
					<th scope="row" class="text-left font-normal px-4 py-2">
						<span class="text-fg">{GM_NAMES[mapping.note_number] ?? 'Non-standard note'}</span>
						<span class="text-fg-subtle font-mono text-xs ml-1">#{mapping.note_number}</span>
					</th>
					<td class="px-4 py-2 text-right text-fg-muted tabular-nums">{noteCounts[mapping.note_number] ?? 0}×</td>
					<td class="px-4 py-2">
						<label class="sr-only" for="instr-{mapping.note_number}">Instrument for note {mapping.note_number}</label>
						<select
							id="instr-{mapping.note_number}"
							bind:value={mappings[i].instrument}
							class="bg-control text-fg border border-line rounded-lg px-2 h-9 text-sm focus:outline-none focus:ring-2 focus:ring-accent"
						>
							{#each INSTRUMENTS as instr}
								<option value={instr}>{instr}</option>
							{/each}
						</select>
					</td>
				</tr>
			{:else}
				<tr>
					<td colspan="3" class="px-4 py-6 text-center text-fg-subtle">No notes found in this track.</td>
				</tr>
			{/each}
		</tbody>
	</table>
</div>

<div class="flex items-center gap-4">
	<Button variant="primary" onclick={handleSave} disabled={saving || mappings.length === 0}>
		{saving ? 'Saving…' : 'Save mapping'}
	</Button>
	{#if message}
		<span class="inline-flex items-center gap-1 text-sm {messageOk ? 'text-good' : 'text-bad'}" role="status">
			{#if messageOk}<Check size={16} aria-hidden="true" />{/if}{message}
		</span>
	{/if}
</div>
