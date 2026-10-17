<script lang="ts">
	import { CircleCheck, CircleAlert, CircleDashed } from '@lucide/svelte';
	import { midiSupported, watchDevices } from '$lib/midi/input';
	import { listMidiInputs } from '$lib/midi/recorder';

	// First-run check of everything recording needs, with how to fix each item.
	type Status = 'ok' | 'problem' | 'pending';

	let supported = $state(true);
	let secure = $state(true);
	let kitCount = $state<number | null>(null);

	async function refreshKits() {
		kitCount = (await listMidiInputs()).length;
	}

	$effect(() => {
		supported = midiSupported();
		secure = window.isSecureContext;
		if (!supported) return;
		refreshKits();
		let unwatch: (() => void) | undefined;
		let cancelled = false;
		watchDevices(refreshKits)
			.then((u) => {
				if (cancelled) u();
				else unwatch = u;
			})
			.catch(() => {});
		return () => {
			cancelled = true;
			unwatch?.();
		};
	});

	const items = $derived<{ label: string; status: Status; hint: string }[]>([
		{
			label: 'Browser can read MIDI',
			status: supported ? 'ok' : 'problem',
			hint: supported ? 'Web MIDI is available.' : 'Use Chrome or Edge — Firefox and Safari cannot read e-drum kits.'
		},
		{
			label: 'Secure connection',
			status: secure ? 'ok' : 'problem',
			hint: secure ? 'This page is secure (HTTPS or localhost).' : 'Open the app via https:// — browsers only allow MIDI on secure pages.'
		},
		{
			label: 'E-drum kit connected',
			status: !supported ? 'problem' : kitCount === null ? 'pending' : kitCount > 0 ? 'ok' : 'problem',
			hint:
				kitCount && kitCount > 0
					? `${kitCount} MIDI ${kitCount === 1 ? 'device' : 'devices'} found.`
					: 'Connect your kit with a USB cable and switch it on — this updates automatically. Allow MIDI access if the browser asks.'
		}
	]);
</script>

<ul class="flex flex-col gap-3" aria-label="Setup check">
	{#each items as item (item.label)}
		<li class="flex gap-3 text-sm">
			<span class="shrink-0 mt-0.5">
				{#if item.status === 'ok'}
					<CircleCheck size={18} class="text-good" aria-label="Done" />
				{:else if item.status === 'problem'}
					<CircleAlert size={18} class="text-warn" aria-label="Needs attention" />
				{:else}
					<CircleDashed size={18} class="text-fg-subtle animate-spin" aria-label="Checking" />
				{/if}
			</span>
			<div>
				<p class="font-medium text-fg">{item.label}</p>
				<p class="text-fg-muted">{item.hint}</p>
			</div>
		</li>
	{/each}
	<li class="flex gap-3 text-sm">
		<span class="shrink-0 mt-0.5"><CircleDashed size={18} class="text-fg-subtle" aria-hidden="true" /></span>
		<div>
			<p class="font-medium text-fg">Optional: set up your kit's pads</p>
			<p class="text-fg-muted">
				Only needed if some pads show up as the wrong instrument.
				<a href="/kit" class="text-accent-text hover:underline">My kit</a>
			</p>
		</div>
	</li>
</ul>
