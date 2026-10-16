<script lang="ts">
	import { CircleCheck } from '@lucide/svelte';
	import { listenToInput, type NoteOn } from '$lib/midi/input';
	import { drumGroup, noteLabel } from '$lib/drums';

	let { deviceId }: { deviceId: string } = $props();

	let last = $state<NoteOn | null>(null);
	let count = $state(0);
	let flash = $state(false);
	let flashTimer: ReturnType<typeof setTimeout> | undefined;

	$effect(() => {
		const id = deviceId;
		last = null;
		count = 0;
		if (!id) return;
		let stop: (() => void) | undefined;
		let cancelled = false;
		listenToInput(id, (hit) => {
			last = hit;
			count++;
			flash = true;
			clearTimeout(flashTimer);
			flashTimer = setTimeout(() => (flash = false), 150);
		})
			.then((s) => {
				if (cancelled) s();
				else stop = s;
			})
			.catch(() => {
				// The device selector already reports missing devices.
			});
		return () => {
			cancelled = true;
			stop?.();
			clearTimeout(flashTimer);
		};
	});
</script>

<div
	class="flex items-center gap-3 rounded-lg border px-3 py-2.5 text-sm transition-colors
		{flash ? 'border-good bg-good/10' : 'border-line'}"
	aria-live="polite"
>
	{#if last}
		<CircleCheck size={18} class="text-good shrink-0" aria-hidden="true" />
		<div class="min-w-0">
			<p class="text-fg">
				<span class="font-medium">{noteLabel(last.note)}</span>
				{#if drumGroup(last.note) === 'Other'}<span class="text-fg-subtle"> (not a standard drum note — see Pad mapping)</span>{/if}
			</p>
			<p class="text-xs text-fg-subtle">Note {last.note} · velocity {last.velocity} · {count} {count === 1 ? 'hit' : 'hits'} so far</p>
		</div>
	{:else}
		<span class="w-2.5 h-2.5 rounded-full bg-control-hover shrink-0 animate-pulse" aria-hidden="true"></span>
		<p class="text-fg-muted">Hit any pad to check that your kit is working.</p>
	{/if}
</div>
