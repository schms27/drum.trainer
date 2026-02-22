<script lang="ts">
	import { listMidiInputs, type MidiDevice } from '$lib/midi/recorder';

	let {
		selectedId = $bindable(''),
		onError
	}: {
		selectedId?: string;
		onError?: (msg: string) => void;
	} = $props();

	let devices = $state<MidiDevice[]>([]);
	let supported = $state(true);
	let secureContext = $state(true);
	let loading = $state(true);

	$effect(() => {
		if (!navigator.requestMIDIAccess) {
			supported = false;
			secureContext = window.isSecureContext;
			loading = false;
			return;
		}
		listMidiInputs().then((list) => {
			devices = list;
			if (list.length > 0 && !selectedId) selectedId = list[0].id;
			loading = false;
		});
	});
</script>

<div class="flex flex-col gap-2">
	<label class="text-sm text-zinc-400 font-medium" for="midi-device">MIDI Input Device</label>

	{#if !supported}
		{#if !secureContext}
			<p class="text-amber-400 text-sm">
				Web MIDI API requires HTTPS. Open the app over <code class="font-mono">https://</code> or
				access it via <code class="font-mono">localhost</code>.
			</p>
		{:else}
			<p class="text-amber-400 text-sm">
				Web MIDI API is not supported in this browser. Please use Chrome or Edge.
			</p>
		{/if}
	{:else if loading}
		<p class="text-zinc-500 text-sm">Scanning for MIDI devices…</p>
	{:else if devices.length === 0}
		<p class="text-zinc-500 text-sm">No MIDI devices found. Connect your e-drum and refresh.</p>
	{:else}
		<select
			id="midi-device"
			bind:value={selectedId}
			class="bg-zinc-700 border border-zinc-600 text-white rounded-lg px-3 py-2 text-sm"
		>
			{#each devices as device (device.id)}
				<option value={device.id}>{device.name}</option>
			{/each}
		</select>
	{/if}
</div>
