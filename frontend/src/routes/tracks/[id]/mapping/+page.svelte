<script lang="ts">
	import type { NoteMapping } from '$lib/types/track';
	import { updateTrackMapping } from '$lib/api/tracks';
	import { INSTRUMENTS } from '$lib/drums';

	let { data } = $props();

	// Local copy of mappings for editing.
	let mappings: NoteMapping[] = $state(data.mappings.map((m) => ({ ...m })));

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

<svelte:head><title>Mapping: {data.track.name} — Drum Trainer</title></svelte:head>

<div class="min-h-screen bg-zinc-900 text-white p-6 max-w-2xl mx-auto">
	<div class="mb-6 flex items-center justify-between">
		<div>
			<h1 class="text-2xl font-bold">Drum Mapping</h1>
			<p class="text-zinc-400 text-sm mt-0.5">{data.track.name}</p>
		</div>
		<a href="/" class="text-zinc-400 hover:text-white text-sm transition">← Back to Tracks</a>
	</div>

	<div class="bg-zinc-800 rounded-xl border border-zinc-700 overflow-hidden mb-4">
		<table class="w-full text-sm">
			<thead>
				<tr class="border-b border-zinc-700 text-zinc-400 text-xs uppercase tracking-wide">
					<th class="text-left px-4 py-3">Note #</th>
					<th class="text-left px-4 py-3">GM Name</th>
					<th class="text-left px-4 py-3">Instrument</th>
				</tr>
			</thead>
			<tbody>
				{#each mappings as mapping, i}
					<tr class="border-b border-zinc-700/50 last:border-0">
						<td class="px-4 py-2 font-mono text-zinc-300">{mapping.note_number}</td>
						<td class="px-4 py-2 text-zinc-400">
							{GM_NAMES[mapping.note_number] ?? '—'}
						</td>
						<td class="px-4 py-2">
							<select
								bind:value={mappings[i].instrument}
								class="bg-zinc-700 border border-zinc-600 rounded px-2 py-1 text-white text-sm focus:outline-none focus:ring-1 focus:ring-indigo-500"
							>
								{#each INSTRUMENTS as instr}
									<option value={instr}>{instr}</option>
								{/each}
							</select>
						</td>
					</tr>
				{/each}
				{#if mappings.length === 0}
					<tr>
						<td colspan="3" class="px-4 py-6 text-center text-zinc-500">
							No notes found in this track.
						</td>
					</tr>
				{/if}
			</tbody>
		</table>
	</div>

	<div class="flex items-center gap-4">
		<button
			onclick={handleSave}
			disabled={saving || mappings.length === 0}
			class="bg-indigo-600 hover:bg-indigo-500 disabled:opacity-50 text-white rounded-lg px-5 py-2 text-sm font-medium transition"
		>
			{saving ? 'Saving…' : 'Save'}
		</button>
		{#if message}
			<span class={messageOk ? 'text-green-400 text-sm' : 'text-red-400 text-sm'}>{message}</span>
		{/if}
	</div>
</div>
