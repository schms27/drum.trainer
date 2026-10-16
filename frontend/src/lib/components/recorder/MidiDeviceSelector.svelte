<script lang="ts">
	import { RefreshCw } from '@lucide/svelte';
	import { listMidiInputs, type MidiDevice } from '$lib/midi/recorder';
	import Notice from '$lib/components/ui/Notice.svelte';
	import Button from '$lib/components/ui/Button.svelte';

	const STORAGE_KEY = 'drum-trainer:default-midi-device';

	let { selectedId = $bindable('') }: { selectedId?: string } = $props();

	let devices = $state<MidiDevice[]>([]);
	let supported = $state(true);
	let secureContext = $state(true);
	let loading = $state(true);

	async function scan() {
		loading = true;
		const list = await listMidiInputs();
		devices = list;
		if (!list.some((d) => d.id === selectedId)) {
			let saved: string | null = null;
			try {
				saved = localStorage.getItem(STORAGE_KEY);
			} catch {
				// Storage unavailable — just pick the first device.
			}
			selectedId = saved && list.some((d) => d.id === saved) ? saved : (list[0]?.id ?? '');
		}
		loading = false;
	}

	$effect(() => {
		if (!navigator.requestMIDIAccess) {
			supported = false;
			secureContext = window.isSecureContext;
			loading = false;
			return;
		}
		scan();
	});

	function handleChange(e: Event) {
		try {
			localStorage.setItem(STORAGE_KEY, (e.target as HTMLSelectElement).value);
		} catch {
			// Not remembered, but still selected.
		}
	}
</script>

{#if !supported}
	<Notice variant="warning">
		{#if !secureContext}
			Your browser only allows MIDI access over a secure connection. Open the app via
			<code class="font-mono">https://</code> (or on <code class="font-mono">localhost</code>).
		{:else}
			This browser can't read MIDI devices. Use Chrome or Edge to record.
		{/if}
	</Notice>
{:else if loading}
	<p class="text-sm text-fg-muted">Looking for MIDI devices…</p>
{:else if devices.length === 0}
	<div class="flex flex-col gap-3">
		<Notice variant="warning">
			No e-drum kit found. Connect it with a USB cable, switch it on, then scan again. If the
			browser asked for permission to use MIDI devices, allow it.
		</Notice>
		<div><Button size="sm" onclick={scan}><RefreshCw size={14} aria-hidden="true" /> Scan again</Button></div>
	</div>
{:else}
	<div class="flex flex-wrap items-center gap-2">
		<label class="sr-only" for="midi-device">MIDI input device</label>
		<select
			id="midi-device"
			bind:value={selectedId}
			onchange={handleChange}
			class="bg-control text-fg rounded-lg px-3 h-10 text-sm border border-line min-w-0 max-w-full focus:outline-none focus:ring-2 focus:ring-accent"
		>
			{#each devices as device (device.id)}
				<option value={device.id}>{device.name}</option>
			{/each}
		</select>
		<Button variant="ghost" size="sm" onclick={scan} aria-label="Scan for MIDI devices again" title="Scan again">
			<RefreshCw size={14} aria-hidden="true" />
		</Button>
	</div>
{/if}
