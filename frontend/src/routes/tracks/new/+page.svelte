<script lang="ts">
	import { goto } from '$app/navigation';
	import { PencilRuler } from '@lucide/svelte';
	import { createBlankTrack } from '$lib/api/tracks';
	import { PATTERNS, patternNotes, type PatternId } from '$lib/editor/patterns';
	import PageHeader from '$lib/components/ui/PageHeader.svelte';
	import Card from '$lib/components/ui/Card.svelte';
	import Button from '$lib/components/ui/Button.svelte';
	import Notice from '$lib/components/ui/Notice.svelte';

	let name = $state('');
	let bpm = $state(100);
	let beatsPerBar = $state(4);
	let bars = $state(4);
	let pattern = $state<PatternId>('rock');
	let creating = $state(false);
	let error = $state('');

	const valid = $derived(
		name.trim().length > 0 &&
			bpm >= 20 &&
			bpm <= 300 &&
			Number.isInteger(beatsPerBar) &&
			beatsPerBar >= 1 &&
			beatsPerBar <= 16 &&
			Number.isInteger(bars) &&
			bars >= 1 &&
			bars <= 999
	);

	async function create(e: SubmitEvent) {
		e.preventDefault();
		if (!valid) return;
		creating = true;
		error = '';
		try {
			const track = await createBlankTrack({
				name: name.trim(),
				bpm,
				time_sig_numerator: beatsPerBar,
				bars,
				notes: patternNotes(pattern, bpm, beatsPerBar, bars)
			});
			goto(`/tracks/${track.id}/edit`);
		} catch (err: unknown) {
			error = err instanceof Error ? err.message : 'Could not create the track';
			creating = false;
		}
	}

	const inputClass =
		'bg-control text-fg rounded-lg px-3 h-10 text-sm border border-line focus:outline-none focus:ring-2 focus:ring-accent';
</script>

<svelte:head><title>New track — Drum Trainer</title></svelte:head>

<PageHeader
	title="New track"
	subtitle="Write a groove, fill or exercise yourself, without a MIDI file."
	back={{ href: '/', label: 'Library' }}
/>

<Card title="Start a track" description="You can change the notes and the length afterwards in the editor.">
	<form class="flex flex-col gap-4 max-w-md" onsubmit={create}>
		<label class="flex flex-col gap-1.5">
			<span class="text-sm font-medium text-fg">Name</span>
			<input class={inputClass} bind:value={name} maxlength="100" placeholder="e.g. Paradiddle warm-up" required />
		</label>
		<div class="grid grid-cols-3 gap-3">
			<label class="flex flex-col gap-1.5">
				<span class="text-sm font-medium text-fg">Tempo (BPM)</span>
				<input class="{inputClass} tabular-nums" type="number" min="20" max="300" bind:value={bpm} required />
			</label>
			<label class="flex flex-col gap-1.5">
				<span class="text-sm font-medium text-fg">Beats per bar</span>
				<input class="{inputClass} tabular-nums" type="number" min="1" max="16" bind:value={beatsPerBar} required />
			</label>
			<label class="flex flex-col gap-1.5">
				<span class="text-sm font-medium text-fg">Bars</span>
				<input class="{inputClass} tabular-nums" type="number" min="1" max="999" bind:value={bars} required />
			</label>
		</div>
		<fieldset class="flex flex-col gap-2">
			<legend class="text-sm font-medium text-fg mb-1.5">Start with</legend>
			{#each PATTERNS as p (p.id)}
				<label class="flex items-center gap-2 text-sm text-fg cursor-pointer">
					<input type="radio" name="pattern" value={p.id} bind:group={pattern} class="accent-[var(--accent)]" />
					{p.label}
				</label>
			{/each}
		</fieldset>
		{#if error}<Notice variant="error">{error}</Notice>{/if}
		<div>
			<Button type="submit" variant="primary" disabled={!valid || creating}>
				<PencilRuler size={16} aria-hidden="true" />
				{creating ? 'Creating…' : 'Create and open the editor'}
			</Button>
		</div>
	</form>
</Card>
